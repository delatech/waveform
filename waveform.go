package waveform

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func Generate(sourcePath string, jsonPath string) {
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	sourceFilename, err := filepath.Abs(sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	filename := filepath.Base(sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	tempfilename := fmt.Sprintf("%s/tmp/%s.raw", currentPath, filename)
	generateRawFile(sourceFilename, tempfilename)

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
