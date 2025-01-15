package server_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sadeepa24/connected_bot/server"
	"go.uber.org/zap"
)

var ctx = context.Background()

var zLogger, _ = zap.NewDevelopment()

func TestServer(t *testing.T) {
	newserver := server.New(context.Background(), &server.ServerOption{}, nil, zLogger)

	go newserver.Start(nil, nil)
	time.Sleep(2 * time.Second)
	res, err := http.Post("http://127.0.0.1:8080/", "text/html", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Println(res.StatusCode)
}
