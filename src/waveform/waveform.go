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
	//GenerateRawFile(sourcePath, tempfilename)

	rawFile, err := os.Open(tempfilename)
	if err != nil {
		log.Fatal(err)
	}

	rawfileInfo, err := rawFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	segmentSize := int(float64(rawfileInfo.Size()) / GetWidth(sourcePath))
	data := make([]byte, segmentSize*NUMBER_OF_BYTES)
	for {
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

		fmt.Println(binary.LittleEndian)

		for index, b := range data {
			word[index%4] = b
			if (index+1)%NUMBER_OF_BYTES == 0 {
				var value int32
				buf := bytes.NewReader(word)
				err := binary.Read(buf, binary.LittleEndian, &value)

				fmt.Println(word, value)

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

		fmt.Printf("min : %d / max: %d\n", min, max)
		break
	}
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
