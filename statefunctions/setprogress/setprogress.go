package setprogress

import (
	"log"
	"strconv"
	"strings"
	"telegram-bot/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func EnterTotalPages(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	totalPages, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a number."))

	}
	if totalPages <= 0 {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a number greater than 0."))

	}
	currentBook := database.GetCurrentBook()
	bookId := currentBook.BookID
	database.SetProgress(database.ReadingProgress{BookID: bookId, UserName: user, Type: database.RegularBook, TotalPages: totalPages})
	database.SetUserStatus(user, "enter_page")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter the page you are currently reading:"))
}

func EnterPage(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	page, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a number."))
		return
	}
	if page <= 0 {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a number greater than 0."))
		return
	}
	userProgress := database.UserProgress(user)
	totalPages := userProgress.TotalPages
	if page > totalPages {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a number less than total pages - "+strconv.Itoa(totalPages)))

	}
	currentBook := database.GetCurrentBook()
	bookId := currentBook.BookID
	progress := float64(page) / float64(totalPages) * 100
	database.SetProgress(database.ReadingProgress{BookID: bookId, UserName: user, Type: database.RegularBook, PageNumber: page, Progress: int(progress), TotalPages: totalPages})
	database.SetUserStatus(user, "")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Thank you!"))
}

func EnterPercent(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	percent, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a number."))
		return
	}
	if percent < 0 || percent > 100 {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a number between 0 and 100."))
		return
	}
	currentBook := database.GetCurrentBook()
	bookId := currentBook.BookID
	database.SetProgress(database.ReadingProgress{BookID: bookId, UserName: user, Type: database.AudioBook, Progress: percent})
	database.SetUserStatus(user, "")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Thank you!"))
}

func EnterBookType(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	message := strings.ToLower(update.Message.Text)
	if strings.Contains(message, "regular") {
		currentBook := database.GetCurrentBook()
		bookId := currentBook.BookID
		database.SetProgress(database.ReadingProgress{BookID: bookId, UserName: user, Type: database.RegularBook})
		database.SetUserStatus(user, "enter_total_pages")
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter total pages of the book:"))
		return
	}
	if strings.Contains(message, "audio") {
		currentBook := database.GetCurrentBook()
		bookId := currentBook.BookID
		database.SetProgress(database.ReadingProgress{BookID: bookId, UserName: user, Type: database.AudioBook})
		database.SetUserStatus(user, "enter_percent")
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter percent of your listening:"))
		return
	}
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, I didn't understand you. Please select the book type - audio or regular:"))
}

func SetProgressDefault(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userProgress := database.UserProgress(user)
	if userProgress == nil {
		database.SetUserStatus(user, "enter_book_type")
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select the book's type (audio or regular):")

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Regular Book"),
				tgbotapi.NewKeyboardButton("Audio Book"),
			),
		)
		keyboard.OneTimeKeyboard = true // Make keyboard disappear after use

		msg.ReplyMarkup = keyboard

		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %s", err)
		}
		return
	}
	if userProgress.Type == database.RegularBook {
		database.SetUserStatus(user, "enter_page")
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter the page you are currently reading:"))
	}
	if userProgress.Type == database.AudioBook {
		database.SetUserStatus(user, "enter_percent")
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter percent of your listening:"))
	}
}
