//go:build darwin && cgo

package audio

/*
#cgo darwin LDFLAGS: -framework CoreAudio -framework CoreFoundation

#include <CoreAudio/CoreAudio.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static AudioDeviceID get_default_output_device() {
    AudioDeviceID devID = kAudioObjectUnknown;
    UInt32 size = sizeof(devID);
    AudioObjectPropertyAddress address = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    AudioObjectGetPropertyData(kAudioObjectSystemObject, &address, 0, NULL, &size, &devID);
    return devID;
}

static int set_default_output_device(AudioDeviceID devID) {
    UInt32 size = sizeof(devID);
    AudioObjectPropertyAddress address = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    OSStatus status = AudioObjectSetPropertyData(kAudioObjectSystemObject, &address, 0, NULL, size, &devID);
    return (status == noErr) ? 0 : -1;
}

static AudioDeviceID validate_device_id(AudioDeviceID targetID) {
    AudioObjectPropertyAddress propertyAddress = {
        kAudioHardwarePropertyDevices,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    UInt32 dataSize = 0;
    AudioObjectGetPropertyDataSize(kAudioObjectSystemObject, &propertyAddress, 0, NULL, &dataSize);
    if (dataSize == 0) return kAudioObjectUnknown;

    int numDevices = dataSize / sizeof(AudioDeviceID);
    AudioDeviceID *deviceIDs = (AudioDeviceID *)malloc(dataSize);
    if (!deviceIDs) return kAudioObjectUnknown;

    AudioObjectGetPropertyData(kAudioObjectSystemObject, &propertyAddress, 0, NULL, &dataSize, deviceIDs);
    AudioDeviceID matchedID = kAudioObjectUnknown;

    for (int i = 0; i < numDevices; i++) {
        if (deviceIDs[i] == targetID) {
            AudioObjectPropertyAddress streamsAddress = {
                kAudioDevicePropertyStreams,
                kAudioDevicePropertyScopeOutput,
                0
            };
            UInt32 streamsSize = 0;
            AudioObjectGetPropertyDataSize(deviceIDs[i], &streamsAddress, 0, NULL, &streamsSize);
            if (streamsSize > 0) {
                matchedID = targetID;
            }
            break;
        }
    }

    free(deviceIDs);
    return matchedID;
}

static AudioDeviceID find_device_by_name(const char *searchName) {
    AudioObjectPropertyAddress propertyAddress = {
        kAudioHardwarePropertyDevices,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    UInt32 dataSize = 0;
    AudioObjectGetPropertyDataSize(kAudioObjectSystemObject, &propertyAddress, 0, NULL, &dataSize);
    if (dataSize == 0) return kAudioObjectUnknown;

    int numDevices = dataSize / sizeof(AudioDeviceID);
    AudioDeviceID *deviceIDs = (AudioDeviceID *)malloc(dataSize);
    if (!deviceIDs) return kAudioObjectUnknown;

    AudioObjectGetPropertyData(kAudioObjectSystemObject, &propertyAddress, 0, NULL, &dataSize, deviceIDs);
    AudioDeviceID matchedID = kAudioObjectUnknown;

    for (int i = 0; i < numDevices; i++) {
        AudioObjectPropertyAddress streamsAddress = {
            kAudioDevicePropertyStreams,
            kAudioDevicePropertyScopeOutput,
            0
        };
        UInt32 streamsSize = 0;
        AudioObjectGetPropertyDataSize(deviceIDs[i], &streamsAddress, 0, NULL, &streamsSize);
        if (streamsSize == 0) continue;

        CFStringRef deviceName = NULL;
        UInt32 nameSize = sizeof(deviceName);
        AudioObjectPropertyAddress nameAddress = {
            kAudioDevicePropertyDeviceNameCFString,
            kAudioObjectPropertyScopeGlobal,
            kAudioObjectPropertyElementMain
        };

        if (AudioObjectGetPropertyData(deviceIDs[i], &nameAddress, 0, NULL, &nameSize, &deviceName) == noErr && deviceName) {
            char nameBuf[256];
            if (CFStringGetCString(deviceName, nameBuf, sizeof(nameBuf), kCFStringEncodingUTF8)) {
                if (strcasestr(nameBuf, searchName) != NULL) {
                    matchedID = deviceIDs[i];
                    CFRelease(deviceName);
                    break;
                }
            }
            CFRelease(deviceName);
        }
    }

    free(deviceIDs);
    return matchedID;
}

static void print_output_devices() {
    AudioObjectPropertyAddress propertyAddress = {
        kAudioHardwarePropertyDevices,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    UInt32 dataSize = 0;
    AudioObjectGetPropertyDataSize(kAudioObjectSystemObject, &propertyAddress, 0, NULL, &dataSize);
    if (dataSize == 0) return;

    int numDevices = dataSize / sizeof(AudioDeviceID);
    AudioDeviceID *deviceIDs = (AudioDeviceID *)malloc(dataSize);
    if (!deviceIDs) return;

    AudioObjectGetPropertyData(kAudioObjectSystemObject, &propertyAddress, 0, NULL, &dataSize, deviceIDs);

    for (int i = 0; i < numDevices; i++) {
        AudioObjectPropertyAddress streamsAddress = {
            kAudioDevicePropertyStreams,
            kAudioDevicePropertyScopeOutput,
            0
        };
        UInt32 streamsSize = 0;
        AudioObjectGetPropertyDataSize(deviceIDs[i], &streamsAddress, 0, NULL, &streamsSize);
        if (streamsSize == 0) continue;

        CFStringRef deviceName = NULL;
        UInt32 nameSize = sizeof(deviceName);
        AudioObjectPropertyAddress nameAddress = {
            kAudioDevicePropertyDeviceNameCFString,
            kAudioObjectPropertyScopeGlobal,
            kAudioObjectPropertyElementMain
        };

        if (AudioObjectGetPropertyData(deviceIDs[i], &nameAddress, 0, NULL, &nameSize, &deviceName) == noErr && deviceName) {
            char nameBuf[256];
            if (CFStringGetCString(deviceName, nameBuf, sizeof(nameBuf), kCFStringEncodingUTF8)) {
                printf("  • ID %u : %s\n", deviceIDs[i], nameBuf);
            }
            CFRelease(deviceName);
        }
    }

    free(deviceIDs);
    fflush(stdout);
}
*/
import "C"

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

