package watchman_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"

	//
	"github.com/gofrs/uuid"
	connected "github.com/sadeepa24/connected_bot"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"github.com/sadeepa24/connected_bot/watchman"
	"go.uber.org/zap"
)

var (
	ctx = context.Background()
	zLogger, _ = zap.NewDevelopment()
	VpsBandwidthForeach = C.Bwidth(5000 * 1024 * 1024 * 1024)/300
)

// you have to edit database mannualy some cases
func TestWatchman(t *testing.T) {
	dbpath := "./newtest.db"
	os.Remove(dbpath)
	
	fmt.Println("starting watchman ")
	testctx, cancle := context.WithCancel(context.Background())
	defer cancle()
	predata := preconfigure(testctx)

	watchman, _ := watchman.New(testctx, predata.ctrl, predata.botapi, predata.db, predata.watchmaconfig, zLogger, predata.msgstore)
	defer watchman.Close()
	
	rnd := Randomizer{
		db: predata.db,
		ctrl: predata.ctrl,
	}

	rnd.RandomizeDb()


	dbch, err := os.ReadFile(dbpath)
	if err == nil {
		dbfile, err := os.Create("before_watchman_start.db")
		if err == nil {
			dbfile.Write(dbch)
		}
	}
	if err := watchman.Start(); err != nil {
		zLogger.Fatal("watchman start failed err ===> " + err.Error())
	}
	dbch, err = os.ReadFile(dbpath)
	if err == nil {
		dbfile, err := os.Create("after_watchman_start.db")
		if err == nil {
			dbfile.Write(dbch)
		}
	}
	zLogger.Info("watchman started")
	watchman.RefreshDb(testctx, true, false)
	dbch, err = os.ReadFile(dbpath)
	if err == nil {
		dbfile, err := os.Create("after_watchman_refresh(count).db")
		if err == nil {
			dbfile.Write(dbch)
		}
	}
	watchman.RefreshDb(testctx, true, true)
	dbch, err = os.ReadFile(dbpath)
	if err == nil {
		dbfile, err := os.Create("after_watchman_refresh(usage_reset).db")
		if err == nil {
			dbfile.Write(dbch)
		}
	}

	watchman.RefreshDb(testctx, true, false)
	dbch, err = os.ReadFile(dbpath)
	if err == nil {
		dbfile, err := os.Create("final_withLasrefresh.db")
		if err == nil {
			dbfile.Write(dbch)
		}
	}

}

type preconfdata struct {
	ctrl          *controller.Controller
	db            *db.Database
	botapi        botapi.BotAPI
	msgstore      *botapi.MessageStore
	watchmaconfig *watchman.Watchmanconfig
}

