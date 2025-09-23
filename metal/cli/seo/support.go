package seo

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/oullin/metal/router"
)

func PrintResponse(rr *httptest.ResponseRecorder) {
	fmt.Println("\n--- Captured Response ---")

	fmt.Printf("Status Code: %d\n", rr.Code)
	fmt.Printf("Response Body: %s\n", rr.Body.String())
	fmt.Printf("Content-Type Header: %s\n", rr.Header().Get("Content-Type"))
}

func Fetch[T any](response *T, handler func() router.StaticRouteResource) error {
	req := httptest.NewRequest("GET", "http://localhost:8080/proxy", nil)
	rr := httptest.NewRecorder()

	maker := handler()

	if err := maker.Handle(rr, req); err != nil {
		return err
	}

	if err := json.Unmarshal(rr.Body.Bytes(), response); err != nil {
		return err
	}

	return nil
}
