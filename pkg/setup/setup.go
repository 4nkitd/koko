package setup

import (
	"archive/tar"
	"compress/bzip2"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	SherpaReleaseURL = "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.6/sherpa-onnx-v1.13.6-osx-universal2-shared.tar.bz2"
	KokoroModelURL   = "https://github.com/k2-fsa/sherpa-onnx/releases/download/tts-models/kokoro-en-v0_19.tar.bz2"
)

func EnsureInstalled(verbose bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot resolve user home directory: %w", err)
	}

	binPath := filepath.Join(homeDir, ".local", "bin", "sherpa-onnx-offline-tts")
	binPathPlay := filepath.Join(homeDir, ".local", "bin", "sherpa-onnx-offline-tts-play")
	modelDir := filepath.Join(homeDir, ".config", "koko", "onnx", "kokoro-en-v0_19")

	binExists := fileExists(binPath) && fileExists(binPathPlay)
	modelExists := fileExists(filepath.Join(modelDir, "model.onnx"))

	if binExists && modelExists {
		return nil
	}

	fmt.Println("🚀 Initializing koko dependencies...")

	if !binExists {
		if err := installSherpaBinary(homeDir, verbose); err != nil {
			return fmt.Errorf("failed to install C++ ONNX engine binary: %w", err)
		}
	}

	if !modelExists {
		if err := installKokoroModel(homeDir, verbose); err != nil {
			return fmt.Errorf("failed to download Kokoro ONNX model: %w", err)
		}
	}

	fmt.Println("✅ Setup complete!")
	return nil
}

func installSherpaBinary(homeDir string, verbose bool) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("unsupported OS: %s (macOS arm64/x64 currently supported)", runtime.GOOS)
	}

	fmt.Println("📦 Downloading sherpa-onnx C++ runtime engine...")
	tmpTar := filepath.Join(os.TempDir(), "sherpa-runtime.tar.bz2")
	defer os.Remove(tmpTar)

	if err := downloadFile(SherpaReleaseURL, tmpTar); err != nil {
		return err
	}

	fmt.Println("⚡ Extracting binary...")
	tmpExtract := filepath.Join(os.TempDir(), "sherpa-extract")
	defer os.RemoveAll(tmpExtract)

	if err := extractTarBz2(tmpTar, tmpExtract); err != nil {
		return err
	}

	binDestDir := filepath.Join(homeDir, ".local", "bin")
	if err := os.MkdirAll(binDestDir, 0755); err != nil {
		return err
	}

	binsToCopy := []string{"sherpa-onnx-offline-tts", "sherpa-onnx-offline-tts-play"}
	for _, binName := range binsToCopy {
		var foundBin string
		_ = filepath.Walk(tmpExtract, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && info.Name() == binName {
				foundBin = path
			}
			return nil
		})

		if foundBin != "" {
			destBin := filepath.Join(binDestDir, binName)
			_ = copyFile(foundBin, destBin)
			_ = os.Chmod(destBin, 0755)
		}
	}

	return nil
}

func installKokoroModel(homeDir string, verbose bool) error {
	fmt.Println("🧠 Downloading Kokoro-82M ONNX model files...")
	tmpTar := filepath.Join(os.TempDir(), "kokoro-model.tar.bz2")
	defer os.Remove(tmpTar)

	if err := downloadFile(KokoroModelURL, tmpTar); err != nil {
		return err
	}

	targetDir := filepath.Join(homeDir, ".config", "koko", "onnx")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	fmt.Println("⚡ Extracting model files...")
	return extractTarBz2(tmpTar, targetDir)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTarBz2(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	bz2Reader := bzip2.NewReader(f)
	tarReader := tar.NewReader(bz2Reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
