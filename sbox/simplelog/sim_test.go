package simplelog_test

import (
	"context"
	"log"
	"testing"

	"github.com/sadeepa24/connected_bot/sbox/simplelog"
)

func TestSimlog(t *testing.T) {
	newlog, err := simplelog.Newsimpllogger(context.Background(), "./hello.txt")

	if err != nil {
		log.Fatal(err)
	}
	newlog.Info("hello this is test string")
	newlog.Info("test intiger", 2)
	newlog.Sync()
}
