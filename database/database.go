package database

import (
	"log"
	"yonatan/labpro/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(databaseURL string) {
	var err error
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema
	err = DB.AutoMigrate(
		&models.User{},
		&models.Course{},
		&models.Module{},
		&models.UserCourse{},
		&models.UserModuleProgress{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create admin user if not exists
	createAdminUser()
}

func createAdminUser() {
	var admin models.User
	result := DB.Where("username = ?", "admin").First(&admin)
	if result.Error == gorm.ErrRecordNotFound {
		admin = models.User{
			Username:  "admin",
			Email:     "admin@labpro.com",
			FirstName: "Admin",
			LastName:  "User",
			Balance:   0,
			IsAdmin:   true,
		}
		admin.SetPassword("admin123")
		DB.Create(&admin)
		log.Println("Admin user created with username: admin, password: admin123")
	}
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
