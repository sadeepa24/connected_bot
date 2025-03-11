package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	//
	"github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"

	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/service"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)

var ctx = context.Background()

var zLogger, _ = zap.NewDevelopment()

var usethisdb = db.New(ctx, nil, "connect.db")

func init() {

}

func TestUsersrv(t *testing.T) {
	usethisdb.InitDb()
	var ctrl ,_ = controller.New(ctx, usethisdb, zLogger, &controller.MetadataConf{}, nil, "")

	groupid := 88890

	var callbacksrv = service.NewCallback(ctx, zLogger, nil, nil)
	var adminsrv *service.Adminsrv
	var defaultsrv *service.Defaultsrv

	testmsg := &tgbotapi.Message{
		MessageID: 1,
		Entities: []tgbotapi.MessageEntity{
			tgbotapi.MessageEntity{
				Type:   "bot_command",
				Length: 6,
			},
		},
		From: &tgbotapi.User{
			ID:           555,
			IsBot:        false,
			FirstName:    "John",
			LastName:     "Doe",
			UserName:     "ss",
			LanguageCode: "en",
		},
		Chat: &tgbotapi.Chat{
			ID:        555,
			FirstName: "John",
			LastName:  "Doe",
			UserName:  "ss",
			Type:      "private",
		},
		Date: 1623142790,
		Text: "/start",
	}

	testchamemjoin := &tgbotapi.ChatMemberUpdated{
		Chat: tgbotapi.Chat{
			ID:        int64(groupid),
			FirstName: "Group",
			LastName:  "Nameu",
			UserName:  "ss",
			Type:      "group",
		},
		From: tgbotapi.User{
			ID:           23245,
			IsBot:        false,
			FirstName:    "Newjoin",
			LastName:     "User",
			UserName:     "joinuserusername",
			LanguageCode: "en",
		},
		OldChatMember: tgbotapi.ChatMember{
			User: &tgbotapi.User{
				ID:           23245,
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
				ID:           23245,
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

	testchamemjoin = nil
	fmt.Println(testchamemjoin)

	testupdate := &tgbotapi.Update{
		Message: testmsg,
	}
	var err error
	var ok bool
	upxx := &update.Updatectx{
		Update: testupdate,
		Ctx: ctx,
	}
	upxx.User, ok, err = ctrl.GetUser(testupdate.Message.From)
	if err != nil || !ok {
		upxx.User, _ = ctrl.Newuser(testupdate.Message.From, testupdate.Message.Chat)
	}

	Userservice := service.NewuserService(ctx, callbacksrv, zLogger, adminsrv, ctrl, defaultsrv, nil, nil)
	//fmt.Println(Userservice)
	//Userservice.Setadminchat(int64(groupid))
	if constbot.Userservicename != Userservice.Name() {
		t.Fail()
	}
	st := time.Now()
	err = Userservice.Exec(upxx)
	fmt.Println("service exec elpsed time ", time.Since(st))
	if err != nil {
		zLogger.Error(err.Error())
	}
}

func TestXraywiz(t *testing.T) {

	usethisdb.InitDb()
	var ctrl, _ = controller.New(ctx, usethisdb, zLogger, &controller.MetadataConf{}, nil, "option.Options{}")

	var callbacksrv = service.NewCallback(ctx, zLogger, nil, nil)
	//var adminsrv *service.Adminsrv
	var defaultsrv *service.Defaultsrv

	testmsg := &tgbotapi.Message{
		MessageID: 1,
		Entities: []tgbotapi.MessageEntity{
			tgbotapi.MessageEntity{
				Type:   "bot_command",
				Length: 6,
			},
		},
		From: &tgbotapi.User{
			ID:           987654321,
			IsBot:        false,
			FirstName:    "John",
			LastName:     "Doe",
			UserName:     "ss",
			LanguageCode: "en",
		},
		Chat: &tgbotapi.Chat{
			ID:        987654321,
			FirstName: "John",
			LastName:  "Doe",
			UserName:  "ss",
			Type:      "private",
		},
		Date: 1623142790,
		Text: "/start",
	}
	testupdate := &tgbotapi.Update{
		Message: testmsg,
	}
	upxx := &update.Updatectx{
		Update: testupdate,
		Ctx: ctx,
	}

	t.Log("testing xray service")
	Xrayserwiz := service.NewXraywiz(ctx, callbacksrv, zLogger,  ctrl, defaultsrv, nil, nil)

	if constbot.Xraywizservicename != Xrayserwiz.Name() {
		t.Fail()
	}

	Xrayserwiz.Exec(upxx)

}
