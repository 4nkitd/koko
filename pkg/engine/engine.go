package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/4nkitd/koko/pkg/audio"
	"github.com/4nkitd/koko/pkg/config"
)

type SynthesizeOptions struct {
	Mode    string
	Text    string
	Voice   string
	Speed   float64
	OutPath string
	Model   string
	Device  string
	Stream  bool
	Verbose bool
}

type Engine struct {
	cfg *config.Config
}

func NewEngine(cfg *config.Config) *Engine {
	return &Engine{cfg: cfg}
}

// VoiceMap maps character names & voices to exact kokoro-en-v0_19 speaker IDs
var VoiceMap = map[string]int{
	"friday":      7, // bf_emma (British female AI)
	"f.r.i.d.a.y": 7,
	"jarvis":      9, // bm_george (British male AI)
	"j.a.r.v.i.s": 9,
	"ironman":     6, // am_michael (American male - Default)
	"stark":       6,
	"tony":        6,
	"adam":        5,
	"af_bella":    1,
	"af_nicole":   2,
	"af_sarah":    3,
	"af_sky":      4,
	"am_adam":     5,
	"am_michael":  6,
	"bf_emma":     7,
	"bf_isabella": 8,
	"bm_george":   9,
	"bm_lewis":    10,
}

var SayVoiceMap = map[string]string{
	"friday":  "Victoria",
	"jarvis":  "Daniel",
	"ironman": "Alex",
	"stark":   "Alex",
	"tony":    "Alex",
}

func (e *Engine) Synthesize(opts SynthesizeOptions) (string, error) {
	voiceName := strings.ToLower(opts.Voice)
	if voiceName == "" {
		voiceName = strings.ToLower(e.cfg.Monotone.Voice)
	}

	sid, exists := VoiceMap[voiceName]
	if !exists {
		sid = 6 // Default to Iron Man
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	modelDir := filepath.Join(homeDir, ".config", "koko", "onnx", "kokoro-en-v0_19")
	modelPath := filepath.Join(modelDir, "model.onnx")
	voicesPath := filepath.Join(modelDir, "voices.bin")
	tokensPath := filepath.Join(modelDir, "tokens.txt")
	dataDir := filepath.Join(modelDir, "espeak-ng-data")

	executable, errLoc := locateSherpaBinary(opts.Stream)
	_, errStat := os.Stat(modelPath)

	if errLoc != nil || errStat != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Native ONNX engine/models unavailable. Falling back to macOS speech engine.\n")
		return e.fallbackSay(opts.Text, voiceName, opts.OutPath)
	}

	// If streaming with specific audio output device, switch default device first
	if opts.Stream && opts.Device != "" {
		devMgr := audio.NewDeviceManager()
		if err := devMgr.SwitchDevice(opts.Device); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: %v. Using default output device.\n", err)
		} else {
			defer devMgr.Restore()
		}
	}

	outDir := os.TempDir()
	prefix := fmt.Sprintf("koko_%d", time.Now().UnixNano())
	targetWav := filepath.Join(outDir, prefix+".wav")
	if opts.OutPath != "" {
		targetWav = opts.OutPath
	}

	numThreads := runtime.NumCPU()
	if numThreads > 8 {
		numThreads = 8
	}

	args := []string{
		"--num-threads=" + strconv.Itoa(numThreads),
		"--kokoro-model=" + modelPath,
		"--kokoro-voices=" + voicesPath,
		"--kokoro-tokens=" + tokensPath,
		"--kokoro-data-dir=" + dataDir,
		"--sid=" + strconv.Itoa(sid),
	}

	if !opts.Stream {
		args = append(args, "--output-filename="+targetWav)
	}

	args = append(args, opts.Text)

	cmd := exec.Command(executable, args...)
	cmd.Env = os.Environ()

	if opts.Verbose || opts.Stream {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if !opts.Stream {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: Native C++ TTS execution error: %v. Falling back to macOS speech engine.\n", err)
			return e.fallbackSay(opts.Text, voiceName, opts.OutPath)
		}
		return "", err
	}

	if opts.Stream {
		return "", nil
	}

	if _, err := os.Stat(targetWav); err != nil {
		return "", fmt.Errorf("generated WAV not found at %s: %w", targetWav, err)
	}

	return targetWav, nil
}

func (e *Engine) fallbackSay(text, voiceName, outPath string) (string, error) {
	sayVoice, exists := SayVoiceMap[voiceName]
	if !exists {
		sayVoice = "Alex"
	}

	outDir := os.TempDir()
	prefix := fmt.Sprintf("koko_fallback_%d", time.Now().UnixNano())
	targetAudio := filepath.Join(outDir, prefix+".aiff")
	if outPath != "" {
		targetAudio = outPath
	}

	args := []string{"-v", sayVoice, text}
	if outPath != "" {
		args = append([]string{"-o", targetAudio}, args...)
	}

	cmd := exec.Command("say", args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("macOS native 'say' fallback failed: %w", err)
	}

	return targetAudio, nil
}

func locateSherpaBinary(stream bool) (string, error) {
	homeDir, _ := os.UserHomeDir()
	binName := "sherpa-onnx-offline-tts"
	if stream {
		binName = "sherpa-onnx-offline-tts-play"
	}

	candidates := []string{
		binName,
		filepath.Join(homeDir, ".local", "bin", binName),
		filepath.Join("/usr/local/bin", binName),
	}

	for _, cand := range candidates {
		if path, err := exec.LookPath(cand); err == nil {
			return path, nil
		}
		if _, err := os.Stat(cand); err == nil {
			return cand, nil
		}
	}

	return "", fmt.Errorf("%s executable not found", binName)
}
