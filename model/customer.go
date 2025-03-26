package model

import (
	"time"
)

type Customer struct {
	CustomerID  int       `gorm:"column:customer_id;primary_key;AUTO_INCREMENT"`
	FirstName   string    `gorm:"column:first_name;NOT NULL"`
	LastName    string    `gorm:"column:last_name;NOT NULL"`
	Email       string    `gorm:"column:email;NOT NULL"`
	PhoneNumber string    `gorm:"column:phone_number"`
	Address     string    `gorm:"column:address"`
	Password    string    `gorm:"column:password;NOT NULL"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

func (m *Customer) TableName() string {
	return "customer"
}

type CustomerResponse struct {
	CustomerID  int    `json:"customer_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}