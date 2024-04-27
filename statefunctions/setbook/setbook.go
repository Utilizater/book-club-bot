package setbook

import (
	"telegram-bot/database"

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
	database.SetUserStatus(user, "")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Thank you!"))
}
