package setbook

import (
	"regexp"
	"telegram-bot/database"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetBookDefault(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	database.SetUserStatus(user, "enter_book_name")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter the name of the book:"))
}

func EnterBookName(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	bookName := update.Message.Text
	database.AddBook(bookName)
	database.SetUserStatus(user, "enter_author")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter the author of the book:"))
}

func EnterAuthor(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	author := update.Message.Text
	currentBook := database.GetCurrentBook()
	database.UpdateBookAuthor(currentBook.BookID, author)
	database.SetUserStatus(user, "enter_finishing_date")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter date of club's meeting. Format 'dd.mm.yyyy'"))
}

func EnterFinishingDate(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	date := update.Message.Text
	re := regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)

	if re.MatchString(date) {
		// Parse the date to check if it's valid
		parsedDate, err := time.Parse("02.01.2006", date)
		if err != nil {
			// If the date is invalid, ask the user to input it again
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid date format. Please enter the date in format dd.mm.yyyy:"))
			return
		}

		// Check if the date is later than today
		currentDate := time.Now()
		if parsedDate.Before(currentDate) || parsedDate.Equal(currentDate) {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "The date must be later than today. Please enter a valid later date in format dd.mm.yyyy:"))
			return
		}

		// If the date is valid and later than today, store it and thank the user
		currentBook := database.GetCurrentBook()
		database.UpdateBookDate(currentBook.BookID, date)
		database.SetUserStatus(user, "")
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Thank you!"))
	} else {
		// If the format is incorrect, ask the user to input it again
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid date format. Please enter the date in format dd.mm.yyyy:"))
	}
}

func UpdateBookDateDefault(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	database.SetUserStatus(user, "enter_finishing_date")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter date of club's meeting. Format 'dd.mm.yyyy'"))
}
