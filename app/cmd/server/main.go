package main

import (
	"birthdayReminder/internal/handler"
	"birthdayReminder/internal/handler/auth"
	"birthdayReminder/internal/notifier"
	"birthdayReminder/internal/repository/subscription"
	"birthdayReminder/internal/repository/user"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbLogin := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbLogin, dbPassword, dbHost, dbPort, dbName)

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer pool.Close()

	userRepo := user.NewRepo(pool)
	subscriptionRepo := subscription.NewRepo(pool)
	tokenManager := &auth.TokenService{}

	notify := notifier.New(userRepo, subscriptionRepo)
	//notify.StartBirthdayNotifier()
	notify.SendBirthdayNotifications()

	router := mux.NewRouter()
	handler.InitRoutes(router, userRepo, subscriptionRepo, tokenManager)

	port := ":8080"
	fmt.Println("Server is running on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
