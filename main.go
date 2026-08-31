package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/4nkitd/koko/pkg/audio"
	"github.com/4nkitd/koko/pkg/config"
	"github.com/4nkitd/koko/pkg/daemon"
	"github.com/4nkitd/koko/pkg/engine"
	"github.com/4nkitd/koko/pkg/server"
	"github.com/4nkitd/koko/pkg/service"
	"github.com/4nkitd/koko/pkg/setup"
	"github.com/4nkitd/koko/pkg/voices"
)

const version = "0.1.0"

func main() {
	voiceFlag := flag.String("voice", "", "Character voice: 'ironman', 'friday', 'jarvis', etc.")
	flag.StringVar(voiceFlag, "v", "", "Character voice (shorthand)")

	deviceFlag := flag.String("device", "", "Target audio output device name (e.g. 'MacBook Pro Speakers', 'AirPods')")
	flag.StringVar(deviceFlag, "d", "", "Target audio output device (shorthand)")

	ironmanFlag := flag.Bool("ironman", false, "Speak using Iron Man / Tony Stark voice")
	starkFlag := flag.Bool("stark", false, "Speak using Iron Man / Tony Stark voice")
	fridayFlag := flag.Bool("friday", false, "Speak using F.R.I.D.A.Y. voice")
	jarvisFlag := flag.Bool("jarvis", false, "Speak using J.A.R.V.I.S. voice")

	speedFlag := flag.Float64("speed", 1.0, "Speech rate multiplier (default: 1.0)")
	flag.Float64Var(speedFlag, "s", 1.0, "Speech rate multiplier (shorthand)")

	outFlag := flag.String("output", "", "Save generated speech to WAV output file path")
	flag.StringVar(outFlag, "o", "", "Save generated speech to WAV output file path (shorthand)")

	focusFlag := flag.Bool("focus", false, "Enable Audio Focus (ducks background audio system-wide)")
	flag.BoolVar(focusFlag, "f", false, "Enable Audio Focus (shorthand)")

	streamFlag := flag.Bool("stream", false, "Enable direct CoreAudio streaming mode for sub-50ms latency")

	portFlag := flag.Int("port", 8848, "Port for koko OpenAI-compatible REST HTTP server")

	noPlayFlag := flag.Bool("no-play", false, "Disable automatic audio playback")
	daemonFlag := flag.Bool("daemon", false, "Start background daemon server for sub-30ms IPC execution")
	listVoicesFlag := flag.Bool("list-voices", false, "List available character voices")
	flag.BoolVar(listVoicesFlag, "l", false, "List available voices (shorthand)")
	listDevicesFlag := flag.Bool("list-devices", false, "List active macOS audio output devices")

	configFlag := flag.String("config", "", "Custom path to config.json")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose debug output")
	versionFlag := flag.Bool("version", false, "Show koko CLI version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: koko [options] [text|command]\n\n")
		fmt.Fprintf(os.Stderr, "High-Performance Standalone Text-To-Speech CLI & Server for macOS.\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  koko server [--port 8848]   Start OpenAI-compatible HTTP REST server\n")
		fmt.Fprintf(os.Stderr, "  koko service install        Install background macOS LaunchAgent service\n")
		fmt.Fprintf(os.Stderr, "  koko service uninstall      Uninstall background macOS LaunchAgent service\n")
		fmt.Fprintf(os.Stderr, "  koko daemon                 Start background Unix socket daemon\n")
		fmt.Fprintf(os.Stderr, "  koko setup                  Initialize ONNX models and runtime\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  koko \"I am Iron Man.\"\n")
		fmt.Fprintf(os.Stderr, "  koko -d \"MacBook Pro Speakers\" \"Output routed to specific speaker.\"\n")
		fmt.Fprintf(os.Stderr, "  koko --stream \"Sub-50ms instant streaming speech synthesis.\"\n")
		fmt.Fprintf(os.Stderr, "  koko --ironman \"Sometimes you gotta run before you can walk.\"\n")
		fmt.Fprintf(os.Stderr, "  koko --friday \"F.R.I.D.A.Y. online.\"\n")
		fmt.Fprintf(os.Stderr, "  koko --jarvis \"Allow me to introduce myself. I am J.A.R.V.I.S.\"\n")
		fmt.Fprintf(os.Stderr, "  koko -f \"Priority notification. Background audio ducked.\"\n")
		fmt.Fprintf(os.Stderr, "  echo \"Status check nominal.\" | koko\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()

	cfg, err := config.LoadConfig(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		switch args[0] {
		case "setup":
			if err := setup.EnsureInstalled(true); err != nil {
				fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "daemon":
			if err := daemon.StartDaemon(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "server":
			if err := setup.EnsureInstalled(*verboseFlag); err != nil {
				fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
				os.Exit(1)
			}
			if err := server.StartServer(cfg, *portFlag); err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "service":
			if len(args) > 1 && args[1] == "uninstall" {
				if err := service.UninstallService(); err != nil {
					fmt.Fprintf(os.Stderr, "Service uninstall error: %v\n", err)
					os.Exit(1)
				}
			} else {
				if err := service.InstallService(); err != nil {
					fmt.Fprintf(os.Stderr, "Service install error: %v\n", err)
					os.Exit(1)
				}
			}
			os.Exit(0)
		}
	}

	if *daemonFlag {
		if err := daemon.StartDaemon(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("koko version %s\n", version)
		os.Exit(0)
	}

	if *listVoicesFlag {
		voices.PrintVoices()
		os.Exit(0)
	}

	if *listDevicesFlag {
		audio.ListDevices()
		os.Exit(0)
	}

	selectedVoice := *voiceFlag
	if *ironmanFlag || *starkFlag {
		selectedVoice = "ironman"
	} else if *fridayFlag {
		selectedVoice = "friday"
	} else if *jarvisFlag {
		selectedVoice = "jarvis"
	}

	if selectedVoice == "" {
		selectedVoice = cfg.Monotone.Voice
	}

	text := strings.Join(args, " ")
	text = strings.TrimSpace(text)

	if text == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			reader := bufio.NewReader(os.Stdin)
			var sb strings.Builder
			for {
				line, err := reader.ReadString('\n')
				sb.WriteString(line)
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
					os.Exit(1)
				}
			}
			text = strings.TrimSpace(sb.String())
		}
	}

	if text == "" {
		flag.Usage()
		os.Exit(1)
	}

	if !*streamFlag {
		daemonReq := daemon.Request{
			Text:    text,
			Voice:   selectedVoice,
			Speed:   *speedFlag,
			OutPath: *outFlag,
			Device:  *deviceFlag,
			Focus:   *focusFlag,
			NoPlay:  *noPlayFlag,
		}

		handledByDaemon, wavPath, daemonErr := daemon.TrySendDaemon(daemonReq)
		if handledByDaemon {
			if daemonErr != nil {
				fmt.Fprintf(os.Stderr, "Daemon execution error: %v\n", daemonErr)
				os.Exit(1)
			}
			if *outFlag != "" {
				fmt.Printf("Saved audio to: %s\n", wavPath)
			}
			os.Exit(0)
		}
	}

	_ = setup.EnsureInstalled(*verboseFlag)

	opts := engine.SynthesizeOptions{
		Text:    text,
		Voice:   selectedVoice,
		Speed:   *speedFlag,
		OutPath: *outFlag,
		Device:  *deviceFlag,
		Focus:   *focusFlag,
		Stream:  *streamFlag,
		Verbose: *verboseFlag,
	}

	eng := engine.NewEngine(cfg)
	wavFile, err := eng.Synthesize(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Speech synthesis failed: %v\n", err)
		os.Exit(1)
	}

	if !*streamFlag {
		shouldPlay := !*noPlayFlag
		if shouldPlay {
			player := audio.NewPlayer()
			if err := player.Play(wavFile, *focusFlag, *deviceFlag); err != nil {
				fmt.Fprintf(os.Stderr, "Audio playback error: %v\n", err)
			}
		}

		if *outFlag == "" {
			audio.Cleanup(wavFile)
		} else {
			fmt.Printf("Saved audio to: %s\n", wavFile)
		}
	}
}
