package main

import (
	"fmt"
	"log"
	"net/http"

	"fingo/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

	// สำหรับการเข้าสู่ระบบ
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
		customerResponse := model.CustomerResponse{
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

	// สำหรับการแก้ไขที่อยู่
	r.PUT("/update_address", func(c *gin.Context) {
		var updateRequest struct {
			Email   string `json:"email"`
			Address string `json:"address"`
		}
		// รับข้อมูลจาก request body
		if err := c.ShouldBindJSON(&updateRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// ตรวจสอบการยืนยันตัวตนและค้นหาลูกค้าจากอีเมล
		var customer model.Customer
		if err := db.Where("email = ?", updateRequest.Email).First(&customer).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer not found"})
			return
		}
		// อัปเดตที่อยู่ของลูกค้า
		customer.Address = updateRequest.Address
		if err := db.Save(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
			return
		}
		// แปลงข้อมูลลูกค้าเป็น CustomerResponse และส่งกลับ
		customerResponse := model.CustomerResponse{
			CustomerID:  customer.CustomerID,
			FirstName:   customer.FirstName,
			LastName:    customer.LastName,
			Email:       customer.Email,
			PhoneNumber: customer.PhoneNumber,
			Address:     customer.Address,
			CreatedAt:   customer.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   customer.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		// ส่งข้อมูลลูกค้าหลังจากอัปเดตที่อยู่
		c.JSON(http.StatusOK, customerResponse)
	})

	// เพิ่มลูกค้า
	r.POST("/auth/register", func(c *gin.Context) {
		var registerRequest struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
			Phone     string `json:"phone"`
			Address   string `json:"address"`
		}
		// รับข้อมูลจาก request body
		if err := c.ShouldBindJSON(&registerRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// แฮชรหัสผ่านที่ได้รับจาก request
		hashedPassword, err := hashPassword(registerRequest.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		// สร้างลูกค้ารายใหม่
		customer := model.Customer{
			FirstName:   registerRequest.FirstName,
			LastName:    registerRequest.LastName,
			Email:       registerRequest.Email,
			Password:    hashedPassword, // เก็บรหัสผ่านที่แฮชแล้ว
			PhoneNumber: registerRequest.Phone,
			Address:     registerRequest.Address,
		}
		// เก็บข้อมูลผู้ใช้ใหม่ในฐานข้อมูล
		if err := db.Create(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register customer"})
			return
		}
		// ส่งข้อมูลลูกค้ากลับโดยไม่รวมรหัสผ่าน
		customerResponse := model.CustomerResponse{
			CustomerID:  customer.CustomerID,
			FirstName:   customer.FirstName,
			LastName:    customer.LastName,
			Email:       customer.Email,
			PhoneNumber: customer.PhoneNumber,
			Address:     customer.Address,
			CreatedAt:   customer.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   customer.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		// ส่งข้อมูลลูกค้าหลังจากลงทะเบียน
		c.JSON(http.StatusOK, customerResponse)
	})

	// สำหรับการเปลี่ยนรหัสผ่าน
	r.PUT("/change_password", func(c *gin.Context) {
		var changePasswordRequest struct {
			Email       string `json:"email"`
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		// รับข้อมูลจาก request body
		if err := c.ShouldBindJSON(&changePasswordRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// ค้นหาลูกค้าจากอีเมล
		var customer model.Customer
		if err := db.Where("email = ?", changePasswordRequest.Email).First(&customer).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer not found"})
			return
		}

		// ตรวจสอบรหัสผ่านเก่าที่แฮชแล้ว
		if err := checkPasswordHash(changePasswordRequest.OldPassword, customer.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password"})
			return
		}

		// แฮชรหัสผ่านใหม่
		hashedNewPassword, err := hashPassword(changePasswordRequest.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}

		// อัพเดตรหัสผ่านใหม่ในฐานข้อมูล
		customer.Password = hashedNewPassword
		if err := db.Save(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		// ส่งข้อมูลกลับว่าเปลี่ยนรหัสผ่านสำเร็จ
		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	})

	//สำหรับเพิ่มสินค้า
	r.POST("/add_product", func(c *gin.Context) {
		var productRequest struct {
			ProductName   string `json:"product_name"`
			Description   string `json:"description"`
			Price         string `json:"price"`
			StockQuantity int    `json:"stock_quantity"`
		}
		// รับข้อมูลจาก request body
		if err := c.ShouldBindJSON(&productRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// สร้างสินค้าใหม่
		newProduct := model.Product{
			ProductName:   productRequest.ProductName,
			Description:   productRequest.Description,
			Price:         productRequest.Price,
			StockQuantity: productRequest.StockQuantity,
		}
		// บันทึกลงฐานข้อมูล
		if err := db.Create(&newProduct).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert product"})
			return
		}
		// ส่งข้อมูลสินค้ากลับ
		c.JSON(http.StatusOK, gin.H{
			"message":    "Product added successfully",
			"product_id": newProduct.ProductID,
		})
	})
	
	// สำหรับเพิ่ม Cart
	r.POST("/add_cart", func(c *gin.Context) {
		var cartRequest struct {
			CustomerID int    `json:"customer_id"`
			CartName   string `json:"cart_name"`
		}
		// อ่านข้อมูล JSON
		if err := c.ShouldBindJSON(&cartRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// เรียกฟังก์ชันเพิ่มข้อมูล Cart
		cart, err := addCart(db, cartRequest.CustomerID, cartRequest.CartName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
			return
		}
		// ส่งข้อมูล Cart ที่สร้างกลับไป
		c.JSON(http.StatusOK, gin.H{
			"cart_id":     cart.CartID,
			"customer_id": cart.CustomerID,
			"cart_name":   cart.CartName,
			"created_at":  cart.CreatedAt,
			"updated_at":  cart.UpdatedAt,
		})
	})

	//สำหรับเพิ่มสินค้าในตะกร้า
	r.POST("/add_cart_item", func(c *gin.Context) {
		var cartItemRequest struct {
			CartID    int `json:"cart_id"`
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
		}

		// อ่านข้อมูล JSON
		if err := c.ShouldBindJSON(&cartItemRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// เรียกฟังก์ชันเพิ่มข้อมูล CartItem
		cartItem, err := addCartItem(db, cartItemRequest.CartID, cartItemRequest.ProductID, cartItemRequest.Quantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
			return
		}

		// ส่งข้อมูล CartItem ที่สร้างกลับไป
		c.JSON(http.StatusOK, gin.H{
			"cart_item_id": cartItem.CartItemID,
			"cart_id":      cartItem.CartID,
			"product_id":   cartItem.ProductID,
			"quantity":     cartItem.Quantity,
			"created_at":   cartItem.CreatedAt,
			"updated_at":   cartItem.UpdatedAt,
		})
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
	// ตรวจสอบรหัสผ่านที่ถูกแฮช
	if err := checkPasswordHash(password, customer.Password); err != nil {
		return nil, fmt.Errorf("incorrect password")
	}
	// ถ้าข้อมูลถูกต้อง ให้ส่งข้อมูลผู้ใช้กลับ
	return &customer, nil
}

// ฟังก์ชันที่ใช้แฮชรหัสผ่าน
func hashPassword(password string) (string, error) {
	// ใช้ bcrypt ในการแฮชรหัสผ่าน
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ฟังก์ชันที่ใช้ตรวจสอบรหัสผ่านที่ถูกแฮช
func checkPasswordHash(password, hash string) error {
	// ใช้ bcrypt.CompareHashAndPassword เพื่อตรวจสอบว่า hash กับรหัสผ่านตรงกันหรือไม่
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ฟังก์ชันเพิ่มข้อมูล Cart
func addCart(db *gorm.DB, customerID int, cartName string) (*model.Cart, error) {
	cart := model.Cart{
		CustomerID: customerID,
		CartName:   cartName,
	}

	// บันทึกข้อมูลลงฐานข้อมูล
	if err := db.Create(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

// ฟังก์ชันเพิ่มสินค้าลงใน CartItem
func addCartItem(db *gorm.DB, cartID, productID, quantity int) (*model.CartItem, error) {
	cartItem := model.CartItem{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
	}

	// บันทึกข้อมูลลงฐานข้อมูล
	if err := db.Create(&cartItem).Error; err != nil {
		return nil, err
	}
	return &cartItem, nil
}