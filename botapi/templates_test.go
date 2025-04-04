package botapi_test

import (
	"fmt"
	"testing"

	"github.com/sadeepa24/connected_bot/botapi"
)

func TestTmpl(t *testing.T) {
	tttt, err := botapi.NewMessageStore("path")
	fmt.Println(err)

	mg := tttt.MsgWithouerro("welcome", "sin", struct {
		Name string
	}{
		Name: "hello",
	})
	fmt.Println(mg)
}

func TestKeyboardMarshell(t *testing.T ) {
	buf, err := botapi.Createkeyboard(&botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{
				{
					Text: "btn1",
					URL: "btnurl1",
				},	
			},
			{
				{
					Text: "btn2",
					URL: `https://t.me`,
				},	
			},
		},
	})
	fmt.Println(string(buf), err)
}