package service

import (
	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/service/events"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
)

func (u *Usersrv) commandEvents(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(upx.Ctx, u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	allevent, err := u.ctrl.LoadEvents(upx.User.TgID)
	if err != nil {
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}
	btns := botapi.NewButtons([]int16{2})
	var (
		enentct   int16
		completed int16
	)
	for _, eve := range u.AllEvents {
		if _, ok := allevent[eve.Name()]; ok {
			btns.Addbutton(eve.Name()+" "+C.GetMsg(C.ButtonSelectEmjoi), eve.Name(), "")
			completed++
			continue
		}
		if eve.Expired() {
			continue
		}
		btns.AddBtcommon(eve.Name())
		enentct++
	}
	btns.AddClose(true)
	Messagesession.Edit(struct {
		*botapi.CommonUser
		AvblCount int16
		Completed int16
	}{
		CommonUser: &botapi.CommonUser{
			Name:     upx.User.Name,
			Username: upx.Chat.UserName,
			TgId:     upx.User.TgID,
		},
		AvblCount: enentct + completed,
		Completed: completed,
	}, btns, C.TmplEventHome)

	callback, err := u.callback.GetcallbackContext(upx.Ctx, btns.ID())
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnClose:
		Messagesession.Edit("closed", nil, "")
		return nil
	default:
		coEventCtx := events.Eventctx{
			Ctx: upx.Ctx,
			Upx: upx,
			Sendreciver: func(msg any) (*tgbotapi.Message, error) {
				_, err := Messagesession.Edit(msg, nil, "")
				if err != nil {
					return nil, err
				}
				mg, err := u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)
				if err == nil {
					Messagesession.Addreply(mg.MessageID)
				}
				return mg, err
			},
			Callbackreciver: func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error) {
				_, err := Messagesession.Edit(msg, btns, "")
				if err != nil {
					return nil, err
				}
				if btns == nil {
					return nil, err
				}
				return u.callback.GetcallbackContext(upx.Ctx, btns.ID())
			},
			Alertsender: func(msg string) {
				Messagesession.SendAlert(msg, nil)
			},
			Btns: btns,
		}
		event := u.AllEvents[callback.Data]
		if _, ok := allevent[callback.Data]; ok {
			return event.ExcuteComplete(coEventCtx)
		}
		return event.Excute(coEventCtx)
	}
}
