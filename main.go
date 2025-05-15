package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type EmailMessage struct {
	Subject string `json:"subject"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type ResponseMessage struct {
	Message string `json:"message"`
}

type SMTPServer struct {
	Host     string
	Port     string
	Username string
	Password string
}

func SendMail(message EmailMessage, server SMTPServer) error {
	body := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n"+
		"<html><body>"+
		"<p><b>Name:</b> %s</p>"+
		"<p><b>Email:</b> %s</p>"+
		"<p><b>Message:</b></p>"+
		"<p>%s</p>"+
		"</body></html>",
		server.Username, message.Email, message.Subject, message.Name, message.Email, message.Message)

	auth := smtp.PlainAuth("", server.Username, server.Password, server.Host)
	addr := fmt.Sprintf("%s:%s", server.Host, server.Port)

	err := smtp.SendMail(addr, auth, server.Username, []string{message.Email}, []byte(body))
	if err != nil {
		return err
	}

	return nil
}

func Retry(message EmailMessage, server SMTPServer, attempts int, interval time.Duration) error {
	for i := 0; i < attempts; i++ {
		err := SendMail(message, server)
		if err == nil {
			return nil
		}
		fmt.Printf("Error sending email (attempt %d/%d): %v\n", i+1, attempts, err)
		time.Sleep(interval)
	}
	return fmt.Errorf("failed after %d attempts", attempts)
}

func MailListener(listenerId int, mailChannel chan EmailMessage, mailServer SMTPServer) {
	for message := range mailChannel {
		err := Retry(message, mailServer, 3, time.Second)
		if err != nil {
			fmt.Println(listenerId, "- Failed to send email after retries:", err)
			return
		}
		fmt.Println(listenerId, "- Email sent successfully!")
	}
}

func CreateMailHandler(mailChannel chan EmailMessage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var message EmailMessage
		err := json.NewDecoder(req.Body).Decode(&message)
		if err != nil {
			http.Error(res, "Error parsing request body", http.StatusBadRequest)
			return
		}
		mailChannel <- message

		response := ResponseMessage{
			Message: "Mail added to mail queue",
		}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		res.WriteHeader(http.StatusOK)
		res.Write(jsonResponse)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	mailChannel := make(chan EmailMessage)
	mailServer := SMTPServer{
		Host:     os.Getenv("MAILER_HOST"),
		Port:     os.Getenv("MAILER_PORT"),
		Username: os.Getenv("MAILER_USER"),
		Password: os.Getenv("MAILER_PASS"),
	}

	var workers int = 10
	if workersEnv := os.Getenv("WORKERS"); workersEnv != "" {
		parsedWorkers, err := strconv.ParseInt(workersEnv, 10, 64)
		if err == nil {
			workers = int(parsedWorkers)
		}
	}

	for i := range workers {
		go MailListener(i, mailChannel, mailServer)
	}

	http.HandleFunc("/", CreateMailHandler(mailChannel))

	port := 3000
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		parsedPort, err := strconv.ParseInt(portEnv, 10, 64)
		if err == nil {
			port = int(parsedPort)
		}
	}

	fmt.Println("Server running on port", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Start server error:", err)
	}
}
