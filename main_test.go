package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Declare db as a global variable so that it references the db initialized in main.go
var testRouter *gin.Engine

// setupIntegrationEnvironment initializes the test DB and sets up the routes for testing
func setupIntegrationEnvironment() *gin.Engine {
	// Use the same connection details as the main app
	dsn := "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Auto-migrate to ensure schema is up-to-date
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Gin engine with routes
	r := gin.Default()
	initializeRoutes(r)

	return r
}

// resetDatabase resets the database by truncating the users table
func resetDatabase() {
	if db == nil {
		log.Fatalf("db is nil in resetDatabase")
	}
	// Reset the state of the database for the users table
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE;")
}

// TestGetUsers tests the /api/v1/users endpoint
func TestGetUsers(t *testing.T) {
	r := setupIntegrationEnvironment()
	defer resetDatabase()

	// Seed the database with test users
	db.Create(&User{Name: "Alice", Email: "alice@example.com"})
	db.Create(&User{Name: "Bob", Email: "bob@example.com"})

	// Perform the GET request to fetch users
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Validate response
	assert.Equal(t, http.StatusOK, w.Code)

	var users []User
	err := json.Unmarshal(w.Body.Bytes(), &users)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users)) // Ensure two users are returned
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, "Bob", users[1].Name)
}

// TestCreateUser tests the POST /api/v1/users endpoint
func TestCreateUser(t *testing.T) {
	r := setupIntegrationEnvironment()
	defer resetDatabase()

	// Define a new user to be created
	newUser := User{Name: "Charlie", Email: "charlie@example.com"}
	jsonData, err := json.Marshal(newUser)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	// Perform the POST request to create a new user
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Validate response
	assert.Equal(t, http.StatusCreated, w.Code)

	var createdUser User
	err = json.Unmarshal(w.Body.Bytes(), &createdUser)
	assert.NoError(t, err)
	assert.Equal(t, "Charlie", createdUser.Name)
	assert.Equal(t, "charlie@example.com", createdUser.Email)
}

// TestGetUser tests the /api/v1/users/:id endpoint
func TestGetUser(t *testing.T) {
	r := setupIntegrationEnvironment()
	defer resetDatabase()

	// Create a user to fetch
	db.Create(&User{Name: "David", Email: "david@example.com"})

	// Fetch the user by ID (assuming ID = 1)
	req, _ := http.NewRequest("GET", "/api/v1/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Validate response
	assert.Equal(t, http.StatusOK, w.Code)

	var user User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "David", user.Name)
	assert.Equal(t, "david@example.com", user.Email)
}

// TestUpdateUser tests the PUT /api/v1/users/:id endpoint
func TestUpdateUser(t *testing.T) {
	r := setupIntegrationEnvironment()
	defer resetDatabase()

	// Create a user to update
	db.Create(&User{Name: "Eve", Email: "eve@example.com"})

	// Define new data for the user
	updatedUser := User{Name: "Eve Updated", Email: "eveupdated@example.com"}
	jsonData, err := json.Marshal(updatedUser)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	// Perform the PUT request to update the user
	req, _ := http.NewRequest("PUT", "/api/v1/users/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Validate response
	assert.Equal(t, http.StatusOK, w.Code)

	var user User
	err = json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "Eve Updated", user.Name)
	assert.Equal(t, "eveupdated@example.com", user.Email)
}

// TestDeleteUser tests the DELETE /api/v1/users/:id endpoint
func TestDeleteUser(t *testing.T) {
	r := setupIntegrationEnvironment()
	defer resetDatabase()

	// Create a user to delete
	db.Create(&User{Name: "Frank", Email: "frank@example.com"})

	// Perform the DELETE request to delete the user by ID (assuming ID = 1)
	req, _ := http.NewRequest("DELETE", "/api/v1/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Validate response
	assert.Equal(t, http.StatusOK, w.Code)

	// Ensure that the user is deleted
	var user User
	err := db.First(&user, 1).Error
	assert.Error(t, err) // User should not be found
}
