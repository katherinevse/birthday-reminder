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
