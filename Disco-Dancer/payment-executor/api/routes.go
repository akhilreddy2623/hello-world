package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	commonHandelers "geico.visualstudio.com/Billing/plutus/common-handlers"
)

func (*Config) routes() http.Handler {
	mux := chi.NewRouter()

	// Security Stuff
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Heartbeat("/health/live"))
	mux.Get("/", commonHandelers.HealthCheckHandler{}.HealthCheck)
	mux.Get("/health/ready", commonHandelers.HealthCheckHandler{}.HealthCheck)

	return mux
}
