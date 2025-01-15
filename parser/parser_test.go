package parser_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/parser"
	option "github.com/sadeepa24/connected_bot/sbox_option/v1"
	"github.com/sadeepa24/connected_bot/service"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"go.uber.org/zap"
)

var ctx = context.Background()

var zLogger, _ = zap.NewDevelopment()

var usethisdb = db.New(ctx, nil, "connect.db")

func TestParser(t *testing.T) {

	usethisdb.InitDb()
	var ctrl, _ = controller.New(ctx, usethisdb, zLogger, &controller.MetadataConf{}, nil, option.Options{})

	var callbacksrv = service.NewCallback(ctx, zLogger, nil, nil)
	var adminsrv *service.Adminsrv
	var defaultsrv = &service.Defaultsrv{}

	botapi := &botapi.Botapi{}

	ctrl.GroupID = 666

	randid := rand.Int63()
	randid = 1
	var groupid int64 = 666

	testmsgbotstart := &tgbotapi.Message{
		MessageID: 1,
		Entities: []tgbotapi.MessageEntity{
			tgbotapi.MessageEntity{
				Type:   "bot_command",
				Length: 6,
			},
		},
		From: &tgbotapi.User{
			ID:           randid,
			IsBot:        false,
			FirstName:    "Newmsg",
			LastName:     "User",
			UserName:     "@newmsguser",
			LanguageCode: "en",
		},
		Chat: &tgbotapi.Chat{
			ID:        randid,
			FirstName: "Newmsg",
			LastName:  "User",
			UserName:  "@newmsguser",
			Type:      "private",
		},
		Date: 1623142790,
		Text: "/start",
	}
	testmsginGroup := &tgbotapi.Message{
		MessageID: 1,
		Entities: []tgbotapi.MessageEntity{
			tgbotapi.MessageEntity{
				Type:   "bot_command",
				Length: 6,
			},
		},

		From: &tgbotapi.User{
			ID:           randid,
			IsBot:        false,
			FirstName:    "Newmsg",
			LastName:     "User",
			UserName:     "@newmsguser",
			LanguageCode: "en",
		},
		Chat: &tgbotapi.Chat{
			ID:        groupid,
			FirstName: "Group",
			LastName:  "Name",
			UserName:  "@groupusername",
			Type:      "group",
		},
		Date: 1623142790,
		Text: "/start",
	}
	testmsgxrawiz := &tgbotapi.Message{
		MessageID: 1,
		Entities: []tgbotapi.MessageEntity{
			tgbotapi.MessageEntity{
				Type:   "bot_command",
				Length: 7,
			},
			tgbotapi.MessageEntity{
				Type:   "mention",
				Length: 5,
			},
		},

		From: &tgbotapi.User{
			ID:           randid,
			IsBot:        false,
			FirstName:    "Newmsg",
			LastName:     "User",
			UserName:     "@newmsguser",
			LanguageCode: "en",
		},
		Chat: &tgbotapi.Chat{
			ID:        randid,
			FirstName: "Newmsg_Create_v2ray",
			LastName:  "User",
			UserName:  "@newmsguser",
			Type:      "private",
		},
		Date: 1623142790,
		Text: "/create",
	}
	testchamemleft := &tgbotapi.ChatMemberUpdated{
		Chat: tgbotapi.Chat{
			ID:        groupid,
			FirstName: "Group",
			LastName:  "Name",
			UserName:  "groupusername",
			Type:      "group",
		},
		From: tgbotapi.User{
			ID:           randid,
			IsBot:        false,
			FirstName:    "Left User",
			LastName:     "User",
			UserName:     "leftuser username",
			LanguageCode: "en",
		},
		OldChatMember: tgbotapi.ChatMember{
			User: &tgbotapi.User{
				ID:           randid,
				IsBot:        false,
				FirstName:    "Newjoin",
				LastName:     "User",
				UserName:     "joinuserusername",
				LanguageCode: "en",
			},
			Status:      "member",
			IsAnonymous: false,
			CustomTitle: "sssss",
		},
		NewChatMember: tgbotapi.ChatMember{
			User: &tgbotapi.User{
				ID:           randid,
				IsBot:        false,
				FirstName:    "Newjoin",
				LastName:     "User",
				UserName:     "joinuserusername",
				LanguageCode: "en",
			},
			Status:      "left",
			IsAnonymous: false,
			CustomTitle: "sssss",
		},
	}
	testchamemjoin := &tgbotapi.ChatMemberUpdated{
		Chat: tgbotapi.Chat{
			ID:        666,
			FirstName: "Group",
			LastName:  "Name",
			UserName:  "ss",
			Type:      "group",
		},
		From: tgbotapi.User{
			ID:           randid,
			IsBot:        false,
			FirstName:    "Newjoin",
			LastName:     "User",
			UserName:     "joinuserusername",
			LanguageCode: "en",
		},
		OldChatMember: tgbotapi.ChatMember{
			User: &tgbotapi.User{
				ID:           randid,
				IsBot:        false,
				FirstName:    "Newjoin",
				LastName:     "User",
				UserName:     "joinuserusername",
				LanguageCode: "en",
			},
			Status:      "left",
			IsAnonymous: false,
			CustomTitle: "sssss",
		},
		NewChatMember: tgbotapi.ChatMember{
			User: &tgbotapi.User{
				ID:           randid,
				IsBot:        false,
				FirstName:    "Newjoin",
				LastName:     "User",
				UserName:     "joinuserusername",
				LanguageCode: "en",
			},
			Status:      "member",
			IsAnonymous: false,
			CustomTitle: "sssss",
		},
	}
	testmsgbotdoesnotrelventGroup := &tgbotapi.Message{
		MessageID: 1,
		Entities: []tgbotapi.MessageEntity{
			tgbotapi.MessageEntity{
				Type:   "bot_command",
				Length: 6,
			},
		},

		From: &tgbotapi.User{
			ID:           randid,
			IsBot:        false,
			FirstName:    "Newmsg",
			LastName:     "User",
			UserName:     "@newmsguser",
			LanguageCode: "en",
		},
		Chat: &tgbotapi.Chat{
			ID:        88,
			FirstName: "Group",
			LastName:  "Name",
			UserName:  "@groupusername",
			Type:      "group",
		},
		Date: 1623142790,
		Text: "/start",
	}

	allupdates := []*tgbotapi.Update{
		&tgbotapi.Update{
			Message: testmsginGroup,
		},
		&tgbotapi.Update{
			Message: testmsgbotstart,
		},
		&tgbotapi.Update{
			Message: testmsgxrawiz,
		},
		&tgbotapi.Update{
			ChatMember: testchamemjoin,
		},
		&tgbotapi.Update{
			ChatMember: testchamemleft,
		},
		&tgbotapi.Update{
			Message: testmsgbotdoesnotrelventGroup,
		},
		&tgbotapi.Update{},
		&tgbotapi.Update{
			EditedMessage: &tgbotapi.Message{},
		},
	}

	Parsesr := parser.New(ctx, ctrl, []service.Service{
		callbacksrv,
		adminsrv,
		defaultsrv,
		&Testservice{},
	}, botapi, zLogger,
	)

	for _, testupdat := range allupdates {
		//upxx := update.Newupdate(ctx, testupdat)
		st := time.Now()
		err := Parsesr.Parse(testupdat)
		fmt.Println("elpsed time to parse the update ", time.Since(st))
		fmt.Print("\n\n")
		if err != nil {
			zLogger.Error(err.Error())
		}

	}

}

type Testservice struct {
}

func (t *Testservice) Init() error {
	fmt.Println("Test service started")
	return nil
}

func (t *Testservice) Exec(upx *update.Updatectx) error {
	fmt.Println("Test servicce exceed")
	fmt.Printf("all results are above\n\n")

	fmt.Println("is upx.User verified ", upx.User.Isverified())
	fmt.Println("is upx user is admin", upx.User.Isadmin())
	fmt.Println("is upx user started bot", upx.User.Isbotstarted())
	fmt.Println("is upx user Newuser", upx.User.IsnewUser())

	fmt.Println("is upx Drop", upx.Drop())
	fmt.Println("upx user service", upx.Service)
	if upx.Update.Message != nil {
		if upx.Update.Message.IsCommand() {
			fmt.Println("upx Recived command ", upx.Update.Message.Command())
		}
	}

	upx = nil
	return nil
}
func (t *Testservice) Name() string {
	fmt.Println("Test service name required")
	return "Testservice"
}
func (t *Testservice) Canhandle(upx *update.Updatectx) (bool, error) {
	fmt.Println("can handle exec")
	return true, nil
}
