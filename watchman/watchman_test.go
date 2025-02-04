package watchman_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/netip"
	"strconv"
	"testing"
	"time"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofrs/uuid"
	connected "github.com/sadeepa24/connected_bot"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	option "github.com/sadeepa24/connected_bot/sbox_option/v1"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"github.com/sadeepa24/connected_bot/watchman"
	"go.uber.org/zap"
)

var ctx = context.Background()

var zLogger, _ = zap.NewDevelopment()

func TestTemp(t *testing.T) {
	addr, _ := netip.ParseAddr("0.0.0.0")
	val := 1
	val2 := 2
	val3 := 3
	val4 := 4

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

			option.Inbound{
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
			option.Outbound{
				Type: "direct",
				Tag:  "direct",
				Id:   &val,
				DirectOptions: option.DirectOutboundOptions{
					DialerOptions: option.DialerOptions{
						BindInterface: "tun0",
					},
				},
			},
			option.Outbound{
				Type: "direct",
				Tag:  "direct2",
				Id:   &val2,
				DirectOptions: option.DirectOutboundOptions{
					DialerOptions: option.DialerOptions{
						BindInterface: "Ethernet 2",
					},
				},
			},
			option.Outbound{
				Type: "block",
				Tag:  "block",
				Id:   &val3,
			},

			option.Outbound{
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
	bt, _ := json.Marshal(newopt)

	fmt.Println(string(bt))

}

func TestAlgo(t *testing.T) {
	var bandwidth float64 = 4000
	var cappedusers float64 = 2
	var verifiedusers float64 = 100
	var captotal float64 = 10
	var addtonal float64 = 0

	mainquota2 := (bandwidth - (captotal + addtonal)) / (verifiedusers - cappedusers)
	fmt.Println(mainquota2)
}

func TestAlgo2(t *testing.T) {
	var gb int = 1024 * 1024 * 1024
	var Mainquota int = 2000 * gb
	var Newquota int = 1000 * gb
	var Gift int = 300 * gb
	st := time.Now()
	k := float64(Mainquota) / float64(Gift)
	fmt.Println(k)
	Newgift := float64(Newquota) / k

	fmt.Println(C.Bwidth(Newgift).BToString())
	fmt.Println(time.Since(st))
}

// you have to edit database mannualy some cases
func TestWatchman(t *testing.T) {
	fmt.Println("starting watchman ")

	testctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	predata := preconfigure(testctx)

	watchman, _ := watchman.New(testctx, predata.ctrl, predata.botapi, predata.db, predata.watchmaconfig, zLogger, predata.msgstore)

	defer watchman.Close()
	if err := watchman.Start(); err != nil {
		zLogger.Fatal("watchman start failed err ===> " + err.Error())
	}

	//insertdummyuser(predata.db, 100, C.Bwidth(predata.ctrl.CommonQuota.Load()), 0)

	//testing predata function
	datadb, err := watchman.PreprosessDb(testctx)
	if err != nil {
		zLogger.Info("predata error watchman")
		zLogger.Fatal(err.Error())
	}
	fmt.Println(datadb.String())
	fmt.Println(predata.ctrl.CheckCount.Load())

	watchman.RefreshDb(testctx, true, false)
	//inserGiftcoupleV2(predata.ctrl, 1, 3)

	return
	//adding 5 verified user

	type dbinfo struct {
		verfiedusercount int
		unverifedcount   int
		usageduser       int
	}

	err = inserGiftcouple(predata.db, 20, predata.ctrl, 20)
	fmt.Println(err)

	start := 15
	tval := start
	for start < tval+2 {
		start++
		insertVerfied2config(predata.db, int64(start), predata.ctrl, int64(start))
	}
	watchman.RefreshDb(testctx, true, false)

	//adding 3 unveirifed user
	tval = start
	for start < tval+3 {
		start++
		insertUnverified(predata.db, int64(start), predata.ctrl, int64(start))
	}

	tval = start
	//adding2 usaged user
	for start < tval+2 {
		start++
		insertUsagedUser(predata.db, int64(start), predata.ctrl, int64(start))
	}

	tval = start
	//adding 4 month limited
	for start < tval+4 {
		start++
		insertMonthlimited(predata.db, int64(start), predata.ctrl, int64(start))
	}

	time.Sleep(2 * time.Minute)
	watchman.RefreshDb(testctx, true, false)

}

type preconfdata struct {
	ctrl          *controller.Controller
	db            *db.Database
	botapi        *botapi.Botapi
	msgstore      *botapi.MessageStore
	watchmaconfig *watchman.Watchmanconfig
}

func preconfigure(ctx context.Context) (data preconfdata) {
	data = preconfdata{}

	options := testingfirst()
	options.Ctx = ctx

	data.db = db.New(options.Ctx, options.Logger, options.Dbpath)
	data.msgstore, _ = botapi.NewMessageStore("")
	data.botapi = botapi.NewBot(options.Ctx, options.Bottoken, options.Botmainurl, data.msgstore)
	data.ctrl, _ = controller.New(options.Ctx, data.db, options.Logger, options.Metadata, data.botapi, options.Sboxoption)
	data.watchmaconfig = options.Watchman

	err := data.db.InitDb()
	if err != nil {
		panic(err)
	}
	err = data.ctrl.Init()
	if err != nil {
		panic(err)
	}
	return
}

func insertdummyuser(dB *db.Database, count int, initquota C.Bwidth, startfrom int) error {
	for i := startfrom; i < count+startfrom; i++ {

		fmt.Println("inserting a user")
		dB.Model(&db.User{}).Create(&db.User{
			CheckID: uint(i),
			TgID:    int64(i),
			Name:    "testName" + strconv.Itoa(i),
			Username: sql.NullString{
				Valid:  true,
				String: "testUserName" + strconv.Itoa(i),
			},
			Lang:              "en",
			IsInGroup:         true,
			IsInChannel:       true,
			IsRemoved:         false,
			GroupBanned:       false,
			ChannelBanned:     false,
			IsBotStarted:      true,
			CalculatedQuota:   initquota,
			RecheckVerificity: false,
		})
	}

	return nil
}

// varified user with 2 configs
func insertVerfied2config(dB *db.Database, checkId int64, ctrl *controller.Controller, userID int64) error {

	randomquota := rand.Int63n(ctrl.CommonQuota.Load())

	err := dB.Model(&db.User{}).Create(&db.User{
		CheckID: uint(checkId),
		TgID:    userID,
		Name:    "verified unused",
		Username: sql.NullString{
			Valid:  true,
			String: "Random user",
		},
		Lang:              "en",
		IsInGroup:         true,
		IsInChannel:       true,
		IsRemoved:         false,
		GroupBanned:       false,
		ChannelBanned:     false,
		IsBotStarted:      true,
		CalculatedQuota:   C.Bwidth(ctrl.CommonQuota.Load()),
		RecheckVerificity: false,
	}).Error

	if err != nil {
		return err
	}

	uid1, _ := uuid.NewV1()
	uid2, _ := uuid.NewV1()
	err = dB.Model(&db.Config{}).Create(&[]db.Config{
		{
			Name:       "unused verified",
			UUID:       uid1.String(),
			UserID:     userID,
			Type:       "vless",
			Active:     true,
			InboundID:  1,
			OutboundID: 1,
			Usage:      0,
			Download:   0,
			Upload:     0,
			LoginLimit: 2,
			Quota:      C.Bwidth(randomquota),
		},
		{
			Name:       "unused verified",
			UUID:       uid2.String(),
			UserID:     userID,
			Active:     true,
			InboundID:  1,
			OutboundID: 1,
			Usage:      0,
			Type:       "vless",
			Download:   0,
			Upload:     0,
			LoginLimit: 2,
			Quota:      C.Bwidth(ctrl.CommonQuota.Load() - randomquota),
		},
	}).Error
	return err

}

func insertUsagedUser(dB *db.Database, checkId int64, ctrl *controller.Controller, userID int64) error {

	randomquota := rand.Int63n(ctrl.CommonQuota.Load())

	err := dB.Model(&db.User{}).Create(&db.User{
		CheckID: uint(checkId),
		TgID:    userID,
		Name:    "usaged user",
		Username: sql.NullString{
			Valid:  true,
			String: "Random user",
		},
		Lang:              "en",
		IsInGroup:         true,
		IsInChannel:       true,
		IsRemoved:         false,
		GroupBanned:       false,
		ChannelBanned:     false,
		IsBotStarted:      true,
		CalculatedQuota:   C.Bwidth(ctrl.CommonQuota.Load()),
		RecheckVerificity: false,
	}).Error

	if err != nil {
		return err
	}

	uid1, _ := uuid.NewV1()
	uid2, _ := uuid.NewV1()

	conf1qt := C.Bwidth(randomquota)
	usage1 := C.Bwidth(rand.Int63n(conf1qt.Int64() / 2))
	dwn1 := C.Bwidth(float64(usage1*3) / 4)

	conf2qt := C.Bwidth(ctrl.CommonQuota.Load() - randomquota)
	usage2 := C.Bwidth(rand.Int63n(conf2qt.Int64() / 2))
	dwn2 := C.Bwidth(float64(usage2*3) / 4)

	err = dB.Model(&db.Config{}).Create(&[]db.Config{
		{
			Name:       "usaged user",
			UUID:       uid1.String(),
			UserID:     userID,
			Type:       "vless",
			Active:     true,
			InboundID:  1,
			OutboundID: 1,
			Usage:      usage1,
			Download:   dwn1,
			Upload:     usage1 - dwn1,
			LoginLimit: 2,
			Quota:      conf1qt,
		},
		{
			Name:       "usaged user",
			UUID:       uid2.String(),
			UserID:     userID,
			Active:     true,
			InboundID:  1,
			OutboundID: 1,
			Usage:      usage2,
			Type:       "vless",
			Download:   dwn2,
			Upload:     usage2 - dwn2,
			LoginLimit: 2,
			Quota:      conf2qt,
		},
	}).Error
	return err
}

func insertMonthlimited(dB *db.Database, checkId int64, ctrl *controller.Controller, userID int64) error {
	return dB.Model(&db.User{}).Create(&db.User{
		CheckID: uint(checkId),
		TgID:    userID,
		Name:    "testName randome",
		Username: sql.NullString{
			Valid:  true,
			String: "Random user",
		},
		Lang:              "en",
		IsInGroup:         false,
		IsInChannel:       false,
		IsRemoved:         false,
		GroupBanned:       false,
		ChannelBanned:     false,
		IsBotStarted:      true,
		IsMonthLimited:    true,
		CalculatedQuota:   C.Bwidth(ctrl.CommonQuota.Load()),
		RecheckVerificity: false,
	}).Error
}

func insertUnverified(dB *db.Database, checkId int64, ctrl *controller.Controller, userID int64) error {
	return dB.Model(&db.User{}).Create(&db.User{
		CheckID: uint(checkId),
		TgID:    userID,
		Name:    "testName randome",
		Username: sql.NullString{
			Valid:  true,
			String: "Random user",
		},
		Lang:              "en",
		IsInGroup:         true,
		IsInChannel:       false,
		IsRemoved:         true,
		GroupBanned:       false,
		ChannelBanned:     false,
		IsBotStarted:      true,
		CalculatedQuota:   C.Bwidth(ctrl.CommonQuota.Load()),
		RecheckVerificity: false,
	}).Error
}

func inserGiftcouple(dB *db.Database, checkId int64, ctrl *controller.Controller, userID int64) error {

	randomquota := rand.Int63n(ctrl.CommonQuota.Load())

	giftquota := rand.Int31n(2000000)

	err := dB.Model(&db.User{}).Create(&db.User{
		CheckID: uint(checkId),
		TgID:    userID,
		Name:    "gift couple",
		Username: sql.NullString{
			Valid:  true,
			String: "Random user",
		},
		Lang:          "en",
		IsInGroup:     true,
		IsInChannel:   true,
		IsRemoved:     false,
		GroupBanned:   false,
		ChannelBanned: false,
		IsBotStarted:  true,
		GiftQuota:     C.Bwidth(giftquota),
		//	Gifttime: time.Now(),
		CalculatedQuota:   C.Bwidth(ctrl.CommonQuota.Load()),
		RecheckVerificity: false,
	}).Error

	if err != nil {
		return err
	}

	err = dB.Model(&db.User{}).Create(&db.User{
		CheckID: uint(500),
		TgID:    500,
		Name:    "gift couple",
		Username: sql.NullString{
			Valid:  true,
			String: "gift couple",
		},
		Lang:          "en",
		IsInGroup:     true,
		IsInChannel:   true,
		IsRemoved:     false,
		GroupBanned:   false,
		ChannelBanned: false,
		IsBotStarted:  true,
		GiftQuota:     -C.Bwidth(giftquota),
		//Gifttime: time.Now(),
		CalculatedQuota:   C.Bwidth(ctrl.CommonQuota.Load()),
		RecheckVerificity: false,
	}).Error

	if err != nil {
		return err
	}

	uid1, _ := uuid.NewV1()
	uid2, _ := uuid.NewV1()
	err = dB.Model(&db.Config{}).Create(&[]db.Config{
		{
			Name:       "gift couple",
			UUID:       uid1.String(),
			UserID:     userID + 500,
			Type:       "vless",
			Active:     true,
			InboundID:  1,
			OutboundID: 1,
			Usage:      0,
			Download:   0,
			Upload:     0,
			LoginLimit: 2,
			Quota:      C.Bwidth(randomquota),
		},
		{
			Name:       "gift couple",
			UUID:       uid2.String(),
			UserID:     userID + 500,
			Active:     true,
			InboundID:  1,
			OutboundID: 1,
			Usage:      0,
			Type:       "vless",
			Download:   0,
			Upload:     0,
			LoginLimit: 2,
			Quota:      C.Bwidth(ctrl.CommonQuota.Load() - randomquota),
		},
	}).Error
	return err

}

func inserGiftcoupleV2(ctrl *controller.Controller, from, to int64) error {
	upx := update.Updatectx{
		Ctx:    context.Background(),
		Cancle: func() {},
	}
	btypeuser, _, err := ctrl.GetUser(&tgbotapi.User{
		ID: from,
	})

	if err != nil {
		zLogger.Error(err.Error())
		return err
	}

	upx.User = btypeuser
	upx.Chat_ID = btypeuser.TgID

	_, err = ctrl.Gift(&upx, int(to), C.Bwidth(32212254720))

	//zLogger.Error(err.Error())
	return nil
}

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
			option.Inbound{
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

			option.Inbound{
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
			option.Outbound{
				Type: "direct",
				Tag:  "direct",
				Id:   &val,
				DirectOptions: option.DirectOutboundOptions{
					DialerOptions: option.DialerOptions{
						BindInterface: "tun0",
					},
				},
			},
			option.Outbound{
				Type: "direct",
				Tag:  "direct2",
				Id:   &val2,
				DirectOptions: option.DirectOutboundOptions{
					DialerOptions: option.DialerOptions{
						BindInterface: "Ethernet 2",
					},
				},
			},
			option.Outbound{
				Type: "block",
				Tag:  "block",
				Id:   &val3,
			},

			option.Outbound{
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
		Bottoken:   "<your token>",
		Botmainurl: "https://api.telegram.org/bot",
		Metadata: &controller.MetadataConf{
			// AllAdmin: []int64{
			// 	1832636256,
			// },
			GroupID:           -1002325676823,
			ChannelID:         -1002400437670,
			Maxconfigcount:    10,
			LoginLimit:        1,
			BandwidthAvelable: "2000GB",
			RefreshRate:       6,
			ForceAdd:          false,
			Botlink:           "https://t.me/connected_test_bot",
			DefaultDomain:     "connected.bot",
			DefaultPublicIp:   "127.0.0.1",
			SudoAdmin:         6695223775,
			WatchMgbuf:        300,
		},
		Logger:     zLogger,
		Sboxoption: newopt,
		//Templates:  botapi.Testtemplts,
	}

	return newoption
}
