package service

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"github.com/sadeepa24/connected_bot/update/bottype"
	"github.com/sagernet/sing-vmess/vless"
	"go.uber.org/zap"
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

	



	adminuser db.User
	adminuserbtype bottype.User
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
	}

}

func (a *Adminsrv) Exec(upx *update.Updatectx) error {
	//Upx.User is nil in this scope
	//admin, ok :=- upx.User.IsAdmin

	upx.User = &a.adminuserbtype
	if upx.Update == nil {
		return nil
	}
	switch {
	case upx.Update.Message != nil:
		return a.handleMessage(upx)
	}

	return fmt.Errorf("admin exec not implemented")
}

func (a *Adminsrv) handleMessage(upx *update.Updatectx) error {
	//Upx.User is nil in this scope

	Messagesession := botapi.NewMsgsession(a.botapi, upx.FromChat().ID, upx.FromChat().ID, "en")

	switch {
	case upx.Update.Message.IsCommand():
		return a.Commandhandler(upx)
	case upx.Update.Message.ForwardFrom != nil:
		forward := upx.Update.Message.ForwardFrom
		_ = forward
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

func (a *Adminsrv) Commandhandler(upx *update.Updatectx) error {
	if upx.Update.Message == nil {
		return nil
	} 
	
	switch upx.Update.Message.Command() {
	case C.CmdUserInfo:
		return a.getuserinfo(upx)
	case C.CmdBrodcast:
		return a.broadcast(upx)
	case C.CmdServerInfo:
		return a.getserverinfo(upx)
	case C.CmdChatSession:
		return a.createchat(upx)
	case C.CmdOverview:
		return a.overview(upx)
	case C.CmdRefreshDb:
		a.ctrl.Addquemg(upx.Ctx, controller.RefreshSignal(1))
		upx.Cancle()
		return nil
	case "meanage":
		return a.meanage(upx)
	}

	return nil
}

func (a *Adminsrv) broadcast(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(a.botapi, upx.User.TgID, upx.User.TgID, upx.User.Lang) 
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

func (a *Adminsrv) getuserinfo(upx *update.Updatectx) error {
	
	

	
	Messagesession := botapi.NewMsgsession(a.botapi, a.ctrl.SudoAdmin, a.ctrl.SudoAdmin, "en")
	
	//TODO: Create Function That construct below three function
	alertsender := func(msg string) {
		Messagesession.SendAlert(msg, nil)
	}
	sendreciver := func(msg any) (*tgbotapi.Message, error) {
		_, err := Messagesession.Edit(msg, nil, "")
		if err != nil {
			return nil, err
		}
		mg, err := a.defaultsrv.ExcpectMsgContext(upx.Ctx, a.ctrl.SudoAdmin, a.ctrl.SudoAdmin)
		if err == nil {
			Messagesession.Addreply(mg.MessageID)
		}
		return mg, err
	}
	callbackreciver := func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error) {
		_, err := Messagesession.Edit(msg, btns, "")
		if err != nil {
			return nil, err
		}
		return a.callback.GetcallbackContext(upx.Ctx, btns.ID())
	}







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
	endusermsg := botapi.NewMsgsession(a.botapi, enduserupx.User.TgID, enduserupx.User.TgID, "en")

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
				alertsender: alertsender,
				sendreciver: sendreciver,
				callbackreciver: callbackreciver,
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

func (a *Adminsrv) createchat(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(a.botapi, upx.User.TgID, upx.User.TgID, upx.User.Lang)

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
			if err != nil {
				if admin {
					a.ctrl.Addquemg(context.Background(), &botapi.Msgcommon{
						Infocontext: &botapi.Infocontext{
							ChatId: upx.User.TgID,
						},
						Text: "chat session ended",
					})
				}
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

//TODO: edit templates. edit configs, edit usermsgs, make restarts
// after change of config, it should restart program

func (a *Adminsrv) RefreshMsgsession() error {
	return nil
}


func (a *Adminsrv) meanage(upx *update.Updatectx) error {

	upx.Ctx, upx.Cancle = context.WithTimeout(a.ctx, 30 * time.Minute)
	defer upx.Cancle() // user should cancel session


	Messagesession := botapi.NewMsgsession(a.botapi, a.ctrl.SudoAdmin, a.ctrl.SudoAdmin, "en")
	
	//TODO: Create Function That construct below three function
	alertsender := func(msg string) {
		Messagesession.SendAlert(msg, nil)
	}
	sendreciver := func(msg any) (*tgbotapi.Message, error) {
		_, err := Messagesession.Edit(msg, nil, "")
		if err != nil {
			return nil, err
		}
		mg, err := a.defaultsrv.ExcpectMsgContext(upx.Ctx, a.ctrl.SudoAdmin, a.ctrl.SudoAdmin)
		if err == nil {
			Messagesession.Addreply(mg.MessageID)
		}
		return mg, err
	}
	callbackreciver := func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error) {
		_, err := Messagesession.Edit(msg, btns, "")
		if err != nil {
			return nil, err
		}
		return a.callback.GetcallbackContext(upx.Ctx, btns.ID())
	}


	_ = sendreciver



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
			btns.AddBtcommon("Singbox Config Change")
			btns.AddBtcommon("edit templates")
			btns.AddBtcommon("edit usermsg")
			
			btns.Addbutton("🔴 Restart", "Restart", "")
			btns.AddClose(true)
			
			if callback, err = callbackreciver("select", btns); err != nil {
				alertsender("select within 1 minitue next time ")
				break mainloop
			}

			switch callback.Data {
			case "Change Config Settings":
				alertsender("very carefull when you changing the config, if you make something wrong program will not restart correctly")
				state = 1
			// case "edit templates":
			// 	state = 2
			// case "edit usermsg":
			// 	state = 3
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
				newcont, err := sendreciver("warning: you must send correct config if program will not restart correctly, also you have to restart program to take effect new config  send new config")
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
			//a.msgstore.


		case 3:

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