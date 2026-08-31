//go:build !darwin || !cgo

package audio

import "fmt"

type DeviceManager struct{}

func NewDeviceManager() *DeviceManager {
	return &DeviceManager{}
}

func (dm *DeviceManager) SwitchDevice(deviceName string) error {
	if deviceName != "" {
		return fmt.Errorf("device selection is supported on macOS")
	}
	return nil
}

func (dm *DeviceManager) Restore() {}

func ListDevices() {
	fmt.Println("Audio device listing is supported on macOS")
}
