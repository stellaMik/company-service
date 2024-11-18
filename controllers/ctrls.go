package controllers

import (
	"company-service/config"
	"company-service/database"
	"company-service/kafka"
	"company-service/middleware"
	"company-service/models"
	"company-service/utils"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

// App struct holds all dependencies for the app
type App struct {
	DB            database.Database
	KafkaProducer kafka.Producer
	Config        *config.Config
}

// NewApp initializes and returns an instance of the App struct
func NewApp(db database.Database, producer kafka.Producer, conf *config.Config) *App {
	return &App{
		DB:            db,
		KafkaProducer: producer,
		Config:        conf,
	}
}

// Login authenticates the user and sends a JWT token
func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Decode the request body
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()            // Ensure the request body is closed
	decoder.DisallowUnknownFields() // Ensure that unknown fields are not allowed
	err := decoder.Decode(&loginRequest)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid input data for login with error: %v", err))
		return
	}

	// Find the user by username
	// Get the user from the database
	user, err := app.DB.GetUserByUsername(loginRequest.Username) //database.GetUserByUsername(loginRequest.Username, app.DB)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(w, http.StatusUnauthorized, "Invalid username or password")
		} else {
			utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error querying the database with error: %v", err))
		}
		return
	}

	// Compare the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	if err != nil {
		utils.SendErrorResponse(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate JWT token
	tokenString, err := middleware.GenerateJWT(user.Username, app.Config.JWTSecret)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Could not generate token with error: %v", err))
		return
	}
	// Set the token in an HTTP-only, Secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",                     // Name of the cookie
		Value:    tokenString,                      // The JWT token value
		HttpOnly: true,                             // Make the cookie HTTP-only to prevent XSS attacks
		Secure:   true,                             // Set to true if using HTTPS
		Path:     "/",                              // Available throughout the entire site
		Expires:  time.Now().Add(15 * time.Minute), // Set the expiration time (e.g., 1 hour)
	})

	// Send the response with the token
	utils.SendJSONResponse(w, http.StatusOK, map[string]string{"message": "Login successful"})
}

// CreateCompany creates a new company record
func (app *App) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var company *models.Company

	//Decode the request body
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()            //Ensure the request body is closed
	decoder.DisallowUnknownFields() //Ensure that are not allowed uknown fields
	err := decoder.Decode(&company)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid input data to create a new company record with error: %v", err))
		return
	}

	// Validate the company input
	if err := utils.ValidateCompanyInput(company); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	//Generate a new UUID
	id := utils.GenerateUUID()
	company.ID = id
	// Fill up the map datastore with id as key and solar panel data as value
	err = app.DB.CreateCompany(company)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	eventMessage := kafka.EventMessage{
		EventType: "company_created",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Company:   company,
	}
	// Publish the event to the message broker
	err = app.KafkaProducer.ProduceEvent(&eventMessage)
	if err != nil {
		// Log the Kafka error for retry or monitoring
		log.Printf("Kafka publish failed with error: %v", err)
		// Continue without failing the HTTP request
		// Include a warning in the JSON response
		response := map[string]interface{}{
			"message": "Company created successfully, but Kafka publishing failed",
			"company": company,
		}
		utils.SendJSONResponse(w, http.StatusCreated, response)
		return
	}
	// Send the response on JSON format
	utils.SendJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Company created successfully",
		"company": company,
	})
}

// GetCompany retrieves a company record with the given ID
func (app *App) GetCompany(w http.ResponseWriter, r *http.Request) {
	//Get UUID parameter
	id, err := utils.GetUUIDParam(r, "id")
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	// Check if the data ID exists in the datastore and return it
	company, err := app.DB.GetCompany(id)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, &company)
}

// UpdateCompany a solar panel record with the given id
func (app *App) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	//Get UUID parameter
	id, err := utils.GetUUIDParam(r, "id")
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	var updatedFields map[string]interface{}
	//Decode the request body
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()            //Ensure the request body is closed
	decoder.DisallowUnknownFields() //Ensure that are not allowed uknown fields
	err = decoder.Decode(&updatedFields)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid input data to update a company record: %v", err))
		return
	}

	// Check if the data ID exists in the datastore and update it
	company, err := app.DB.UpdateCompany(id, updatedFields)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			utils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	eventMessage := kafka.EventMessage{
		EventType: "company_updated",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Company:   company,
	}
	// Publish the event to the message broker
	err = app.KafkaProducer.ProduceEvent(&eventMessage)
	if err != nil {
		// Log the Kafka error for retry or monitoring
		log.Printf("Kafka publish failed with error: %v", err)
		// Continue without failing the HTTP request
		// Include a warning in the JSON response
		response := map[string]interface{}{
			"message": "Company updated successfully, but Kafka publishing failed",
			"company": &company,
		}
		utils.SendJSONResponse(w, http.StatusOK, response)
		return
	}
	// Send the response on JSON format
	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Company updated successfully",
		"company": &company,
	})
}

// DeleteCompany a solar panel record with the given id
func (app *App) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	//Get UUID parameter
	id, err := utils.GetUUIDParam(r, "id")
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if the data ID exists in the datastore and delete it
	err = app.DB.DeleteCompany(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			utils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	parsedUUID, err := utils.GenerateUUIDFromString(id)
	if err != nil {
		log.Printf("Could not parse UUID from string: %v", err)
	}
	eventMessage := kafka.EventMessage{
		EventType: "company_deleted",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Company:   &models.Company{ID: parsedUUID},
	}
	// Publish the event to the message broker
	err = app.KafkaProducer.ProduceEvent(&eventMessage)
	if err != nil {
		// Log the Kafka error for retry or monitoring
		log.Printf("Kafka publish failed with error: %v", err)
	}
	w.WriteHeader(http.StatusNoContent)
}
