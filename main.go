package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"telegram-bot/commandhandler"
	"telegram-bot/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Telegram bot is running!")
		})

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		fmt.Println("Starting HTTP server on port", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed  to start HTTP server: %v", err)
		}
	}()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	fmt.Println("TELEGRAM_TOKEN:", os.Getenv("TELEGRAM_TOKEN"))

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			// log.Println("update.Message.Chat.ID!", update.Message.Chat.ID)
			username := update.Message.From.UserName
			isUserBelongsToClub := database.IsUserBelongsToClub(username)
			if !isUserBelongsToClub {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not a member of the club. Please contact @alexeygav to join the club.")
				bot.Send(msg)
				continue
			}

			commandhandler.HandleCommand(bot, update, username)
		}
	}
}
