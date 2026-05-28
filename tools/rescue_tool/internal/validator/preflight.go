package validator

import (
	"fmt"
	"strings"
	"sync"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
)

// CheckStatus represents the result of a single validation check
type CheckStatus int

const (
	CheckPending CheckStatus = iota
	CheckPass
	CheckFail
	CheckWarn
)

func (s CheckStatus) Emoji() string {
	switch s {
	case CheckPass:
		return "✅"
	case CheckFail:
		return "❌"
	case CheckWarn:
		return "⚠️"
	default:
		return "⏳"
	}
}

// CheckResult holds one validation check result
type CheckResult struct {
	Name    string
	Status  CheckStatus
	Value   string
	Detail  string
}

// PreflightResult holds all validation results
type PreflightResult struct {
	Checks     []CheckResult
	AllPassed  bool
	HasWarning bool
}

// Validator runs pre-flash prerequisite checks
type Validator struct {
	deviceMgr *adb.DeviceManager
}

// NewValidator creates a new pre-flash validator
func NewValidator(dm *adb.DeviceManager) *Validator {
	return &Validator{deviceMgr: dm}
}

// RunChecks performs all pre-flash validation checks
func (v *Validator) RunChecks() (*PreflightResult, error) {
	info := v.deviceMgr.Detect()
	if info.Mode != adb.ModeAndroid {
		return nil, fmt.Errorf("device tidak terhubung di Android mode. Hubungkan device dulu via USB dengan ADB debugging aktif.")
	}

	adbClient := v.deviceMgr.GetADBClient()
	result := &PreflightResult{
		AllPassed: true,
	}

	type checkFn struct {
		name string
		fn   func() CheckResult
	}

	checks := []checkFn{
		{
			name: "Device Codename",
			fn: func() CheckResult {
				val, _ := adbClient.GetProp("ro.product.device")
				val = strings.TrimSpace(val)
				if val == "fire" {
					return CheckResult{
						Name:   "Device Codename",
						Status: CheckPass,
						Value:  val,
						Detail: "Codename 'fire' — Redmi 12 terdeteksi ✓",
					}
				}
				return CheckResult{
					Name:   "Device Codename",
					Status: CheckWarn,
					Value:  val,
					Detail: fmt.Sprintf("Codename '%s' — Pastikan Custom Kernel yang akan Anda flash mendukung codename perangkat ini.", val),
				}
			},
		},
		{
			name: "Android Version",
			fn: func() CheckResult {
				val, _ := adbClient.GetProp("ro.build.version.release")
				val = strings.TrimSpace(val)
				if val == "15" {
					return CheckResult{
						Name:   "Android Version",
						Status: CheckPass,
						Value:  "Android " + val,
						Detail: "Android 15 — compatible ✓",
					}
				}
				return CheckResult{
					Name:   "Android Version",
					Status: CheckWarn,
					Value:  "Android " + val,
					Detail: fmt.Sprintf("Android %s terdeteksi. Pastikan Custom Kernel yang akan Anda flash mendukung versi Android ini.", val),
				}
			},
		},
		{
			name: "Panel Variant",
			fn: func() CheckResult {
				device := strings.TrimSpace(info.Codename)

				val, _ := adbClient.GetProp("ro.boot.lcm_name")
				val = strings.TrimSpace(val)
				upper := strings.ToUpper(val)

				if device == "fire" {
					if strings.Contains(upper, "LC0C") || strings.Contains(upper, "LC0D") {
						return CheckResult{
							Name:   "Panel Variant",
							Status: CheckWarn,
							Value:  val,
							Detail: "Panel LC0C/LC0D — beberapa user melaporkan masalah display. Proceed with caution.",
						}
					}
					if val == "" {
						return CheckResult{
							Name:   "Panel Variant",
							Status: CheckWarn,
							Value:  "Unknown",
							Detail: "Tidak bisa mendeteksi panel variant.",
						}
					}
				} else {
					if val == "" {
						return CheckResult{
							Name:   "Panel Variant",
							Status: CheckPass,
							Value:  "N/A",
							Detail: "Deteksi panel dilewati (hanya berlaku untuk Redmi 12) ✓",
						}
					}
				}

				return CheckResult{
					Name:   "Panel Variant",
					Status: CheckPass,
					Value:  val,
					Detail: "Panel variant compatible ✓",
				}
			},
		},
		{
			name: "Verified Boot (vbmeta)",
			fn: func() CheckResult {
				// Check dm-verity / vbmeta status
				val, _ := adbClient.Shell("getprop ro.boot.verifiedbootstate 2>/dev/null")
				val = strings.TrimSpace(val)
				if val == "orange" {
					return CheckResult{
						Name:   "Verified Boot",
						Status: CheckPass,
						Value:  "Unlocked (orange)",
						Detail: "Bootloader unlocked & vbmeta disabled ✓",
					}
				}
				if val == "green" {
					return CheckResult{
						Name:   "Verified Boot",
						Status: CheckFail,
						Value:  "Locked (green)",
						Detail: "Bootloader LOCKED! Kamu harus unlock bootloader dan disable vbmeta verification dulu.",
					}
				}
				return CheckResult{
					Name:   "Verified Boot",
					Status: CheckWarn,
					Value:  val,
					Detail: "Status vbmeta tidak standar. Pastikan vbmeta verification sudah disabled.",
				}
			},
		},
		{
			name: "ROM Build",
			fn: func() CheckResult {
				val, _ := adbClient.GetProp("ro.build.display.id")
				val = strings.TrimSpace(val)
				return CheckResult{
					Name:   "ROM Build",
					Status: CheckPass,
					Value:  val,
					Detail: "ROM build ID tercatat untuk referensi.",
				}
			},
		},
	}

	// Run checks concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]CheckResult, len(checks))

	for i, c := range checks {
		wg.Add(1)
		go func(idx int, check checkFn) {
			defer wg.Done()
			r := check.fn()
			mu.Lock()
			results[idx] = r
			mu.Unlock()
		}(i, c)
	}
	wg.Wait()

	for _, r := range results {
		result.Checks = append(result.Checks, r)
		if r.Status == CheckFail {
			result.AllPassed = false
		}
		if r.Status == CheckWarn {
			result.HasWarning = true
		}
	}

	logger.Info("Pre-flight validation: allPassed=%v, hasWarning=%v, checks=%d",
		result.AllPassed, result.HasWarning, len(result.Checks))

	return result, nil
}
