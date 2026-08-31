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

var MediaApps = []string{
	"Spotify",
	"Music",
	"VLC",
	"QuickTime Player",
	"TV",
	"Podcasts",
}

var Browsers = []string{
	"Google Chrome",
	"Safari",
	"Arc",
	"Brave Browser",
	"Microsoft Edge",
}

func (fm *FocusManager) ApplyFocus() float32 {
	fm.duckedApps = []AppVolumeInfo{}

	// 1. Duck desktop media players to 10% volume (System master volume left 100% untouched)
	for _, appName := range MediaApps {
		if isAppRunning(appName) {
			origVol := getAppVolume(appName)
			if origVol > 10 {
				fm.duckedApps = append(fm.duckedApps, AppVolumeInfo{
					AppName:    appName,
					OrigVolume: origVol,
				})
				setAppVolume(appName, 10)
			}
		}
	}

	// 2. Duck browser media elements (YouTube, Twitch, web audio) across open tabs to 15% volume
	for _, browserName := range Browsers {
		if isAppRunning(browserName) {
			duckBrowserMedia(browserName)
		}
	}

	fm.isFocused = true
	return 1.0
}

func (fm *FocusManager) Restore() {
	if !fm.isFocused {
		return
	}

	// 1. Restore desktop media player volumes
	for _, appInfo := range fm.duckedApps {
		if isAppRunning(appInfo.AppName) {
			setAppVolume(appInfo.AppName, appInfo.OrigVolume)
		}
	}

	// 2. Restore browser media tab volumes
	for _, browserName := range Browsers {
		if isAppRunning(browserName) {
			restoreBrowserMedia(browserName)
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

func duckBrowserMedia(browserName string) {
	var script string
	switch browserName {
	case "Google Chrome", "Arc", "Brave Browser", "Microsoft Edge":
		script = `tell application "` + browserName + `"
			repeat with w in windows
				repeat with t in tabs of w
					try
						execute t javascript "document.querySelectorAll('video, audio').forEach(e => { if(!e.dataset.origVol) e.dataset.origVol = e.volume; e.volume = 0.15; });"
					end try
				end repeat
			end repeat
		end tell`
	case "Safari":
		script = `tell application "Safari"
			repeat with w in windows
				repeat with t in tabs of w
					try
						do JavaScript "document.querySelectorAll('video, audio').forEach(e => { if(!e.dataset.origVol) e.dataset.origVol = e.volume; e.volume = 0.15; });" in t
					end try
				end repeat
			end repeat
		end tell`
	}

	if script != "" {
		_ = exec.Command("osascript", "-e", script).Run()
	}
}

func restoreBrowserMedia(browserName string) {
	var script string
	switch browserName {
	case "Google Chrome", "Arc", "Brave Browser", "Microsoft Edge":
		script = `tell application "` + browserName + `"
			repeat with w in windows
				repeat with t in tabs of w
					try
						execute t javascript "document.querySelectorAll('video, audio').forEach(e => { if(e.dataset.origVol) e.volume = parseFloat(e.dataset.origVol); });"
					end try
				end repeat
			end repeat
		end tell`
	case "Safari":
		script = `tell application "Safari"
			repeat with w in windows
				repeat with t in tabs of w
					try
						do JavaScript "document.querySelectorAll('video, audio').forEach(e => { if(e.dataset.origVol) e.volume = parseFloat(e.dataset.origVol); });" in t
					end try
				end repeat
			end repeat
		end tell`
	}

	if script != "" {
		_ = exec.Command("osascript", "-e", script).Run()
	}
}
