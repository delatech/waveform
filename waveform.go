package waveform

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	BufferSize             = 300
	MaxAudioValue  int64   = 65536
	MinAudioValue  int64   = -65536
	NumberOfBytes          = 4
	PixelPerSecond float64 = (1000 / 30.0) * 6
)

// generateRawData calls sox on the file with filename in and write to a file with filename out
func generateRawData(in string, out string) error {
	if _, err := os.Stat(in); os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(out); os.IsNotExist(err) {
		return err
	}

	cmd := exec.Command("sox", in, "-t", "raw", "-r", "8000", "-c", "1", "-e", "signed-integer", "-L", out)
	return cmd.Run()
}

func getWidth(in string) float64 {
	return math.Ceil((getDuration(in) * 1000) / PixelPerSecond)
}

func getDuration(in string) float64 {
	cmd := exec.Command("soxi", "-D", in)

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

func Generate(sourcePath string, w io.Writer) {
	sourceFilename, err := filepath.Abs(sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	rawFile, err := ioutil.TempFile(os.TempDir(), "wv")
	defer rawFile.Close()
	defer os.Remove(rawFile.Name())
	if err != nil {
		log.Fatalln(err)
	}
	if err := generateRawData(sourceFilename, rawFile.Name()); err != nil {
		log.Fatalln(err)
	}

	minimumValues, maximumValues := extractMinMaxValues(sourcePath, rawFile)

	var tojson interface{}
	percents := convertToPercentage(minimumValues, maximumValues)
	ints := make([]int64, len(percents))
	for k, v := range percents {
		ints[k] = int64(v * 128)
	}
	tojson = ints

	result, err := json.Marshal(tojson)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := w.Write(result); err != nil {
		log.Fatalln(err)
	}
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

func extractMinMaxValues(sourcePath string, rawFile *os.File) ([]int64, []int64) {
	rawfileInfo, err := rawFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	widthInFloat := getWidth(sourcePath)
	segmentSize := int(float64(rawfileInfo.Size())/widthInFloat+0.5) / NumberOfBytes
	width := int(widthInFloat)
	maximumValues := make([]int64, width)
	minimumValues := make([]int64, width)
	segmentByteSize := segmentSize * NumberOfBytes

	var wg sync.WaitGroup

	for position := 0; position < width; position++ {
		if position%BufferSize == 0 {
			wg.Add(1)
			go func(index int) {
				lastIndex := index + BufferSize
				if lastIndex > width {
					lastIndex = width
				}

				mins, maxs := getMinMaxValuesWithIndexFromFile(rawFile, index, lastIndex, segmentByteSize)
				for i := 0; i < len(mins); i++ {
					minimumValues[index+i] = mins[i]
					maximumValues[index+i] = maxs[i]
				}
				wg.Done()
			}(position)
		}
	}

	wg.Wait()

	return minimumValues, maximumValues
}

func getMinMaxValuesWithIndexFromFile(file *os.File, startIndex int, lastIndex int, segmentByteSize int) ([]int64, []int64) {
	numberOfSegments := lastIndex - startIndex
	data := make([]byte, segmentByteSize*numberOfSegments)
	mins := make([]int64, numberOfSegments)
	maxs := make([]int64, numberOfSegments)
	n, err := file.ReadAt(data, int64(startIndex*segmentByteSize))
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		log.Fatal(err)
	}

	var start int
	var last int
	for i := 0; i < numberOfSegments; i++ {
		start = i * segmentByteSize
		last = (i + 1) * segmentByteSize

		if last > n {
			last = n
		}

		min, max := getMinMaxValue(data[start:last], last-start)
		mins[i] = min
		maxs[i] = max
	}

	return mins, maxs
}

func getMinMaxValue(data []byte, dataLength int) (int64, int64) {
	max := MinAudioValue
	min := MaxAudioValue

	word := make([]byte, NumberOfBytes*2)
	for index := 0; index < dataLength; index++ {
		word[index%NumberOfBytes] = data[index]
		if (index+1)%NumberOfBytes == 0 {
			for j := 0; j < NumberOfBytes; j++ {
				word[NumberOfBytes+j] = 0
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
		}
	}
	return min, max
}
