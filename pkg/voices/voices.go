package voices

import (
	"fmt"
	"strings"
)

type VoiceInfo struct {
	Name        string
	Gender      string
	Language    string
	Description string
}

var AvailableVoices = []VoiceInfo{
	{Name: "friday", Gender: "Female", Language: "British/Irish English", Description: "F.R.I.D.A.Y. tactical AI assistant voice (Default)"},
	{Name: "jarvis", Gender: "Male", Language: "British English", Description: "J.A.R.V.I.S. AI assistant voice (Paul Bettany style)"},
	{Name: "ironman", Gender: "Male", Language: "American English", Description: "Iron Man / Tony Stark voice (Robert Downey Jr. style)"},
	{Name: "am_michael", Gender: "Male", Language: "American English", Description: "Clear, balanced American male voice"},
	{Name: "af_bella", Gender: "Female", Language: "American English", Description: "Bright American female voice"},
	{Name: "af_sarah", Gender: "Female", Language: "American English", Description: "Expressive American female voice"},
	{Name: "am_adam", Gender: "Male", Language: "American English", Description: "Deep American male voice"},
	{Name: "bf_emma", Gender: "Female", Language: "British English", Description: "Classic British female voice"},
	{Name: "bm_george", Gender: "Male", Language: "British English", Description: "Distinguished British male voice"},
}

func PrintVoices() {
	fmt.Println("=== Native ONNX C++ Voices (Zero Python) ===")
	for _, v := range AvailableVoices {
		fmt.Printf("  • %-12s [%s / %s] - %s\n", v.Name, v.Gender, v.Language, v.Description)
	}
	fmt.Println()
}

func FormatVoiceList() string {
	var sb strings.Builder
	sb.WriteString("Available Voices:\n")
	for _, v := range AvailableVoices {
		sb.WriteString(fmt.Sprintf("  - %s (%s, %s): %s\n", v.Name, v.Gender, v.Language, v.Description))
	}
	return sb.String()
}
