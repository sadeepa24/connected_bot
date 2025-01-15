package common

import (
	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/controller"

	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"go.uber.org/zap"
)

type Sendreciver func(msg any) (*tgbotapi.Message, error)
type Callbackreciver func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error)
type Alertsender func(msg string)

type OptionExcutors struct {
	//Common
	Callbackreciver Callbackreciver
	Sendreciver     Sendreciver
	Alertsender     Alertsender
	Upx             *update.Updatectx
	Btns            *botapi.Buttons
	Usersession     *controller.CtrlSession
	MessageSession  *botapi.Msgsession
	Ctrl            *controller.Controller
	Logger 			*zap.Logger

	//For Exec Rule addr
}

type Initer interface {
	Init() 
}
