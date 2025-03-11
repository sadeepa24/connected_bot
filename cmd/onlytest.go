package main

import (
	"context"

	"go.uber.org/zap"
)

var ctx = context.Background()

var zLogger, _ = zap.NewDevelopment()
/*
func testingfirst() connected.Botoptions {
	addr, _ := netip.ParseAddr("0.0.0.0")
	val := 1
	val2 := 2
	val3 := 3
	val4 := 4

	_ = val4
	newopt := option.Options{
		Log: &option.LogOptions{
			Disabled: true,
		},
		DNS: &option.DNSOptions{
			ReverseMapping: false,
		},
		Inbounds: []option.Inbound{
			{
				Type: "vless",
				Tag:  "default",
				Id:   &val,

				VLESSOptions: option.VLESSInboundOptions{
					Users: []option.VLESSUser{
						option.VLESSUser{
							Name:     "testt",
							UUID:     "1c7a5143-bfeb-4cfd-b733-1f5e96edc949",
							Maxlogin: 3,
						},
					},
					ListenOptions: option.ListenOptions{
						Listen:      option.NewListenAddress(addr),
						ListenPort:  443,
						TCPFastOpen: true,
						InboundOptions: option.InboundOptions{
							SniffEnabled:              true,
							SniffOverrideDestination:  false,
							SniffTimeout:              option.Duration(100 * time.Millisecond),
							UDPDisableDomainUnmapping: true,
							//	DomainStrategy: option.DomainStrategy(3),
						},
					},
					Transport: &option.V2RayTransportOptions{
						Type: "ws",
						WebsocketOptions: option.V2RayWebsocketOptions{
							Path: "/",
						},
					},
				},
			},

			{
				Type: "vless",
				Tag:  "inbn_2",
				Id:   &val2,

				VLESSOptions: option.VLESSInboundOptions{

					Users: []option.VLESSUser{
						option.VLESSUser{
							Name:     "routetest",
							UUID:     "1c7a5143-bfeb-4cfd-b733-1f5e96edc949",
							Maxlogin: 3,
						},
					},
					ListenOptions: option.ListenOptions{
						Listen:      option.NewListenAddress(addr),
						ListenPort:  444,
						TCPFastOpen: true,
						InboundOptions: option.InboundOptions{
							SniffEnabled:              true,
							SniffOverrideDestination:  false,
							SniffTimeout:              option.Duration(100 * time.Millisecond),
							UDPDisableDomainUnmapping: true,
							//	DomainStrategy: option.DomainStrategy(3),
						},
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
			{
				Type: "direct",
				Tag:  "direct",
				Id:   &val,
				DirectOptions: option.DirectOutboundOptions{
					DialerOptions: option.DialerOptions{
						BindInterface: "tun0",
					},
				},
			},
			{
				Type: "direct",
				Tag:  "direct2",
				Id:   &val2,
				DirectOptions: option.DirectOutboundOptions{
					DialerOptions: option.DialerOptions{
						BindInterface: "Ethernet 2",
					},
				},
			},
			{
				Type: "block",
				Tag:  "block",
				Id:   &val3,
			},

			{
				Id:          &val4,
				Type:        "vless",
				Tag:         "vlessout",
				Custom_info: "This is outbound from the server",
				VLESSOptions: option.VLESSOutboundOptions{
					UUID: "",
					ServerOptions: option.ServerOptions{
						Server:     "104.27.206.92",
						ServerPort: 443,
					},
					OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
						TLS: &option.OutboundTLSOptions{
							ServerName: "linkedin.ghostnet.site",
							Insecure:   true,
							Enabled:    true,
						},
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

		Route: &option.RouteOptions{
			AutoDetectInterface: true,
			Final:               "direct",
			Rules: []option.Rule{

				option.Rule{
					Type: "default",
					DefaultOptions: option.DefaultRule{
						AuthUser: option.Listable[string]{
							"ttt",
							"hello",
						},
						Outbound: "direct",
					},
				},
				option.Rule{
					Type: "botrule",
					DefaultOptions: option.DefaultRule{
						AuthUser: option.Listable[string]{
							"routetest",
						},
						Outbound: "direct",
					},
				},
				option.Rule{
					Type: "botrule",
					DefaultOptions: option.DefaultRule{
						AuthUser: option.Listable[string]{
							"routetest",
						},
						Outbound: "vlessout",
					},
				},

				option.Rule{
					Type: "botrule",
					DefaultOptions: option.DefaultRule{
						AuthUser: option.Listable[string]{
							"routetest",
						},
						Outbound: "direct2",
					},
				},

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

	newoption := connected.Botoptions{
		Watchman: &watchman.Watchmanconfig{
		},
		Dbpath:     "./newtest.db",
		Ctx:        ctx,
		Bottoken:   "7450429117:AAG-GSYKGsylucObfp8FmsCxNnus8L7EtHo",
		Botmainurl: "https://api.telegram.org/bot",

		WebHookServerOption: &server.ServerOption{
			Addr:       "127.0.0.1:5566",
			HttpPath:   "/",
			Cert:       "./tls/certificate.pem",
			Key:        "./tls/private_key.pem",
			ServerName: "localhost",
		},

		Metadata: &controller.MetadataConf{
			// AllAdmin: []int64{
			// 	1832636256,
			// },
			GroupID:           -1002325676823,
			ChannelID:         -1002400437670,
			Maxconfigcount:    10,
			LoginLimit:        1,
			BandwidthAvelable: "2000GB",
			RefreshRate:       2,
			Botlink:           "https://t.me/connected_test_bot",
			DefaultDomain:     "connected.bot",
			DefaultPublicIp:   "127.0.0.1",
			SudoAdmin:         6695223775,
			StorePath:         "./store.json",
			ConfigFolder:      "./configs/",
		},
		Logger:     zLogger,
		Sboxoption: newopt,
		//Templates:  botapi.Testtemplts,
	}

	return newoption
}
	*/
