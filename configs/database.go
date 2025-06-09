package config

import (
	"fmt"
	"log"
	"os"

	"auth.alexmust/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	fmt.Printf("Database connection details:\nHost: %s\nPort: %s\nUser: %s\nDB: %s\n",
		host, port, user, dbname)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	if err := database.Raw("SELECT 1").Error; err != nil {
		log.Fatal("Database connection test failed:", err)
	}

	fmt.Printf("Attempting to migrate User model: %+v\n", models.User{})

	err = database.AutoMigrate(&models.User{})
	if err != nil {
		log.Printf("Migration error details: %v\n", err)
		log.Fatal("Failed to migrate database:", err)
	}
	
	fmt.Println("Миграция успешно выполнена")
	DB = database
	fmt.Println("Успешное подключение к базе данных")
}
