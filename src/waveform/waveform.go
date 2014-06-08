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
	MAX_AUDIO_VALUE  int64   = 65536
	MIN_AUDIO_VALUE  int64   = -65536
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

	result, err := json.Marshal(percents)
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

func extractMinMaxValues(sourcePath string, rawFile *os.File) ([]int64, []int64) {
	rawfileInfo, err := rawFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	width := GetWidth(sourcePath)
	segmentSize := int(float64(rawfileInfo.Size())/width+0.5) / NUMBER_OF_BYTES
	maximumValues := make([]int64, int(width))
	minimumValues := make([]int64, int(width))
	data := make([]byte, segmentSize*NUMBER_OF_BYTES)

	fmt.Printf("segment size: %d\n", segmentSize)
	for position := 0; position < int(width); position++ {
		max := MIN_AUDIO_VALUE
		min := MAX_AUDIO_VALUE

		n, err := rawFile.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		word := make([]byte, NUMBER_OF_BYTES*2)
		for index := 0; index < n; index++ {
			word[index%NUMBER_OF_BYTES] = data[index]
			if (index+1)%NUMBER_OF_BYTES == 0 {
				for j := 0; j < NUMBER_OF_BYTES; j++ {
					word[NUMBER_OF_BYTES+j] = 0
				}

				var value int32
				var valueInInt64 int64
				buf := bytes.NewReader(word)
				err := binary.Read(buf, binary.LittleEndian, &value)

				valueInInt64 = int64(value)
				if err != nil {
					log.Fatal(err)
				}

				if valueInInt64 < min {
					min = valueInInt64
				}

				if valueInInt64 > max {
					max = valueInInt64
				}

				if position == 21 {
					fmt.Printf("[%d] %d = [%d:%d]\n", (position*segmentSize + index), valueInInt64, min, max)
				}
			}
		}

		minimumValues[position] = min
		maximumValues[position] = max
	}
	return minimumValues, maximumValues
}

func convertToPercentage(minimumValues []int64, maximumValues []int64) []float64 {
	width := len(maximumValues)
	heightsInInt64 := make([]int64, width)
	heights := make([]float64, width)
	highestHeight := maximumValues[0] - minimumValues[0]
	heightsInInt64[0] = 0
	for i := 1; i < width; i++ {
		heightsInInt64[i] = maximumValues[i] - minimumValues[i]
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
