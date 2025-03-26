package main

import (
	"fingo/model"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func login(db *gorm.DB, email, password string) (*model.Customer, error) {
	var customer model.Customer

	// Find the customer by email
	if err := db.Where("email = ?", email).First(&customer).Error; err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Check if the provided password matches
	if customer.Password != password {
		return nil, fmt.Errorf("incorrect password")
	}

	// If authentication is successful, return the customer
	return &customer, nil
}

func main() {
	dsn := "cp_65011212046:65011212046@csmsu@tcp(202.28.34.197:3306)/cp_65011212046?collation=utf8mb4_unicode_ci&parseTime=true"
	dialactor := mysql.Open(dsn)
	_, err := gorm.Open(dialactor)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection successful")
}