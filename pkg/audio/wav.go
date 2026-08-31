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

	// Read until data chunk is found
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
			for i := range samples {
				switch header.BitsPerSample {
				case 16:
					raw := int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
					samples[i] = float32(raw) / 32768.0
				case 32:
					if header.AudioFormat == 3 { // float32
						bits := binary.LittleEndian.Uint32(data[i*4 : i*4+4])
						samples[i] = math.Float32frombits(bits)
					} else { // int32
						raw := int32(binary.LittleEndian.Uint32(data[i*4 : i*4+4]))
						samples[i] = float32(raw) / 2147483648.0
					}
				}
			}

			return samples, int(header.SampleRate), nil
		}

		// Skip unknown chunk
		if _, err := f.Seek(int64(subchunkHeader.Size), io.SeekCurrent); err != nil {
			return nil, 0, err
		}
	}

	return nil, 0, fmt.Errorf("data chunk not found in WAV file")
}
