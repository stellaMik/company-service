package models

import (
	"github.com/google/uuid"
	"time"
)

type CompanyType string

type Company struct {
	ID          uuid.UUID  `json:"id" gorm:"primary_key"` //gorm:"type:char(36);primary_key"`
	Name        string     `json:"name" gorm:"size:15;unique;not null"`
	Description string     `json:"description" gorm:"size:3000"`
	Employees   int        `json:"employees" gorm:"not null"`
	Registered  bool       `json:"registered" gorm:"not null"`
	Type        string     `json:"type" gorm:"type:enum('Corporations','NonProfit','Cooperative','Sole Proprietorship');not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deletedAt" gorm:"index"`
}
