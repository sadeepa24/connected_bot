package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"github.com/sadeepa24/connected_bot/update/bottype"
	"github.com/sagernet/sing-vmess/vless"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Adminsrv struct {
	ctx      context.Context
	callback *Callback
	defaultsrv *Defaultsrv 
	logger   *zap.Logger
	ctrl     *controller.Controller
	botapi   botapi.BotAPI
	xraywiz *Xraywiz
	msgstore *botapi.MessageStore

	templateEditin *atomic.Bool


	adminuser db.User
	adminuserbtype bottype.User

	modeUser *atomic.Bool // true mode- user false - admin
}

func NewAdminsrv(
	ctx context.Context,
	logger *zap.Logger,
	callback *Callback,
	defaulsrv *Defaultsrv,
	xraywiz *Xraywiz,
	botapi botapi.BotAPI,
	ctrl *controller.Controller,
	msgstore *botapi.MessageStore,
) *Adminsrv {
	return &Adminsrv{
		ctx:      ctx,
		callback: callback,
		botapi:   botapi,
		xraywiz: xraywiz,
		ctrl:     ctrl,
		defaultsrv: defaulsrv,
		logger:   logger,
		msgstore: msgstore,
		templateEditin: new(atomic.Bool),
		modeUser: new(atomic.Bool),
	}

}

func (a *Adminsrv) Exec(upx *update.Updatectx) error {
	//Upx.User is nil in this scope
	//admin, ok :=- upx.User.IsAdmin
	upx.User = &a.adminuserbtype
	if upx.Update == nil {
		return nil
	}
	upx.Ctx, upx.Cancle = context.WithTimeout(a.ctx, 30 * time.Minute) //admin has more time to deal with things
	switch {
	case upx.Update.Message != nil:
		return a.handleMessage(upx)
	}

	return fmt.Errorf("admin exec not implemented")
}

func (a *Adminsrv) handleMessage(upx *update.Updatectx) error {
	//Upx.User is nil in this scope
	Messagesession := botapi.NewMsgsession(upx.Ctx, a.botapi, upx.FromChat().ID, upx.FromChat().ID, "en")
	switch {
	case upx.Update.Message.IsCommand():
		return a.Commandhandler(upx, Messagesession)
	case upx.Update.Message.ForwardFrom != nil:
		forward := upx.Update.Message.ForwardFrom
		_ = forward
		//TODO: implement later
	case upx.Update.Message.ReplyToMessage != nil:
		replyMg := upx.Update.Message.ReplyToMessage

		parts := strings.Split(replyMg.Text, ",")
		if len(parts) == 0 {
			return nil
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}
		Messagesession.CopyMessageTo(int64(id), int64(upx.Update.Message.MessageID))

	default:
		upx.Cancle()
	}
	return nil
}

func (a *Adminsrv) Name() string {
	return C.Adminservicename
}

func (a *Adminsrv) Init() error {
	a.adminuser = db.User{
		TgID: a.ctrl.SudoAdmin,

	}
	a.adminuserbtype = bottype.User{
		User: &db.User{
			TgID: a.ctrl.SudoAdmin,
		},
		Newuser: false,
		Tguser: &tgbotapi.User{
			ID: a.ctrl.SudoAdmin,
		},
	}
	

	return nil
}

func (a *Adminsrv) Canhandle(upx *update.Updatectx) (bool, error) {
	return upx.Service == C.Adminservicename, nil
}

func (a *Adminsrv) Commandhandler(upx *update.Updatectx, Messagesession *botapi.Msgsession) error {
	calls := common.Tgcalls{
		//TODO: Create Function That construct below three function
		Alertsender: func(msg string) {
			Messagesession.SendAlert(msg, nil)
		},
		Sendreciver: func(msg any) (*tgbotapi.Message, error) {
			_, err := Messagesession.Edit(msg, nil, "")
			if err != nil {
				return nil, err
			}
			mg, err := a.defaultsrv.ExcpectMsgContext(upx.Ctx, a.ctrl.SudoAdmin, a.ctrl.SudoAdmin)
			if err == nil {
				Messagesession.Addreply(mg.MessageID)
			}
			return mg, err
		},
		Callbackreciver: func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error) {
			_, err := Messagesession.Edit(msg, btns, "")
			if err != nil {
				return nil, err
			}
			return a.callback.GetcallbackContext(upx.Ctx, btns.ID())
		},
	}

	
	
	

	switch upx.Update.Message.Command() {
	case C.CmdUserInfo:
		return a.getuserinfo(upx, Messagesession, calls)
	case C.CmdBrodcast:
		return a.broadcast(upx)
	case C.CmdServerInfo:
		return a.getserverinfo(upx)
	case C.CmdChatSession:
		return a.createchat(upx, Messagesession, calls)
	case C.CmdOverview:
		return a.overview(upx)
	case C.CmdRefreshDb:
		a.ctrl.Addquemg(upx.Ctx, controller.RefreshSignal(1))
		upx.Cancle()
		return nil
	case "manage":
		return a.manage(Messagesession, calls)
	case "template":
		return a.editTemplate(upx, Messagesession, calls)
	default:
		upx.Cancle()
	}

	return nil
}

