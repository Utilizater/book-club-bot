package setuser

import (
	"telegram-bot/database"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetUserDefault(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	database.SetUserStatus(user, "enter_nickname")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter user telegram nick name:"))
}

func EnterUserNickName(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	nickName := update.Message.Text
	if !utils.IsValidTelegramNickname(nickName) {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a valid nickname."))
		return
	}
	database.AddUser(nickName, "")
	database.SetUserStatus(user, "enter_username")

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter full user name:"))
}

func EnterUserName(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userName := update.Message.Text
	database.SetUserFullName(userName)
	database.SetUserStatus(user, "")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Thank you!"))
}
