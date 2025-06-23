package main

import (
	_ "github.com/lib/pq"
	"github.com/oullin/boost"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/http/middleware"
	"log"
	"log/slog"
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
var validator *pkg.Validator

func init() {
	secrets, validate := boost.Spark("./.env")

	environment = secrets
	validator = validate
}

func main() {
	dbConnection := boost.MakeDbConnection(environment)
	logs := boost.MakeLogs(environment)
	localSentry := boost.MakeSentry(environment)

	defer (*logs).Close()
	defer (*dbConnection).Close()

	mux := baseHttp.NewServeMux()

	_ = boost.MakeApp(mux, boost.App{
		Validator:    validator,
		Logs:         logs,
		DbConnection: dbConnection,
		Env:          environment,
		Mux:          mux,
		Sentry:       localSentry,
	})

	pipelines := middleware.Pipeline{
		Env: environment,
	}

	tokenMid := middleware.MakeTokenMiddleware(environment.App.Credentials)

	userHandler := http.MakeApiHandler(
		pipelines.Chain(
			handleGetUser,
			middleware.UsernameCheck,
			tokenMid.Handle,
		),
	)

	mux.HandleFunc("GET /profile", userHandler)

	(*dbConnection).Ping()
	slog.Info("Starting new server on :" + environment.Network.HttpPort)

	if err := baseHttp.ListenAndServe(environment.Network.GetHostURL(), mux); err != nil {
		slog.Error("Error starting server", "error", err)
		panic("Error starting server." + err.Error())
	}
}
