package routes

import (
	"birthdayReminder/app/internal/handler"
	"birthdayReminder/app/internal/repository"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
)

func InitRoutes(router *mux.Router, db *pgxpool.Pool) {
	r := repository.New(db)
	h := handler.New(r)
	router.HandleFunc("/api/registration", h.Register).Methods("POST")
	router.HandleFunc("/api/login", h.Login).Methods("POST")
	router.HandleFunc("/api/subscribe", h.Subscribe).Methods("POST")
	router.HandleFunc("/api/available", h.GetAvailableUsers).Methods("GET")
	router.HandleFunc("/api/unsubscribe", h.Unsubscribe).Methods("POST")
}
