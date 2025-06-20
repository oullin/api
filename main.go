package main

import (
    "encoding/json"
    _ "github.com/lib/pq"
    "log"
    "net/http"
    "os"
)

const file = "./storage/fixture/profile.json"

// ErrorResponse defines the structure for a JSON error response object.
type ErrorResponse struct {
    Error string `json:"error"`
}

// apiError represents an application-level error, including an HTTP status code.
type apiError struct {
    Message string
    Status  int
}

// Error makes apiError satisfy the standard error interface.
func (e *apiError) Error() string {
    return e.Message
}

// apiHandler is a custom type for handlers that return an apiError.
type apiHandler func(http.ResponseWriter, *http.Request) *apiError

// middleware is a function that takes an apiHandler and returns one.
type middleware func(apiHandler) apiHandler

// Chain applies a list of middlewares to a final apiHandler.
// It builds the chain in reverse, so the first middleware in the list
// is the outermost one, executing first.
func chain(h apiHandler, mws ...middleware) apiHandler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

// makeApiHandler is a wrapper that converts our custom apiHandler into a standard
// http.HandlerFunc. It centrally handles turning an apiError into a JSON response.
func makeApiHandler(fn apiHandler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := fn(w, r); err != nil {
            log.Printf("API Error: %s, Status: %d", err.Message, err.Status)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(err.Status)
            resp := ErrorResponse{Error: err.Message}
            if jsonErr := json.NewEncoder(w).Encode(resp); jsonErr != nil {
                log.Printf("Could not encode error response: %v", jsonErr)
            }
        }
    }
}

// authMiddleware now correctly wraps an apiHandler and returns an apiHandler.
// This allows it to return a structured apiError, integrating it cleanly
// into our unified error handling pattern.
func authMiddleware(next apiHandler) apiHandler {
    return func(w http.ResponseWriter, r *http.Request) *apiError {
        username := r.Header.Get("X-Username")
        if username != "gocanto" {
            log.Printf("Unauthorized access attempt by user: '%s'", username)
            // Return a structured error instead of writing directly to the response.
            return &apiError{
                Message: "Unauthorized",
                Status:  http.StatusUnauthorized,
            }
        }

        log.Println("Successfully authenticated user: gocanto")
        // If auth succeeds, call the next handler in the chain.
        return next(w, r)
    }
}

// tokenCheckMiddleware is a secondary middleware to check for a specific token.
// It also wraps an apiHandler and returns an apiHandler, allowing it to be chained.
func tokenCheckMiddleware(next apiHandler) apiHandler {
    return func(w http.ResponseWriter, r *http.Request) *apiError {
        token := r.Header.Get("X-Token")
        if token != "3" {
            log.Printf("Forbidden: Invalid token received ('%s')", token)
            // Using 403 Forbidden to differentiate from 401 Unauthorized.
            return &apiError{
                Message: "Forbidden",
                Status:  http.StatusForbidden,
            }
        }

        log.Println("Token validation successful")
        // If the token is valid, proceed to the next handler.
        return next(w, r)
    }
}

// handleGetUser returns an *apiError on failure and nil on success.
func handleGetUser(w http.ResponseWriter, r *http.Request) *apiError {
    jsonBytes, err := os.ReadFile(file)
    if err != nil {
        log.Printf("Error reading profile file: %v", err)
        return &apiError{
            Message: "Internal Server Error: could not read profile data",
            Status:  http.StatusInternalServerError,
        }
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, err = w.Write(jsonBytes)
    if err != nil {
        log.Printf("Error writing response: %v", err)
    }

    return nil // A nil return indicates success.
}

func main() {
    // Create a new ServeMux, which is the standard practice for new Go services.
    mux := http.NewServeMux()

    // Using the chain function makes adding new middlewares much cleaner.
    // The execution order is left-to-right: authMiddleware, then tokenCheckMiddleware.
    userHandler := makeApiHandler(chain(handleGetUser, authMiddleware, tokenCheckMiddleware))
    mux.HandleFunc("GET /profile", userHandler)

    addr := ":8080"
    log.Printf("Server starting on %s", addr)
    log.Println("Ensure you have a 'store/profile.json' file relative to the executable.")

    // Start the HTTP server with the new mux.
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatalf("Could not start server: %s\n", err)
    }
}
