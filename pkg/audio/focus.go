//go:build darwin && cgo

package audio

/*
#cgo darwin LDFLAGS: -framework CoreAudio -framework AudioToolbox

#include <CoreAudio/CoreAudio.h>

static float get_master_volume() {
    AudioDeviceID defaultOutputDeviceID = kAudioObjectUnknown;
    UInt32 size = sizeof(defaultOutputDeviceID);
    AudioObjectPropertyAddress propertyAddress = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(
        kAudioObjectSystemObject,
        &propertyAddress,
        0, NULL, &size, &defaultOutputDeviceID
    );
    if (status != noErr) return -1.0f;

    Float32 volume = 0.0f;
    size = sizeof(volume);
    AudioObjectPropertyAddress volumeAddress = {
        kAudioDevicePropertyVolumeScalar,
        kAudioDevicePropertyScopeOutput,
        0
    };

    status = AudioObjectGetPropertyData(
        defaultOutputDeviceID,
        &volumeAddress,
        0, NULL, &size, &volume
    );
    if (status != noErr) {
        volumeAddress.mElement = 1;
        status = AudioObjectGetPropertyData(
            defaultOutputDeviceID,
            &volumeAddress,
            0, NULL, &size, &volume
        );
    }
    if (status != noErr) return -1.0f;
    return volume;
}

static int set_master_volume(float volume) {
    AudioDeviceID defaultOutputDeviceID = kAudioObjectUnknown;
    UInt32 size = sizeof(defaultOutputDeviceID);
    AudioObjectPropertyAddress propertyAddress = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(
        kAudioObjectSystemObject,
        &propertyAddress,
        0, NULL, &size, &defaultOutputDeviceID
    );
    if (status != noErr) return -1;

    AudioObjectPropertyAddress volumeAddress = {
        kAudioDevicePropertyVolumeScalar,
        kAudioDevicePropertyScopeOutput,
        0
    };

    size = sizeof(volume);
    status = AudioObjectSetPropertyData(
        defaultOutputDeviceID,
        &volumeAddress,
        0, NULL, size, &volume
    );

    volumeAddress.mElement = 1;
    AudioObjectSetPropertyData(defaultOutputDeviceID, &volumeAddress, 0, NULL, size, &volume);
    volumeAddress.mElement = 2;
    AudioObjectSetPropertyData(defaultOutputDeviceID, &volumeAddress, 0, NULL, size, &volume);

    return 0;
}
*/
import "C"

type FocusManager struct {
	origVolume float32
	gainBoost  float32
	isFocused  bool
}

func NewFocusManager() *FocusManager {
	return &FocusManager{
		gainBoost: 2.5,
	}
}

func (fm *FocusManager) ApplyFocus() float32 {
	vol := float32(C.get_master_volume())
	if vol > 0 {
		fm.origVolume = vol
		// Duck master volume by 60% (e.g. 80% dips down to 32%) for background audio
		duckVolume := vol * 0.40
		if duckVolume < 0.25 {
			duckVolume = 0.25
		}

		if vol > duckVolume {
			// 2.5x gain boost ensures koko remains 100% full volume over ducked background
			fm.gainBoost = vol / duckVolume
			C.set_master_volume(C.float(duckVolume))
		} else {
			fm.gainBoost = 2.0
		}

		fm.isFocused = true
		return fm.gainBoost
	}
	return 2.5
}

func (fm *FocusManager) Restore() {
	if !fm.isFocused {
		return
	}
	if fm.origVolume > 0 {
		C.set_master_volume(C.float(fm.origVolume))
	}
	fm.isFocused = false
}

func (fm *FocusManager) GetGainBoost() float32 {
	if fm.gainBoost <= 0 {
		return 2.5
	}
	return fm.gainBoost
}

func GetCurrentMasterVolume() float32 {
	return float32(C.get_master_volume())
}
