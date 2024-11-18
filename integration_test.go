package company_service

import (
	"bytes"
	"company-service/config"
	"company-service/controllers"
	"company-service/database"
	"company-service/kafka"
	"company-service/middleware"
	"company-service/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Global variables for DB and Kafka

func TestIntegration(t *testing.T) {

	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Could not load config: %v", err)
	}
	db, err := database.InitDB(conf)
	if err != nil {
		t.Fatalf("Could not connect to database: %v", err)
	}
	kafkaProducer, err := kafka.NewKafkaProducer(conf.KafkaURL, conf.KafkaTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer")
	}
	kafkaConsumer, err := kafka.NewKafkaConsumer(conf.KafkaURL, conf.KafkaGroupId, conf.KafkaTopic)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer")
	}
	var dbInterface database.Database = db
	var kafkaProducerInterface kafka.Producer = kafkaProducer
	newApp := controllers.App{DB: dbInterface, KafkaProducer: kafkaProducerInterface, Config: conf}
	consumedMessages := make(chan kafka.EventMessage, 10)
	var wg sync.WaitGroup

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	// Public routes: Login
	apiRouter.HandleFunc("/login", newApp.Login).Methods("POST")
	apiRouter.HandleFunc("/companies", middleware.JwtMiddleware(newApp.CreateCompany, conf)).Methods("POST")
	apiRouter.HandleFunc("/companies/{id}", newApp.GetCompany).Methods("GET")
	apiRouter.HandleFunc("/companies/{id}", middleware.JwtMiddleware(newApp.UpdateCompany, conf)).Methods("PATCH")
	apiRouter.HandleFunc("/companies/{id}", middleware.JwtMiddleware(newApp.DeleteCompany, conf)).Methods("DELETE")

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
		err = db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection")
		}
		kafkaProducer.Close()
		kafkaConsumer.Close()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafkaConsumer.ConsumeMessage(ctx, conf.KafkaTopic, consumedMessages)
	}()

	// Step 0: Login
	login := map[string]interface{}{
		"username": conf.User,
		"password": conf.Password,
	}
	loginJSON, _ := json.Marshal(login)
	req, err := http.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(loginJSON))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	cookies := rr.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Step 1: Create a Company
	company := models.Company{
		Name:        "Test Company",
		Description: "This is an integration test company.",
		Employees:   50,
		Registered:  true,
		Type:        "Corporations",
	}
	expectedMessage := "Company created successfully"

	companyJSON, _ := json.Marshal(company)
	req, err = http.NewRequest(http.MethodPost, "/api/companies", bytes.NewBuffer(companyJSON))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookies[0])

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdCompany struct {
		Message string         `json:"message"`
		Company models.Company `json:"company"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &createdCompany)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	if createdCompany.Company.ID == uuid.Nil {
		t.Errorf("Expected a non-empty UUID for ID, but got: %v", createdCompany.Company.ID)
	} else if _, err := uuid.Parse(createdCompany.Company.ID.String()); err != nil {
		t.Errorf("Expected a valid UUID for ID, but got: %v", createdCompany.Company.ID)
	}

	if createdCompany.Message != expectedMessage {
		t.Errorf("Expected Message: %v, but got: %v", expectedMessage, createdCompany.Message)
	}

	if !compareCompanyIgnoringGenerated(company, createdCompany.Company) {
		t.Errorf("Expected Name: %v, but got: %v", company, createdCompany.Company)
	}

	select {
	case consumedMessage := <-consumedMessages:
		assert.Equal(t, "company_created", consumedMessage.EventType)
		assert.Equal(t, createdCompany.Company, *consumedMessage.Company)
	case <-time.After(60 * time.Second):
		t.Errorf("Timed out waiting for message on topic %s", conf.KafkaTopic)
	}
	// Step 4: Retrieve the Company
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/companies/%s", createdCompany.Company.ID.String()), nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var requestedCompany models.Company
	err = json.Unmarshal(rr.Body.Bytes(), &requestedCompany)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, createdCompany.Company, requestedCompany)

	// Step 5: Update the Company
	updateFields := map[string]interface{}{
		"name": "Updated Name",
	}
	updatedCompanyJSON, _ := json.Marshal(updateFields)
	expectedMessage = "Company updated successfully"

	req, err = http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/companies/%s", createdCompany.Company.ID.String()), bytes.NewBuffer(updatedCompanyJSON))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookies[0])

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var updatedCompany struct {
		Message string         `json:"message"`
		Company models.Company `json:"company"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &updatedCompany)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, expectedMessage, updatedCompany.Message)
	assert.Equal(t, updateFields["name"], updatedCompany.Company.Name)

	select {
	case consumedMessage := <-consumedMessages:
		assert.Equal(t, "company_updated", consumedMessage.EventType)
		assert.Equal(t, updatedCompany.Company, *consumedMessage.Company)
	case <-time.After(30 * time.Second):
		t.Errorf("Timed out waiting for message on topic %s", conf.KafkaTopic)
	}

	// Step 7: Delete the Company
	req, err = http.NewRequest("DELETE", fmt.Sprintf("/api/companies/%s", createdCompany.Company.ID.String()), nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.AddCookie(cookies[0])

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	select {
	case consumedMessage := <-consumedMessages:
		assert.Equal(t, "company_deleted", consumedMessage.EventType)
		assert.Equal(t, createdCompany.Company.ID.String(), consumedMessage.Company.ID.String())
	case <-time.After(30 * time.Second):
		t.Errorf("Timed out waiting for message on topic %s", conf.KafkaTopic)
	}

	// Step 10: Verify the Company was deleted
	req, err = http.NewRequest("GET", fmt.Sprintf("/api/companies/%s", createdCompany.Company.ID.String()), nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func compareCompanyIgnoringGenerated(expected, actual models.Company) bool {
	return expected.Name == actual.Name &&
		expected.Description == actual.Description &&
		expected.Employees == actual.Employees &&
		expected.Registered == actual.Registered &&
		expected.Type == actual.Type
}
