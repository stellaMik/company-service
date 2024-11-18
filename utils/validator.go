package utils

import (
	"company-service/models"
	"fmt"
)

// ValidateCompanyInput validates the input fields of a company.
func ValidateCompanyInput(company *models.Company) error {
	// Validate Name
	if company.Name == "" || len(company.Name) > 15 {
		return fmt.Errorf("invalid 'Name': it is required and must be at most 15 characters")
	}

	// Validate Employees
	if company.Employees <= 0 {
		return fmt.Errorf("invalid 'Employees': it must be a positive number")
	}
	// Validate Registered
	if !company.Registered {
		return fmt.Errorf("invalid 'Registered': it is required and must be true")
	}

	// Validate Type
	if company.Type == "" {
		return fmt.Errorf("invalid 'Type': it is required")
	}
	validTypes := map[string]bool{
		"Corporations":        true,
		"NonProfit":           true,
		"Cooperative":         true,
		"Sole Proprietorship": true,
	}
	if !validTypes[company.Type] {
		return fmt.Errorf("invalid 'Type': must be one of 'Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship'")
	}

	return nil
}