func (a *Adminsrv) broadcast(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(upx.Ctx, a.botapi, upx.User.TgID, upx.User.TgID, upx.User.Lang) 
	Messagesession.Edit("send brodcast message", nil, "")
	message, err := a.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)

	if err != nil {
		return err
	}
	Messagesession.Addreply(message.MessageID)
	
	btns := botapi.NewButtons([]int16{2})
	btns.AddBtcommon("to verified")
	btns.AddBtcommon("to all")
	btns.AddBtcommon("to unverified")
	btns.AddCloseBack()


	Messagesession.Edit("select target user type", btns, "")

	callback, err := a.callback.GetcallbackContext(upx.Ctx, btns.ID())
	if err != nil {
		return err
	}
	Messagesession.SendAlert("broadcasting message", nil,)
	Messagesession.Edit("📣", nil, "")

	var userlist = []int64{}
	switch callback.Data {
	case "to verified":
		err = a.ctrl.GetVerifiedUserList(&userlist) 
	case "to all":
		err = a.ctrl.GetUserList(&userlist) 
	case "to unverified":
		err = a.ctrl.GetUnVerifiedUserList(&userlist) 
	}
	if err != nil {
		Messagesession.SendAlert("fetching user list failed try again", nil)
		return err
	}

	var sendrfunc func(to int64, mgid int64) error
	if message.IsForwaded() {
		sendrfunc = func(to, mgid int64) error {
			return Messagesession.ForwardMgTo(to, mgid, int64(message.ForwardFromMessageID))
		}
	} else {
		sendrfunc = Messagesession.CopyMessageTo
	}

	for _, user := range userlist {
		sendrfunc(user, int64(message.MessageID))
	}
	Messagesession.Edit(fmt.Sprintf("Broadcast Done, Message Sent To %d users", len(userlist))  , nil, "")

	return nil
}

