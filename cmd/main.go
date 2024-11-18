package main

import (
	"company-service/config"
	"company-service/controllers"
	"company-service/database"
	"company-service/kafka"
	"company-service/middleware"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load the configuration from .env file.
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	db, err := database.InitDB(conf)
	if err != nil {
		log.Fatalf("Failed to connect to database")
	}

	kafkaProducer, err := kafka.NewKafkaProducer(conf.KafkaURL, conf.KafkaTopic)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka")
	}
	var dbInterface database.Database = db
	var kafkaProducerInterface kafka.Producer = kafkaProducer

	// Setting up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		err = dbInterface.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection")
		}
		kafkaProducerInterface.Close()
	}()

	newApp := controllers.App{DB: dbInterface, KafkaProducer: kafkaProducerInterface, Config: conf}

	//Router and endpoint setup code
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api").Subrouter()
	// Public routes: Login
	apiRouter.HandleFunc("/login", newApp.Login).Methods("POST")
	apiRouter.HandleFunc("/companies", middleware.JwtMiddleware(newApp.CreateCompany, conf)).Methods("POST")
	apiRouter.HandleFunc("/companies/{id}", newApp.GetCompany).Methods("GET")
	apiRouter.HandleFunc("/companies/{id}", middleware.JwtMiddleware(newApp.UpdateCompany, conf)).Methods("PATCH")
	apiRouter.HandleFunc("/companies/{id}", middleware.JwtMiddleware(newApp.DeleteCompany, conf)).Methods("DELETE")

	// Create an HTTP server with a graceful shutdown capability
	server := &http.Server{
		Addr:    ":" + conf.APIPort,
		Handler: router,
	}
	// Start the server in a goroutine
	go func() {
		log.Println("Start company service API on port", conf.APIPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until termination signal is received
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	// Give a timeout for the server shutdown (optional)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
