package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const plistLabel = "com.4nkitd.koko"

func InstallService() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to locate home directory: %w", err)
	}

	execPath, err := exec.LookPath("koko")
	if err != nil {
		execPath = filepath.Join(homeDir, ".local", "bin", "koko")
	}

	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	plistPath := filepath.Join(launchAgentsDir, plistLabel+".plist")

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/koko_daemon.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/koko_daemon.err</string>
</dict>
</plist>
`, plistLabel, execPath)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write LaunchAgent plist: %w", err)
	}

	_ = exec.Command("launchctl", "unload", plistPath).Run()
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load LaunchAgent via launchctl: %w", err)
	}

	fmt.Printf("✅ koko daemon service successfully installed as LaunchAgent (%s)\n", plistPath)
	fmt.Println("🚀 koko daemon will now start automatically on user login.")
	return nil
}

func UninstallService() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to locate home directory: %w", err)
	}

	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", plistLabel+".plist")

	if _, err := os.Stat(plistPath); err == nil {
		_ = exec.Command("launchctl", "unload", plistPath).Run()
		_ = os.Remove(plistPath)
		fmt.Println("✅ koko daemon LaunchAgent service successfully uninstalled.")
	} else {
		fmt.Println("koko LaunchAgent service was not installed.")
	}

	return nil
}
