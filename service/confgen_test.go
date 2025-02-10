package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/sadeepa24/connected_bot/builder"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	option "github.com/sagernet/sing-box/option"
)

func TestBuilder(t *testing.T) {
	store, err := builder.NewConfStore("./store.json")
	if err != nil {
		zLogger.Fatal(err.Error())
	}

	Builder, err := builder.NewBuilder(context.Background(), "firstets.json", store, zLogger)
	if err != nil {
		zLogger.Fatal(err.Error())
	}
	uid, _ := uuid.NewV4()

	//Builder.AddDnsServer(option.HostsDNSServerOptions{})
	err = Builder.AddOutbound(db.Config{
		UUID: uid.String(),
		Name: "tests",
	}, sbox.Inboud{
		Type:          "vless",
		Tag:           "sboxin",
		ListenAddres:  "127.0.0.1",
		Domain:        "testcom",
		Listenport:    80,
		Tlsenabled:    true,
		Transporttype: "ws",
		Option: &option.Inbound{
			Type: "vless",
			Options: option.VLESSInboundOptions{
				Transport:                  &option.V2RayTransportOptions{},
				InboundTLSOptionsContainer: option.InboundTLSOptionsContainer{},
			},
		},
	}, "hello.com")
	if err != nil {
		zLogger.Error(err.Error())
	}

	Builder.AddRouteRule("blockads", "")
	Builder.AddDnsRule("testName", "")
	//fmt.Println(Builder.Export())

	err = Builder.Close()
	if err != nil {
		zLogger.Error(err.Error())
	}

	//Builder.Export()

}

func TestStore(t *testing.T) {
	store, err := builder.NewConfStore("./store.json")
	if err != nil {
		zLogger.Fatal(err.Error())
	}
	srv, err := store.DnsServerbyTag("block")
	if err != nil {
		zLogger.Fatal(err.Error())
	}

	fmt.Println(srv.Server.Address)
	fmt.Println(srv.Server.Detour)
	fmt.Println(srv.Server.Tag)

	//fmt.Println(store.DnsServer("block"))
}
