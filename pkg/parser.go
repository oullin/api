package pkg

import (
	"encoding/json"
	"fmt"
	"os"
)

func ParseJsonFile[T any](filePath string) (T, error) {
	// We must declare a variable of type T to hold the result.
	// This will also be the zero value of T if an error occurs.
	var result T

	// Read the entire file into a byte slice.
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Wrap the error with context for clearer debugging.
		return result, fmt.Errorf("could not read file %s: %w", filePath, err)
	}

	// Unmarshal the JSON data into the 'result' variable.
	// We pass a pointer to 'result' so json.Unmarshal can populate it.
	if err := json.Unmarshal(content, &result); err != nil {
		return result, fmt.Errorf("could not unmarshal json from %s: %w", filePath, err)
	}

	// If successful, return the populated struct and a nil error.
	return result, nil
}
