//go:build darwin && cgo

package audio

import (
	"os/exec"
	"strconv"
	"strings"
)

type AppVolumeInfo struct {
	AppName    string
	OrigVolume int
}

type FocusManager struct {
	duckedApps []AppVolumeInfo
	isFocused  bool
}

func NewFocusManager() *FocusManager {
	return &FocusManager{
		duckedApps: []AppVolumeInfo{},
	}
}

// Media apps that support direct volume control via AppleScript
var MediaApps = []string{
	"Spotify",
	"Music",
	"VLC",
	"QuickTime Player",
	"TV",
	"Podcasts",
}

func (fm *FocusManager) ApplyFocus() float32 {
	fm.duckedApps = []AppVolumeInfo{}

	for _, appName := range MediaApps {
		if isAppRunning(appName) {
			origVol := getAppVolume(appName)
			if origVol > 10 {
				fm.duckedApps = append(fm.duckedApps, AppVolumeInfo{
					AppName:    appName,
					OrigVolume: origVol,
				})
				setAppVolume(appName, 10) // Lower app volume to 10%
			}
		}
	}

	fm.isFocused = true
	// System volume is untouched (1.0 gain for koko)
	return 1.0
}

func (fm *FocusManager) Restore() {
	if !fm.isFocused {
		return
	}

	for _, appInfo := range fm.duckedApps {
		if isAppRunning(appInfo.AppName) {
			setAppVolume(appInfo.AppName, appInfo.OrigVolume)
		}
	}

	fm.isFocused = false
}

func (fm *FocusManager) GetGainBoost() float32 {
	return 1.0
}

func GetCurrentMasterVolume() float32 {
	return 1.0
}

func isAppRunning(appName string) bool {
	script := "application \"" + appName + "\" is running"
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

func getAppVolume(appName string) int {
	var script string
	switch appName {
	case "Spotify":
		script = "tell application \"Spotify\" to get sound volume"
	case "Music":
		script = "tell application \"Music\" to get sound volume"
	case "TV":
		script = "tell application \"TV\" to get sound volume"
	default:
		return 80
	}

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return 80
	}

	vol, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 80
	}
	return vol
}

func setAppVolume(appName string, volume int) {
	var script string
	switch appName {
	case "Spotify":
		script = "tell application \"Spotify\" to set sound volume to " + strconv.Itoa(volume)
	case "Music":
		script = "tell application \"Music\" to set sound volume to " + strconv.Itoa(volume)
	case "TV":
		script = "tell application \"TV\" to set sound volume to " + strconv.Itoa(volume)
	case "QuickTime Player":
		gain := float64(volume) / 100.0
		script = "tell application \"QuickTime Player\" to set volume of document 1 to " + strconv.FormatFloat(gain, 'f', 2, 64)
	}

	if script != "" {
		_ = exec.Command("osascript", "-e", script).Run()
	}
}
