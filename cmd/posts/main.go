package main

import (
	"log"
	"net/http"
	"os"

	"posts/internal/db"
	"posts/internal/handlers"
	"posts/internal/metrics"
	"posts/internal/middleware"

	"github.com/gorilla/mux"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	db.Connect()
	metrics.RegisterMetrics()

	r := mux.NewRouter()

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Application endpoints
	r.HandleFunc("/posts", handlers.CreatePostHandler).Methods("POST")
	r.HandleFunc("/posts/{id:[0-9]+}", handlers.GetPostHandler).Methods("GET")

	// Apply middlewares
	loggedRouter := middleware.LoggingMiddleware(r)
	metricsRouter := middleware.MetricsMiddleware(loggedRouter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Starting posts-service on port", port)
	if err := http.ListenAndServe(":"+port, metricsRouter); err != nil {
		log.Fatal(err)
	}
}
