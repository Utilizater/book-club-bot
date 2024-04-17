package commandhandler

import (
	"fmt"
	"log"
	"strconv"
	"telegram-bot/database"
	"telegram-bot/statemachine"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// var chatStates = make(map[int64]string)

func HandleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, username string) {
	log.Printf("Received message: %s", update.Message.Text)
	log.Printf("Command: %s", update.Message.Command())
	log.Printf("Arguments: %s", update.Message.CommandArguments())

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	userStatus := database.UserStatus(username)

	if userStatus != "" {
		msg.Text = statemachine.FuncMap[userStatus](username, userStatus, bot, update)
		return
	}

	isUserAdmin := database.IsUserAdmin(username)

	switch update.Message.Command() {
	case "help":
		msg.Text = help(isUserAdmin)
	case "addBook":
		msg.Text = addBook(update, isUserAdmin)
	case "getUserList":
		msg.Text = getUserList()
	case "setProgress":
		setProgress(update, username, bot)
		return
	case "getCurrentBook":
		msg.Text = getCurrentBook()
	case "getGroupProgress":
		msg.Text = getGroupProgress()
	case "removeBook":
		msg.Text = removeBook(update.Message.CommandArguments(), isUserAdmin)
	case "addUser":
		msg.Text = addUser(update, isUserAdmin)
	case "removeUser":
		msg.Text = removeUser(update, isUserAdmin)
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
	applicationVersion := "0.1"
	if isUserAdmin {
		return "Here are the commands you can use: \n/help\n/addBook\n/getUserList\n/setProgress\n/getCurrentBook\n/getGroupProgress\n/removeBook\n/addUser\n/removeUser\n/getBookList\n\n applicationVersion: " + applicationVersion
	}
	return "Here are the commands you can use: \n/help\n/setProgress\n/getCurrentBook\n"
}

func addBook(update tgbotapi.Update, isUserAdmin bool) string {
	if !isUserAdmin {
		return "You are not authorized to add a book."
	}
	args := utils.NormalizeQuotes(update.Message.CommandArguments())
	var title, author string
	_, err := fmt.Sscanf(args, "title: %q, author: %q", &title, &author)

	if err != nil {
		return "Failed to parse input. Please use the format: /addBook title: \"The Great Gatsby\", author: \"F. Scott Fitzgerald\""
	}

	database.AddBook(title, author)

	return "Book added successfully: " + title + " by " + author
}

func getUserList() string {
	userList := database.UserList()
	usersText := "\n"
	for _, user := range userList {
		usersText += user.UserName + " : " + user.Name + "\n"
	}

	return "Here is the list of users: " + usersText
}

func setProgress(update tgbotapi.Update, username string, bot *tgbotapi.BotAPI) {
	statemachine.SetProgress(username, "", bot, update)
}

func getCurrentBook() string {
	book := database.GetCurrentBook()
	return "The current book is: " + book.Title + " by " + book.Author + "(id - " + book.BookID + ")"
}

func getGroupProgress() string {
	return database.GroupProgress()
}

func removeBook(BookID string, isUserAdmin bool) string {
	if !isUserAdmin {
		return "You are not authorized to remove a book."
	}
	database.RemoveBook(BookID)
	return "Done"
}

func addUser(update tgbotapi.Update, isUserAdmin bool) string {
	if !isUserAdmin {
		return "You are not authorized to add a user."
	}
	args := utils.NormalizeQuotes(update.Message.CommandArguments())
	var userName, name string
	_, err := fmt.Sscanf(args, "nickname: %q, name: %q", &userName, &name)

	if err != nil {
		return "Failed to parse input. Please use the format: /addUser nickname: \"alexeygav\", name: \"Alexey Gavrilov\""
	}

	database.AddUser(userName, name)
	return "Done"
}

func removeUser(update tgbotapi.Update, isUserAdmin bool) string {
	if !isUserAdmin {
		return "You are not authorized to remove a user."
	}
	args := utils.NormalizeQuotes(update.Message.CommandArguments())
	var userName string
	_, err := fmt.Sscanf(args, "%d", &userName)

	if err != nil {
		return "Failed to parse input. Please use the format: /removeUser alexeygav"
	}

	database.RemoveUser(userName)
	return "Done"
}

func bookList() string {
	bookList := database.BookList()
	booksText := "\n"
	for _, book := range bookList {
		booksText += book.Title + " by " + book.Author + "(id - " + book.BookID + "; active - " + strconv.FormatBool(book.Active) + ")\n"
	}

	return "Here is the list of books: " + booksText
}