func (a *Adminsrv) getuserinfo(upx *update.Updatectx, Messagesession *botapi.Msgsession, calls common.Tgcalls) error {
	
	alertsender := calls.Alertsender
	sendreciver := calls.Sendreciver
	callbackreciver := calls.Callbackreciver

	message, err := sendreciver("send user id or username")
	if err != nil {
		return err
	}

	enduserupx := update.Updatectx{
		User: &bottype.User{},
		Ctx: upx.Ctx,
		Cancle: upx.Cancle,
	}

	id, err := strconv.Atoi(message.Text)
	if err != nil {
		message.Text = strings.ReplaceAll(message.Text, "@", "")
		enduserupx.User.User, err = a.ctrl.GetUserByUserName(message.Text)
	} else {
		enduserupx.User.User, err = a.ctrl.GetUserById(int64(id))
	}

	if err != nil {
		Messagesession.SendAlert(fmt.Sprintf("failed fetching target user from db - %s", err.Error()), nil)
		return nil
	}

	endusersession, err := controller.NewctrlSession(a.ctrl, &enduserupx, false)

	if err != nil {
		Messagesession.SendAlert(fmt.Sprintf("failed creating target userssion err - %s", err.Error()), nil)
		return nil
	}

	defer endusersession.Close()
	endusermsg := botapi.NewMsgsession(upx.Ctx, a.botapi, enduserupx.User.TgID, enduserupx.User.TgID, "en")

	var (
		state int
		callback *tgbotapi.CallbackQuery
		confid int
	)

	btns := botapi.NewButtons([]int16{2})

	main:
	for {
		// 0 initiate
		// 1 user info
		// 2 show configs
		// 3 config info
		btns.Reset([]int16{2})
		if upx.Ctx.Err() != nil {
			break main
		}
		switch state {
		case 0:
			btns.AddBtcommon("User Info")
			btns.AddBtcommon("Config Info")
			btns.AddBtcommon("Builder")
			btns.AddBtcommon("Configure")		
			btns.AddClose(false)

			callback, err = callbackreciver("select", btns)
			if err != nil {
				break main
			}

			switch callback.Data {
			case "User Info":
				state = 1
			case "Config Info":
				state = 2
			case "Builder":
				state = 4
			case "Configure":
				state = 5
			case C.BtnClose:
				break main
			}

		case 1:
			if enduserupx.User.Restricted {
				btns.Addbutton("Remove Restrict 🟢", "res", "")
			} else {
				btns.Addbutton("Restrict User 🔴",  "res","" )
			}
			if enduserupx.User.IsMonthLimited {
				btns.AddBtcommon("Remove Monthlimit")
			}
			btns.AddBtcommon(C.BtnBack)
			btns.AddBtcommon("Distribute")
			btns.AddClose(false)
			Messagesession.Edit(userinfo{
				CommonUser: &botapi.CommonUser{
					Name:     enduserupx.User.Name,
					TgId:     enduserupx.User.TgID,
					Username: enduserupx.User.User.Username.String,
				},
				GiftQuota: enduserupx.User.GiftQuota.BToString(),
				Joined:    enduserupx.User.Joined.Format("2006-01-02 15:04:05"),
				Dedicated: C.Bwidth(a.ctrl.CommonQuota.Load()).BToString(),
				TQuota:    (endusersession.GetUser().CalculatedQuota + enduserupx.User.AdditionalQuota).BToString(),
				LeftQuota: endusersession.LeftQuota().BToString(),
				TUsage:    endusersession.TotalUsage().BToString(),
				ConfCount: endusersession.GetUser().ConfigCount,
				CapEndin:  upx.User.Captime.AddDate(0, 0, 30).String(),

				Disendin:     ((a.ctrl.ResetCount - a.ctrl.CheckCount.Load()) * a.ctrl.RefreshRate) / 24,
				UsageResetIn: ((a.ctrl.ResetCount - a.ctrl.CheckCount.Load()) * a.ctrl.RefreshRate) / 24,

				Iscapped:       enduserupx.User.IsCapped,
				IsMonthLimited: enduserupx.User.IsMonthLimited,
				Isdisuser:      enduserupx.User.IsDistributedUser,

				

				JoinedPlace: enduserupx.User.CheckID,

			}, btns, C.TmpUserInfo)

			if callback, err = a.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}
			
			switch callback.Data {
			case C.BtnBack:
				state = 0
			case C.BtnClose:
				break main
			case "Distribute":
				if enduserupx.User.IsCapped || enduserupx.User.IsDistributedUser {
					alertsender("can't distribute, either user capped or already distributed")
					continue
				}
				
				endusersession.GetUser().IsDistributedUser = true
				endusersession.DeactivateAll()
				endusermsg.SendAlert("you'r quota has being Distributed By Admin ", nil)
			
			case "res":
				if endusersession.GetUser().Restricted {
					endusersession.RemoveRestrict()
					endusermsg.SendAlert("✅ admin removed you'r restriction, you can use service again 🎉", nil)
					Messagesession.SendAlert("make a db refresh to change bandiwdth, it will automatically change in next refresh cycle", nil)
				
				} else {
					endusersession.Restrict()
					endusermsg.SendAlert("🔴 you have restricted by admin you may have to contact admin to remove this restriction ", nil)
					Messagesession.SendAlert("make a db refresh to change bandiwdth, it will automatically change in next refresh cycle", nil)
				}
			case "Remove Monthlimit":
				endusersession.GetUser().IsMonthLimited = false
				endusersession.ActivateAll()
				endusermsg.SendAlert("🎉you'r monthlimitation removed by admin 🍾", nil)
				Messagesession.SendAlert("make a db refresh to change bandiwdth, it will automatically change in next refresh cycle", nil)
			}
				
		case 2:
			if endusersession.GetUser().ConfigCount <= 0{
				alertsender("user does not have created configs")
				state = 0
				break
			}
			for _, conf := range  endusersession.GetUser().Configs {
				btns.Addbutton(conf.Name, strconv.Itoa(int(conf.Id)), "")
			}
			btns.AddCloseBack()

			callback, err := callbackreciver("select config", btns)
			if err != nil {
				break main
			}
			switch callback.Data {
			case C.BtnBack:
				state = 0
			case C.BtnClose:
				break main
			default:
				confid, err = strconv.Atoi(callback.Data)
				if err != nil {
					continue
				}
				state = 3

			}

		case 3:

			selectedconfig, err := endusersession.GetConfig(int64(confid))
			if err != nil {
				continue
			}

			status, err := endusersession.Getstatus(int64(confid))

			if err != nil {
				if errors.Is(err, C.ErrContextDead) {
					return err
				} else if errors.Is(err, vless.ErrUserNotFound) {
					status = sbox.Sboxstatus{
						Download:  0,
						Upload:    0,
						Online_ip: map[netip.Addr]int64{},
					}
				} else {
					Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.GetMsg(C.Msgconfcannotfind)), true)
					Messagesession.DeleteAllMsg()
					return err
				}

			}
			sboxin, _ := a.ctrl.Getinbound(int(selectedconfig.InboundID))
			sboxout, _ := a.ctrl.Getoutbound(int(selectedconfig.OutboundID))


			btns.AddBtcommon(C.BtnCloseConn)
			btns.AddCloseBack()

			if callback, err = callbackreciver(botapi.UpMessage{
				Template: configinfo{
					CommonUser: &botapi.CommonUser{
						Name:     enduserupx.User.Name,
						Username: enduserupx.User.User.Username.String,
						TgId:     enduserupx.User.TgID,
					},
	
					TotalQuota:     selectedconfig.Quota.BToString(),
					ConfigName:     selectedconfig.Name,
					ConfigType:     selectedconfig.Type,
					ConfigUUID:     selectedconfig.UUID,
					Loginlimit: selectedconfig.LoginLimit,
					UsedPresenTage: float64(int(((selectedconfig.Usage+status.FullUsage()).Float64()/selectedconfig.Quota.Float64())*100*1000)) / 1000,
					//UsedPresenTage: (((selectedconfig.Usage + status.FullUsage()).Float64()/selectedconfig.Quota.Float64()))*100,
	
					ResetDays: ((a.ctrl.ResetCount - a.ctrl.CheckCount.Load()) * a.ctrl.RefreshRate) / 24,
	
					ConfigDownload: (selectedconfig.Download + status.Download).BToString(),
					ConfigUpload:   (selectedconfig.Upload + status.Upload).BToString(),
	
					ConfigDownloadtd: (status.Download).BToString(),
					ConfigUploadtd:   (status.Upload).BToString(),
	
					ConfigUsagetd: (status.Download + status.Upload).BToString(),
					ConfigUsage:   (status.Download + status.Upload + selectedconfig.Usage).BToString(),
					PublicIp: sboxin.PublicIp,
					PublicDomain: sboxin.Domain,
	
					InName:         sboxin.Name,
					InType:         sboxin.Type,
					InPort:         sboxin.Port(),
					InAddr:         sboxin.Laddr(),
					InInfo:         sboxin.Custom_info,
					TranstPortType: sboxin.TransortType(),
					TlsEnabled:     sboxin.TlsIsEnabled(),
					UsageDuration:  time.Since(a.ctrl.GetLastRefreshtime()).Round(1 * time.Second).String(),
					SupportInfo:    sboxin.Support,
	
					OutName: sboxout.Name,
					OutType: sboxout.Type,
					OutInfo: sboxout.Custom_info,
					Latency: sboxout.Latency.Load(),
	
					Online: len(status.Online_ip),
					IpMap:  status.Online_ip,
					//TODO: fill here
				},
				TemplateName: C.TmpConfigInfo,
				Lang: "en",
			}, btns); err != nil {
				a.logger.Error(err.Error())
				continue

			}

			switch callback.Data {
			case C.BtnClose:
				break main
			case C.BtnBack:
				state = 2
				continue
			case C.BtnCloseConn:
				endusersession.ConfigCloseConn(int64(confid))
			}

		case 4:
			if _, ok := a.xraywiz.builds.Load(enduserupx.User.TgID); ok {
				Messagesession.SendAlert("user have already opend a builder session please wait until he closes it", nil)
				return nil
			}
			
			tra := alltransportadder(sendreciver, callbackreciver, alertsender, btns)			
			buildstate := BuildState{
				ctx: upx.Ctx,
				State: 0,
				Messagesession: Messagesession,
				dbuser: endusersession.GetUser(),
				userID: endusersession.GetUser().TgID,
				btns: btns,
				wiz: a.xraywiz,
				sendreciver: sendreciver,
				callbackreciver: callbackreciver,
				alertsender: alertsender,
				otadders: alloutadders(sendreciver, callbackreciver, alertsender, btns, tra),
				changers: allchangers(sendreciver, callbackreciver, alertsender, btns, tra),

			}
			a.xraywiz.builds.Store(enduserupx.User.TgID, struct{}{})
			err = buildstate.run()
			if buildstate.Builder != nil {
				buildstate.Builder.Close()
			}
			a.xraywiz.builds.Delete(enduserupx.User.TgID)
			if err != nil {
				a.logger.Error(err.Error())
				if errors.Is(err, C.ErrContextDead) {
					tmpctx, cancle := context.WithTimeout(a.ctx, 10*time.Second)
					Messagesession.SetNewcontext(tmpctx)
					Messagesession.SendAlert("context timeouts", nil)
					cancle()
					break main
				}
			}
			state = 0

		case 5:
			
			configState := &configState{
				ctx:            upx.Ctx,
				State:          stconfhome,
				//upx:            upx,
				userId: upx.User.TgID,
				dbuser: endusersession.GetUser(),
				btns:           botapi.NewButtons([]int16{2}),
				Usersession:    endusersession,
				wiz:            a.xraywiz,
				Messagesession: Messagesession,
				Tgcalls: common.Tgcalls{
					Alertsender: alertsender,
					Sendreciver: sendreciver,
					Callbackreciver: callbackreciver,
				},
			}

			conformbtns := botapi.NewButtons([]int16{1, 1})
			conformbtns.Addbutton(C.BtnConform, C.BtnConform, "")
			conformbtns.Addbutton(C.BtnCancle, C.BtnCancle, "")
		
			configState.conform = func(msg any, name string) (bool, error) {
				if _, err = Messagesession.Edit(msg, conformbtns, name); err != nil {
					return false, err
				}
				var callback *tgbotapi.CallbackQuery
				if callback, err = a.callback.GetcallbackContext(upx.Ctx, conformbtns.ID()); err != nil {
					return false, err
				}
				switch callback.Data {
				case C.BtnConform:
					return true, nil
				case C.BtnCancle:
					return false, nil
				default:
					return false, nil
				}
			}

			configState.run()
			if err != nil {
				a.logger.Error(err.Error())
				if errors.Is(err, C.ErrContextDead) {
					tmpctx, cancle := context.WithTimeout(a.ctx, 10*time.Second)
					Messagesession.SetNewcontext(tmpctx)
					Messagesession.SendAlert("context timeouts", nil)
					cancle()
					break main
				}
			}
			state = 0

		}
	}

	Messagesession.DeleteAllMsg()

	return nil

}

