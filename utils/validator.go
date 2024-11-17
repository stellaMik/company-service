package utils

import (
	"CRUD-API/models"
	"fmt"
)

func ValidateCompanyInput(company *models.Company) error {
	// Validate Name
	if company.Name == "" || len(company.Name) > 15 {
		return fmt.Errorf("Invalid 'Name': it is required and must be at most 15 characters")
	}

	// Validate Employees
	if company.Employees <= 0 {
		return fmt.Errorf("Invalid 'Employees': it must be a positive number")
	}
	// Validate Registered
	if !company.Registered {
		return fmt.Errorf("Invalid 'Registered': it is required and must be true")
	}

	// Validate Type
	if company.Type == "" {
		return fmt.Errorf("Invalid 'Type': it is required")
	}
	validTypes := map[string]bool{
		"Corporations":        true,
		"NonProfit":           true,
		"Cooperative":         true,
		"Sole Proprietorship": true,
	}
	if !validTypes[company.Type] {
		return fmt.Errorf("Invalid 'Type': must be one of 'Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship'")
	}

	// Optional: Additional checks (if needed)

	return nil
}
