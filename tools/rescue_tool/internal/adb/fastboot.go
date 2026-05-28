package adb

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
)

// FastbootClient wraps fastboot.exe interactions
type FastbootClient struct {
	BinaryPath   string
	DeviceSerial string
	Timeout      time.Duration
}

// NewFastbootClient creates a new Fastboot client
func NewFastbootClient(binaryPath string) *FastbootClient {
	return &FastbootClient{
		BinaryPath: binaryPath,
		Timeout:    DefaultTimeout,
	}
}

// Run executes a fastboot command with timeout
func (f *FastbootClient) Run(args ...string) (string, error) {
	if f.DeviceSerial != "" {
		args = append([]string{"-s", f.DeviceSerial}, args...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), f.Timeout)
	defer cancel()

	logger.Info("Fastboot exec: %s %s", f.BinaryPath, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, f.BinaryPath, args...)
	PrepareCmd(cmd)
	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))

	if ctx.Err() == context.DeadlineExceeded {
		logger.Error("Fastboot command timed out after %v", f.Timeout)
		return result, fmt.Errorf("fastboot command timed out after %v", f.Timeout)
	}

	if err != nil {
		logger.Error("Fastboot command failed: %v — output: %s", err, result)
		return result, fmt.Errorf("fastboot error: %w — %s", err, result)
	}

	return result, nil
}

// Flash flashes an image to a partition
func (f *FastbootClient) Flash(partition, imagePath string) (string, error) {
	// Use longer timeout for flashing
	origTimeout := f.Timeout
	f.Timeout = 120 * time.Second
	defer func() { f.Timeout = origTimeout }()

	return f.Run("flash", partition, imagePath)
}

// Reboot reboots the device from fastboot mode
func (f *FastbootClient) Reboot() error {
	_, err := f.Run("reboot")
	return err
}

// GetDevices returns a list of devices in fastboot mode
func (f *FastbootClient) GetDevices() ([]string, error) {
	output, err := f.Run("devices")
	if err != nil {
		// fastboot devices may return exit code 1 with no devices
		// but still output empty, which is fine
		if strings.TrimSpace(output) == "" {
			return nil, nil
		}
		return nil, err
	}

	var devices []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "fastboot" {
			devices = append(devices, parts[0])
		}
	}
	return devices, nil
}

// GetVar reads a fastboot variable
func (f *FastbootClient) GetVar(name string) (string, error) {
	output, err := f.Run("getvar", name)
	if err != nil {
		// fastboot getvar outputs to stderr, check combined output
		if strings.Contains(output, name+":") {
			parts := strings.SplitN(output, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// IsAvailable checks if the fastboot binary exists
func (f *FastbootClient) IsAvailable() bool {
	_, err := exec.LookPath(f.BinaryPath)
	return err == nil
}
