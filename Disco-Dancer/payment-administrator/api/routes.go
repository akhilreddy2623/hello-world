package main

import (
	"net/http"

	_ "geico.visualstudio.com/Billing/plutus/payment-administrator-api/docs"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-api/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	commonHandelers "geico.visualstudio.com/Billing/plutus/common-handlers"
	httpSwagger "github.com/swaggo/http-swagger/v2"
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
	mux.Get("/swagger/*", httpSwagger.Handler())
	mux.Get("/", commonHandelers.HealthCheckHandler{}.HealthCheck)
	mux.Get("/health/ready", commonHandelers.HealthCheckHandler{}.HealthCheck)
	mux.Post("/payment", handlers.PaymentHandler{}.MakePaymentHandler)
	return mux
}
