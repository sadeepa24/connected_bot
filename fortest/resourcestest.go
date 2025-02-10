package fortest

import (
	"context"
	"io"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/sbox"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"go.uber.org/zap"
)

var ctx = context.Background()

var zLogger, _ = zap.NewDevelopment()

type TestSboxStruct struct {
}

func Newtestsbox() *TestSboxStruct { return &TestSboxStruct{} }

// func (t *TestSboxStruct) Start() error { return nil }
// func (t *TestSboxStruct) Close() error { return nil }

func (t *TestSboxStruct) Close() error { return nil }
func (t *TestSboxStruct) Start() error {
	zLogger.Info("Sbox controller start called")
	return nil
}
func (t *TestSboxStruct) AddUser(user *sbox.Userconfig) (*sbox.Sboxstatus, error) {
	zLogger.Info("Sbox Newuser adding called")
	return &sbox.Sboxstatus{
		Download: 2,
		Upload:   2,
	}, nil
}
func (t *TestSboxStruct) RemoveUser(user *sbox.Userconfig) (*sbox.Sboxstatus, error) { return nil, nil }

// Do not going to update database usage it will automaticaly doing by watchman
func (t *TestSboxStruct) GetStatus(user *sbox.Userconfig) (*sbox.Sboxstatus, error) { return nil, nil }
func (t *TestSboxStruct) AddInboud()                                                {}
func (t *TestSboxStruct) RemoveInboud() error                                       { return nil }
func (t *TestSboxStruct) InboundStatus(tag string) error                            { return nil }
func (t *TestSboxStruct) ShareLinkEncode(user *sbox.Userconfig, str string) (string, error) {
	return "", nil
}

func (t *TestSboxStruct) GetAllInbound() ([]sbox.Inboud, error) {
	return []sbox.Inboud{}, nil
}
func (t *TestSboxStruct) AddInbound() error {
	return nil
}

func (t *TestSboxStruct) GetAllOutbound() ([]sbox.Outbound, error) {
	return []sbox.Outbound{}, nil
}
func (t *TestSboxStruct) AddOutbound() error {
	return nil
}
func (t *TestSboxStruct) RemoveOutboud() error {
	return nil
}
func (t *TestSboxStruct) OutboundStatus(string) error {
	return nil
}

type TestBOTAPI struct {
}

func NewTESTBOTAPI() *TestBOTAPI { return &TestBOTAPI{} }

func (t *TestBOTAPI) Makerequest(ctx context.Context, s, srt string, readcloser io.ReadCloser) (*tgbotapi.APIResponse, error) {
	return nil, nil
}
func (t *TestBOTAPI) SendContext(ctx context.Context, msg *botapi.Msgcommon) (*tgbotapi.Message, error) {
	zLogger.Info("Called BOTAPI send context")
	return &tgbotapi.Message{
		MessageID: 5566,
		From: &tgbotapi.User{
			ID: 16535,
		},
		Chat: &tgbotapi.Chat{
			ID: 32434,
		},
	}, nil

}
func (t *TestBOTAPI) AnswereCallbackCtx(ctx context.Context, Callbackanswere *botapi.Callbackanswere) error {
	return nil
}
func (t *TestBOTAPI) GetchatmemberCtx(ctx context.Context, Userid int64, Chatid int64) (*tgbotapi.ChatMember, bool, error) {
	zLogger.Info("Getchat member called BOTAPI")
	return &tgbotapi.ChatMember{
		Status: "member",
		User: &tgbotapi.User{
			ID:        Userid,
			IsBot:     false,
			FirstName: "Dummy",
			LastName:  "User",
			UserName:  "@dummyuser",
		},
	}, true, nil
}
func (t *TestBOTAPI) Createkeyboard(keyboard *botapi.InlineKeyboardMarkup) ([]byte, error) {
	return []byte{}, nil
}
func (t *TestBOTAPI) Send(msg *botapi.Msgcommon) (*tgbotapi.Message, error) {
	return t.SendContext(ctx, msg)
}
