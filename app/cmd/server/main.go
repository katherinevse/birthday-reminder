package main

import (
	"birthdayReminder/app/internal/notifications"
	"birthdayReminder/app/internal/routes"
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

	notifications.StartBirthdayNotifier(pool)

	router := mux.NewRouter()
	routes.InitRoutes(router, pool)

	port := ":8080"
	fmt.Println("Server is running on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
