package main

import (
	"birthdayReminder/app/internal/handler"
	"birthdayReminder/app/internal/handler/auth"
	"birthdayReminder/app/internal/notifier"
	"birthdayReminder/app/internal/repository/subscription"
	"birthdayReminder/app/internal/repository/user"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/birthdayReminder"

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer pool.Close()

	userRepo := user.NewRepo(pool)
	subscriptionRepo := subscription.NewRepo(pool)
	tokenManager := &auth.TokenService{}

	notify := notifier.New(userRepo, subscriptionRepo)
	notify.StartBirthdayNotifier()

	router := mux.NewRouter()
	handler.InitRoutes(router, userRepo, subscriptionRepo, tokenManager)

	port := ":8080"
	fmt.Println("Server is running on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
