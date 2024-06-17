package notifier

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"log"
	"net/smtp"
	"os"
	"time"
)

type Notifier struct {
	userRepo         UserRepository
	subscriptionRepo SubscriptionRepository
}

func New(userRepo UserRepository, subscriptionRepo SubscriptionRepository) Notifier {
	return Notifier{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

func (n *Notifier) StartBirthdayNotifier() {
	log.Println("Initializing the scheduler")
	s := gocron.NewScheduler(time.UTC)

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return
	}

	s.ChangeLocation(loc)

	_, err = s.Every(1).Day().At("12:15").Do(n.sendBirthdayNotifications)
	if err != nil {
		fmt.Println(err)
	}

	s.StartAsync()
}

func (n *Notifier) sendBirthdayNotifications() {
	users, err := n.userRepo.GetUsersWithBirthdayTomorrow()
	if err != nil {
		log.Println("Error fetching users with birthday tomorrow:", err)
		return
	}

	if len(users) == 0 {
		log.Println("No users found with birthday tomorrow.")
		return
	}

	for _, user := range users {
		subscribers, err := n.userRepo.GetSubscribers(user.ID)
		if err != nil {
			log.Println("Error fetching subscribers for user ID:", user.ID, "-", err)
			continue
		}

		if len(subscribers) == 0 {
			log.Printf("No subscribers found for user %s (ID: %d)\n", user.Name, user.ID)
			continue
		}

		subject := "Happy Birthday Notification"
		message := "Завтра день рождения у " + user.Name + "!" +
			"Не забудьте поздравить!"
		for _, subscriber := range subscribers {
			if err := n.sendMessage(subscriber.Email, subject, message); err != nil {
				log.Println("Error sending email to", subscriber.Email, ":", err)
				continue
			}
			log.Printf("Sent birthday notification to %s for user %s (ID: %d)\n", subscriber.Email, user.Name, user.ID)
		}
	}
}

func (n *Notifier) sendMessage(email, subject, message string) error {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	smtpHost := "smtp.mail.ru"
	smtpPort := "587"
	from := os.Getenv("SMTP_USERNAME")
	Password := os.Getenv("SMTP_PASSWORD")

	//авторизация
	auth := smtp.PlainAuth("", from, Password, smtpHost)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", email, subject, message))

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{email}, msg)
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}
	return nil

}
