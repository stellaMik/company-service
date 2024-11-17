package database

import (
	"company-service/config"
	"company-service/models"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type GormDatabase struct {
	db *gorm.DB
}
type Database interface {
	GetUserByUsername(username string) (*models.User, error)
	CreateCompany(company *models.Company) error
	GetCompany(id string) (*models.Company, error)
	UpdateCompany(id string, fields map[string]interface{}) (*models.Company, error)
	DeleteCompany(id string) error
	CreateDefaultUser(conf *config.Config) error
	CheckIfExistsByID(id string) (*models.Company, error)
	Close() error
}

func InitDB(conf *config.Config) (*GormDatabase, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.DBUser, conf.DBPassword, conf.DBHost, conf.DBPort, conf.DBName)

	// Open the database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	err = db.AutoMigrate(&models.Company{}, &models.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate the database: %w", err)
	}
	newDB := &GormDatabase{db: db}
	err = newDB.CreateDefaultUser(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create default admin user: %w", err)
	}

	return newDB, nil
}

func (g *GormDatabase) CreateDefaultUser(conf *config.Config) error {
	var user models.User

	// Check if the admin user already exists
	err := g.db.Where("username = ?", conf.User).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Hash the password before saving it
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(conf.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("Could not hash the password: %w ", err)
		}
		// Create the default admin user with hashed password
		defaultUser := models.User{
			Username: conf.User,
			Password: string(hashedPassword),
		}

		// Create the admin user in the database
		if err := g.db.Create(&defaultUser).Error; err != nil {
			return fmt.Errorf("Could not create default admin user: %w ", err)
		}
	} else if err != nil {
		// Handle any other errors
		return fmt.Errorf("Could not query the database: %w ", err)
	}
	return nil
}

// GetUserByUsername retrieves a user by their username from the database.
func (g *GormDatabase) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := g.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (g *GormDatabase) CreateCompany(company *models.Company) error {
	if err := g.db.Create(company).Error; err != nil {
		return fmt.Errorf("Could not create a new company record with error: %v", err)
	}
	return nil
}
func (g *GormDatabase) GetCompany(id string) (*models.Company, error) {
	company, err := g.CheckIfExistsByID(id)
	if err != nil {
		return nil, err
	}
	return company, nil
}
func (g *GormDatabase) UpdateCompany(id string, updatedFeilds map[string]interface{}) (*models.Company, error) {
	company, err := g.CheckIfExistsByID(id)
	if err != nil {
		return nil, err
	}
	if err = g.db.Model(&company).Updates(updatedFeilds).Error; err != nil {
		return nil, fmt.Errorf("Could not update company ", err)
	}
	return company, nil
}
func (g *GormDatabase) DeleteCompany(id string) error {
	_, err := g.CheckIfExistsByID(id)
	if err != nil {
		return err
	}
	result := g.db.Delete(&models.Company{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("Could not delete company ", result.Error)
	}
	return nil
}

func (g *GormDatabase) CheckIfExistsByID(id string) (*models.Company, error) {
	// Check if the record exists by ID
	var company models.Company
	if err := g.db.First(&company, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return a custom error if the record is not found
			return nil, fmt.Errorf("company with ID %s not found", id)
		}
		// Return other errors (e.g., database issues)
		return nil, fmt.Errorf("error checking company existence: %v", err)
	}
	// Return nil if the record exists
	return &company, nil
}

// Close the database connection
func (g *GormDatabase) Close() error {
	db, err := g.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
