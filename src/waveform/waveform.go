package waveform

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	PIXEL_PER_SECOND float64 = 1000 / 30.0
	TEMP_FILE_DIR    string  = "tmp"
	MAX_AUDIO_VALUE  int32   = 65536
	MIN_AUDIO_VALUE  int32   = -65536
	NUMBER_OF_BYTES  int     = 4
)

func Generate(sourcePath string) {
	filename := filepath.Base(sourcePath)
	tempfilename := fmt.Sprintf("%s/%s.raw", TEMP_FILE_DIR, filename)
	GenerateRawFile(sourcePath, tempfilename)

	originalFile, err := os.Open(sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	originalFileInfo, err := originalFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("original file size: %d\n", originalFileInfo.Size())

	rawFile, err := os.Open(tempfilename)
	if err != nil {
		log.Fatal(err)
	}

	minimumValues, maximumValues := extractMinMaxValues(sourcePath, rawFile)
	fmt.Println(maximumValues)
	fmt.Println(minimumValues)
	percents := convertToPercentage(minimumValues, maximumValues)
	fmt.Println(percents)
}

func extractMinMaxValues(sourcePath string, rawFile *os.File) ([]int32, []int32) {
	rawfileInfo, err := rawFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("raw file size: %d\n", rawfileInfo.Size())

	width := GetWidth(sourcePath)
	segmentSize := int(float64(rawfileInfo.Size()) / width)
	maximumValues := make([]int32, int(width))
	minimumValues := make([]int32, int(width))

	data := make([]byte, segmentSize*NUMBER_OF_BYTES)
	fmt.Printf("width: %d\n", int(width))
	fmt.Printf("segmentSize: %d\n", segmentSize)
	fmt.Printf("raw file size: %d\n", rawfileInfo.Size())
	for position := 0; position < int(width); position++ {
		max := MIN_AUDIO_VALUE
		min := MAX_AUDIO_VALUE

		data = data[:cap(data)]
		n, err := rawFile.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		if err != nil {
			log.Fatal(err)
		}

		data = data[:n]
		word := make([]byte, NUMBER_OF_BYTES*2)

		for index, b := range data {
			word[index%4] = b
			if (index+1)%NUMBER_OF_BYTES == 0 {
				var value int32
				buf := bytes.NewReader(word)
				err := binary.Read(buf, binary.LittleEndian, &value)

				if err != nil {
					log.Fatal(err)
				}

				if value < min {
					min = value
				}

				if value > max {
					max = value
				}
			}
		}

		minimumValues[position] = min
		maximumValues[position] = max
	}
	return minimumValues, maximumValues
}

func convertToPercentage(minimumValues []int32, maximumValues []int32) []float64 {
	width := len(maximumValues)
	heights_in_int32 := make([]int32, width)
	heights := make([]float64, width)
	highestHeight := maximumValues[0] - minimumValues[0]
	heights_in_int32[0] = 0
	for i := 1; i < width; i++ {
		heights_in_int32[i] = maximumValues[i] - minimumValues[i]
		if highestHeight < heights_in_int32[i] {
			highestHeight = heights_in_int32[i]
		}
	}

	highestHeight_in_float64 := float64(highestHeight)
	for i := 0; i < width; i++ {
		heights[i] = float64(heights_in_int32[i]) / highestHeight_in_float64
	}
	return heights
}

func GenerateRawFile(sourcePath string, tempFilePath string) {
	cmd := exec.Command("sox", sourcePath, "-t", "raw", "-r", "44100", "-c", "1", "-e", "signed-integer", "-L", tempFilePath)
	_, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
		return
	}
}

func GetWidth(sourcePath string) float64 {
	return math.Ceil((GetDuration(sourcePath) * 1000) / PIXEL_PER_SECOND)
}

func GetDuration(sourcePath string) float64 {
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
