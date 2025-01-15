package botapi_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/sadeepa24/connected_bot/botapi"
)

func TestBotapi(t *testing.T) {
	newbot := botapi.NewBot(context.Background(), "7450429117:AAG-GSYKGsylucObfp8FmsCxNnus8L7EtHo", "https://api.telegram.org/bot", nil)
	_, err := newbot.SendContext(context.Background(), &botapi.Msgcommon{
		Text: "Hello",
		Infocontext: &botapi.Infocontext{
			ChatId:  5413731343,
			User_id: 5413731343,
		},
	})

	res1, is, err := newbot.GetchatmemberCtx(context.Background(), 2090841797, -1002325676823)
	fmt.Println(res1.Status == "member", is, err)

}

func TestTexttmp(t *testing.T) {

	mk, err := botapi.NewMessageStore("sss")
	if err != nil {
		panic(err)
	}

	mmmm, err := mk.GetMessage("temp", "sin", struct {
		Test string
	}{
		Test: "this is parsed val",
	})

	fmt.Println(mmmm, err)

	//botapi.Testtmpl()

}

func TestCallBackdata(t *testing.T) {
	newdata := botapi.Callbackdata{
		Data:   "hellodata",
		Uniqid: rand.Int63(),
	}

	fmt.Println(newdata.StringV2(), newdata.String()[:])

	filldata := &botapi.Callbackdata{}
	err := filldata.FillV2(newdata.StringV2())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(filldata.Data, filldata.Uniqid)
}