func (a *Adminsrv) getserverinfo(upx *update.Updatectx) error {
	var memorystate = runtime.MemStats{}
	runtime.ReadMemStats(&memorystate)

	info := fmt.Sprintf(`

	Memory:
	- Total Allocated: %f MB
	- Total 

	Prosess:
	- CPU %d
	- Goroutine %d

	Debug:
	- Lookups %d
	- HeapObjects %d
	- StackInuse %d
	- Frees %d

	`, 
	//memory
	float64(memorystate.Sys)/(1024*1024), 
	
	//prosess
	runtime.NumCPU(),
	runtime.NumGoroutine(),

	//debug
	memorystate.Lookups,
	memorystate.HeapObjects,
	memorystate.StackInuse,
	memorystate.Frees,

)
	a.ctrl.Addquemg(upx.Ctx, &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId: a.ctrl.SudoAdmin,
		},
		Text: info,
	})

	return nil
}

func (a *Adminsrv) createchat(upx *update.Updatectx, Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {
	Messagesession.Edit("send target user", nil, "")
	message, err := a.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)
	if err != nil {
		return err
	}

	var dbuser *db.User

	id, err := strconv.Atoi(message.Text)
	if err != nil {
		message.Text = strings.ReplaceAll(message.Text, "@", "")
		dbuser, err = a.ctrl.GetUserByUserName(message.Text)
	} else {
		dbuser, err = a.ctrl.GetUserById(int64(id))
	}


	if err != nil {
		return err
	}

	if dbuser.TgID == a.adminuser.TgID {
		calls.Alertsender("You can't chat weith your self 😅")
		return nil
	}
	a.ctrl.Addquemg(upx.Ctx, &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId: dbuser.TgID,
		},
		Text: "admin created chat session with you, you can't use any command or anything until he ends the session ",
	})


	canclechan := make(chan any)

	mgcoping := func (src, dst int64, admin bool)  {
		mgcopy:
		for {
			select {
			case <-canclechan:
				break mgcopy
			default:
			}
			mg, err := a.defaultsrv.ExcpectMsgContext(upx.Ctx, src, src) // this will check context automatically
			if err != nil && admin{
				a.ctrl.Addquemg(context.Background(), &botapi.Msgcommon{
					Infocontext: &botapi.Infocontext{
						ChatId: upx.User.TgID,
					},
					Text: "chat session ended",
				})
				break
			}
			if mg.IsCommand() {
				if mg.Command() == "cancel" && admin {
					upx.Cancle()
					canclechan <- struct{}{}
					break
				}
			}
			Messagesession.CopyMessageRawTo(dst, int64(mg.MessageID), src)
		}
		if !admin {
			a.ctrl.Addquemg(context.Background(), &botapi.Msgcommon{
				Infocontext: &botapi.Infocontext{
					ChatId: dbuser.TgID,
				},
				Text: "admin closed chat session",
			})
		}

	}
	
	go mgcoping(upx.User.TgID, dbuser.TgID, true)
	go mgcoping(dbuser.TgID, upx.User.TgID, false)
	
	
	return nil
}

