package audio

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type Player struct{}

func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) Play(filePath string, useFocus bool) error {
	focusMgr := NewFocusManager()
	if useFocus {
		focusMgr.ApplyFocus()
		defer focusMgr.Restore()
	}

	args := []string{}
	if useFocus {
		// Boost TTS playback gain so speech is crisp over ducked background audio
		args = append(args, "-v", "2.5")
	}
	args = append(args, filePath)

	cmd := exec.Command("afplay", args...)
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
		return fmt.Errorf("playback interrupted by signal: %v", sig)
	}
}

func Cleanup(filePath string) {
	if filePath != "" {
		_ = os.Remove(filePath)
	}
}
