package main

import (
	"CRUD-API/config"
	"CRUD-API/controllers"
	"CRUD-API/database"
	"CRUD-API/kafka"
	"CRUD-API/middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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

	//Start server
	log.Println("Start company service API on port", conf.APIPort)
	// Use the port from the configuration.
	if err := http.ListenAndServe(":"+conf.APIPort, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
