package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

func ReadWavPCM(filePath string) ([]float32, int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	var header struct {
		ChunkID       [4]byte
		ChunkSize     uint32
		Format        [4]byte
		Subchunk1ID   [4]byte
		Subchunk1Size uint32
		AudioFormat   uint16
		NumChannels   uint16
		SampleRate    uint32
		ByteRate      uint32
		BlockAlign    uint16
		BitsPerSample uint16
	}

	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return nil, 0, fmt.Errorf("failed to read WAV header: %w", err)
	}

	for {
		var subchunkHeader struct {
			ID   [4]byte
			Size uint32
		}
		if err := binary.Read(f, binary.LittleEndian, &subchunkHeader); err != nil {
			if err == io.EOF {
				break
			}
			return nil, 0, err
		}

		if string(subchunkHeader.ID[:]) == "data" {
			data := make([]byte, subchunkHeader.Size)
			if _, err := io.ReadFull(f, data); err != nil {
				return nil, 0, err
			}

			samples := make([]float32, subchunkHeader.Size/uint32(header.BlockAlign))
			var maxPeak float32 = 0.0

			for i := range samples {
				switch header.BitsPerSample {
				case 16:
					raw := int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
					val := float32(raw) / 32768.0
					samples[i] = val
					if absVal := float32(math.Abs(float64(val))); absVal > maxPeak {
						maxPeak = absVal
					}
				case 32:
					if header.AudioFormat == 3 {
						bits := binary.LittleEndian.Uint32(data[i*4 : i*4+4])
						val := math.Float32frombits(bits)
						samples[i] = val
						if absVal := float32(math.Abs(float64(val))); absVal > maxPeak {
							maxPeak = absVal
						}
					} else {
						raw := int32(binary.LittleEndian.Uint32(data[i*4 : i*4+4]))
						val := float32(raw) / 2147483648.0
						samples[i] = val
						if absVal := float32(math.Abs(float64(val))); absVal > maxPeak {
							maxPeak = absVal
						}
					}
				}
			}

			// Automatic 0 dBFS Peak Normalization for maximum audio clarity & volume
			if maxPeak > 0 && maxPeak < 0.98 {
				scale := float32(0.98) / maxPeak
				for i := range samples {
					samples[i] *= scale
				}
			}

			return samples, int(header.SampleRate), nil
		}

		if _, err := f.Seek(int64(subchunkHeader.Size), io.SeekCurrent); err != nil {
			return nil, 0, err
		}
	}

	return nil, 0, fmt.Errorf("data chunk not found in WAV file")
}

func NormalizeWavFile(filePath string) error {
	samples, sampleRate, err := ReadWavPCM(filePath)
	if err != nil || len(samples) == 0 {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	numSamples := len(samples)
	dataSize := uint32(numSamples * 2)

	// Write 16-bit PCM WAV header
	_ = binary.Write(f, binary.LittleEndian, [4]byte{'R', 'I', 'F', 'F'})
	_ = binary.Write(f, binary.LittleEndian, uint32(36+dataSize))
	_ = binary.Write(f, binary.LittleEndian, [4]byte{'W', 'A', 'V', 'E'})
	_ = binary.Write(f, binary.LittleEndian, [4]byte{'f', 'm', 't', ' '})
	_ = binary.Write(f, binary.LittleEndian, uint32(16)) // Subchunk1Size
	_ = binary.Write(f, binary.LittleEndian, uint16(1))  // PCM format
	_ = binary.Write(f, binary.LittleEndian, uint16(1))  // Mono
	_ = binary.Write(f, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(f, binary.LittleEndian, uint32(sampleRate*2))
	_ = binary.Write(f, binary.LittleEndian, uint16(2))  // BlockAlign
	_ = binary.Write(f, binary.LittleEndian, uint16(16)) // 16 bits
	_ = binary.Write(f, binary.LittleEndian, [4]byte{'d', 'a', 't', 'a'})
	_ = binary.Write(f, binary.LittleEndian, dataSize)

	pcmData := make([]byte, dataSize)
	for i, s := range samples {
		if s > 1.0 {
			s = 1.0
		}
		if s < -1.0 {
			s = -1.0
		}
		val := int16(s * 32767.0)
		binary.LittleEndian.PutUint16(pcmData[i*2:i*2+2], uint16(val))
	}

	_, err = f.Write(pcmData)
	return err
}
