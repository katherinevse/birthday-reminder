package repository

import (
	"birthdayReminder/app/internal/models"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

//TODO struct создать

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type Repo struct {
	db DBPool
}

func New(db DBPool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateUser(user *models.User, hashedPassword []byte) error {
	query := `INSERT INTO users (name, email, password, date_of_birth) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(context.Background(), query, user.Name, user.Email, hashedPassword, user.DateOfBirth)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, name, email, password, date_of_birth FROM users WHERE email=$1`
	err := r.db.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.DateOfBirth)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repo) CreateSubscription(userID int, relatedUserID int) error {
	// Проверяем, существует ли уже подписка
	queryCheck := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE user_id=$1 AND related_user_id=$2)`
	var exists bool
	err := r.db.QueryRow(context.Background(), queryCheck, userID, relatedUserID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("subscription already exists")
	}

	// Если подписка не существует, создаем новую
	queryInsert := `INSERT INTO subscriptions (user_id, related_user_id) VALUES ($1, $2)`
	_, err = r.db.Exec(context.Background(), queryInsert, userID, relatedUserID)
	return err
}

func (r *Repo) GetAvailableUsersForSubscription(userID int) ([]models.User, error) {
	query := `
		SELECT id, name, email, date_of_birth
		FROM users
		WHERE id != $1
		AND id NOT IN (
			SELECT related_user_id
			FROM subscriptions
			WHERE user_id = $1
		)
	`
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.DateOfBirth); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repo) UnsubscribeUser(userID int, relatedUserID int) error {
	// Проверяем, существует ли подписка
	queryCheck := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE user_id=$1 AND related_user_id=$2)`
	var exists bool
	err := r.db.QueryRow(context.Background(), queryCheck, userID, relatedUserID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("subscription does not exist")
	}

	// Удаляем подписку только если она существует
	queryDelete := `DELETE FROM subscriptions WHERE user_id=$1 AND related_user_id=$2`
	_, err = r.db.Exec(context.Background(), queryDelete, userID, relatedUserID)
	if err != nil {
		return err
	}

	return nil
}

//user := GetUserByUsername()
//return models.RegisterResponse{
//	password: user.Password,
//}