type DeviceManager struct {
	origDeviceID C.AudioDeviceID
	isSwitched   bool
}

func NewDeviceManager() *DeviceManager {
	return &DeviceManager{}
}

func (dm *DeviceManager) SwitchDevice(deviceArg string) error {
	deviceArg = strings.TrimSpace(deviceArg)
	if deviceArg == "" {
		return nil
	}

	var matchedID C.AudioDeviceID = C.kAudioObjectUnknown

	// Check if deviceArg is a numeric Device ID (e.g. "74", "86", "101")
	if devNum, err := strconv.ParseUint(deviceArg, 10, 32); err == nil {
		matchedID = C.validate_device_id(C.AudioDeviceID(devNum))
	}

	// Fallback to name search if numeric match failed
	if matchedID == C.kAudioObjectUnknown {
		cSearch := C.CString(deviceArg)
		defer C.free(unsafe.Pointer(cSearch))
		matchedID = C.find_device_by_name(cSearch)
	}

	if matchedID == C.kAudioObjectUnknown {
		return fmt.Errorf("audio output device ID/Name '%s' not found", deviceArg)
	}

	dm.origDeviceID = C.get_default_output_device()
	if C.set_default_output_device(matchedID) == 0 {
		dm.isSwitched = true
	} else {
		return fmt.Errorf("failed to switch default audio output device to ID '%s'", deviceArg)
	}

	return nil
}

func (dm *DeviceManager) Restore() {
	if !dm.isSwitched {
		return
	}
	if dm.origDeviceID != C.kAudioObjectUnknown {
		C.set_default_output_device(dm.origDeviceID)
	}
	dm.isSwitched = false
}

func ListDevices() {
	fmt.Println("=== Available Audio Output Devices ===")
	C.print_output_devices()
	fmt.Println()
}
