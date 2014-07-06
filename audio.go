package waveform

import (
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	PIXEL_PER_SECOND float64 = 1000 / 30.0
)

func generateRawFile(sourcePath string, tempFilePath string) {
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		log.Fatalf("Source file does not exists: %s", sourcePath)
		return
	}

	cmd := exec.Command("sox", sourcePath, "-t", "raw", "-r", "44100", "-c", "1", "-e", "signed-integer", "-L", tempFilePath)
	_, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
		return
	}
}

func getWidth(sourcePath string) float64 {
	return math.Ceil((getDuration(sourcePath) * 1000) / PIXEL_PER_SECOND)
}

func getDuration(sourcePath string) float64 {
	cmd := exec.Command("soxi", "-D", sourcePath)

	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	result, err := strconv.ParseFloat(strings.TrimSpace(string(output[0:])), 64)
	if err != nil {
		log.Fatal(err)
	}
	return result
}
