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
		gainBoost: 1.8,
	}
}

func (fm *FocusManager) ApplyFocus() float32 {
	vol := float32(C.get_master_volume())
	if vol > 0 {
		fm.origVolume = vol

		// Duck master volume system-wide to 50% of original (drops all background audio: YouTube, Chrome, Spotify, games) by ~12dB
		duckVolume := vol * 0.50
		if duckVolume < 0.35 {
			duckVolume = 0.35
		}

		if vol > duckVolume {
			// Gain compensation factor so koko remains at 100% full volume (e.g. 2.0x gain boost)
			fm.gainBoost = vol / duckVolume
			C.set_master_volume(C.float(duckVolume))
		} else {
			fm.gainBoost = 1.5
		}

		fm.isFocused = true
		return fm.gainBoost
	}
	return 1.8
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
		return 1.8
	}
	return fm.gainBoost
}

func GetCurrentMasterVolume() float32 {
	return float32(C.get_master_volume())
}
