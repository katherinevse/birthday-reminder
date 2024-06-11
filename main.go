package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetDefault("PORT", ":	9999")
	viper.SetDefault("DB_HOST", "10.0.100.7")
	viper.SetDefault("DB_USER", "tg_bot")
	viper.SetDefault("DB_PASSWORD", "0GlFBsXh97XOWvqtdWQBggOn7jSUpwdZ")
	viper.SetDefault("DB_NAME", "test_seller_db")
	viper.SetDefault("DB_PORT", "5432")
}

func main() {
	r := chi.NewRouter()
}
