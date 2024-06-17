package handler

import (
	"birthdayReminder/app/internal/handler/auth"
	"birthdayReminder/app/internal/handler/available_user"
	"birthdayReminder/app/internal/handler/login"
	"birthdayReminder/app/internal/handler/subscribe"
	"birthdayReminder/app/internal/repository/user"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// TODO - проверить формат даты
// TODO добавить логи

type Handler struct {
	JWTSecretKey     string
	userRepo         UserRepository
	subscriptionRepo SubscriptionRepository
}

func New(userRepo UserRepository, subscriptionRepo SubscriptionRepository) *Handler {
	return &Handler{JWTSecretKey: os.Getenv("JWT_SECRET_KEY"), userRepo: userRepo, subscriptionRepo: subscriptionRepo}
}

// Register /api/registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var user user.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing request body: %v", err)
		}
	}(r.Body)

	if user.Name == "" || user.Email == "" || user.Password == "" || user.DateOfBirth.IsZero() {
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	err = h.userRepo.CreateUser(&user, hashedPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving user to database: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("User registered successfully"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

// Login /api/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var reqBody login.Dto
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing request body: %v", err)
		}
	}(r.Body)

	dbUser, err := h.userRepo.GetUserByEmail(reqBody.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(reqBody.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := auth.GenerateJWT(dbUser.ID, h.JWTSecretKey)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   true,
	})

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Login successful"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

// Subscribe /api/subscribe
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling subscription request")

	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseJWT(tokenStr, h.JWTSecretKey)
	if err != nil {
		log.Println("Error parsing JWT:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var reqBody subscribe.RequestDto
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing request body: %v", err)

		}
	}(r.Body)

	if err := h.subscriptionRepo.CreateSubscription(claims.UserID, reqBody.RelatedUserID); err != nil {
		log.Println("Error creating subscription:", err)
		http.Error(w, "Error creating subscription", http.StatusInternalServerError)
		return
	}

	log.Println("Subscription created successfully")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("Subscription created successfully"))
	if err != nil {
		log.Printf("Error writing response: %v", err)

		return
	}
}

// GetAvailableUsers /api/available
func (h *Handler) GetAvailableUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling get available users request")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseJWT(authHeader, h.JWTSecretKey)
	if err != nil {
		log.Println("Error parsing JWT:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.UserID == 0 {
		log.Println("Invalid user ID in token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := h.userRepo.GetAvailableUsersForSubscription(claims.UserID)
	if err != nil {
		log.Println("Error fetching available users:", err)
		http.Error(w, "Error fetching available users", http.StatusInternalServerError)
		return
	}

	availableUsers := make([]available_user.ResponseDto, 0, len(users))
	for _, user := range users {
		availableUsers = append(availableUsers, available_user.ResponseDto{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			DateOfBirth: user.DateOfBirth,
		})
	}

	response, err := json.Marshal(availableUsers)
	if err != nil {
		log.Println("Error marshalling response:", err)
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		log.Printf("Error writing response: %v", err)

		return
	}
}

// Unsubscribe /api/unsubscribe
func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling unsubscribe request")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseJWT(authHeader, h.JWTSecretKey)
	if err != nil {
		log.Println("Error parsing JWT:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.UserID == 0 {
		log.Println("Invalid user ID in token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var reqBody subscribe.RequestDto
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing request body: %v", err)

		}
	}(r.Body)

	err = h.subscriptionRepo.UnsubscribeUser(claims.UserID, reqBody.RelatedUserID)
	if err != nil {
		if err.Error() == "subscription does not exist" {
			http.Error(w, "You are not subscribed to this user", http.StatusBadRequest)
		} else {
			log.Println("Error unsubscribing user:", err)
			http.Error(w, "Error unsubscribing user", http.StatusInternalServerError)
		}
		return
	}

	log.Println("Unsubscribed successfully")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Unsubscribed successfully"))
	if err != nil {
		log.Printf("Error writing response: %v", err)

		return
	}
}
