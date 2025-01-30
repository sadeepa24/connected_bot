package service

import (
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/update"
)

// var _ Event = AplusConfig{}
func (u *Usersrv) commandPoints(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(upx.Ctx, u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	btns := botapi.NewButtons([]int16{2})
	btns.Addbutton(C.BtnInfo, C.BtnInfo, "")
	btns.Addbutton(C.BtnBuy, C.BtnBuy, "")
	btns.AddClose(true)

	Messagesession.Edit(struct {
		Count int64
		*botapi.CommonUser
	}{
		Count: upx.User.Points,
		CommonUser: &botapi.CommonUser{
			Name:     upx.User.Name,
			Username: upx.Chat.UserName,
			TgId:     upx.User.TgID,
		},
	}, btns, C.TmplPoints)

	for {
		if upx.Ctx.Err() != nil {
			return C.ErrContextDead
		}

		callback, err := u.callback.GetcallbackContext(upx.Ctx, btns.ID())
		if err != nil {
			return err
		}

		switch callback.Data {
		case C.BtnClose:
			Messagesession.DeleteAllMsg()
			return nil
		case C.BtnInfo:
			//TODO:
			Messagesession.Callbackanswere(callback.ID, "You need points to claim events. 🎯 Every month, you'll receive 10 points! 🎁", true)
			//continue
		case C.BtnBuy:
			//TODO:
			Messagesession.Callbackanswere(callback.ID, "The Buy option is coming soon! 🛍️ Stay tuned for updates! ⏳", true)

		}
		// reply, err := u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)
		// if err != nil {
		// 	return err
		// }
		// Messagesession.Addreply(reply.MessageID)

	}

}