func preconfigure(ctx context.Context) (data preconfdata) {
	data = preconfdata{}

	//options := testingfirst()
	options := connected.Botoptions{}
	options.Ctx = ctx

	options.Metadata = &controller.MetadataConf{
		WatchMgbuf: 100,
	}

	data.db = db.New(options.Ctx, options.Logger, options.Dbpath)
	data.msgstore, _ = botapi.NewMessageStore("./store.json")
	data.botapi = botapi.NewBot(options.Ctx, options.Bottoken, options.Botmainurl, data.msgstore)
	data.botapi = &TestBotapiWatchman{
		dotrace: false,
	}

	data.ctrl, _ = controller.New(options.Ctx, data.db, options.Logger, options.Metadata, data.botapi, "./sbox.json")
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

	giftquota := rand.Int31n(20000)

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
		CheckID: uint(checkId)+1,
		TgID:    checkId+1,
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
/* //TODO: update this test confgi to v1.12.0
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
	
		SboxConfPath: "./path.json",
		Bottoken:   "<your token>",
		Botmainurl: "https://api.telegram.org/bot",
		Metadata: &controller.MetadataConf{
			// AllAdmin: []int64{
			// 	1832636256,
			// },
			ConfigFolder: "./confs",
			GroupID:           -1002325676823,
			ChannelID:         -1002400437670,
			Maxconfigcount:    10,
			LoginLimit:        1,
			BandwidthAvelable: "2000GB",
			RefreshRate:       6,
			Botlink:           "https://t.me/connected_test_bot",
			DefaultDomain:     "connected.bot",
			DefaultPublicIp:   "127.0.0.1",
			SudoAdmin:         6695223775,
			WatchMgbuf:        300,
			StorePath: "./store.json",
		},
		Logger:     zLogger,
		Sboxoption: newopt,
		//Templates:  botapi.Testtemplts,
	}

	return newoption
}
*/

type Randomizer struct {
	db *db.Database
	ctrl *controller.Controller

}

/*

unverifieduse who
-- only in chan (with and without usage)
-- only in group (with and without usage)
-- left both (with and without usage)
-- banned (with and without usage)
-- 

verifieduser who
-- has 
---- no config
---- only one config
---- many config
---- one config full used
---- many config full used
---- config quota exceeded
---- capped quota
---- has deleted config
---- has many deleted config
---- has many deleted and many usaged config
---- has many deleted and many usaged config and many unused config

-- is 
---- distributed and hasn't any config
---- distributed and have configs
---- distributed and has many usaged config

*/


func (r *Randomizer) ConfigUsed(user *db.User, active bool) C.Bwidth {
	uid,_ := uuid.NewV4()
	quota := C.Bwidth(5000)
	up := C.Bwidth(rand.Int31n(1500))
	dwn :=  C.Bwidth(rand.Int31n(1000))
	
	
	r.db.Create(&db.Config{
		Quota: quota,
		Download: dwn,
		Upload: up,
		Usage: dwn+up,
		LoginLimit: 5,
		UserID: user.TgID,
		UUID: uid.String(),
		Type: "vless",
		Name: user.Name,
		Active: active,
		InboundID: 1,
		OutboundID: 1,

	})
	user.ConfigCount++
	user.UsedQuota += quota
	user.MonthUsage += dwn+up

	return quota
}
func (r *Randomizer) ConfigUnUsed(user *db.User, active bool) C.Bwidth {
	uid,_ := uuid.NewV4()
	quota := C.Bwidth(5000)
	r.db.Create(&db.Config{
		Quota: quota,
		Download: 0,
		Upload: 0,
		Usage: 0,
		LoginLimit: int16(rand.Int31n(5)),
		UserID: user.TgID,
		UUID: uid.String(),
		Type: "vless",
		Name: user.Name,
		Active: active,
		InboundID: 1,
		OutboundID: 1,
	})
	user.ConfigCount++
	user.UsedQuota += quota
	user.MonthUsage += 0
	return quota
}
func (r *Randomizer) ConfigFullUsed(user *db.User, active bool) C.Bwidth {
	uid,_ := uuid.NewV4()
	quota := C.Bwidth(5000)
	dwn := C.Bwidth(rand.Int31n(4000))
	r.db.Create(&db.Config{
		Quota: quota,
		Download: dwn,
		Upload: quota-dwn,
		Usage: quota,
		LoginLimit: int16(rand.Int31n(5)),
		UserID: user.TgID,
		UUID: uid.String(),
		Type: "vless",
		Name: user.Name,
		Active: active,
		InboundID: 1,
		OutboundID: 1,
	})
	user.ConfigCount++
	user.UsedQuota += quota
	user.MonthUsage += quota

	return quota
}
func (r *Randomizer) ConfigOverUsed(user *db.User, active bool) C.Bwidth {
	uid,_ := uuid.NewV4()
	quota := C.Bwidth(5000)
	over := C.Bwidth(rand.Int31n(3000))
	
	total := quota+over
	dwn := C.Bwidth(rand.Int31n(4000))
	r.db.Create(&db.Config{
		Quota: quota,
		Download: dwn,
		Upload: total-dwn,
		Usage: total,
		LoginLimit: int16(rand.Int31n(5)),
		UserID: user.TgID,
		UUID: uid.String(),
		Type: "vless",
		Name: user.Name,
		Active: active,
		InboundID: 1,
		OutboundID: 1,
	})
	user.ConfigCount++
	return quota
}


//TODO: complete this later 😪
func (r *Randomizer) RandomizeDb() {
	var totaluser int

	r.ctrl.CommonQuota.Swap(VpsBandwidthForeach.Int64())

	//30 unverified user randomized
	for i := 0; i < 30; i++ {
		user := &db.User{
			Name: "unverified "+strconv.Itoa(i),
			TgID: int64(i),
			//Username: "unverified "+strconv.Itoa(i),
			CalculatedQuota: VpsBandwidthForeach,
			Lang: "en",
			IsBotStarted: true,
			Points: 5,
		}


		//user who isn't in chan or group
		if i%4 == 0 {
			user.IsInChannel = false
			user.IsInGroup = false
		}

		//user who has config and used them for while and left
		if i%10 == 0 {
			user.ConfigCount = int16(i%5)
			if user.ConfigCount > 0 {
				for i := 0; i < i%5; i++ {
					r.ConfigUsed(user, false)
				}
			}
		}

		//user who has config and notused them and left
		if i%6 == 0 {
			user.ConfigCount = int16(i%5)
			if user.ConfigCount > 0 {
				for i := 0; i < i%5; i++ {
					r.ConfigUnUsed(user, false)
					
				}
			}
		
		}




		r.db.Create(user)
	}

	totaluser+= 30
	var giftcpl bool
	for i := 30; i < 330; i++ {
		if giftcpl {
			giftcpl = false
			continue
		}
		
		user := &db.User{
			Name: "verified "+strconv.Itoa(i),
			TgID: int64(i),
			Lang: "en",
			IsInChannel: true,
			IsBotStarted: true,
			IsInGroup: true,
			Points: 10,
			CalculatedQuota: VpsBandwidthForeach,
		}

		if i%2 == 0 {
			confcount := rand.Int31n(5)
			if confcount == 0 {
				confcount++
			}
			for j := 0; j < int(confcount); j++ {
				r.ConfigUsed(user, true)	
			}
		}

		if i%5 == 0 {
			confcount := rand.Int31n(5)
			if confcount == 0 {
				confcount++
			}
			user.Configs = []db.Config{}
			user.ConfigCount = 0
			user.UsedQuota = 0
			user.MonthUsage = 0
			for j := 0; j < int(confcount); j++ {
				r.ConfigFullUsed(user, true)	
			}

		}
		if i%50 == 0 {
			confcount := rand.Int31n(5)
			if confcount == 0 {
				confcount++
			}
			user.Configs = []db.Config{}
			user.ConfigCount = 0
			user.UsedQuota = 0
			user.MonthUsage = 0
			for j := 0; j < int(confcount); j++ {
				r.ConfigOverUsed(user, true)	
			}

		}

		if i%30 == 0 {
			giftcpl = true
			inserGiftcouple(r.db, int64(i), r.ctrl, int64(i))
		}

	


		r.db.Create(user)
	}




	//r.db.Create()
}









type TestBotapiWatchman struct {
	botapi.BotAPI
	dotrace bool
}

func (t *TestBotapiWatchman) Makerequest(ctx context.Context, method, endpoint string, body *botapi.BotReader) (*tgbotapi.APIResponse, error){
	if t.dotrace {
		zLogger.Debug("call to botapi's makerequest", zap.Stack("trace"))
	}
	
	
	return &tgbotapi.APIResponse{}, nil
}

func (t *TestBotapiWatchman) SendRawReq(req *http.Request) (*tgbotapi.APIResponse, error){
	if t.dotrace {

		zLogger.Debug("call to botapi's serndrawreq", zap.Stack("trace"))
	}
	
	return &tgbotapi.APIResponse{}, nil
}

func (t *TestBotapiWatchman) SendContext(ctx context.Context, msg *botapi.Msgcommon) (*tgbotapi.Message, error){
	if t.dotrace {
		zLogger.Debug("call to botapi's sendcontext", zap.Stack("trace"))
	}
	

	return &tgbotapi.Message{}, nil
}

func (t *TestBotapiWatchman) AnswereCallbackCtx(ctx context.Context, Callbackanswere *botapi.Callbackanswere) error{
	if t.dotrace {
		zLogger.Debug("call to botapi's answerecallback", zap.Stack("trace"))
	}
	

	return nil
}

func (t *TestBotapiWatchman) GetchatmemberCtx(ctx context.Context, Userid int64, Chatid int64) (*tgbotapi.ChatMember, bool, error){
	if t.dotrace {

		zLogger.Debug("call to botapi's GetchatmemberCtx", zap.Stack("trace"))
	}
	
	return &tgbotapi.ChatMember{}, true, nil
}


func (t *TestBotapiWatchman) Send(msg *botapi.Msgcommon) (*tgbotapi.Message, error){
	if t.dotrace {
		zLogger.Debug("call to botapi's Send", zap.Stack("trace"))
	}
	
	
	return &tgbotapi.Message{}, nil
}

func (t *TestBotapiWatchman) SendError(error, int64){
	if t.dotrace {
		zLogger.Debug("call to botapi's SendError", zap.Stack("trace"))
	}
	
	
}

func (t *TestBotapiWatchman) DeleteMsg(ctx context.Context, msgid int64, chatid int64) error{
	if t.dotrace {
		zLogger.Debug("call to botapi's DeleteMsg", zap.Stack("trace"))
	}
	
	
	return nil
}

func (t *TestBotapiWatchman) GetMgStore() *botapi.MessageStore{
	if t.dotrace {
		zLogger.Debug("call to botapi's GetMgStore", zap.Stack("trace"))
	}
	
	
	return nil
}

func (t *TestBotapiWatchman) SetWebhook(webhookurl, secret, ip_addr string, allowd_ob []string) error{
	if t.dotrace {
		zLogger.Debug("call to botapi's SetWebhook", zap.Stack("trace"))
	}
	
	
	return nil
}

func (t *TestBotapiWatchman) CreateFullUrl(endpoint string) string{
	if t.dotrace {
		zLogger.Debug("call to botapi's CreateFullUrl", zap.Stack("trace"))
	}
	
	
	return "nil"
}

func (t *TestBotapiWatchman) GetFile(file_Id string) (io.ReadCloser, error){
	if t.dotrace {
		zLogger.Debug("call to botapi's GetFile", zap.Stack("trace"))
	}
	return io.ReadCloser(nil), nil
}


func TestUserCountIncrease(t *testing.T) {
	count := 3000
	lasrefreshCount := 0
	for i := 0; i < count; i++ {
		if float32(lasrefreshCount) + (float32(lasrefreshCount)/4)*3 < float32(i) {
			fmt.Println(i)
			lasrefreshCount = i
		}
	}
}