//go:build !darwin || !cgo

package audio

type FocusManager struct{}

func NewFocusManager() *FocusManager {
	return &FocusManager{}
}

func (fm *FocusManager) ApplyFocus() float32 {
	return 1.0
}

func (fm *FocusManager) Restore() {}

func (fm *FocusManager) GetGainBoost() float32 {
	return 1.0
}

func GetCurrentMasterVolume() float32 {
	return 1.0
}
