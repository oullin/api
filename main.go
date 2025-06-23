package main

import (
	_ "github.com/lib/pq"
	"github.com/oullin/boost"
	"github.com/oullin/env"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/http/middleware"
	"log"
	baseHttp "net/http"
	"os"
)

const file = "./storage/fixture/profile.json"

func handleGetUser(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	jsonBytes, err := os.ReadFile(file)
	if err != nil {
		log.Printf("Error reading profile file: %v", err)
		return &http.ApiError{
			Message: "Internal Server Error: could not read profile data",
			Status:  baseHttp.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(baseHttp.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}

	return nil // A nil return indicates success.
}

var environment *env.Environment

//var validator *pkg.Validator

func init() {
	secrets, _ := boost.Spark("./.env")

	environment = secrets
	//validator = validate
}

func main() {
	// Create a new ServeMux, which is the standard practice for new Go services.
	mux := baseHttp.NewServeMux()
	pipelines := middleware.Pipeline{
		Env: environment,
	}

	tokenMid := middleware.MakeTokenMiddleware(environment.App.Credentials)

	// Using the chain function makes adding new middlewares much cleaner.
	// The execution order is left-to-right: authMiddleware, then tokenCheckMiddleware.
	userHandler := http.MakeApiHandler(
		pipelines.Chain(
			handleGetUser,
			middleware.UsernameCheck,
			tokenMid.Handle,
		),
	)

	mux.HandleFunc("GET /profile", userHandler)

	addr := ":8080"
	log.Printf("Server starting on %s", addr)
	log.Println("Ensure you have a 'store/profile.json' file relative to the executable.")

	// Start the HTTP server with the new mux.
	if err := baseHttp.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
