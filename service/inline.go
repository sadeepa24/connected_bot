package service

import (
	"context"
	"errors"
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
		return errors.New("no inline quary found")
	}
	quary := upx.Update.InlineQuery
	//var sendquary io.Reader 
	answere := &tgbotapi.AnswerInlineQuery{
		InlineQueryId: quary.ID,
		Is_personal: true,
		ChacheTIme: 3600,
	}

	posts := a.ctrl.GetInlinePost()
	if len(posts) == 0 {
		return nil
	}
	btns := botapi.NewButtons([]int16{2, 1})
	btns.AddUrlbutton("Channel", a.ctrl.Channelink)
	btns.AddUrlbutton("Group", a.ctrl.GroupLink)
	btns.AddUrlbutton("Bot", a.ctrl.Botlink)
	btns.SetOveride()

	for i, post := range posts {
		message, err := a.botapi.GetMgStore().GetMessage(post.Name, "en", struct{}{})
		if err != nil {
			continue
		}
		if message.Keyboard != nil {
			btns.OverideKeyboard(message.Keyboard)
		}

		if message.Includemed {
			switch message.MedType {
				case constbot.MedPhoto:
					answere.Results = append(answere.Results, tgbotapi.InlineQueryResultCachedPhoto{
						ParseMode: message.ParseMode,
						Caption: message.Msg,
						ID: strconv.Itoa(i),
						Type: "photo",
						Description: "post.Dis",
						Title: "post Title",
						ReplyMarkup: btns.GetkeyboardCopy(),
						PhotoID: message.MediaId,
					})	
				case constbot.MedVideo:
					answere.Results = append(answere.Results, tgbotapi.InlineQueryResultCachedVideo{
						ParseMode: message.ParseMode,
						Caption: message.Msg,
						Type: "video",
						Description: post.Dis,
						Title: post.Title,
						ReplyMarkup: btns.GetkeyboardCopy(),
					})
			}
		}

		if message.Keyboard != nil {
			btns.ResetNoOveride([]int16{2})
			btns.AddUrlbutton("Channel", a.ctrl.Channelink)
			btns.AddUrlbutton("Group", a.ctrl.GroupLink)
			btns.AddUrlbutton("Bot", a.ctrl.Botlink)
		}
	}

	if len(answere.Results) == 0 {
		return errors.New("inline quary found with no answere")
	}
	_, err := a.botapi.Makerequest(upx.Ctx, "POST", constbot.ApiMethodAnswereInline, &botapi.BotReader{RealOb: answere})
	if err != nil {
		a.logger.Error("Inline Quary Send Error: " + err.Error())
	}
	return nil
}



func (a *InlineService) SendUsage(upx *update.Updatectx) error {
	//TODO: may be later
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