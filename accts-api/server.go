package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"accts-api/api"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {

	appContext, appContextCancel := context.WithCancel(context.Background())
	backendApi := api.NewApiAccess(appContext)

	// defer functions are executed in LIFO order
	// defer the
	defer appContextCancel()
	defer backendApi.AccountsDB.Close()

	r := mux.NewRouter()
	r.HandleFunc("/health", backendApi.HealthCheckHandler)
	r.HandleFunc("/login", backendApi.LoginUser).Methods("POST")
	r.HandleFunc("/register", backendApi.RegisterUser).Methods("POST")

	// Protected routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(backendApi.AuthMiddleware)
	apiRouter.HandleFunc("/create", backendApi.CreateRewardsAccount).Methods("POST")
	apiRouter.HandleFunc("/{accountId}/details", backendApi.GetRewardsAccount)
	apiRouter.HandleFunc("/list/transactions/{accountId}", backendApi.GetAccountTransactions)
	apiRouter.HandleFunc("/update/balance/{accountId}", backendApi.UpdateRewardsBalance).Methods("PUT")

	cors := cors.New(cors.Options{
		AllowedHeaders: []string{"content-type", "access-control-allow-headers", "access-control-allow-methods", "access-control-allow-origin", "authorization"},
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "HEAD", "POST", "PUT"},
	})
	routerWithCors := cors.Handler(r)

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      routerWithCors, // Pass our instance of gorilla/mux in.
		BaseContext: func(l net.Listener) context.Context {
			reqContext := context.WithValue(appContext, api.ContextKeyRequestID, uuid.NewString())
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
