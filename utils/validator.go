package utils

import (
	"company-service/models"
	"fmt"
)

// ValidateCompanyUpdate validates the updated fields of a company.
func ValidateCompanyUpdate(updatedFields map[string]interface{}) error {
	// Iterate over the updated fields and validate each one
	for field, value := range updatedFields {
		switch field {
		case "name":
			// Validate 'name' field: Ensure it is not empty and has a reasonable length
			if name, ok := value.(string); ok {
				if len(name) > 15 {
					return fmt.Errorf("name cannot be longer than 15 characters")
				}
				if len(name) == 0 {
					return fmt.Errorf("name cannot be empty")
				}
			} else {
				return fmt.Errorf("invalid type for 'name'. It should be a string")
			}
		case "description":
			// Validate 'description' field: Ensure it does not exceed the max length of 3000 characters
			if description, ok := value.(string); ok {
				if len(description) > 3000 {
					return fmt.Errorf("description cannot be longer than 3000 characters")
				}
			} else {
				return fmt.Errorf("invalid type for 'description'. It should be a string")
			}
		case "employees":
			// Validate 'employees' field: Ensure it is a non-negative integer
			if employees, ok := value.(float64); ok { // JSON decodes numbers as float64
				if employees < 0 {
					return fmt.Errorf("employees count cannot be negative")
				}
			} else {
				return fmt.Errorf("invalid type for 'employees'. It should be an integer")
			}
		case "type":
			// Validate 'type' field: Ensure it matches one of the allowed values
			if companyType, ok := value.(string); ok {
				validTypes := map[string]bool{"Corporations": true, "NonProfit": true, "Cooperative": true, "Sole Proprietorship": true}
				if !validTypes[companyType] {
					return fmt.Errorf("invalid 'type'. Allowed values are 'Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship'")
				}
			} else {
				return fmt.Errorf("invalid type for 'type'. It should be a string")
			}
		}
	}
	return nil
}

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
