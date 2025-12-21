package logger

import (
	"strconv"

	"github.com/go-telegram/bot/models"
)

const colorBrightCyan = "\033[96m"

type TgLogger struct {
	chatID int64
}

func ForUpdate(u *models.Update) TgLogger {
	var chatID int64
	if u != nil && u.Message != nil {
		chatID = u.Message.Chat.ID
	}
	return TgLogger{chatID: chatID}
}

func ForUser(userID int64) TgLogger {
	return TgLogger{chatID: userID}
}

func (l TgLogger) prefix() string {
	if l.chatID == 0 {
		return ""
	}
	return colorBrightCyan + "[" + strconv.FormatInt(l.chatID, 10) + "]" + colorReset
}

func (l TgLogger) Infof(format string, args ...any) {
	Infof(l.prefix()+" "+format, args...)
}

func (l TgLogger) Debugf(format string, args ...any) {
	Debugf(l.prefix()+" "+format, args...)
}

func (l TgLogger) Warnf(format string, args ...any) {
	Warnf(l.prefix()+" "+format, args...)
}

func (l TgLogger) Errorf(format string, args ...any) {
	Errorf(l.prefix()+" "+format, args...)
}