func (a *Adminsrv) overview(upx *update.Updatectx) error {
	overview := a.ctrl.Overview
	overview.Mu.RLock()
	defer overview.Mu.RUnlock()


	a.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
		DestinatioID: upx.User.TgID,
		TemplateName: "overview",
		Template: struct{
			BandwidthAvailable string
			MonthTotal string
			AllTime string
			VerifiedUserCount int64
			TotalUser int32
			CappedUser int64
			DistributedUser int64
			Restricte int64
			QuotaForEach string
			LastRefresh time.Time
		}{
			BandwidthAvailable: overview.BandwidthAvailable.BToString(),
			AllTime: overview.AllTime.BToString(),
			QuotaForEach: overview.QuotaForEach.BToString(),
			MonthTotal: overview.MonthTotal.BToString(),
			TotalUser: overview.TotalUser,
			CappedUser: overview.CappedUser,
			Restricte: overview.Restricted,
			DistributedUser: overview.DistributedUser,
			LastRefresh: overview.LastRefresh,
			VerifiedUserCount: overview.VerifiedUserCount,
			
		},
		Lang: "en",
	})


	return nil
}

func (a *Adminsrv) editTemplate(upx *update.Updatectx, Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {
	// This Editing Does Not Affect Running Templates Due Running Template is in memory, only loads at start up 
	// Admin need to restart after editig
	// I don't add realtime changes due to syncing overhead for small feture, it does not worth
	
	

	if a.templateEditin.Swap(true) {
		calls.Alertsender("Already opend template editor")
		upx.Cancle()
		return nil
	}

	defer a.templateEditin.Swap(false)

	


	path := a.msgstore.GetPath() //TODO: GetPath Later


	file, err := os.ReadFile(path)
	if err != nil {
		calls.Alertsender("template file opening err - " + err.Error())
		return nil
	}
	
	var Templates map[string]map[string]*botapi.MgItem

	switch {
	case strings.Contains(path, ".yaml"):
		err = yaml.Unmarshal(file, &Templates)
	case strings.Contains(path, ".json"):
		err = json.Unmarshal(file, &Templates)
	}
	if err != nil {
		calls.Alertsender("Unmarshaling err - " + err.Error())
		return nil
	}
	nameslice := C.MapToSliceKey(Templates)
	btnpereach := 16
	maxpages := len(nameslice)/btnpereach
	currentpage := 0
	btns := botapi.NewButtons([]int16{2})
	var (
		callback *tgbotapi.CallbackQuery
		replymg *tgbotapi.Message
		state int16
	)

	t := []*botapi.MgItem{}

	for _, v := range Templates {
		for _, s := range v {
			t = append(t, s)
		}
	}

	calls.Alertsender("You will Recive All Media Before initing Template Editor")
	botapi.TemplateInit(a.botapi, a.adminuser.TgID, a.logger, t)
	t = nil

	defer func ()  {
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			calls.Alertsender("file opening err - " + err.Error())
		}

		output, err := yaml.Marshal(Templates)
		if err != nil {
			calls.Alertsender("yaml marshling err - " + err.Error())
		}

		_, err = file.Write(output)
		if err != nil {
			calls.Alertsender("file writing err - " + err.Error())
		}
		if err = file.Close(); err != nil {
			calls.Alertsender("file closing err - " + err.Error())
		}
		calls.Alertsender("succesfully save new template, you need to restart program to take effect new template")
	}()


	selecttmpl:
	for {
		btns.Reset([]int16{2})

		switch state {
		case 0:
			btns.AddBtcommon("Upload Media")
			btns.AddBtcommon("Create New Template")
			btns.AddBtcommon("Edit Templates")
			btns.AddClose(false)
	
			if callback, err = calls.Callbackreciver("select option", btns); err != nil {
				return err
			}
	
			switch callback.Data {
			case "Upload Media":
				state = 1
			case "Create New Template":
				state = 2
			case "Edit Templates":
				state = 3
			case C.BtnClose:
				break selecttmpl
			}

		case 1: //Upload Img
			
			replymg, err = calls.Sendreciver("send you'r media (only support photo or video),  media should below 20MB")
			if err != nil {
				return err
			}
			var fileid string
		
			switch {
			case replymg.Document != nil:
				fileid = replymg.Document.FileID
			case replymg.Video != nil:
				fileid = replymg.Video.FileID
			case replymg.Photo != nil:
				if len(replymg.Photo) == 0 {
					continue selecttmpl
				} 
				fileid = replymg.Photo[len(replymg.Photo)-1].FileID

			}

			replymg, err = calls.Sendreciver("send new file name for the file name with extetion ex - example.mp4")

			filename := replymg.Text

			file, err := a.botapi.GetFile(fileid)
				
			if err != nil {
				calls.Alertsender("file reciving err - " + err.Error())
				continue selecttmpl
			}
			endfile, err := os.OpenFile("./res/" + filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			
			if err != nil {
				calls.Alertsender("opening new file err" + err.Error())
				endfile.Close()
				file.Close()
				continue selecttmpl
			}

			_, err = io.Copy(endfile, file)

			if err != nil {
				endfile.Close()
				file.Close()
				calls.Alertsender("file saving err" + err.Error())
				continue selecttmpl
			}
			endfile.Close()
			file.Close()
			state = 0


		case 2:
			btns.AddBtcommon("Add Help Pages")
			btns.AddBtcommon("Add Inline Post")
			btns.AddCloseBack()

			callback, err  = calls.Callbackreciver("select option", btns)
			if err != nil {
				return err
			}

			switch callback.Data {
			case "Add Help Pages":

				//TODO: add later
				calls.Alertsender("not avbl yet")


			case "Add Inline Post":
				replymg, err = calls.Sendreciver("Send Name for New template")
				if err != nil {
					return err
				}
				if _, ok := Templates[replymg.Text]; ok {
					calls.Alertsender("their is alredy template with this name try again")
					continue selecttmpl
				}
				Templates[replymg.Text] = map[string]*botapi.MgItem{
					"en": {

					},
				}
				calls.Alertsender("New Template Created, Now You can Edit It, Also You have to Add this Template Name Into config.json's metadata.inline_post inorder to recive")

			case C.BtnClose:
				return nil
			case C.BtnBack:
				state = 0


			}


		case 3:
			btns.Reset([]int16{2})
			if currentpage < maxpages {
				for _, name := range  nameslice[(btnpereach*currentpage): (btnpereach*currentpage)+btnpereach] {
					btns.AddBtcommon(name)
				}
			} else {
				for _, name := range  nameslice[(btnpereach*currentpage): (len(nameslice) - (btnpereach*currentpage))  + (btnpereach*currentpage)] {
					btns.AddBtcommon(name)
				}
			}
			if currentpage+1 < maxpages || (currentpage+1 == maxpages &&  (len(nameslice)%btnpereach > 0)){
				btns.AddBtcommon("next")
			}
			btns.AddCloseBack()
	
			if callback, err = calls.Callbackreciver("select template", btns); err != nil {
				return err
			}
	
			switch callback.Data {
			case "next":
				currentpage++
				continue selecttmpl
			case "back":
				if currentpage == 0 {
					state = 0
					continue selecttmpl
				}
				currentpage--
				continue selecttmpl
			case C.BtnClose:
				
				rep, err := calls.Sendreciver("you'r current template will save,  do you want to continue, then send ok")
				if err != nil {
					return err
				}
			
				if rep.Text != "ok" {
					calls.Alertsender("Closing Canceld Continue Editing")
				}
				calls.Alertsender("Editor Closed, Saving Template..")
	
				break selecttmpl
			}
	
	
			btns.Reset([]int16{2})
	
			selectedtemplate := Templates[callback.Data]
	
			for langcode := range selectedtemplate {
				btns.AddBtcommon(langcode)
			}
	
			btns.Addbutton("create lang template","crtt", "")
	
			if callback, err = calls.Callbackreciver("select langcode or create template using new langcode", btns); err != nil {
				return err
			}
	
	
			var replymg *tgbotapi.Message
	
			if callback.Data == "crtt" {
				replymg, err = calls.Sendreciver("send new langcode, if you send exting code current item will replace with new template boilerplate")
				if err != nil {
					return err
				}
				selectedtemplate[replymg.Text] = &botapi.MgItem{}
				callback.Data = replymg.Text
			}
	
			selectedItem := selectedtemplate[callback.Data]
	
	
			editable := []string{
				"msg_template", 
				"alt_med_url", 
				"parse_mode", 
				"include_media", 
				"media_type", 
				"media_id", 
				"continue_media", 
				"disabled", 
				"skip_text", 
				"contin_skip_text", 
				"alt_med_path",  
				"supercontinue", 
	
			}
	
			var mode = "prv"
			
	
			var msghook = func (original *botapi.MgItem) any {
				switch mode {
				case "dt":
					kk, err := json.MarshalIndent(original, "", " ")
					if err != nil {
						return "Errpr"		
					}
					return botapi.Htmlstring("<pre>" + string(kk) + "</pre>")
				default:
					return &botapi.Message{
						Msg: original.Msgtmpl            ,
						MediaId: original.MediaId,
						MedType: original.Mediatype,
						ParseMode: original.ParseMode,
						Includemed: original.Includemed,
						ContinueMed: false,
						SuperContinue: false,
					}
				}
	
				
			}
	
			itemchange:
			for {
				
				
				btns.Reset([]int16{2})
	
				switch mode {
				case "prv":
					btns.Addbutton("As Detail 💠", "modebt", "")
				case "dt":
					btns.Addbutton("As Preview 💠", "modebt", "")
				}
				for _, editname := range editable {
					btns.AddBtcommon(editname)
				}
				btns.AddBtcommon("Done")

	
				if callback, err  = calls.Callbackreciver(msghook(selectedItem), btns); err != nil {
					if errors.Is(err, C.ErrContextDead) {
						return err
					}
					mode = "dt"
					calls.Alertsender("tg rendering error template check you'r template again err " + err.Error())
					continue itemchange
					
				}
	
				switch callback.Data {
				case "modebt":
					switch mode {
					case "prv":
						mode = "dt"
					case "dt":
						mode = "prv"	
					}
				case "Done":
					break itemchange
				case "parse_mode":	
					btns.Reset([]int16{2})
					btns.AddBtcommon("html")
					btns.AddBtcommon("markdown")
					btns.AddBtcommon("markdown2")
					btns.AddBtcommon("none")
	
	
					callback, err = calls.Callbackreciver("Select parse mode", btns)
					if err != nil {
						return err
					}
					switch callback.Data {
					case "html":
						selectedItem.ChangeField("parse_mode", "HTML")
					case "markdown":
						selectedItem.ChangeField("parse_mode", "Markdown")
					case "markdown2":
						selectedItem.ChangeField("parse_mode", "MarkdownV2")
					default:
						selectedItem.ChangeField("parse_mode", "")
					}
				case "alt_med_path":
					dirs, err := os.ReadDir("./res")
					if err != nil {
						continue itemchange
					}

					s := " Select Name From Below (All Files In res Folder) \n\n"

					for _, dir := range dirs {
						s = s + dir.Name() + "\n"
					}

					replymg, err = calls.Sendreciver(s)

					if err != nil {
						return err
					}
					if err = selectedItem.ChangeField("alt_med_path", "./res/"+replymg.Text); err != nil {
						calls.Alertsender(" field changing failed err - "+ err.Error())
					}

					calls.Alertsender("If this Newly Uploded Media You will Need to Restart Editor to See it In preview Mode")

				default:
					value, err := calls.Sendreciver("send you'r new value for, send /cancel to cancel " + callback.Data)
					if err != nil {
						return err
					}
					if value.Text == "/cancel" {
						continue itemchange
					}
					if err = selectedItem.ChangeField(callback.Data, value.Text); err != nil {
						calls.Alertsender(" field changing failed err - "+ err.Error())
					}
	
				}
	
			}

		}
	}
	Messagesession.DeleteAllMsg()
	return nil

}

func (a *Adminsrv) setUserMod() {
	a.modeUser.Store(true)
}

func (a *Adminsrv) setAdminMod() {
	a.modeUser.Store(false)
}

func (a *Adminsrv) SwapMode() {
	a.modeUser.Swap(!a.modeUser.Load())
}
func (a *Adminsrv) AdminMode() bool {
	return !a.modeUser.Load()
}




//TODO: edit templates. edit configs, edit usermsgs, make restarts
// after change of config, it should restart program

func (a *Adminsrv) RefreshMsgsession() error {
	return nil
}


func (a *Adminsrv) manage(Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {


	// Messagesession := botapi.NewMsgsession(a.botapi, a.ctrl.SudoAdmin, a.ctrl.SudoAdmin, "en")
	
	// //TODO: Create Function That construct below three function
	// alertsender := func(msg string) {
	// 	Messagesession.SendAlert(msg, nil)
	// }
	// sendreciver := func(msg any) (*tgbotapi.Message, error) {
	// 	_, err := Messagesession.Edit(msg, nil, "")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	mg, err := a.defaultsrv.ExcpectMsgContext(upx.Ctx, a.ctrl.SudoAdmin, a.ctrl.SudoAdmin)
	// 	if err == nil {
	// 		Messagesession.Addreply(mg.MessageID)
	// 	}
	// 	return mg, err
	// }
	// callbackreciver := func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error) {
	// 	_, err := Messagesession.Edit(msg, btns, "")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return a.callback.GetcallbackContext(upx.Ctx, btns.ID())
	// }
	callbackreciver := calls.Callbackreciver
	sendreciver := calls.Sendreciver
	alertsender := calls.Alertsender
	
	
	btns := botapi.NewButtons([]int16{2})


	var (
		state int
		callback *tgbotapi.CallbackQuery
		err error
	)

	mainloop:
	for {

		btns.Reset([]int16{2})

		switch state {
		case 0:
			btns.AddBtcommon("Change Config Settings")
			btns.AddBtcommon("Reset Usage")
			btns.Addbutton("🔴 Restart", "Restart", "")
			btns.AddClose(true)
			
			if callback, err = callbackreciver("select", btns); err != nil {
				alertsender("select within 1 minitue next time ")
				break mainloop
			}

			switch callback.Data {
			case "Reset Usage":

				calls.Alertsender("warning: If you Reset Usages New 30Days Cycle Begin From Here")
				reply, err := calls.Sendreciver("if you want to continue send ok")
				if err != nil {
					return err
				}
				if reply.Text != "ok" {
					calls.Alertsender("canceld usage reset")
					continue mainloop
				}
				calls.Alertsender("Usage Reset Added, If you want to undo this You have backup DB")
				a.ctrl.Addquemg(a.ctx, controller.ForceResetUsage(1))
				break mainloop

			case "Change Config Settings":
				alertsender("very carefull when you changing the config, if you make something wrong program will not restart correctly")
				state = 1
			case "Restart":
				Messagesession.DeleteAllMsg()
				err = sendSIGHUP()
				if err != nil {
					Messagesession.SendAlert("Restart Signal Sending Failed "+ err.Error(), nil)
				}
				break mainloop
			default:
				Messagesession.DeleteAllMsg()
				alertsender("not Available yet")
				break mainloop
			}
		case 1:
			config, err := os.ReadFile("config.json")
			if err != nil {
				Messagesession.SendAlert("config open err" + err.Error(), nil)
				state = 0
				continue
			}

			cont := "<code>" + string(config) + "</code>"

			btns.AddBtcommon("Replace")
			btns.AddBack(false)
			btns.AddClose(false)
			

			if callback, err = callbackreciver(botapi.Htmlstring(cont), btns); err != nil {
				fmt.Println(err)
				break mainloop 
			}

			switch callback.Data {
			case C.BtnClose:
				break mainloop
			case C.BtnBack:
				state = 0 
				continue
			case "Replace":
				//TODO:ADD warning
				newcont, err := sendreciver("warning: you must send correct config, if not program will not restart correctly, also you have to restart program to take effect new config,  send new config")
				if err != nil {
					break mainloop

				}
				file, err := os.OpenFile("config.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					Messagesession.SendAlert("file opening err " + err.Error(), nil)
				}
				file.Truncate(0)
				_, err = file.Write([]byte(newcont.Text))
				file.Close()
				if err != nil {
					Messagesession.SendAlert("file opening err " + err.Error(), nil)
				}
				Messagesession.SendAlert("you need to restart program to take effect", nil)
				state = 0

			}

		case 2:


		default:
			Messagesession.DeleteAllMsg()
			break mainloop
		
		}


	}
	
	return err

}


func sendSIGHUP() error{
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		return errors.New("Error finding process: " + err.Error())
	}
	return process.Signal(syscall.SIGHUP)
}