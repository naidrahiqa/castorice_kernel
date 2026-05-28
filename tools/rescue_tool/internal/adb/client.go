package adb

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
)

const DefaultTimeout = 30 * time.Second

// Client wraps adb.exe interactions
type Client struct {
	BinaryPath   string
	DeviceSerial string // empty = first device
	Timeout      time.Duration
}

// NewClient creates a new ADB client with the given binary path
func NewClient(binaryPath string) *Client {
	return &Client{
		BinaryPath: binaryPath,
		Timeout:    DefaultTimeout,
	}
}

// Run executes an ADB command with timeout and returns stdout output
func (c *Client) Run(args ...string) (string, error) {
	if c.DeviceSerial != "" {
		args = append([]string{"-s", c.DeviceSerial}, args...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	logger.Info("ADB exec: %s %s", c.BinaryPath, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, c.BinaryPath, args...)
	PrepareCmd(cmd)
	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))

	if ctx.Err() == context.DeadlineExceeded {
		logger.Error("ADB command timed out after %v", c.Timeout)
		return result, fmt.Errorf("adb command timed out after %v", c.Timeout)
	}

	if err != nil {
		logger.Error("ADB command failed: %v — output: %s", err, result)
		return result, fmt.Errorf("adb error: %w — %s", err, result)
	}

	return result, nil
}

// Shell runs `adb shell <cmd>` and returns the output
func (c *Client) Shell(cmd string) (string, error) {
	return c.Run("shell", cmd)
}

// Pull copies a file from the device to local filesystem
func (c *Client) Pull(remote, local string) error {
	_, err := c.Run("pull", remote, local)
	return err
}

// GetProp reads a system property via `adb shell getprop <key>`
func (c *Client) GetProp(key string) (string, error) {
	return c.Shell("getprop " + key)
}

// WaitForDevice blocks until a device is connected or timeout expires
func (c *Client) WaitForDevice(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{"wait-for-device"}
	if c.DeviceSerial != "" {
		args = append([]string{"-s", c.DeviceSerial}, args...)
	}

	cmd := exec.CommandContext(ctx, c.BinaryPath, args...)
	PrepareCmd(cmd)
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timed out waiting for device (%v)", timeout)
		}
		return fmt.Errorf("wait-for-device failed: %w", err)
	}
	return nil
}

// GetDevices returns a list of connected device serials
func (c *Client) GetDevices() ([]string, error) {
	output, err := c.Run("devices")
	if err != nil {
		return nil, err
	}

	var devices []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of") || strings.HasPrefix(line, "*") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "device" {
			devices = append(devices, parts[0])
		}
	}
	return devices, nil
}

// IsAvailable checks if the adb binary exists and is executable
func (c *Client) IsAvailable() bool {
	_, err := exec.LookPath(c.BinaryPath)
	return err == nil
}
