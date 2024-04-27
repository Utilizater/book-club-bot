package removeuser

import (
	"telegram-bot/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func RemoveUserDefault(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	database.SetUserStatus(user, "enter_nickname_to_remove")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Enter user telegram nick name:"))
}

func RemoveUser(user string, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userNickName := update.Message.Text
	database.RemoveUser(userNickName)
	database.SetUserStatus(user, "")
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "User removed successfully!"))
}
