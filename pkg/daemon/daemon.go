package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/4nkitd/koko/pkg/audio"
	"github.com/4nkitd/koko/pkg/config"
	"github.com/4nkitd/koko/pkg/engine"
)

const SocketPath = "/tmp/koko_daemon.sock"

type Request struct {
	Text    string  `json:"text"`
	Voice   string  `json:"voice"`
	Speed   float64 `json:"speed"`
	OutPath string  `json:"out_path"`
	Focus   bool    `json:"focus"`
	NoPlay  bool    `json:"no_play"`
}

type Response struct {
	Success  bool   `json:"success"`
	WavPath  string `json:"wav_path"`
	ErrorMsg string `json:"error_msg"`
}

func StartDaemon(cfg *config.Config) error {
	_ = os.Remove(SocketPath)

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return fmt.Errorf("failed to bind daemon Unix socket at %s: %w", SocketPath, err)
	}
	defer listener.Close()
	defer os.Remove(SocketPath)

	_ = os.Chmod(SocketPath, 0700)

	fmt.Printf("🚀 koko daemon server listening on %s (IPC ready for sub-30ms execution)\n", SocketPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	eng := engine.NewEngine(cfg)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleConn(conn, eng)
		}
	}()

	<-sigChan
	fmt.Println("\nStopping koko daemon...")
	return nil
}

func handleConn(conn net.Conn, eng *engine.Engine) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var req Request
	if err := decoder.Decode(&req); err != nil {
		_ = encoder.Encode(Response{Success: false, ErrorMsg: err.Error()})
		return
	}

	opts := engine.SynthesizeOptions{
		Text:    req.Text,
		Voice:   req.Voice,
		Speed:   req.Speed,
		OutPath: req.OutPath,
	}

	wavFile, err := eng.Synthesize(opts)
	if err != nil {
		_ = encoder.Encode(Response{Success: false, ErrorMsg: err.Error()})
		return
	}

	if !req.NoPlay {
		player := audio.NewPlayer()
		_ = player.Play(wavFile, req.Focus)
	}

	_ = encoder.Encode(Response{Success: true, WavPath: wavFile})

	if req.OutPath == "" {
		audio.Cleanup(wavFile)
	}
}

func TrySendDaemon(req Request) (bool, string, error) {
	if _, err := os.Stat(SocketPath); err != nil {
		return false, "", nil
	}

	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return false, "", nil
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return false, "", err
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return false, "", err
	}

	if !resp.Success {
		return true, "", fmt.Errorf("%s", resp.ErrorMsg)
	}

	return true, resp.WavPath, nil
}
