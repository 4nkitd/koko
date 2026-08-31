//go:build darwin && cgo

package audio

/*
#cgo darwin LDFLAGS: -framework AudioToolbox -framework CoreAudio -framework CoreFoundation

#include <AudioToolbox/AudioToolbox.h>
#include <CoreAudio/CoreAudio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>

typedef struct {
    AudioQueueRef queue;
    AudioQueueBufferRef buffers[3];
    float gain;
    float *pcm_samples;
    int total_samples;
    int current_sample;
    int done;
} StreamState;

static void stream_audio_callback(void *userData, AudioQueueRef inAQ, AudioQueueBufferRef inBuffer) {
    StreamState *state = (StreamState *)userData;
    int samplesPerBuffer = inBuffer->mAudioDataBytesCapacity / sizeof(float);
    float *outBuffer = (float *)inBuffer->mAudioData;
    int numToCopy = state->total_samples - state->current_sample;

    if (numToCopy > samplesPerBuffer) {
        numToCopy = samplesPerBuffer;
    }

    if (numToCopy > 0) {
        for (int i = 0; i < numToCopy; i++) {
            float rawSample = state->pcm_samples[state->current_sample + i] * state->gain;
            // Soft-saturation curve (tanh) for maximum loudness without clipping distortion
            float sample = tanhf(rawSample);
            outBuffer[i] = sample;
        }
        state->current_sample += numToCopy;
        inBuffer->mAudioDataByteSize = numToCopy * sizeof(float);
        AudioQueueEnqueueBuffer(inAQ, inBuffer, 0, NULL);
    } else {
        inBuffer->mAudioDataByteSize = 0;
        state->done = 1;
    }
}

static StreamState* create_stream_player(float *samples, int numSamples, int sampleRate, float gain) {
    StreamState *state = (StreamState *)malloc(sizeof(StreamState));
    memset(state, 0, sizeof(StreamState));
    state->pcm_samples = samples;
    state->total_samples = numSamples;
    state->gain = gain;

    AudioStreamBasicDescription fmt;
    memset(&fmt, 0, sizeof(fmt));
    fmt.mSampleRate = (double)sampleRate;
    fmt.mFormatID = kAudioFormatLinearPCM;
    fmt.mFormatFlags = kAudioFormatFlagIsFloat | kAudioFormatFlagIsPacked;
    fmt.mFramesPerPacket = 1;
    fmt.mChannelsPerFrame = 1;
    fmt.mBitsPerChannel = 32;
    fmt.mBytesPerPacket = 4;
    fmt.mBytesPerFrame = 4;

    OSStatus status = AudioQueueNewOutput(&fmt, stream_audio_callback, state, NULL, NULL, 0, &state->queue);
    if (status != noErr) {
        free(state);
        return NULL;
    }

    int bufferSize = 4096 * sizeof(float);
    for (int i = 0; i < 3; i++) {
        AudioQueueAllocateBuffer(state->queue, bufferSize, &state->buffers[i]);
        stream_audio_callback(state, state->queue, state->buffers[i]);
    }

    AudioQueueStart(state->queue, NULL);
    return state;
}

static void wait_and_free_stream_player(StreamState *state) {
    if (!state) return;
    while (!state->done) {
        usleep(10000);
    }
    AudioQueueStop(state->queue, true);
    AudioQueueDispose(state->queue, true);
    free(state);
}
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"
)

type StreamPlayer struct{}

func NewStreamPlayer() *StreamPlayer {
	return &StreamPlayer{}
}

func (sp *StreamPlayer) PlayPCM(samples []float32, sampleRate int, useFocus bool, targetDevice string) error {
	if len(samples) == 0 {
		return fmt.Errorf("empty PCM audio buffer")
	}

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

	cSamples := (*C.float)(unsafe.Pointer(&samples[0]))
	cState := C.create_stream_player(cSamples, C.int(len(samples)), C.int(sampleRate), C.float(gainBoost))
	if cState == nil {
		return fmt.Errorf("failed to create CoreAudio streaming player")
	}

	C.wait_and_free_stream_player(cState)
	return nil
}
