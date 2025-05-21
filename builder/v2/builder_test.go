package builder_test

import (
	"context"
	"fmt"
	"testing"

	//option "github.com/sadeepa24/connected_bot/builder/sbox_option/v2"
	"github.com/sadeepa24/connected_bot/builder/v2"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/experimental/deprecated"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/service"
)

type Conec struct{}

func (c *Conec) Select(options []string, msg any) (string, error) {
	fmt.Println(msg)
	for i, option := range options {
		fmt.Printf("%d: %s \n", i+1, option)
	}

	var choice int
	fmt.Print("\nEnter choice: ")
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(options) {
		fmt.Println("Invalid input.")
		return "", nil
	}

	return options[choice-1], nil
}

func (c *Conec) ReciveVal(msg string) (string, error) {
	fmt.Print("Enter a value: " + msg)
	var choice string
	fmt.Scanln(&choice)
	return choice, nil
}

func (c *Conec) AlertSend(msg string)  error {
	fmt.Println("ALERT:", msg)
	return nil
}
















func TestBuilder(t *testing.T) {
	ctx := context.Background()
	
	globalCtx := service.ContextWith(ctx, deprecated.NewStderrManager(log.StdLogger()))
	globalCtx = box.Context(globalCtx, include.InboundRegistry(), include.OutboundRegistry(), include.EndpointRegistry())

	Builder, _ := builder.NewBuilder(&Conec{})
	Builder.SetInbound(map[string]*option.Inbound{
		"test": {
			Tag: "test",
			Type: "vless",
			Options: &option.VLESSInboundOptions{
				Users: []option.VLESSUser{
					{
						Name: "testname",
						UUID: "testuuid",
					},
				},
			},
		},
	}) //TODO: remove



	err := Builder.AddOutbound(
		`{
 "type": "vless",
 "tag": "rrdfb d",
 "detour": "direct",
 "bind_interface": "eth1",
 "connect_timeout": "300ms",
 "tcp_fast_open": true,
 "server": "127.0.0.1",
 "server_port": 443,
 "uuid": "jdhdgetfchnde",
 "tls": {
  "enabled": true,
  "server_name": "tlssrvname.com",
  "insecure": true,
  "min_version": "1.1",
  "max_version": "1.3"
 },
 "multiplex": {
  "enabled": true
 },
 "transport": {
  "type": "ws",
  "path": "/websocketopts",
  "headers": {
   "host": "websocket.host"
  }
 }
}
`,
	)

	if err != nil {
		fmt.Println(err)
	}
}