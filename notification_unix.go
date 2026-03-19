//go:build !windows

package main

// getNotificationIconPath returns empty string on non-Windows platforms
// as beeep handles icons differently on macOS and Linux
func getNotificationIconPath() string {
	return ""
}
