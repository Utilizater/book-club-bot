package statemachine

import (
	"log"
	"telegram-bot/statefunctions/removeuser"
	"telegram-bot/statefunctions/setbook"
	"telegram-bot/statefunctions/setprogress"
	"telegram-bot/statefunctions/setuser"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type FuncType func(string, string, *tgbotapi.BotAPI, tgbotapi.Update)

var FuncMap = map[string]FuncType{
	"enter_page":               SetProgress,
	"enter_percent":            SetProgress,
	"enter_book_type":          SetProgress,
	"enter_total_pages":        SetProgress,
	"enter_book_name":          SetBook,
	"enter_author":             SetBook,
	"enter_finishing_date":     SetBook,
	"enter_nickname":           AddUser,
	"enter_username":           AddUser,
	"enter_name":               AddUser,
	"enter_nickname_to_remove": RemoveUser,
}

func SetProgress(user string, userStatus string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch userStatus {
	case "":
		setprogress.SetProgressDefault(user, bot, update)
	case "enter_page":
		setprogress.EnterPage(user, bot, update)
	case "enter_percent":
		setprogress.EnterPercent(user, bot, update)
	case "enter_book_type":
		setprogress.EnterBookType(user, bot, update)
	case "enter_total_pages":
		setprogress.EnterTotalPages(user, bot, update)
	default:
		log.Fatal("There is no status - " + userStatus)
	}
}

func SetBook(user string, userStatus string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch userStatus {
	case "":
		setbook.SetBookDefault(user, bot, update)
	case "enter_book_name":
		setbook.EnterBookName(user, bot, update)
	case "enter_author":
		setbook.EnterAuthor(user, bot, update)
	case "enter_finishing_date":
		setbook.EnterFinishingDate(user, bot, update)
	default:
		log.Fatal("There is no status - " + userStatus)
	}
}

func AddUser(user string, userStatus string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch userStatus {
	case "":
		setuser.SetUserDefault(user, bot, update)
	case "enter_username":
		setuser.EnterUserName(user, bot, update)
	case "enter_nickname":
		setuser.EnterUserNickName(user, bot, update)
	default:
		log.Fatal("There is no status - " + userStatus)
	}
}

func RemoveUser(user string, userStatus string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch userStatus {
	case "":
		removeuser.RemoveUserDefault(user, bot, update)
	case "enter_nickname_to_remove":
		removeuser.RemoveUser(user, bot, update)
	default:
		log.Fatal("There is no status - " + userStatus)
	}
}

func UpdateMeetingDate(user string, userStatus string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch userStatus {
	case "":
		setbook.UpdateBookDateDefault(user, bot, update)
	}
}
