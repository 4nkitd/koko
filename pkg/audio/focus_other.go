//go:build !darwin || !cgo

package audio

type FocusManager struct{}

func NewFocusManager() *FocusManager {
	return &FocusManager{}
}

func (fm *FocusManager) ApplyFocus() {}

func (fm *FocusManager) Restore() {}
