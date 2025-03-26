package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"fingo/model" // ปรับตามโครงสร้างโปรเจ็กต์ของคุณ
)

// โครงสร้างข้อมูลที่ไม่รวมรหัสผ่าน
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

func main() {
	// เชื่อมต่อกับฐานข้อมูล MySQL
	dsn := "cp_65011212046:65011212046@csmsu@tcp(202.28.34.197:3306)/cp_65011212046?collation=utf8mb4_unicode_ci&parseTime=true"
	dialactor := mysql.Open(dsn)
	db, err := gorm.Open(dialactor, &gorm.Config{})
	if err != nil {
		log.Fatal("Connection failed:", err)
	}
	fmt.Println("Connection successful")

	// ตั้งค่า Gin
	r := gin.Default()

	// สร้าง endpoint สำหรับการเข้าสู่ระบบ
	r.POST("/auth/login", func(c *gin.Context) {
		var loginRequest struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		// รับข้อมูลจาก request body
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// เรียกฟังก์ชัน login เพื่อตรวจสอบข้อมูลผู้ใช้
		customer, err := login(db, loginRequest.Email, loginRequest.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// แปลงข้อมูลลูกค้าเป็น CustomerResponse และส่งกลับ
		customerResponse := CustomerResponse{
			CustomerID:  customer.CustomerID,
			FirstName:   customer.FirstName,
			LastName:    customer.LastName,
			Email:       customer.Email,
			PhoneNumber: customer.PhoneNumber,
			Address:     customer.Address,
			CreatedAt:   customer.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   customer.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// ส่งข้อมูลลูกค้ากลับ โดยไม่มี key ข้างหน้า
		c.JSON(http.StatusOK, customerResponse)
	})

	// เริ่มต้นเซิร์ฟเวอร์ที่พอร์ต 8080
	r.Run(":8080")
}

// ฟังก์ชัน login ที่ตรวจสอบอีเมลและรหัสผ่าน
func login(db *gorm.DB, email, password string) (*model.Customer, error) {
	var customer model.Customer

	// ค้นหาผู้ใช้จากอีเมล
	if err := db.Where("email = ?", email).First(&customer).Error; err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// ตรวจสอบรหัสผ่าน
	if customer.Password != password {
		return nil, fmt.Errorf("incorrect password")
	}

	// ถ้าข้อมูลถูกต้อง ให้ส่งข้อมูลผู้ใช้กลับ
	return &customer, nil
}
