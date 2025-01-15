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
