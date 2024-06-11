package main

import (
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

	dbpool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer dbpool.Close()

	router := mux.NewRouter()
	routes.InitRoutes(router, dbpool)

	port := ":8080"
	fmt.Println("Server is running on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
