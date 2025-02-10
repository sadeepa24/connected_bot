package service

import (
	"context"
	"strconv"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)

type InlineService struct {
	ctx      context.Context
	logger   *zap.Logger
	ctrl     *controller.Controller
	botapi   botapi.BotAPI
}

func NewInline(
	ctx context.Context,
	logger *zap.Logger,
	botapi botapi.BotAPI,
	ctrl *controller.Controller,
) *InlineService {
	return &InlineService{
		ctx:      ctx,
		botapi:   botapi,
		ctrl:     ctrl,
		logger:   logger,
	}

}

func (a *InlineService) Exec(upx *update.Updatectx) error {
	if upx.Update.InlineQuery == nil {
		return nil
	}

	if upx.Update.InlineQuery.Query != "" {
		return nil
	}
	quary := upx.Update.InlineQuery

	//var sendquary io.Reader 
	answere := tgbotapi.AnswerInlineQuery{
		InlineQueryId: quary.ID,
	}

	a.ctrl.GetInlinePost()


	posts := a.ctrl.GetInlinePost()
	btns := botapi.NewButtons([]int16{2, 1})

	btns.AddUrlbutton("Channel", a.ctrl.Channelink)
	btns.AddUrlbutton("Group", a.ctrl.GroupLink)
	btns.AddUrlbutton("Bot", a.ctrl.Botlink)

	for i, post := range posts {
		message, err := a.botapi.GetMgStore().GetMessage(post, "en", struct{}{})
		if err != nil {
			continue
		}
		if message.Includemed {
			switch message.MedType {
				case constbot.MedPhoto:
					answere.Results = append(answere.Results, tgbotapi.InlineQueryResultCachedPhoto{
						ParseMode: message.ParseMode,
						Caption: message.Msg,
						ID: strconv.Itoa(i),
						Type: "photo",
						ReplyMarkup: struct{
							Keyboard [][]botapi.InlineKeyboardButton `json:"inline_keyboard,omitempty"`
						}{
							Keyboard: btns.Getkeyboard().Inline_keyboard,
						},
						PhotoID: message.MediaId,
					})
					
				case constbot.MedVideo:
					answere.Results = append(answere.Results, tgbotapi.InlineQueryResultCachedVideo{
						ParseMode: message.ParseMode,
						Caption: message.Msg,
						Type: "video",
						ReplyMarkup: struct{
							Keyboard [][]botapi.InlineKeyboardButton `json:"inline_keyboard,omitempty"`
						}{
							Keyboard: btns.Getkeyboard().Inline_keyboard,
						},
					})
			}
		} else {
		}
	}

	a.botapi.Makerequest(upx.Ctx, "POST", constbot.ApiMethodAnswereInline, &answere)

	return nil
}


func (a *InlineService) Name() string {
	return constbot.InlineServiceName
}







func (a *InlineService) Init() error {
	return nil
}

func (a *InlineService) Canhandle(upx *update.Updatectx) (bool, error) {
	return upx.Service == constbot.InlineServiceName, nil
}