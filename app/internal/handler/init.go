package handler

import (
	"birthdayReminder/app/internal/handler/auth"
	"github.com/gorilla/mux"
)

func InitRoutes(router *mux.Router, userRepo UserRepository, subscriptionRepo SubscriptionRepository, tokenManager auth.TokenManager) {
	h := New(userRepo, subscriptionRepo, tokenManager)
	router.HandleFunc("/api/registration", h.Register).Methods("POST")
	router.HandleFunc("/api/login", h.Login).Methods("POST")
	router.HandleFunc("/api/subscribe", h.Subscribe).Methods("POST")
	router.HandleFunc("/api/available", h.GetAvailableUsers).Methods("GET")
	router.HandleFunc("/api/unsubscribe", h.Unsubscribe).Methods("POST")
}
