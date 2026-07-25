package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"accts-api/adapters/httpapi"
	"accts-api/adapters/postgres"
	"accts-api/application"

	"github.com/google/uuid"
	"github.com/rs/cors"
)

func main() {

	appContext, appContextCancel := context.WithCancel(context.Background())
	dbAdapter, err := postgres.New(appContext)
	if err != nil {
		log.Fatalf("failed to start service: %v", err)
	}

	// defer functions are executed in LIFO order
	// defer the
	defer appContextCancel()
	defer dbAdapter.Close()

	appServices := application.NewServices(application.Dependencies{
		Stores:     dbAdapter.Stores(),
		UnitOfWork: dbAdapter,
	})
	r := httpapi.NewRouter(httpapi.Services{
		Partners:         appServices.Partners,
		Programs:         appServices.Programs,
		Members:          appServices.Members,
		Transactions:     appServices.Transactions,
		RewardProcessing: appServices.RewardProcessing,
		Ledger:           appServices.Ledger,
		Reporting:        appServices.Reporting,
	})

	cors := cors.New(cors.Options{
		AllowedHeaders: []string{"content-type"},
		AllowedOrigins: allowedOrigins(),
		AllowedMethods: []string{"GET", "HEAD", "POST", "PUT", "OPTIONS"},
	})
	routerWithCors := cors.Handler(r)

	srv := &http.Server{
		Addr: getenv("PAISA_HTTP_ADDR", "0.0.0.0:8080"),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      routerWithCors, // Pass our instance of gorilla/mux in.
		BaseContext: func(l net.Listener) context.Context {
			reqContext := context.WithValue(appContext, "requestID", uuid.NewString())
			return reqContext //Pass the app level context to the downstream apis
		},
	}

	log.Printf("Server accepting requests")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGTERM, syscall.SIGINT)

	<-gracefulShutdown
	log.Printf("Done!")
}

func allowedOrigins() []string {
	value := os.Getenv("PAISA_ALLOWED_ORIGINS")
	if value == "" {
		return []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}

	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			origins = append(origins, part)
		}
	}
	return origins
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
