package audio

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
)

type Player struct{}

func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) Play(filePath string, useFocus bool, targetDevice string) error {
	deviceMgr := NewDeviceManager()
	if targetDevice != "" {
		if err := deviceMgr.SwitchDevice(targetDevice); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: %v. Using default output device.\n", err)
		} else {
			defer deviceMgr.Restore()
		}
	}

	focusMgr := NewFocusManager()
	var gainBoost float32 = 1.0

	if useFocus {
		gainBoost = focusMgr.ApplyFocus()
		defer focusMgr.Restore()
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		args := []string{}
		if useFocus {
			// Apply inverse gain boost compensation so koko remains loud (100%) while background audio is ducked (20%)
			gainStr := strconv.FormatFloat(float64(gainBoost), 'f', 2, 64)
			args = append(args, "-v", gainStr)
		}
		args = append(args, filePath)
		cmd = exec.Command("afplay", args...)
	case "windows":
		psCmd := fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", filePath)
		cmd = exec.Command("powershell", "-c", psCmd)
	default:
		if path, err := exec.LookPath("paplay"); err == nil {
			cmd = exec.Command(path, filePath)
		} else if path, err := exec.LookPath("aplay"); err == nil {
			cmd = exec.Command(path, filePath)
		} else if path, err := exec.LookPath("ffplay"); err == nil {
			cmd = exec.Command(path, "-nodisp", "-autoexit", filePath)
		} else if path, err := exec.LookPath("mpv"); err == nil {
			cmd = exec.Command(path, "--no-video", filePath)
		} else {
			return fmt.Errorf("no audio player found (install paplay, aplay, or ffplay)")
		}
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		signal.Stop(sigChan)
		if err != nil {
			return fmt.Errorf("playback failed: %w", err)
		}
		return nil
	case sig := <-sigChan:
		signal.Stop(sigChan)
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		if useFocus {
			focusMgr.Restore()
		}
		if targetDevice != "" {
			deviceMgr.Restore()
		}
		return fmt.Errorf("playback interrupted by signal: %v", sig)
	}
}

func Cleanup(filePath string) {
	if filePath != "" {
		_ = os.Remove(filePath)
	}
}
