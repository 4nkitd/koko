<div align="center">

# 🗣️ koko

### Fast, standalone, zero-dependency Text-to-Speech CLI & Server for macOS.

[![Release](https://img.shields.io/github/v/release/4nkitd/koko?color=00ADD8&style=flat-square)](https://github.com/4nkitd/koko/releases)
[![Homebrew](https://img.shields.io/badge/homebrew-4nkitd%2Ftap%2Fkoko-orange?style=flat-square)](https://github.com/4nkitd/homebrew-tap)
[![License: MIT](https://img.shields.io/github/license/4nkitd/koko?style=flat-square)](LICENSE)

Get up and running with high-speed neural text-to-speech on your Mac in seconds.

[Quickstart](#quickstart) • [REST Server](#-openai-compatible-rest-server-http-v1audiospeech) • [Character Voices](#character-voices) • [Benchmarks](#-performance-benchmarks) • [Audio Focus](#audio-focus) • [Daemon Mode](#daemon-mode) • [AI Integration](#-ai-agent-integration-llmstxt)

---

</div>

## Quickstart

### Install via Homebrew (Recommended)

```bash
brew install 4nkitd/tap/koko
```

### Install via Go

```bash
go install github.com/4nkitd/koko@latest
```

### Run your first command

```bash
koko "Sometimes you gotta run before you can walk."
```

---

## 🔌 OpenAI-Compatible REST Server (`http://v1/audio/speech`)

Launch `koko` as a local OpenAI-compatible HTTP REST server:

```bash
koko server --port 8848
```

Any app, web frontend, VS Code extension, or script designed for OpenAI TTS can stream audio locally with zero code changes:

```bash
curl -X POST http://localhost:8848/v1/audio/speech \
  -H "Content-Type: application/json" \
  -d '{"input": "OpenAI API compatibility test passed.", "voice": "ironman"}' \
  --output response.wav
```

---

## 🚦 macOS Auto-Starting Service (`koko service install`)

Automatically register `koko daemon` as a native macOS LaunchAgent service so it stays pre-warmed on system login:

```bash
# Install & load LaunchAgent service
koko service install

# Uninstall service
koko service uninstall
```

---

## 📊 Performance Benchmarks

`koko` was benchmarked directly against PyTorch/Python TTS implementations (`mlx_audio`) running the **exact same neural model (`Kokoro-82M`)** and text input (*"Sometimes you gotta run before you can walk."*), measured via macOS hardware performance counters (`/usr/bin/time -l`):

| Metric / Hardware Resource | PyTorch / Python (`mlx_audio`) | `koko` CLI (Native C++ ONNX) | Improvement Factor |
| :--- | :--- | :--- | :--- |
| ⏱️ **Wall Clock Execution Time** | **`2.66 s`** | **`0.70 s`** | **3.8x Faster** |
| 💻 **CPU Cycles Elapsed** | **`14.8 Billion`** | **`15.6 Million`** | **948x Fewer Cycles** |
| ⚙️ **CPU Instructions Retired** | **`22.7 Billion`** | **`41.4 Million`** | **550x Fewer Instructions** |
| 🛠️ **OS System Overhead (`sys`)** | **`2.31 s`** | **`0.10 s`** | **23x Less OS Overhead** |
| 🧠 **Peak Memory Footprint** | **`2,352 MB` (2.35 GB)** | **`3.6 MB`** | **653x Less Memory** |
| 🔄 **Context Switches** | **`12,424`** | **`375`** | **33x Less Thread Thrashing** |

---

## 🔍 Open Source Tool Comparison

| Feature / Metric | 🗣️ `koko` (Ours) | 🥧 Piper TTS | 📻 Pocket TTS | ☁️ ElevenLabs CLI | 🍎 macOS `say` |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Engine / Architecture** | **Kokoro-82M C++ ONNX** | VITS + ONNX | Flow-matching | Cloud REST API | Legacy Synthesizer |
| **Execution Latency** | **`0.70s`** *(<30ms Daemon)* | ~1.20s | ~0.60s | 300ms + Net Latency | 0.10s |
| **Voice Quality** | ⭐⭐⭐⭐⭐ (Studio Neural) | ⭐⭐⭐ (Robotic) | ⭐⭐⭐ (Muffled) | ⭐⭐⭐⭐⭐ (Studio) | ⭐⭐ (Legacy) |
| **System Audio Focus (`-f`)** | ✅ **Yes (<1ms HAL)** | ❌ No | ❌ No | ❌ No | ❌ No |
| **Zero Python Dependencies** | ✅ **100% Native** | ❌ Requires Python | ❌ Needs wrapper | ❌ Requires Python | ✅ Native |
| **Licensing** | **MIT License** | ⚠️ GPL-3.0 | MIT License | 💳 Paid Subscription | Closed / Native |

---

## 🤖 AI Agent Integration (`llms.txt`)

`koko` includes machine-readable instruction files (`llms.txt` & `llm-instructions.txt`) designed for AI coding agents (**OpenCode, Claude Code, Cursor, Windsurf, Aider**).

### Give this prompt to your AI Agent:

> *"Please read https://raw.githubusercontent.com/4nkitd/koko/main/llms.txt and configure yourself to use `koko` for speaking updates back to me."*

Your AI agent will automatically install `koko` and use it to speak responses, task completions, and status updates out loud.

---

## Character Voices

`koko` comes built-in with instant character flags modeled after iconic AI assistants and characters:

```bash
koko --ironman "I am Iron Man."
koko --friday  "F.R.I.D.A.Y. online and operational."
koko --jarvis  "Allow me to introduce myself. I am J.A.R.V.I.S."
```

| Flag | Voice Preset | Gender | Accent / Style |
| :--- | :--- | :--- | :--- |
| `--ironman` / `--stark` | Tony Stark *(Default)* | Male | American English |
| `--friday` | F.R.I.D.A.Y. | Female | British / Irish English |
| `--jarvis` | J.A.R.V.I.S. | Male | British English |
| `-v am_michael` | Michael | Male | American English |
| `-v am_adam` | Adam | Male | Deep American English |
| `-v bf_emma` | Emma | Female | British English |
| `-v bm_george` | George | Male | British English |

List all available voices at any time:

```bash
koko --list-voices
```

---

## Audio Focus (`-f` / `--focus`)

Ducks background audio (music, YouTube, video games, browser tabs, media players) system-wide using native macOS CoreAudio HAL APIs in **<1ms**, speaks your text loudly and clearly, then automatically restores your original volume levels.

```bash
koko -f "Priority notification. All background audio ducked."
```

> **How it works**: Unlike fragile script-based implementations, `koko` uses zero application-specific hacks. It operates directly at the macOS CoreAudio hardware layer (`AudioObjectSetPropertyData`), attenuating all active audio outputs across the entire OS during speech playback.

---

## Daemon Mode (`< 30ms` IPC Execution)

For voice agents, terminal scripts, or IDE extensions requiring ultra-low latency playback, launch `koko` in background daemon mode:

```bash
# Start background daemon server
koko daemon
```

When daemon mode is running, `koko` keeps model weights pre-warmed in memory and routes commands over a local Unix domain socket (`/tmp/koko_daemon.sock`).

```bash
# Executed via IPC in under 30ms
koko "Sub 30ms instant response."
```

---

## Terminal Piping & Scripting

Pipe standard input directly into `koko` from any shell command, log parser, or AI pipeline:

```bash
echo "Build succeeded in 4.2 seconds." | koko
```

```bash
git push 2>&1 | tail -n 1 | koko --friday
```

Save generated speech to a `.wav` file without playing:

```bash
koko -o output.wav --no-play "Saving audio output directly to file."
```

---

## Technical Architecture

- **Zero Python Dependencies**: 100% compiled standalone Go binary with native C++ SIMD inference (`sherpa-onnx`). No virtual environments, PyTorch, or Hugging Face Hub runtime needed.
- **Model Backbone**: **Kokoro-82M** (StyleTTS 2 architecture) exported to ONNX (~345MB).
- **CoreAudio HAL**: Direct Cgo linking against Apple's `-framework CoreAudio` for hardware-level volume scalar manipulation.
- **Auto-Setup**: Missing model files are self-provisioned automatically on turn 1.

---

## License

[MIT License](LICENSE) © 2026 Ankit Yadav
