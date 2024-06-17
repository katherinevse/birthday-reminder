package handler

import (
	"birthdayReminder/app/internal/repository/user"
)

type Repository interface {
	CreateUser(user *user.User, hashedPassword []byte) error
	GetUserByEmail(email string) (*user.User, error)
	CreateSubscription(userID int, relatedUserID int) error
	GetAvailableUsersForSubscription(userID int) ([]user.User, error)
	UnsubscribeUser(userID int, relatedUserID int) error
	GetUsersWithBirthdayTomorrow() ([]user.User, error)
	GetSubscribers(userID int) ([]user.User, error)
}
