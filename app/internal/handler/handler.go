package handler

import (
	"birthdayReminder/app/internal/auth"
	"birthdayReminder/app/internal/config"
	"birthdayReminder/app/internal/models"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

type Repository interface {
	CreateUser(user *models.User, hashedPassword []byte) error
	GetUserByEmail(email string) (*models.User, error)
	CreateSubscription(userID int, relatedUserID int) error
	GetAvailableUsersForSubscription(userID int) ([]models.User, error)
	UnsubscribeUser(userID int, relatedUserID int) error
}

type Handler struct {
	config *config.Config
	repo   Repository
}

func New(repo Repository) *Handler {
	return &Handler{config: &config.Config{JWTSecretKey: "kjghfdf"}, repo: repo}
}

// /api/registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if user.Name == "" || user.Email == "" || user.Password == "" || user.DateOfBirth.IsZero() {
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Вызываем метод для выполнения запроса к базе данных
	err = h.repo.CreateUser(&user, hashedPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving user to database: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

// /api/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var reqBody models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	dbUser, err := h.repo.GetUserByEmail(reqBody.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(reqBody.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := auth.GenerateJWT(dbUser.ID, h.config.JWTSecretKey)
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
	w.Write([]byte("Login successful"))

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
	w.Write([]byte("Login successful"))
}

// /api/subscribe
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling subscription request")

	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseJWT(tokenStr, h.config.JWTSecretKey)
	if err != nil {
		log.Println("Error parsing JWT:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var reqBody models.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := h.repo.CreateSubscription(claims.UserID, reqBody.RelatedUserID); err != nil {
		log.Println("Error creating subscription:", err)
		http.Error(w, "Error creating subscription", http.StatusInternalServerError)
		return
	}

	log.Println("Subscription created successfully")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Subscription created successfully"))
}

// /api/available
func (h *Handler) GetAvailableUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling get available users request")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseJWT(authHeader, h.config.JWTSecretKey)
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

	users, err := h.repo.GetAvailableUsersForSubscription(claims.UserID)
	if err != nil {
		log.Println("Error fetching available users:", err)
		http.Error(w, "Error fetching available users", http.StatusInternalServerError)
		return
	}

	availableUsers := make([]models.AvailableUserResponse, 0, len(users))
	for _, user := range users {
		availableUsers = append(availableUsers, models.AvailableUserResponse{
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
	w.Write(response)
}

// /api/unsubscribe
func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling unsubscribe request")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseJWT(authHeader, h.config.JWTSecretKey)
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

	var reqBody models.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = h.repo.UnsubscribeUser(claims.UserID, reqBody.RelatedUserID)
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
	w.Write([]byte("Unsubscribed successfully"))
}
