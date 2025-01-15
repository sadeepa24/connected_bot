package singapi_test

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sadeepa24/connected_bot/sbox/singapi"
	"github.com/sagernet/sing-box/option"
)

func TestSing(t *testing.T) {
	addr, _ := netip.ParseAddr("0.0.0.0")
	val := 1
	newopt := option.Options{
		Log: &option.LogOptions{
			Disabled: true,
		},
		DNS: &option.DNSOptions{
			ReverseMapping: false,
		},
		Inbounds: []option.Inbound{
			option.Inbound{
				Type: "vless",
				Tag:  "test_in",
				Id:   &val,

				VLESSOptions: option.VLESSInboundOptions{
					Users: []option.VLESSUser{
						option.VLESSUser{
							Name:     "testt",
							UUID:     "1c6a5143-bfeb-4cfd-b733-1f5e96edc949",
							Maxlogin: 3,
						},
					},
					ListenOptions: option.ListenOptions{
						Listen:      option.NewListenAddress(addr),
						ListenPort:  443,
						TCPFastOpen: true,
					},
					Transport: &option.V2RayTransportOptions{
						Type: "ws",
						WebsocketOptions: option.V2RayWebsocketOptions{
							Path: "/",
						},
					},
				},
			},
		},
		Outbounds: []option.Outbound{
			option.Outbound{
				Type: "direct",
				Tag:  "direct",
				DirectOptions: option.DirectOutboundOptions{
					DialerOptions: option.DialerOptions{
						BindInterface: "tun0",
					},
				},
			},
		},
		Route: &option.RouteOptions{
			AutoDetectInterface: true,
			Final:               "direct",
			Rules: []option.Rule{
				option.Rule{
					Type: "default",
					DefaultOptions: option.DefaultRule{
						Protocol: option.Listable[string]{},
						Outbound: "direct",
					},
				},
			},
		},
	}

	instance, _ := singapi.NewsingAPI(context.Background(), newopt, nil, )
	err := instance.Start()

	if err != nil {
		log.Fatal(err)
	}

	// time.Sleep(2 * time.Minute)
	// ruid := "1c7a5143-bfeb-4cfd-b733-1f5e96edc949"
	uid, _ := uuid.NewV4()
	instance.AddUser(&sbox.Userconfig{
		Vlessgroup: &sbox.Vlessgroup{
			UUID: uid,
		},
		Inboundtag: "test_in",
		Usage:      0,
		Quota:      100,
		LoginLimit: 2,
	})
	go func() {
		time.Sleep(2 * time.Minute)
		instance.RemoveUser(&sbox.Userconfig{
			Vlessgroup: &sbox.Vlessgroup{
				UUID: uid,
			},
			Inboundtag: "test_in",
			Usage:      0,
			Quota:      100,
			LoginLimit: 2,
		})
	}()

	for {
		status, err := instance.GetstatusUser(&sbox.Userconfig{
			Vlessgroup: &sbox.Vlessgroup{
				UUID: uid,
			},
			Inboundtag: "test_in",
			Usage:      0,
			Quota:      100,
			LoginLimit: 2,
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(status)
		time.Sleep(800 * time.Millisecond)
	}

	//time.Sleep(5 * time.Minute)

}
