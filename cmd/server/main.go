// cmd/server/main.go

package main

import (
    "log"
    "net/http"
    "transcription-service/internal/api"
)

func main() {
    err := api.InitSuperTokens()
    if err != nil {
        log.Fatalf("Failed to initialize SuperTokens: %v", err)
    }

    router := api.NewRouter()

    // Listen on all interfaces
    log.Printf("Starting server on port %s", api.PORT)
    if err := http.ListenAndServe("0.0.0.0:" + api.PORT, router); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
