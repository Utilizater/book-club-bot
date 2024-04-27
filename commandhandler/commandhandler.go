package commandhandler

import (
	"fmt"
	"log"
	"strconv"
	"telegram-bot/database"
	"telegram-bot/statemachine"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, username string) {
	log.Printf("Received message: %s", update.Message.Text)
	log.Printf("Command: %s", update.Message.Command())
	log.Printf("Arguments: %s", update.Message.CommandArguments())

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	userStatus := database.UserStatus(username)
	fmt.Println("User status: ", userStatus)
	if userStatus != "" {
		statemachine.FuncMap[userStatus](username, userStatus, bot, update)
		return
	}

	isUserAdmin := database.IsUserAdmin(username)

	switch update.Message.Command() {
	case "help":
		msg.Text = help(isUserAdmin)
	case "addBook":
		statemachine.SetBook(username, "", bot, update)
		return
	case "getUserList":
		msg.Text = getUserList()
	case "setProgress":
		statemachine.SetProgress(username, "", bot, update)
		return
	case "getCurrentBook":
		msg.Text = getCurrentBook()
	case "getGroupProgress":
		msg.Text = getGroupProgress()
	// case "removeBook":
	// 	msg.Text = removeBook(update.Message.CommandArguments(), isUserAdmin)
	case "addUser":
		statemachine.AddUser(username, "", bot, update)
		return
	case "removeUser":
		statemachine.RemoveUser(username, "", bot, update)
		return
	case "getBookList":
		msg.Text = bookList()
	default:
		msg.Text = "I don't recognize that command. Use /help to see the list of commands."
	}

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func help(isUserAdmin bool) string {
	applicationVersion := "0.3"
	if isUserAdmin {
		return "Here are the commands you can use: \n/help\n/addBook\n/getUserList\n/setProgress\n/getCurrentBook\n/getGroupProgress\n/addUser\n/removeUser\n/getBookList\n\n applicationVersion: " + applicationVersion
	}
	return "Here are the commands you can use: \n/help\n/setProgress\n/getCurrentBook\n/getGroupProgress"
}

func getUserList() string {
	userList := database.UserList()
	usersText := "\n"
	for _, user := range userList {
		usersText += user.UserName + " : " + user.FullName + "\n"
	}

	return "Here is the list of users: " + usersText
}

func getCurrentBook() string {
	book := database.GetCurrentBook()
	return "The current book is: " + book.Title + " by " + book.Author + "(id - " + book.BookID + ")"
}

func getGroupProgress() string {
	return database.GroupProgress()
}

// func removeBook(BookID string, isUserAdmin bool) string {
// 	if !isUserAdmin {
// 		return "You are not authorized to remove a book."
// 	}
// 	database.RemoveBook(BookID)
// 	return "Done"
// }

func bookList() string {
	bookList := database.BookList()
	booksText := "\n"
	for _, book := range bookList {
		booksText += book.Title + " by " + book.Author + "(id - " + book.BookID + "; active - " + strconv.FormatBool(book.Active) + ")\n"
	}

	return "Here is the list of books: " + booksText
}
