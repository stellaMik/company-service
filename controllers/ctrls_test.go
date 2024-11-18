package controllers_test

import (
	"bytes"
	"company-service/config"
	"company-service/controllers"
	"company-service/mocks"
	"company-service/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  interface{}
		mockSetup    func(mockDB *mocks.MockDatabase)
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name: "Valid login",
			requestBody: map[string]string{
				"username": "user2",
				"password": "test2",
			},
			mockSetup: func(mockDB *mocks.MockDatabase) {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test2"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}
				mockDB.On("GetUserByUsername", "user2").Return(&models.User{
					Username: "user2",
					Password: string(hashedPassword),
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{"message": "Login successful"},
		},
		{
			name: "Invalid username",
			requestBody: map[string]string{
				"username": "invaliduser",
				"password": "test2",
			},
			mockSetup: func(mockDB *mocks.MockDatabase) {
				mockDB.On("GetUserByUsername", "invaliduser").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]string{"error": "Invalid username or password"},
		},
		{
			name: "Invalid password",
			requestBody: map[string]string{
				"username": "user2",
				"password": "invalidpassword",
			},
			mockSetup: func(mockDB *mocks.MockDatabase) {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test2"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}
				mockDB.On("GetUserByUsername", "user2").Return(&models.User{
					Username: "user2",
					Password: string(hashedPassword),
				}, nil)
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]string{"error": "Invalid username or password"},
		},
		{
			name: "Invalid input data uknown field",
			requestBody: map[string]string{
				"username": "user2",
				"pass":     "test2",
			},
			mockSetup:    func(mockDB *mocks.MockDatabase) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]string{"error": "Invalid input data for login with error: json: unknown field \"pass\""},
		},
		{
			name: "Database error",
			requestBody: map[string]string{
				"username": "user2",
				"password": "test2",
			},
			mockSetup: func(mockDB *mocks.MockDatabase) {
				mockDB.On("GetUserByUsername", "user2").Return(nil, errors.New("database error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]string{"error": "Error querying the database with error: database error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mocks.MockDatabase)
			app := controllers.NewApp(mockDB, nil, &config.Config{JWTSecret: "secretTest", Password: "test2", User: "user2"})
			// Set up mocks
			tt.mockSetup(mockDB)

			// Prepare request and response recorder
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Call the handler
			app.Login(rec, req)

			// Assertions
			assert.Equal(t, tt.expectedCode, rec.Code)
			// Deserialize the response body to a map
			var resp map[string]string
			err := json.NewDecoder(rec.Body).Decode(&resp)
			if err != nil {
				t.Fatalf("Failed to decode response body: %v", err)
			}
			assert.Equal(t, tt.expectedBody, resp)
			// Assert that all expectations were met
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCreateCompany(t *testing.T) {
	mockKafka := new(mocks.MockKafkaProducer)
	tests := []struct {
		name         string
		requestBody  interface{}
		mockSetup    func(mockDB *mocks.MockDatabase, mockKa *mocks.MockKafkaProducer)
		expectedCode int
		expectedBody struct {
			Message string         `json:"message"`
			Company models.Company `json:"company"`
		}
		expectedError map[string]string
	}{
		{
			name: "Valid company creation",
			requestBody: map[string]interface{}{
				"name":        "Tech company",
				"description": "A leading technology company.",
				"employees":   50,
				"registered":  true,
				"type":        "NonProfit",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("CheckIfExistsByName", mock.AnythingOfType("string")).Return(false)
				mockDB.On("CreateCompany", mock.AnythingOfType("*models.Company")).Return(nil)
				mockKafka.On("ProduceEvent", mock.AnythingOfType("*kafka.EventMessage")).Return(nil)
				// Mock ProduceEvent
			},
			expectedCode: http.StatusCreated,
			expectedBody: struct {
				Message string         `json:"message"`
				Company models.Company `json:"company"`
			}{
				Message: "Company created successfully",
				Company: models.Company{
					Name:        "Tech company",
					Description: "A leading technology company.",
					Employees:   50,
					Registered:  true,
					Type:        "NonProfit",
					CreatedAt:   time.Time{},
					UpdatedAt:   time.Time{},
					DeletedAt:   nil,
				},
			},
		},
		{
			name: "Invalid input data - missing required field",
			requestBody: map[string]interface{}{
				"name": "Tech company",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("CheckIfExistsByName", mock.AnythingOfType("string")).Return(false)
				mockDB.On("CreateCompany", mock.AnythingOfType("*models.Company")).Return(nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: map[string]string{"error": "invalid 'Employees': it must be a positive number"},
		},
		{
			name: "Invalid input data - invalid type",
			requestBody: map[string]interface{}{
				"name":       "Tech company",
				"employees":  50,
				"registered": true,
				"type":       "InvalidType",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("CheckIfExistsByName", mock.AnythingOfType("string")).Return(false)
				mockDB.On("CreateCompany", mock.AnythingOfType("*models.Company")).Return(nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: map[string]string{"error": "invalid 'Type': must be one of 'Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship'"},
		},
		{
			name: "Invalid input data - invalid name",
			requestBody: map[string]interface{}{
				"name": "This is a very long name for a company",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("CheckIfExistsByName", mock.AnythingOfType("string")).Return(false)
				mockDB.On("CreateCompany", mock.AnythingOfType("*models.Company")).Return(nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: map[string]string{"error": "invalid 'Name': it is required and must be at most 15 characters"},
		},
		{
			name: "Invalid input data - duplicate name",
			requestBody: map[string]interface{}{
				"name":       "Tech company",
				"employees":  50,
				"registered": true,
				"type":       "NonProfit",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("CheckIfExistsByName", mock.AnythingOfType("string")).Return(true)
				mockDB.On("CreateCompany", mock.AnythingOfType("*models.Company")).Return(nil)
			},
			expectedCode:  http.StatusConflict,
			expectedError: map[string]string{"error": "The company with the same name already exists"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mocks.MockDatabase)
			app := controllers.NewApp(mockDB, mockKafka, &config.Config{})
			tt.mockSetup(mockDB, mockKafka)

			// Prepare request and response recorder
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/companies", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Call the handler
			app.CreateCompany(rec, req)

			// Assertions
			assert.Equal(t, tt.expectedCode, rec.Code)
			if tt.expectedError != nil {
				var resp map[string]string
				err := json.NewDecoder(rec.Body).Decode(&resp)
				if err != nil {
					t.Fatalf("Error decoding response body: %v", err)
				}
				assert.Equal(t, tt.expectedError, resp)
				return
			}

			var actualResponse struct {
				Message string         `json:"message"`
				Company models.Company `json:"company"`
			}

			err := json.NewDecoder(rec.Body).Decode(&actualResponse)
			if err != nil {
				t.Fatalf("Error decoding response body: %v", err)
			}
			// Validate that the actual ID is a valid UUID (not empty or invalid)
			if actualResponse.Company.ID == uuid.Nil {
				t.Errorf("Expected a non-empty UUID for ID, but got: %v", actualResponse.Company.ID)
			} else if _, err := uuid.Parse(actualResponse.Company.ID.String()); err != nil {
				t.Errorf("Expected a valid UUID for ID, but got: %v", actualResponse.Company.ID)
			}

			if actualResponse.Message != tt.expectedBody.Message {
				t.Errorf("Expected Message: %v, but got: %v", tt.expectedBody.Message, actualResponse.Message)
			}

			if !compareCompanyIgnoringID(tt.expectedBody.Company, actualResponse.Company) {
				t.Errorf("Expected Name: %v, but got: %v", tt.expectedBody, actualResponse.Company)
			} //assert.Equal(t, tt.expectedBody, actualResponse)

			mockDB.AssertExpectations(t)
		})
	}
}
func TestGetCompany(t *testing.T) {
	validUUID := uuid.New() // A valid UUID
	tests := []struct {
		name          string
		id            string
		mockSetup     func(mockDB *mocks.MockDatabase)
		expectedCode  int
		expectedBody  interface{}
		expectedError map[string]string
	}{
		{
			name: "Valid company retrieval",
			id:   validUUID.String(), // Simulating a valid UUID as a string,
			mockSetup: func(mockDB *mocks.MockDatabase) {
				mockDB.On("GetCompany", validUUID.String()).Return(&models.Company{
					ID:          validUUID,
					Name:        "Tech company",
					Description: "A leading technology company.",
					Employees:   50,
					Registered:  true,
					Type:        "NonProfit",
					CreatedAt:   time.Time{},
					UpdatedAt:   time.Time{},
					DeletedAt:   nil,
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: models.Company{
				ID:          validUUID,
				Name:        "Tech company",
				Description: "A leading technology company.",
				Employees:   50,
				Registered:  true,
				Type:        "NonProfit",
				CreatedAt:   time.Time{},
				UpdatedAt:   time.Time{},
				DeletedAt:   nil,
			},
		},
		{
			name: "Invalid uuid",
			id:   "invaliduuid",
			mockSetup: func(mockDB *mocks.MockDatabase) {
				mockDB.On("GetCompany", "invaliduuid").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedCode: http.StatusBadRequest,
			expectedError: map[string]string{
				"error": "the parameter id is not UUID",
			},
		},
		{
			name: "Company not found",
			id:   validUUID.String(),
			mockSetup: func(mockDB *mocks.MockDatabase) {
				mockDB.On("GetCompany", validUUID.String()).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedCode: http.StatusNotFound,
			expectedError: map[string]string{
				"error": "record not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mocks.MockDatabase)
			app := controllers.NewApp(mockDB, nil, &config.Config{})
			tt.mockSetup(mockDB)

			// Set up the router and handler for this test
			router := mux.NewRouter()
			router.HandleFunc("/api/companies/{id}", app.GetCompany).Methods(http.MethodGet)

			// Prepare request and response recorder
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/companies/%s", tt.id), nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Serve the request through the router
			router.ServeHTTP(rec, req)

			// Assertions
			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != nil {
				var resp map[string]string
				err := json.NewDecoder(rec.Body).Decode(&resp)
				if err != nil {
					t.Fatalf("Error decoding response body: %v", err)
				}
				assert.Equal(t, tt.expectedError, resp)
				return
			}
			var actualResponse models.Company
			err := json.NewDecoder(rec.Body).Decode(&actualResponse)
			if err != nil {
				t.Fatalf("Error decoding response body: %v", err)
			}

			// Validate the company fields
			assert.Equal(t, tt.expectedBody, actualResponse)

			mockDB.AssertExpectations(t)
		})

	}
}

func TestApp_UpdateCompany(t *testing.T) {
	validUUID := uuid.New() // A valid UUID
	tests := []struct {
		name          string
		id            string
		requestBody   interface{}
		mockSetup     func(mockDB *mocks.MockDatabase, mockKa *mocks.MockKafkaProducer)
		expectedCode  int
		expectedBody  interface{}
		expectedError map[string]string
	}{
		{
			name: "Valid company update",
			id:   validUUID.String(), // Simulating a valid UUID as a string,
			requestBody: map[string]interface{}{
				"name":        "Tech company",
				"description": "A leading technology company.",
				"employees":   50,
				"registered":  true,
				"type":        "NonProfit",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("UpdateCompany", validUUID.String(), mock.AnythingOfType("map[string]interface {}")).Return(&models.Company{
					ID:          validUUID,
					Name:        "Tech company",
					Description: "A leading technology company.",
					Employees:   50,
					Registered:  true,
					Type:        "NonProfit",
					CreatedAt:   time.Time{},
					UpdatedAt:   time.Time{},
					DeletedAt:   nil,
				}, nil,
					mockKafka.On("ProduceEvent", mock.AnythingOfType("*kafka.EventMessage")).Return(nil)) // Mock ProduceEvent
			},
			expectedCode: http.StatusOK,
			expectedBody: struct {
				Message string         `json:"message"`
				Company models.Company `json:"company"`
			}{
				Message: "Company updated successfully",
				Company: models.Company{
					ID:          validUUID,
					Name:        "Tech company",
					Description: "A leading technology company.",
					Employees:   50,
					Registered:  true,
					Type:        "NonProfit",
					CreatedAt:   time.Time{},
					UpdatedAt:   time.Time{},
					DeletedAt:   nil,
				},
			},
		},
		{
			name: "Invalid uuid",
			id:   "invaliduuid",
			requestBody: map[string]interface{}{
				"name":        "Tech company",
				"description": "A leading technology company.",
				"employees":   50,
				"registered":  true,
				"type":        "NonProfit",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("UpdateCompany", "invaliduuid", mock.AnythingOfType("map[string]interface {}")).Return(nil, gorm.ErrRecordNotFound,
					mockKafka.On("ProduceEvent", mock.AnythingOfType("map[string]interface {}")).Return(nil)) // Mock ProduceEvent
			},
			expectedCode: http.StatusBadRequest,
			expectedError: map[string]string{
				"error": "the parameter id is not UUID",
			},
		},
		{
			name: "Company not found",
			id:   validUUID.String(),
			requestBody: map[string]interface{}{
				"name":        "Tech company",
				"description": "A leading technology company.",
				"employees":   50,
				"registered":  true,
				"type":        "NonProfit",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("UpdateCompany", validUUID.String(), mock.AnythingOfType("map[string]interface {}")).Return(nil, gorm.ErrRecordNotFound,
					mockKafka.On("ProduceEvent", mock.AnythingOfType("map[string]interface {}")).Return(nil)) // Mock ProduceEvent
			},
			expectedCode: http.StatusNotFound,
			expectedError: map[string]string{
				"error": "record not found",
			},
		},
		{
			name: "Invalid input data - invalid type",
			id:   validUUID.String(),
			requestBody: map[string]interface{}{
				"name":       "Tech company",
				"employees":  50,
				"registered": true,
				"type":       "InvalidType",
			},
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("UpdateCompany", validUUID.String(), mock.AnythingOfType("map[string]interface {}")).Return(nil, errors.New("invalid 'type'. Allowed values are 'Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship'"),
					mockKafka.On("ProduceEvent", mock.AnythingOfType("map[string]interface {}")).Return(nil)) // Mock ProduceEvent
			},
			expectedCode: http.StatusBadRequest,
			expectedError: map[string]string{
				"error": "invalid 'type'. Allowed values are 'Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship'",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mocks.MockDatabase)
			mockKafka := new(mocks.MockKafkaProducer)
			app := controllers.NewApp(mockDB, mockKafka, &config.Config{})
			tt.mockSetup(mockDB, mockKafka)

			// Set up the router and handler for this test
			router := mux.NewRouter()
			router.HandleFunc("/api/companies/{id}", app.UpdateCompany).Methods(http.MethodPatch)

			// Prepare request and response recorder
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/companies/%s", tt.id), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Serve the request through the router
			router.ServeHTTP(rec, req)

			// Assertions
			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != nil {
				var resp map[string]string
				err := json.NewDecoder(rec.Body).Decode(&resp)
				if err != nil {
					t.Fatalf("Error decoding response body: %v", err)
				}
				assert.Equal(t, tt.expectedError, resp)
				return
			}
			var actualResponse struct {
				Message string         `json:"message"`
				Company models.Company `json:"company"`
			}
			err := json.NewDecoder(rec.Body).Decode(&actualResponse)
			if err != nil {
				t.Fatalf("Error decoding response body: %v", err)
			}

			// Validate the company fields
			assert.Equal(t, tt.expectedBody, actualResponse)

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeleteCompany(t *testing.T) {
	validUUID := uuid.New() // A valid UUID
	tests := []struct {
		name         string
		id           string
		mockSetup    func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer)
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name: "Valid company deletion",
			id:   validUUID.String(), // Simulating a valid UUID as a string,
			mockSetup: func(mockDB *mocks.MockDatabase, mockKafka *mocks.MockKafkaProducer) {
				mockDB.On("DeleteCompany", validUUID.String()).Return(nil,
					mockKafka.On("ProduceEvent", mock.AnythingOfType("*kafka.EventMessage")).Return(nil)) // Mock ProduceEvent
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name: "Invalid uuid",
			id:   "invaliduuid",
			mockSetup: func(mockDB *mocks.MockDatabase, mockKa *mocks.MockKafkaProducer) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]string{
				"error": "the parameter id is not UUID",
			},
		},
		{
			name: "Company not found",
			id:   validUUID.String(),
			mockSetup: func(mockDB *mocks.MockDatabase, mockKa *mocks.MockKafkaProducer) {
				mockDB.On("DeleteCompany", validUUID.String()).Return(gorm.ErrRecordNotFound,
					mockKa.On("ProduceEvent", mock.AnythingOfType("map[string]interface {}")).Return(nil)) // Mock ProduceEvent
			},
			expectedCode: http.StatusNotFound,
			expectedBody: map[string]string{
				"error": "record not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mocks.MockDatabase)
			mockKafka := new(mocks.MockKafkaProducer)
			app := controllers.NewApp(mockDB, mockKafka, &config.Config{})
			tt.mockSetup(mockDB, mockKafka)

			// Set up the router and handler for this test
			router := mux.NewRouter()
			router.HandleFunc("/api/companies/{id}", app.DeleteCompany).Methods(http.MethodDelete)

			// Prepare request and response recorder
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/companies/%s", tt.id), nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Serve the request through the router
			router.ServeHTTP(rec, req)

			// Assertions
			assert.Equal(t, tt.expectedCode, rec.Code)
			if tt.expectedBody != nil {

				var actualResponse map[string]string
				err := json.NewDecoder(rec.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatalf("Error decoding response body: %v", err)
				}
				assert.Equal(t, tt.expectedBody, actualResponse)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

// Helper function to compare Company structs, ignoring the ID field
func compareCompanyIgnoringID(expected, actual models.Company) bool {
	return expected.Name == actual.Name &&
		expected.Description == actual.Description &&
		expected.Employees == actual.Employees &&
		expected.Registered == actual.Registered &&
		expected.Type == actual.Type &&
		expected.CreatedAt == actual.CreatedAt &&
		expected.UpdatedAt == actual.UpdatedAt
}
