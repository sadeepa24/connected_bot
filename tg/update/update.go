package update

import (
	"context"
	"fmt"

	"github.com/sadeepa24/connected_bot/db"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update/bottype"
)

type Updatectx struct {
	Ctx        context.Context
	Cancle     context.CancelFunc
	iscallback bool

	Service string
	Update  *tgbotapi.Update
	User    *bottype.User
	Newuser bool
	//Configs []db.Config

	Serviceset bool
	drop       bool
	ShouldSave bool

	Chat_ID int64
	Chat    *tgbotapi.Chat

	Command string // only use inside setuser method in parser
}

func (u Updatectx) IsCommand(cmd string) bool {
	return cmd == u.Command
}

func (u *Updatectx) Dbuser() *db.User {
	return u.User.User
}

func (u *Updatectx) FromChat() *tgbotapi.Chat {
	if u.Update == nil {
		return nil
	}
	return u.Update.FromChat()
}

func (u *Updatectx) FromUser() *tgbotapi.User {
	if u.Update == nil {
		return nil
	}
	switch {
	case u.Update.Message != nil:
		return u.Update.Message.From
	case u.Update.EditedMessage != nil:
		return u.Update.EditedMessage.From
	case u.Update.ChannelPost != nil:
		return u.Update.ChannelPost.From
	case u.Update.EditedChannelPost != nil:
		return u.Update.EditedChannelPost.From
	case u.Update.CallbackQuery != nil:
		return u.Update.CallbackQuery.From
	case u.Update.ChatMember != nil:
		return &u.Update.ChatMember.From
	case u.Update.MyChatMember != nil:
		return &u.Update.MyChatMember.From
	default:
		return nil
	}
}

func (u *Updatectx) Setcallback() {
	u.iscallback = true
}

func (u Updatectx) Iscallback() (bool, error) {
	return u.iscallback, nil
}

func (u *Updatectx) Setservice(srvname string) error {
	if u.Serviceset {
		return fmt.Errorf("already service seted")
	}
	u.Serviceset = true
	u.Service = srvname
	return nil
}

func (u *Updatectx) SetDrop(drp bool) {
	u.drop = true
}

func (u *Updatectx) Drop() bool {
	return u.drop
}

