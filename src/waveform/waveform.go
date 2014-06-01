package waveform

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
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

func Generate(sourcePath string, jsonPath string) {
	filename := filepath.Base(sourcePath)
	tempfilename := fmt.Sprintf("%s/%s.raw", TEMP_FILE_DIR, filename)
	GenerateRawFile(sourcePath, tempfilename)

	rawFile, err := os.Open(tempfilename)
	if err != nil {
		log.Fatal(err)
	}

	minimumValues, maximumValues := extractMinMaxValues(sourcePath, rawFile)
	percents := convertToPercentage(minimumValues, maximumValues)

	type Waves struct{ Waves []float64 }
	waves := Waves{Waves: percents}

	result, err := json.Marshal(waves)
	if err != nil {
		log.Fatal(err)
	}

	jsonFile, err := os.Create(jsonPath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := jsonFile.Write(result); err != nil {
		log.Fatal(err)
	}
}

func extractMinMaxValues(sourcePath string, rawFile *os.File) ([]int32, []int32) {
	rawfileInfo, err := rawFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	width := GetWidth(sourcePath)
	segmentSize := int(float64(rawfileInfo.Size()) / width)
	maximumValues := make([]int32, int(width))
	minimumValues := make([]int32, int(width))

	data := make([]byte, segmentSize)
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

		data = data[:n]
		word := make([]byte, NUMBER_OF_BYTES*2)

		for index, b := range data {
			word[index%NUMBER_OF_BYTES] = b
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

				for i := 0; i < NUMBER_OF_BYTES*2; i++ {
					word[i] = 0
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
	heightsInInt64 := make([]int64, width)
	heights := make([]float64, width)
	highestHeight := int64(maximumValues[0]) - int64(minimumValues[0])
	heightsInInt64[0] = 0
	for i := 1; i < width; i++ {
		heightsInInt64[i] = int64(maximumValues[i]) - int64(minimumValues[i])
		if highestHeight < heightsInInt64[i] {
			highestHeight = heightsInInt64[i]
		}
	}

	highestHeightInFloat64 := float64(highestHeight)
	for i := 0; i < width; i++ {
		heights[i] = float64(heightsInInt64[i]) / highestHeightInFloat64
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
