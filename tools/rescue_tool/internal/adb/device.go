package adb

import (
	"strings"
	"sync"
	"time"

	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
)

// DeviceMode represents the current device connection state
type DeviceMode int

const (
	ModeNone     DeviceMode = iota
	ModeAndroid
	ModeFastboot
)

func (m DeviceMode) String() string {
	switch m {
	case ModeAndroid:
		return "Android (ADB)"
	case ModeFastboot:
		return "Fastboot"
	default:
		return "Not Connected"
	}
}

// DeviceInfo holds all detected device information
type DeviceInfo struct {
	Mode           DeviceMode
	Serial         string
	Model          string
	Codename       string
	AndroidVersion string
	ROMVersion     string
	PanelVariant   string
	KSUVersion     string
	SUSFSActive    bool
	KernelVersion  string
}

// DeviceManager handles device detection and polling
type DeviceManager struct {
	adb      *Client
	fastboot *FastbootClient
	mu       sync.RWMutex
	current  DeviceInfo
}

func NewDeviceManager(adb *Client, fastboot *FastbootClient) *DeviceManager {
	return &DeviceManager{adb: adb, fastboot: fastboot}
}

// Detect probes ADB and Fastboot to determine current device state
func (dm *DeviceManager) Detect() DeviceInfo {
	info := DeviceInfo{Mode: ModeNone}

	if dm.adb.IsAvailable() {
		devices, err := dm.adb.GetDevices()
		if err == nil && len(devices) > 0 {
			info.Mode = ModeAndroid
			info.Serial = devices[0]
			dm.fillAndroidInfo(&info)
			dm.mu.Lock()
			dm.current = info
			dm.mu.Unlock()
			return info
		}
	}

	if dm.fastboot.IsAvailable() {
		devices, err := dm.fastboot.GetDevices()
		if err == nil && len(devices) > 0 {
			info.Mode = ModeFastboot
			info.Serial = devices[0]
			dm.fillFastbootInfo(&info)
			dm.mu.Lock()
			dm.current = info
			dm.mu.Unlock()
			return info
		}
	}

	dm.mu.Lock()
	dm.current = info
	dm.mu.Unlock()
	return info
}

func (dm *DeviceManager) GetCurrent() DeviceInfo {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.current
}

func (dm *DeviceManager) GetADBClient() *Client       { return dm.adb }
func (dm *DeviceManager) GetFastbootClient() *FastbootClient { return dm.fastboot }

func (dm *DeviceManager) fillAndroidInfo(info *DeviceInfo) {
	props := map[string]*string{
		"ro.product.model":         &info.Model,
		"ro.product.device":        &info.Codename,
		"ro.build.version.release": &info.AndroidVersion,
		"ro.build.display.id":      &info.ROMVersion,
		"ro.boot.lcm_name":        &info.PanelVariant,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for key, target := range props {
		wg.Add(1)
		go func(k string, t *string) {
			defer wg.Done()
			if val, err := dm.adb.GetProp(k); err == nil && val != "" {
				mu.Lock()
				*t = val
				mu.Unlock()
			}
		}(key, target)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Try normal first, then try su if it fails
		v, err := dm.adb.Shell("cat /proc/ksu/version 2>/dev/null || su -c 'cat /proc/ksu/version' 2>/dev/null || echo ''")
		if err == nil {
			v = strings.TrimSpace(v)
			if v != "" && v != "''" && !strings.Contains(v, "Permission denied") && !strings.Contains(v, "not found") {
				mu.Lock()
				info.KSUVersion = v
				mu.Unlock()
				return
			}
		}

		// Fallback: check if KernelSU is in uname
		unameVal, err := dm.adb.Shell("uname -a 2>/dev/null || uname -r 2>/dev/null || echo ''")
		if err == nil {
			unameLower := strings.ToLower(unameVal)
			if strings.Contains(unameLower, "kernelsu") || strings.Contains(unameLower, "ksu") {
				mu.Lock()
				info.KSUVersion = "Active (Kernel)"
				mu.Unlock()
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		v, err := dm.adb.Shell("cat /proc/susfs/version 2>/dev/null || su -c 'cat /proc/susfs/version' 2>/dev/null || echo ''")
		if err == nil {
			v = strings.TrimSpace(v)
			if v != "" && v != "''" && !strings.Contains(v, "Permission denied") && !strings.Contains(v, "not found") {
				mu.Lock()
				info.SUSFSActive = true
				mu.Unlock()
				return
			}
		}

		// Fallback: check if SUSFS is in uname
		unameVal, err := dm.adb.Shell("uname -a 2>/dev/null || uname -r 2>/dev/null || echo ''")
		if err == nil {
			unameLower := strings.ToLower(unameVal)
			if strings.Contains(unameLower, "susfs") || strings.Contains(unameLower, "sus") {
				mu.Lock()
				info.SUSFSActive = true
				mu.Unlock()
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if v, err := dm.adb.Shell("uname -r 2>/dev/null || echo ''"); err == nil {
			mu.Lock()
			info.KernelVersion = strings.TrimSpace(v)
			mu.Unlock()
		}
	}()

	wg.Wait()
	logger.Info("Device: %s (%s) Android %s Panel: %s KSU: %s",
		info.Model, info.Codename, info.AndroidVersion, info.PanelVariant, info.KSUVersion)
}

func (dm *DeviceManager) fillFastbootInfo(info *DeviceInfo) {
	if product, err := dm.fastboot.GetVar("product"); err == nil {
		info.Codename = product
	}
	logger.Info("Fastboot device: serial=%s codename=%s", info.Serial, info.Codename)
}

func (dm *DeviceManager) PollInterval() time.Duration {
	return 2 * time.Second
}
