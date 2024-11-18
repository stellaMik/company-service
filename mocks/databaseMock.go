package mocks

import (
	"company-service/config"
	"company-service/models"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock implementation of the Database interface
type MockDatabase struct {
	mock.Mock
}

// Implement the methods of the Database interface

func (m *MockDatabase) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	// Handle the case where the user is nil and return a nil *models.User
	return nil, args.Error(1)
}

func (m *MockDatabase) CreateCompany(company *models.Company) error {
	args := m.Called(company)
	return args.Error(0)
}

func (m *MockDatabase) GetCompany(id string) (*models.Company, error) {
	args := m.Called(id)
	if company, ok := args.Get(0).(*models.Company); ok {
		return company, args.Error(1) // Return the company if it's correctly asserted
	}

	// If the company is nil or not found, return nil and the error
	return nil, args.Error(1)
}

func (m *MockDatabase) UpdateCompany(id string, fields map[string]interface{}) (*models.Company, error) {
	args := m.Called(id, fields)
	if company, ok := args.Get(0).(*models.Company); ok {
		return company, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDatabase) DeleteCompany(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockDatabase) GetIfExistsByID(id string) (*models.Company, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Company), args.Error(1)
}
func (m *MockDatabase) CreateDefaultUser(conf *config.Config) error {
	args := m.Called(conf)
	return args.Error(0)
}
func (m *MockDatabase) CheckIfExistsByName(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}
func (m *MockDatabase) Close() error {
	m.Called()
	return nil
}
