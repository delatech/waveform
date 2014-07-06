package waveform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	var expected []float64
	var result []float64

	currentPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to read current path", err)
	}

	expactedJsonPath := fmt.Sprintf("%s/test/fixtures/expected.json", currentPath)
	resultJsonPath := fmt.Sprintf("%s/test/generated/result.json", currentPath)
	sourceFilePath := fmt.Sprintf("%s/test/fixtures/source.mp3", currentPath)

	expectedBytes, err := ioutil.ReadFile(expactedJsonPath)
	if err != nil {
		t.Fatalf("Failed to read: %s / %v", expactedJsonPath, err)
	}

	err = json.Unmarshal(expectedBytes, &expected)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %s / %v", expactedJsonPath)
	}

	Generate(sourceFilePath, resultJsonPath)

	resultBytes, err := ioutil.ReadFile(resultJsonPath)
	if err != nil {
		t.Fatalf("Failed to read: %s / %v", resultJsonPath, err)
	}

	err = json.Unmarshal(resultBytes, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %s / %v", "tmp/test/result.json")
	}

	if len(expected) != len(result) {
		t.Fatalf("Length are not matched. Expected: %d, Result: %d", len(expected), len(result))
	}

	for i := 0; i < len(expected); i++ {
		if math.Abs(expected[i]-result[i]) > 0.0001 {
			t.Fatalf("Value[%d] is not matched. Expected: %d, Result: %d", i, expected[i], result[i])
		}
	}
}
