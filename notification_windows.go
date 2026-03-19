//go:build windows

package main

import (
	"os"
	"path/filepath"
)

// Embedded icon data for Windows notifications (minimal valid ICO header)
// This is a fallback minimal valid ICO file for toast notifications
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
		// Try to find icon in multiple locations
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)

		// List of possible icon locations (in order of preference)
		iconLocations := []string{
			filepath.Join(exeDir, "build", "windows", "icon.ico"),       // Standard Wails location
			filepath.Join(exeDir, "build", "icons", "icon.ico"),         // Legacy location
			filepath.Join(exeDir, "icon.ico"),                           // Same directory as exe
			filepath.Join(exeDir, "..", "build", "windows", "icon.ico"), // Dev mode
		}

		var iconFound bool
		for _, srcIcon := range iconLocations {
			if data, err := os.ReadFile(srcIcon); err == nil {
				os.WriteFile(iconPath, data, 0644)
				iconFound = true
				break
			}
		}

		if !iconFound {
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
