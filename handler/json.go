package handler

import (
	"fmt"
	baseHttp "net/http"
)

func writeJSON(content []byte, w baseHttp.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(baseHttp.StatusOK)

	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("error writing response: %v", err)
	}

	return nil
}
