# `koko` — High-Performance Standalone Text-To-Speech CLI

[![Release](https://img.shields.io/badge/release-v0.0.1-blue.svg)](https://github.com/4nkitd/koko/releases)
[![Go Version](https://img.shields.io/badge/go-1.27+-00ADD8.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`koko` is a lightweight, zero-dependency command-line Text-To-Speech tool for macOS. Powered natively by the C++ ONNX engine (`sherpa-onnx`) and the **Kokoro-82M** model, it delivers sub-second speech synthesis (**~0.4s**) without embedding Python, heavy virtual environments, or external API servers.

---

## ✨ Key Features

- **🦾 Iron Man Default Voice**: Default character voice set to Iron Man (`am_michael`).
- **⚡ Sub-Second & Sub-30ms Daemon**: Sub-second C++ ONNX execution, plus optional background daemon server (`koko daemon`) for **<30ms instant IPC response time**.
- **📦 Zero Python Dependencies**: 100% independent Go binary build. No Python, PyTorch, or Hugging Face Hub runtime needed.
- **🛡️ Fallback Warning Protection**: Automatic fallback to macOS native speech engine (`say`) with a clear user warning if ONNX models are offline/unavailable.
- **🎯 System-Wide Audio Focus (`-f` / `--focus`)**: Ducks system-wide background audio (`<1ms` CoreAudio HAL) during speech playback, then automatically restores original volume levels.
- **🎙️ Shorthand Character Flags**:
  - `--ironman` / `--stark`: Iron Man character voice *(Default)*
  - `--friday`: F.R.I.D.A.Y. tactical AI assistant voice
  - `--jarvis`: J.A.R.V.I.S. AI assistant voice (Paul Bettany style)
- **🍺 Homebrew Tap Support**: One-command installation via `brew install 4nkitd/tap/koko`.

---

## 🚀 Quick Start & Installation

### Option 1: Install via Homebrew Tap
```bash
brew install 4nkitd/tap/koko
```

### Option 2: Install via Go
```bash
go install github.com/4nkitd/koko@v0.0.1
```

### Option 3: Build from Source
```bash
git clone https://github.com/4nkitd/koko.git
cd koko
make install
```

---

## 💻 Usage

```bash
# Speak text using default voice (Iron Man)
koko "Sometimes you gotta run before you can walk."

# Shorthand character flags
koko --ironman "I am Iron Man."
koko --friday "F.R.I.D.A.Y. online."
koko --jarvis "Allow me to introduce myself. I am J.A.R.V.I.S."

# Speak text with Audio Focus enabled (ducks background audio system-wide)
koko -f "Priority notification. Background audio ducked."

# Piped text from standard input
echo "Status check nominal." | koko

# Launch sub-30ms IPC daemon server
koko daemon

# Save speech directly to WAV file without playing
koko -o response.wav --no-play "Saving audio output directly to file."

# List all available character voices
koko --list-voices
```

---

## 🛠️ Character Voice Reference

| Character Flag | Gender | Language | Description |
| :--- | :--- | :--- | :--- |
| `--ironman` | Male | American English | Iron Man / Tony Stark voice *(Default)* |
| `--friday` | Female | British/Irish English | F.R.I.D.A.Y. AI assistant |
| `--jarvis` | Male | British English | J.A.R.V.I.S. AI assistant |
| `-v am_michael` | Male | American English | Clear American male voice |
| `-v am_adam` | Male | American English | Deep American male voice |
| `-v bf_emma` | Female | British English | Classic British female voice |
| `-v bm_george` | Male | British English | Distinguished British male voice |

---

## 📄 License

[MIT License](LICENSE) © 2026 Ankit Yadav
