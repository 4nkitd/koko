<div align="center">

# 🗣️ koko

### Fast, standalone, zero-dependency Text-to-Speech CLI for macOS.

[![Release](https://img.shields.io/github/v/release/4nkitd/koko?color=00ADD8&style=flat-square)](https://github.com/4nkitd/koko/releases)
[![Homebrew](https://img.shields.io/badge/homebrew-4nkitd%2Ftap%2Fkoko-orange?style=flat-square)](https://github.com/4nkitd/homebrew-tap)
[![License: MIT](https://img.shields.io/github/license/4nkitd/koko?style=flat-square)](LICENSE)

Get up and running with high-speed neural text-to-speech on your Mac in seconds.

[Quickstart](#quickstart) • [Character Voices](#character-voices) • [Audio Focus](#audio-focus) • [Daemon Mode](#daemon-mode) • [Documentation](#technical-architecture)

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
