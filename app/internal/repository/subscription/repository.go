package subscription

import (
	"context"
	"errors"
)

type Repo struct {
	db DBPool
}

func NewRepo(db DBPool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateSubscription(userID int, relatedUserID int) error {
	//проверяем, существует ли подписка
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
