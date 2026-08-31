//go:build !darwin || !cgo

package audio

import "fmt"

type StreamPlayer struct{}

func NewStreamPlayer() *StreamPlayer {
	return &StreamPlayer{}
}

func (sp *StreamPlayer) PlayPCM(samples []float32, sampleRate int, useFocus bool, targetDevice string) error {
	return fmt.Errorf("streaming audio player supported on macOS")
}
