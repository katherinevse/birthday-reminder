package available_user

import "time"

type ResponseDto struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	DateOfBirth time.Time `json:"date_of_birth"`
}
