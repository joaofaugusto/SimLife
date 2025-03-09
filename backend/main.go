package main

import (
	"SimLife/models"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Connect to PostgresSQL
	dsn := "host=localhost user=postgres password=123 dbname=simlife port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		return
	}
	var jwtSecret = []byte("4b08286222a69603f47f766d7f95a93d") // Replace with a secure secret
	// Initialize Gorilla router
	r := mux.NewRouter()

	// CORS middleware (for React)
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000", "http://192.168.1.19:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// Registration endpoint
	r.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
			return
		}

		// Backend validation: Password length
		if len(user.PasswordHash) < 5 {
			http.Error(w, `{"error":"Password must be at least 5 characters"}`, http.StatusBadRequest)
			return
		}

		// Check if username exists
		var existingUser models.User
		if result := db.Where("username = ?", user.Username).First(&existingUser); result.Error == nil {
			http.Error(w, `{"error":"Username already exists"}`, http.StatusConflict)
			return
		}

		// Save to DB (password auto-hashed via BeforeSave hook)
		if result := db.Create(&user); result.Error != nil {
			http.Error(w, `{"error":"Failed to create user"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(map[string]string{
			"message": "User created",
		})
		if err != nil {
			return
		}
	}).Methods("POST")

	// Login endpoint
	r.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		var loginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
			http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
			return
		}
		fmt.Printf("Login attempt - Username: %s, Password: %s\n", loginRequest.Username, loginRequest.Password)
		// Find user by username
		var user models.User
		if result := db.Where("username = ?", loginRequest.Username).First(&user); result.Error != nil {
			fmt.Printf("User not found: %s\n", loginRequest.Username)
			http.Error(w, `{"error":"Invalid username or password"}`, http.StatusUnauthorized)
			return
		}
		fmt.Printf("User found: %+v\n", user)
		fmt.Printf("Stored hash: %s\n", user.PasswordHash)
		// Compare password hash
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Password)); err != nil {
			fmt.Printf("Password mismatch for user: %s, Error: %v\n", loginRequest.Username, err)
			http.Error(w, `{"error":"Invalid username or password"}`, http.StatusUnauthorized)
			return
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID.String(),
			"username": user.Username,
			"exp":      time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
			return
		}

		// Return token to client
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{
			"token":    tokenString,
			"user_id":  user.ID.String(),
			"username": user.Username,
		})
		if err != nil {
			return
		}
	}).Methods("POST")

	// Start server with CORS middleware
	err = http.ListenAndServe(":8080", corsMiddleware(r))
	if err != nil {
		return
	}
}
