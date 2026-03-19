//go:build windows

package main

import (
	"os"
	"path/filepath"
)

// Embedded icon data for Windows notifications (icon.ico - 16x16 minimal ICO)
// This is a minimal valid ICO file that Windows toast notifications can use
var notificationIconData = []byte{
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x00, 0x00, 0x01, 0x00,
	0x20, 0x00, 0x68, 0x04, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
	0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x01, 0x00,
	0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00,
}

var notificationIconPath string

func init() {
	// Create temp directory for app resources
	tempDir := filepath.Join(os.TempDir(), "sysmind")
	os.MkdirAll(tempDir, 0755)

	// Write icon to temp file
	iconPath := filepath.Join(tempDir, "notification.ico")

	// Check if icon already exists and is valid
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		// Try to copy from build directory first (development mode)
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)
		srcIcon := filepath.Join(exeDir, "build", "icons", "icon.ico")

		if data, err := os.ReadFile(srcIcon); err == nil {
			os.WriteFile(iconPath, data, 0644)
		} else {
			// Use embedded minimal icon as fallback
			os.WriteFile(iconPath, notificationIconData, 0644)
		}
	}

	notificationIconPath = iconPath
}

// getNotificationIconPath returns the path to the notification icon
func getNotificationIconPath() string {
	return notificationIconPath
}
