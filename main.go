package main

import (
	"fmt"
	_ "gin-gorm/docs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID    int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name  string `json:"name" gorm:"type:varchar(100);not null"`
	Email string `json:"email" gorm:"type:varchar(100);uniqueIndex;not null"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// Global variable to hold the DB connection
var db *gorm.DB
var err error

// @title User API
// @version 1.0
// @description This is a simple API for managing users in a PostgreSQL database.
// @host localhost:8000
// BasePath /api/v1 Because of this in url "/api/v1" was repeating and causing the error.
// @contact.name API Support
// @contact.url http://localhost:8000/support   // Local URL for your development environment
// @contact.email support@localhost.com
func main() {
	// Database connection setup
	dsn := "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	fmt.Println("Successfully connected to the database.")

	// Auto-migrate your database schema (optional but recommended)
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	// Create Gin router and initialize routes
	r := gin.Default()
	initializeRoutes(r)

	// Start the server
	r.Run(":8080")
}

// initializeRoutes sets up your routes
func initializeRoutes(r *gin.Engine) {
	r.GET("/api/v1/users", getUsers)
	r.GET("/api/v1/users/:id", getUser)
	r.POST("/api/v1/users", createUser)
	r.PUT("/api/v1/users/:id", updateUser)
	r.DELETE("/api/v1/users/:id", deleteUser)
}

// Initialize DB connection
func initDB() {

	dsn := os.Getenv("DATABASE_URL")
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}

	// Auto-migrate the User struct to create the 'users' table
	db.AutoMigrate(&User{})
}

// Fetch all users
// @Summary Get all users
// @Description Retrieve a list of all users in the database
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {array} User
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users [get]
func getUsers(c *gin.Context) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error fetching users"})
		return
	}
	c.JSON(200, users)
}

// Fetch a single user by ID
// @Summary Get user by ID
// @Description Retrieve a single user's details by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID" // The ID of the user to retrieve
// @Success 200 {object} User // The user object returned in the response
// @Failure 400 {object} ErrorResponse // Bad request if the ID is invalid
// @Failure 404 {object} ErrorResponse // User not found
// @Failure 500 {object} ErrorResponse // Internal server error
// @Router /api/v1/users/{id} [get]
func getUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
		return
	}
	c.JSON(200, user)
}

// Create a new user
// @Summary Create a new user
// @Description Create a new user by providing a name and email
// @Tags Users
// @Accept  json
// @Produce  json
// @Param user body User true "New user information"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users [post]
func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid input"})
		return
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to create user"})
		return
	}

	c.JSON(201, user)
}

// Update an existing user
// @Summary Update an existing user
// @Description Update a user's name and email by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID" // This is the ID parameter from the URL path
// @Param user body User true "Updated user information" // The request body (updated user data)
// @Success 200 {object} User // The updated user object returned in the response
// @Failure 400 {object} ErrorResponse // Bad request if the input is invalid
// @Failure 404 {object} ErrorResponse // If the user is not found
// @Failure 500 {object} ErrorResponse // Internal server error
// @Router /api/v1/users/{id} [put]
func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid input"})
		return
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update user"})
		return
	}

	c.JSON(200, user)
}

// Delete a user by ID
// @Summary Delete a user
// @Description Delete a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID" // ID of the user to delete
// @Success 200 {string} string "User deleted" // Success message
// @Failure 404 {object} ErrorResponse // If the user is not found
// @Failure 500 {object} ErrorResponse // Internal server error
// @Router /api/v1/users/{id} [delete]
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
		return
	}

	if err := db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to delete user"})
		return
	}

	c.JSON(200, gin.H{"message": "User deleted"})
}
