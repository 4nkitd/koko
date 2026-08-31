package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/4nkitd/koko/pkg/audio"
	"github.com/4nkitd/koko/pkg/config"
	"github.com/4nkitd/koko/pkg/engine"
)

type SpeechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	Speed          float64 `json:"speed"`
	ResponseFormat string  `json:"response_format"`
}

func StartServer(cfg *config.Config, port int) error {
	eng := engine.NewEngine(cfg)

	http.HandleFunc("/v1/audio/speech", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
			return
		}

		var req SpeechRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON body: %v", err), http.StatusBadRequest)
			return
		}

		if req.Input == "" {
			http.Error(w, "Field 'input' is required.", http.StatusBadRequest)
			return
		}

		voice := req.Voice
		if voice == "" {
			voice = "ironman"
		}

		speed := req.Speed
		if speed <= 0 {
			speed = 1.0
		}

		opts := engine.SynthesizeOptions{
			Text:  req.Input,
			Voice: voice,
			Speed: speed,
		}

		wavFile, err := eng.Synthesize(opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Synthesis error: %v", err), http.StatusInternalServerError)
			return
		}
		defer audio.Cleanup(wavFile)

		f, err := os.Open(wavFile)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read audio output: %v", err), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Transfer-Encoding", "chunked")
		_, _ = io.Copy(w, f)
	})

	http.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data": []map[string]string{
				{"id": "tts-1", "object": "model", "owned_by": "koko"},
				{"id": "tts-1-hd", "object": "model", "owned_by": "koko"},
				{"id": "kokoro-82m", "object": "model", "owned_by": "koko"},
			},
		})
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🚀 koko OpenAI-compatible REST server listening on http://localhost%s (POST /v1/audio/speech)\n", addr)
	return http.ListenAndServe(addr, nil)
}
