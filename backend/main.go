package main

import (
	"SimLife/middleware"
	"SimLife/models"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
		handlers.AllowedOrigins(
			[]string{
				"http://localhost:3000", "http://192.168.1.19:3000",
			},
		),
		handlers.AllowedMethods(
			[]string{
				"GET", "POST", "PUT", "DELETE", "OPTIONS",
			},
		),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// Registration endpoint
	r.HandleFunc(
		"/api/register", func(w http.ResponseWriter, r *http.Request) {
			var user models.User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				http.Error(
					w, `{"error":"Invalid request"}`, http.StatusBadRequest,
				)
				return
			}

			// Backend validation: Password length
			if len(user.PasswordHash) < 5 {
				http.Error(
					w, `{"error":"Password must be at least 5 characters"}`,
					http.StatusBadRequest,
				)
				return
			}

			// Check if username exists
			var existingUser models.User
			if result := db.Where(
				"username = ?", user.Username,
			).First(&existingUser); result.Error == nil {
				http.Error(
					w, `{"error":"Username already exists"}`,
					http.StatusConflict,
				)
				return
			}

			// Save to DB (password auto-hashed via BeforeSave hook)
			if result := db.Create(&user); result.Error != nil {
				http.Error(
					w, `{"error":"Failed to create user"}`,
					http.StatusInternalServerError,
				)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			err := json.NewEncoder(w).Encode(
				map[string]string{
					"message": "User created",
				},
			)
			if err != nil {
				return
			}
		},
	).Methods("POST")

	// Login endpoint
	r.HandleFunc(
		"/api/login", func(w http.ResponseWriter, r *http.Request) {
			var loginRequest struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
				http.Error(
					w, `{"error":"Invalid request"}`, http.StatusBadRequest,
				)
				return
			}
			fmt.Printf(
				"Login attempt - Username: %s, Password: %s\n",
				loginRequest.Username, loginRequest.Password,
			)
			// Find user by username
			var user models.User
			if result := db.Where(
				"username = ?", loginRequest.Username,
			).First(&user); result.Error != nil {
				fmt.Printf("User not found: %s\n", loginRequest.Username)
				http.Error(
					w, `{"error":"Invalid username or password"}`,
					http.StatusUnauthorized,
				)
				return
			}
			fmt.Printf("User found: %+v\n", user)
			fmt.Printf("Stored hash: %s\n", user.PasswordHash)
			// Compare password hash
			if err := bcrypt.CompareHashAndPassword(
				[]byte(user.PasswordHash), []byte(loginRequest.Password),
			); err != nil {
				fmt.Printf(
					"Password mismatch for user: %s, Error: %v\n",
					loginRequest.Username, err,
				)
				http.Error(
					w, `{"error":"Invalid username or password"}`,
					http.StatusUnauthorized,
				)
				return
			}

			// Generate JWT token
			token := jwt.NewWithClaims(
				jwt.SigningMethodHS256, jwt.MapClaims{
					"user_id":  user.ID.String(),
					"username": user.Username,
					"exp":      time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
				},
			)

			tokenString, err := token.SignedString(jwtSecret)
			if err != nil {
				http.Error(
					w, `{"error":"Failed to generate token"}`,
					http.StatusInternalServerError,
				)
				return
			}

			// Return token to client
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(
				map[string]string{
					"token":    tokenString,
					"user_id":  user.ID.String(),
					"username": user.Username,
				},
			)
			if err != nil {
				return
			}
		},
	).Methods("POST")

	r.HandleFunc(
		"/api/banks", func(w http.ResponseWriter, r *http.Request) {
			var banks []models.Bank
			if result := db.Find(&banks); result.Error != nil {
				http.Error(
					w, `{"error":"Failed to fetch banks"}`,
					http.StatusInternalServerError,
				)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(banks)
			if err != nil {
				return
			}
		},
	).Methods("GET")

	// 2. Create a new bank account (protected route)
	r.Handle(
		"/api/bank-accounts", middleware.AuthMiddleware(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					// Extract user ID from JWT
					claims := r.Context().Value("user").(jwt.MapClaims)
					userID, _ := uuid.Parse(claims["user_id"].(string))

					var bankAccountRequest struct {
						BankID        string  `json:"bank_id"`
						AgencyNumber  string  `json:"agency_number"`
						AccountNumber string  `json:"account_number"`
						Balance       float64 `json:"balance"`
					}

					if err := json.NewDecoder(r.Body).Decode(&bankAccountRequest); err != nil {
						http.Error(
							w, `{"error":"Invalid request"}`,
							http.StatusBadRequest,
						)
						return
					}

					// Validate bank exists
					var bank models.Bank
					bankUUID, _ := uuid.Parse(bankAccountRequest.BankID)
					if result := db.First(
						&bank, "id = ?", bankUUID,
					); result.Error != nil {
						http.Error(
							w, `{"error":"Invalid bank"}`,
							http.StatusBadRequest,
						)
						return
					}

					// Create bank account
					bankAccount := models.BankAccount{
						UserID:        userID,
						BankID:        bankUUID,
						AgencyNumber:  bankAccountRequest.AgencyNumber,
						AccountNumber: bankAccountRequest.AccountNumber,
						Balance:       bankAccountRequest.Balance,
					}

					if result := db.Create(&bankAccount); result.Error != nil {
						http.Error(
							w, `{"error":"Failed to create account"}`,
							http.StatusInternalServerError,
						)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					err := json.NewEncoder(w).Encode(bankAccount)
					if err != nil {
						return
					}
				},
			),
		),
	).Methods("POST")

	r.Handle(
		"/api/bank-accounts", middleware.AuthMiddleware(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					// Extract user ID from JWT
					claims := r.Context().Value("user").(jwt.MapClaims)
					userID, _ := uuid.Parse(claims["user_id"].(string))

					// Fetch user's bank accounts
					var accounts []models.BankAccount
					if result := db.Where(
						"user_id = ?", userID,
					).Find(&accounts); result.Error != nil {
						http.Error(
							w, `{"error":"Failed to fetch accounts"}`,
							http.StatusInternalServerError,
						)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					err := json.NewEncoder(w).Encode(accounts)
					if err != nil {
						return
					}
				},
			),
		),
	).Methods("GET")

	r.HandleFunc(
		"/api/transaction-categories",
		func(w http.ResponseWriter, r *http.Request) {
			var categories []models.TransactionCategory
			if result := db.Find(&categories); result.Error != nil {
				http.Error(
					w, `{"error":"Failed to fetch categories"}`,
					http.StatusInternalServerError,
				)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(categories)
			if err != nil {
				return
			}
		},
	).Methods("GET")

	// 2. Create transaction (protected route)
	r.Handle(
		"/api/transactions", middleware.AuthMiddleware(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					claims := r.Context().Value("user").(jwt.MapClaims)
					userID, _ := uuid.Parse(claims["user_id"].(string))

					var transactionRequest struct {
						CategoryID    string  `json:"category_id"`
						FromAccountID string  `json:"from_account_id"` // Required for debit/transfer
						ToAccountID   string  `json:"to_account_id"`   // Required for credit/transfer
						Amount        float64 `json:"amount"`
						Description   string  `json:"description"`
					}

					if err := json.NewDecoder(r.Body).Decode(&transactionRequest); err != nil {
						http.Error(
							w, `{"error":"Invalid request"}`,
							http.StatusBadRequest,
						)
						return
					}

					// Start database transaction
					tx := db.Begin()

					// Get category type (credit/debit/transfer)
					var category models.TransactionCategory
					categoryID, _ := uuid.Parse(transactionRequest.CategoryID)
					if result := tx.First(
						&category, "id = ?", categoryID,
					); result.Error != nil {
						tx.Rollback()
						http.Error(
							w, `{"error":"Invalid category"}`,
							http.StatusBadRequest,
						)
						return
					}

					// Update balances based on transaction type
					switch category.Type {
					case "transfer":
						// Deduct from "from_account"
						if err := tx.Model(&models.BankAccount{}).
							Where("id = ?", transactionRequest.FromAccountID).
							Update(
								"balance",
								gorm.Expr(
									"balance - ?", transactionRequest.Amount,
								),
							).
							Error; err != nil {
							tx.Rollback()
							http.Error(
								w,
								`{"error":"Failed to update sender balance"}`,
								http.StatusInternalServerError,
							)
							return
						}

						// Add to "to_account"
						if err := tx.Model(&models.BankAccount{}).
							Where("id = ?", transactionRequest.ToAccountID).
							Update(
								"balance",
								gorm.Expr(
									"balance + ?", transactionRequest.Amount,
								),
							).
							Error; err != nil {
							tx.Rollback()
							http.Error(
								w,
								`{"error":"Failed to update receiver balance"}`,
								http.StatusInternalServerError,
							)
							return
						}

					case "debit":
						if transactionRequest.FromAccountID == "" {
							http.Error(
								w,
								`{"error":"From account required for debit"}`,
								http.StatusBadRequest,
							)
							return
						}

						// Verify account exists
						var account models.BankAccount
						if result := tx.First(
							&account, "id = ?",
							transactionRequest.FromAccountID,
						); result.Error != nil {
							tx.Rollback()
							http.Error(
								w, `{"error":"Sender account not found"}`,
								http.StatusBadRequest,
							)
							return
						}
						// Deduct from "from_account"
						if err := tx.Model(&models.BankAccount{}).
							Where("id = ?", transactionRequest.FromAccountID).
							Update(
								"balance",
								gorm.Expr(
									"balance - ?", transactionRequest.Amount,
								),
							).
							Error; err != nil {
							tx.Rollback()
							http.Error(
								w, `{"error":"Failed to update balance"}`,
								http.StatusInternalServerError,
							)
							return
						}

					case "credit":
						if transactionRequest.ToAccountID == "" {
							http.Error(
								w, `{"error":"To account required for credit"}`,
								http.StatusBadRequest,
							)
							return
						}
						// Add to "to_account"
						if err := tx.Model(&models.BankAccount{}).
							Where("id = ?", transactionRequest.ToAccountID).
							Update(
								"balance",
								gorm.Expr(
									"balance + ?", transactionRequest.Amount,
								),
							).
							Error; err != nil {
							tx.Rollback()
							http.Error(
								w, `{"error":"Failed to update balance"}`,
								http.StatusInternalServerError,
							)
							return
						}
					}
					var fromAccountID, toAccountID *uuid.UUID
					var err error

					// Parse FromAccountID only if provided
					if transactionRequest.FromAccountID != "" {
						parsedID, err := uuid.Parse(transactionRequest.FromAccountID)
						if err != nil {
							http.Error(
								w, `{"error":"Invalid sender account ID"}`,
								http.StatusBadRequest,
							)
							return
						}
						fromAccountID = &parsedID // Assign pointer
					}

					// Parse ToAccountID only if non-empty
					if transactionRequest.ToAccountID != "" {
						parsedID, err := uuid.Parse(transactionRequest.ToAccountID)
						if err != nil {
							http.Error(
								w, `{"error":"Invalid receiver account ID"}`,
								http.StatusBadRequest,
							)
							return
						}
						toAccountID = &parsedID // Assign pointer
					}
					// Create transaction record
					transaction := models.Transaction{
						UserID:        userID,
						CategoryID:    categoryID,
						FromAccountID: fromAccountID, // Can be nil
						ToAccountID:   toAccountID,   // Can be nil
						Amount:        transactionRequest.Amount,
						Description:   transactionRequest.Description,
					}

					if result := tx.Create(&transaction); result.Error != nil {
						tx.Rollback()
						http.Error(
							w, `{"error":"Failed to create transaction"}`,
							http.StatusInternalServerError,
						)
						return
					}

					if transactionRequest.FromAccountID == "" {
						transaction.FromAccountID = nil
					}
					if transactionRequest.ToAccountID == "" {
						transaction.ToAccountID = nil
					}

					tx.Commit()
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					err = json.NewEncoder(w).Encode(transaction)
					if err != nil {
						return
					}
				},
			),
		),
	).Methods("POST")

	// Add this to your main.go file, next to the other API route handlers
	r.Handle(
		"/api/transactions", middleware.AuthMiddleware(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if r.Method == "GET" {
						// Extract user ID from JWT
						claims := r.Context().Value("user").(jwt.MapClaims)
						userID, _ := uuid.Parse(claims["user_id"].(string))

						// Fetch user's transactions with related data
						var transactions []models.Transaction
						if result := db.Preload("Category").
							Preload("FromAccount").
							Preload("ToAccount").
							Where("user_id = ?", userID).
							Order("transaction_date DESC").
							Find(&transactions); result.Error != nil {
							http.Error(
								w, `{"error":"Failed to fetch transactions"}`,
								http.StatusInternalServerError,
							)
							return
						}

						w.Header().Set("Content-Type", "application/json")
						err := json.NewEncoder(w).Encode(transactions)
						if err != nil {
							return
						}
					}
				},
			),
		),
	).Methods("GET")

	// Start server with CORS middleware
	err = http.ListenAndServe(":8080", corsMiddleware(r))
	if err != nil {
		return
	}
}
