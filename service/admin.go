package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/builder/v2"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"github.com/sadeepa24/connected_bot/tg/update/bottype"
	"github.com/sadeepa24/walker"
	"github.com/sagernet/sing-box/connectedbot/opts"
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

	inspecifcUse *atomic.Bool  // remove in future after implmenting callback with id
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
		inspecifcUse: new(atomic.Bool),
	}

}

func (a *Adminsrv) Exec(upx *update.Updatectx) error {
	//Upx.User is nil in this scope
	//admin, ok :=- upx.User.IsAdmin
	upx.User = &a.adminuserbtype
	if upx.Update == nil {
		return nil
	}
	upx.Ctx, upx.Cancle  = context.WithTimeout(a.ctrl.GetBaseContext(), 30 * time.Minute) //admin has more time to deal with things
	switch {
	case upx.Update.Message != nil:
		return a.handleMessage(upx)
	}

	return errors.New("admin exec not implemented")
}
func (a *Adminsrv) handleMessage(upx *update.Updatectx) error {
	//Upx.User is nil in this scope
	Messagesession := botapi.NewMsgsession(upx.Ctx, a.botapi, upx.FromChat().ID, upx.FromChat().ID, "en")
	switch {
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
	case upx.Update.Message.IsCommand():
		return a.Commandhandler(upx, Messagesession)
	case upx.Update.Message.ForwardFrom != nil:
		forward := upx.Update.Message.ForwardFrom
		_ = forward
		//TODO: implement later
		upx.Cancle() //emove when implimeting
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
			if msg != nil {
				if _, err := Messagesession.Edit(msg, nil, ""); err != nil {
					return nil, err
				}
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
		return a.broadcast(upx, Messagesession)
	case C.CmdServerInfo:
		return a.getserverinfo()
	case C.CmdChatSession:
		return a.createchat(upx, Messagesession, calls)
	case C.CmdOverview:
		return a.overview(calls)
	case C.CmdRefreshDb:
		a.ctrl.Addquemg(controller.RefreshSignal(1))
		upx.Cancle()
		return nil
	case C.CmdFClose:
		return a.fclose(Messagesession, calls)
	case "users":
		return a.userLists(upx, Messagesession, calls)
	case "vpnconfig", "editconf":
		return a.vpnConfig(Messagesession, calls)
	case "activeconfs":
		return a.activeUserStatus(upx, Messagesession, calls)
	case "manage":
		return a.manage(Messagesession, calls)
	case "template":
		return a.editTemplate(upx, Messagesession, calls)
	case "botconfig":
		return a.botconfig(Messagesession, calls)
	case "restart":
		a.Restart(Messagesession)
	default:
		upx.Cancle()
	}

	return nil
}




func (a *Adminsrv) broadcast(upx *update.Updatectx, Messagesession *botapi.Msgsession) error {
	btns := botapi.NewButtons([]int16{2})
	for _, btname := range a.ctrl.AvailableUserList() {
		btns.AddBtcommon(btname)
	}


	btns.AddClose(true)
	Messagesession.Edit("select target user type", btns, "")
	callback, err := a.callback.GetcallbackContext(upx.Ctx, btns.ID())
	if err != nil {
		return err
	}
	if callback.Data == C.BtnClose {
		Messagesession.Edit("Broadcast Canceled", nil, "")
		return nil
	}
	var userlist = []int64{}
	err = a.ctrl.GetUserList(callback.Data, &userlist) 
	if err != nil {
		Messagesession.EditText("fetching user list failed try again", nil)
		return err
	}
	if len(userlist) == 0 {
		Messagesession.EditText(" no user found", nil)
		return nil
	}
	Messagesession.Edit("send brodcast message", nil, "")
	message, err := a.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)
	if err != nil {
		return err
	}
	Messagesession.SendAlert("broadcasting message", nil,)
	Messagesession.Edit("📣", nil, "")


	var sendrfunc func(to int64, mgid int64) error
	if message.IsForwaded() {
		sendrfunc = func(to, mgid int64) error {
			return Messagesession.ForwardMgTo(to, mgid, int64(message.ForwardFromMessageID))
		}
	} else {
		sendrfunc = Messagesession.CopyMessageTo
	}

	var erroreduser int
	for _, user := range userlist {
		if err := sendrfunc(user, int64(message.MessageID)); err != nil {
			erroreduser++
		}
	}
	Messagesession.DeleteAllMsg()
	Messagesession.SendAlert(fmt.Sprintf("Broadcast Done, Message Sent successfully to %d users from %d", len(userlist)-erroreduser, len(userlist) )  , nil)

	return nil
}
func (a *Adminsrv) getuserinfo(upx *update.Updatectx, Messagesession *botapi.Msgsession, calls common.Tgcalls) error {
	message, err := calls.Sendreciver("send user id or username")
	if err != nil {
		return err
	}
	defer Messagesession.DeleteAllMsg()
	return a.loaduserinfo(upx, Messagesession, calls, message.Text)
}
func (a *Adminsrv) getserverinfo() error {
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
	a.ctrl.Addquemg( &botapi.Msgcommon{
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
		calls.Alertsender("chat creation failed" + err.Error())
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
	a.ctrl.Addquemg(&botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId: dbuser.TgID,
		},
		Text: "admin created chat session with you, you can't use any command or anything until he ends the session ",
	})
	calls.Alertsender("chat started if you want to end chat you must /cancel session, if not it will hold for 30 min")
	chatctx, chatcancel := context.WithTimeout(a.ctx, 30 * time.Minute)
	Messagesession.SetNewcontext(chatctx)

	mgcoping := func (src, dst int64, admin bool)  {
		mgcopy:
		for {
			select {
			case <- chatctx.Done():
				break mgcopy
			default:
			}
			mg, err := a.defaultsrv.ExcpectMsgContext(chatctx, src, src) // this will check context automatically
			if err != nil{
				if admin {
					a.ctrl.Addquemg(&botapi.Msgcommon{
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
					chatcancel()
					break
				}
			}
			Messagesession.CopyMessageRawTo(dst, int64(mg.MessageID), src)
		}
		if !admin {
			a.ctrl.Addquemg(&botapi.Msgcommon{
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
func (a *Adminsrv) overview(calls common.Tgcalls ) error {
	calls.Alertsender(a.ctrl.Overview.String() + fmt.Sprintf("Updates Since Last Refresh: %d", a.ctrl.UpdateCounter.Load()))
	return nil
}
func (a *Adminsrv) SwapMode() {
	a.modeUser.Store(!a.modeUser.Load()) //this is okay with this, no heavy concurrent calls to this
}
func (a *Adminsrv) AdminMode() bool {
	return !a.modeUser.Load()
}
func(a *Adminsrv) Restart(Messagesession *botapi.Msgsession) {
	err := common.SendSIGHUP()
	if err != nil {
		Messagesession.SendAlert("Restart Signal Sending Failed "+ err.Error(), nil)
	}
}



const (
	btnpereach = 16
)
var editable = []string{
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
	"btnconf",
}


func (a *Adminsrv) loaduserinfo(upx *update.Updatectx, Messagesession *botapi.Msgsession, calls common.Tgcalls, IDorUserName string) error {
	if a.inspecifcUse.Swap(true) {
		calls.Alertsender("Already in getuser close it first and use this")
		return nil
	}
	defer a.inspecifcUse.Swap(false)
	alertsender := calls.Alertsender
	sendreciver := calls.Sendreciver
	callbackreciver := calls.Callbackreciver
	enduserupx := update.Updatectx{
		User: &bottype.User{},
		Ctx: upx.Ctx,
		Cancle: upx.Cancle,
	}

	id, err := strconv.Atoi(IDorUserName)
	if err != nil {
		IDorUserName = strings.ReplaceAll(IDorUserName, "@", "")
		enduserupx.User.User, err = a.ctrl.GetUserByUserName(IDorUserName)
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
			btns.AddBtcommon("User Settings")
			btns.AddBtcommon("Builder")
			btns.AddBtcommon("Configure")
			btns.AddBtcommon("Gifts")	
			btns.AddClose(true)

			callback, err = callbackreciver("select", btns)
			if err != nil {
				break main
			}

			switch callback.Data {
			case "User Settings":
				state = 1
			case "Builder":
				state = 4
			case "Configure":
				state = 5
			case "User Info":
				state = 6
			case "Gifts":
				state = 7
			case C.BtnClose:
				break main
			}
		case 1:
			if enduserupx.User.Restricted {
				btns.Addbutton("🟢 Remove Restrict ", "res", "")
				btns.Addbutton("Restrict Reason", "reason", "")
			} else {
				btns.Addbutton("🔴 Restrict User ",  "res","" )
			}
			if enduserupx.User.IsMonthLimited {
				btns.AddBtcommon("Remove Monthlimit")
			}
			if enduserupx.User.Templimited {
				btns.AddBtcommon("Remove Templimit")
			}
			if enduserupx.User.IsCapped {
				btns.AddBtcommon("Remove Cap")
			}
			if enduserupx.User.IsDistributedUser {
				btns.Addbutton("Remove Distribute",  "Distribute","" )
			} else {
				btns.Addbutton("🔴 Distribute",  "Distribute","" )
			}
			btns.Addbutton("Add Additional Bandwidth", "addbw", "")
			if enduserupx.User.AdditionalQuota > 0 {
				btns.Addbutton("Remove Additional Bandwidth", "rembw", "")
			}
			btns.AddBtcommon("Change Point Count")
			if enduserupx.User.Verified() {
				btns.AddBtcommon("Create Config")
			}
			btns.Passline()
			btns.Addbutton("🔴🔴 Advanced Change 🔴🔴", "Advanced Change", "")
			btns.AddCloseBack()
			tusage := endusersession.TotalUsage()
			Messagesession.Edit(userinfo{
				CappedQuota: enduserupx.User.CappedQuota.BToString(),
				IsTemplimited: enduserupx.User.Templimited,
				TempLimitRate: enduserupx.User.WarnRatio,
				IsVerified: enduserupx.User.Verified(),
				Paused: enduserupx.User.IsPaused,
				CommonUser: &botapi.CommonUser{
					Name:     enduserupx.User.Name,
					TgId:     enduserupx.User.TgID,
					Username: enduserupx.User.User.Username,
				},
				NonUseCycle: upx.User.EmptyCycle,
				UsagePercentage: ((tusage * 100)/(endusersession.GetUser().CalculatedQuota)).Float64(),
				GiftQuota: enduserupx.User.GiftQuota.BToString(),
				Joined:    enduserupx.User.Joined.Format("2006-01-02 15:04:05"),
				Dedicated: a.ctrl.CommonQuota.Load().BToString(),
				TQuota:    (endusersession.GetUser().CalculatedQuota).BToString(),
				LeftQuota: endusersession.LeftQuota().BToString(),
				TUsage:    tusage.BToString(),
				ConfCount: endusersession.GetUser().ConfigCount,
				CapEndin:  enduserupx.User.Captime.AddDate(0, 0, int(enduserupx.User.CapDays)).String(),
				CapDays: enduserupx.User.CapDays,
				AlltimeUsage: (upx.User.AlltimeUsage+tusage).BToString(),
				Points: enduserupx.User.Points,
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
				if enduserupx.User.IsDistributedUser {
					endusersession.GetUser().IsDistributedUser = false
					err = endusersession.ActivateAll()
					if err != nil {
						calls.Alertsender("activation failed " + err.Error())
						endusersession.DeactivateAll()
						endusersession.GetUser().IsDistributedUser = true
						continue
					}
					alertsender("distribution removed")
					endusermsg.SendAlert("you'r distribution removed by admin", nil)
					continue
				}
				if enduserupx.User.IsCapped {
					alertsender("can't distribute, user is capped")
					continue
				}
				endusersession.GetUser().IsDistributedUser = true
				endusersession.DeactivateAll()
				endusermsg.SendAlert("you'r quota has being Distributed By Admin ", nil)
			case "res":
				if endusersession.GetUser().Restricted {
					endusersession.RemoveRestrict()
					if err != nil {
						calls.Alertsender("failed: " + err.Error())
						continue
					}
					endusermsg.SendAlert("✅ admin removed you'r restriction, you can use service again 🎉", nil)
					Messagesession.SendAlert("make a db refresh to change bandiwdth, it will automatically change in next refresh cycle", nil)
				
				} else {
					reason, err := calls.Sendreciver("send reason for restriction")
					if err != nil {
						return err
					}
					err = endusersession.Restrict(reason.Text)
					if err != nil {
						calls.Alertsender("failed: " + err.Error())
						continue
					}
					endusermsg.SendAlert("🔴 you have restricted by admin you may have to contact admin to remove this restriction ", nil)
					Messagesession.SendAlert("make a db refresh to change bandiwdth, it will automatically change in next refresh cycle", nil)
				}
			case "reason":
				restr, err := endusersession.RestrictInfo()
				if err != nil {
					calls.Alertsender("fetch err: " + err.Error())
					continue
				}
				calls.Alertsender(restr.Reason)
			case "Remove Monthlimit":
				endusersession.GetUser().IsMonthLimited = false
				err = endusersession.ActivateAll()
				if err != nil {
					Messagesession.SendAlert("config activation failed" + err.Error(), nil)
					continue
				}
				endusermsg.SendAlert("🎉you'r monthlimitation removed by admin 🍾", nil)
				Messagesession.SendAlert("make a db refresh to change bandiwdth, it will automatically change in next refresh cycle", nil)
			case "Remove Templimit":
				endusersession.GetUser().Templimited = false
				endusersession.GetUser().WarnRatio = a.ctrl.GetWarnRate()
				err = endusersession.ActivateAll()
				if err != nil {
					Messagesession.SendAlert("config activation failed" + err.Error(), nil)
					continue
				}
				endusermsg.SendAlert("🎉 Your temporary limitation has been removed and warning rate reset by admin 🍾", nil)
				Messagesession.SendAlert("Temporary limitation removed and warning rate reset successfully", nil)
			case "Remove Cap":
				enduserupx.User.IsCapped = false
				enduserupx.User.CappedQuota = 0
				if err = a.ctrl.RecalculateConfigquotas(endusersession.GetUser()); err != nil {
					Messagesession.SendAlert(C.GetMsg(C.MsgcapRecalFail), nil)
				}
				endusermsg.SendAlert("🎉 Your Cap Removed By Admin🍾", nil)	
			case "Create Config":
				opts := common.OptionExcutors{
					Tgcalls: calls,
					Btns: btns,
					Usersession: endusersession,
					MessageSession: Messagesession,
					Ctrl: a.ctrl,
					Logger: a.logger,

				}
				err := CreateConfig(opts)
				if err != nil {
					Messagesession.SendAlert(err.Error(), nil)
				}
				endusermsg.SendAlert("Admin Created Config For you Check /getinfo to see it", nil)
			case "Change Point Count":
				pt, err := common.ReciveInt(calls, 10000, 0)
				if err != nil {
					Messagesession.SendAlert(err.Error(),  nil)
				}
				endusersession.GetUser().Points = int64(pt)
				endusermsg.SendAlert("admin changed you'r point count to" + strconv.Itoa(pt), nil)
			case "Advanced Change":
				Messagesession.SendAlert("⚠️ Proceed with Caution! 🔴 Modifying these values can have serious consequences. Only make changes if you are absolutely certain of what you’re doing.  🔒 Some fields are protected and must never be altered. ", nil)
				Messagesession.ResetState()
				time.Sleep(1 * time.Second)
				pre := botapi.PreBuuf{
					Buffer: &bytes.Buffer{},
				}
				conec := builder.NewConnector(calls)
				wlkr, err := walker.NewWalker(endusersession.GetUser())
				if err != nil {
					Messagesession.SendAlert("walker create failed: " + err.Error(), nil)
					continue main
				}
				wlkr.SetValue = setvaluefunc(conec, nil)
				wlkr.CanSetCheck = func(val reflect.Value, nextItemPath string, wlkr *walker.Walker) bool {
					return false
				}
				err = builder.AnyFieldChange(wlkr, endusersession.GetUser(), conec, func(item any) any {
					itm, ok := wlkr.CurrentPtrIface()
					if ok {
						pre.Reset()
						enc := json.NewEncoder(&pre)
						enc.SetIndent("", " ")
						err = enc.Encode(itm)
						if err != nil {
							return "json Encode Err " + err.Error()
						}
						if pre.Len() == 0 {
							return "marshall zero"
						}
						return &pre
					} else {
						return "current item cannot be marshal"
					}
				}, a.logger)
				if err != nil {
					Messagesession.SendAlert("field change got err: " + err.Error(), nil)
					return err
				}
				endusersession.RecalculateConfigquotas()
				endusersession.Save()
				endusersession.DeactivateAll()
				err = endusersession.ActivateAll()
				if err != nil {
					Messagesession.SendAlert("config reactivate failed: "+ err.Error(), nil)
				}
			case "addbw":
				btns.Reset([]int16{1})
				btns.Addbutton("Add Bandwidth", "addbw", "")
				btns.AddCloseBack()
				Messagesession.Edit("send additional bandwidth", btns, "")
				callback, err = a.callback.GetcallbackContext(upx.Ctx, btns.ID())
				if err != nil {
					return err
				}
				if callback.Data == C.BtnClose {
					Messagesession.SendAlert("add bandwidth canceled", nil)
					continue main
				}
				if callback.Data == C.BtnBack {
					state = 1
					continue main
				}
				Messagesession.Edit("send additional bandwidth", nil, "")
				bw, err := common.ReciveBandwidth(calls, a.ctrl.BandwidthAvelable, 0)
				if err != nil {
					Messagesession.SendAlert("bandwidth recive failed: " + err.Error(), nil)
					continue main
				}
				if bw <= 0 {
					Messagesession.SendAlert("bandwidth must be greater than 0", nil)
					continue main
				}
				if err = endusersession.Addadditional(bw.GbtoByte()); err != nil {
					Messagesession.SendAlert("bandwidth add failed: " + err.Error(), nil)
					continue main
				}
				endusermsg.SendAlert("admin added additional bandwidth of " + bw.GbtoByte().BToString() + " to you'r account", nil)
				Messagesession.SendAlert("make a db refresh to change bandiwdth, it will automatically change in next refresh cycle", nil)
			case "rembw":
				if err = endusersession.Removeadditional(); err != nil {
					Messagesession.SendAlert("bandwidth remove failed: " + err.Error(), nil)
					continue main
				}
				Messagesession.SendAlert("done", nil) 
				endusermsg.SendAlert("admin removed you'r additional bandwidth", nil)
				Messagesession.SendAlert("make a db refresh to change bandiwdth, it will automatically change in next refresh cycle", nil)
			}
		case 4:
			Messagesession.ResetState()
			if _, ok := a.xraywiz.builds.LoadOrStore(enduserupx.User.TgID, true); ok {
				Messagesession.SendAlert("user have already opend a builder session please wait until he closes it", nil)
				state = 0 
				continue
			}
			build := &confBuilder{
				Tgcalls: calls,
				disableTimeout: true,
				ctx: upx.Ctx,
				State: 0,
				Messagesession: Messagesession,
				btns: btns,
				wiz: a.xraywiz,
				store: a.xraywiz.bulderstore,
				counter: new(atomic.Int32),
				userId: enduserupx.Dbuser().TgID,
			}
			build.run()
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

			err = configState.run()
			if err != nil {
				a.logger.Error("admin: configure run failed ", zap.Error(err))
				if errors.Is(err, C.ErrContextDead) {
					tmpctx, cancle := context.WithTimeout(a.ctx, 10*time.Second)
					Messagesession.SetNewcontext(tmpctx)
					Messagesession.SendAlert("context timeouts", nil)
					cancle()
					break main
				}
			}
			state = 0
		case 6:
			userinfost := &getinfo{
				state:  0,
				callback: a.callback,
				Messagesession: Messagesession,
				calls: common.Tgcalls{
					Alertsender: alertsender,
					Sendreciver: sendreciver,
					Callbackreciver: callbackreciver,
				},
				btns: btns,
				ctx: upx.Ctx,
				dbuser: endusersession.GetUser(),
				ctrl: a.ctrl,
				Usersession: endusersession,
			}
			err = userinfost.run()
			if err != nil {
				return err
			}
			state = 0
		case 7:
			gifts, err := endusersession.AllGifts(true)
			if err != nil || len(gifts) == 0  {
				if err != nil {
					Messagesession.SendAlert("fetch failed: " + err.Error(), nil)
				} else {
					Messagesession.SendAlert("no gifts", nil)
				}
				state = 0
				continue main
			}
			for i := range gifts {
				btns.AddBtcommon(strconv.Itoa(i))
			}
			btns.AddCloseBack()
			callback, err = calls.Callbackreciver("select option", btns)
			if err != nil {
				return err
			}
			
			switch callback.Data {
			case C.BtnClose:
				return nil
			case C.BtnBack:
				state = 0
				continue main
			default:
				giftnum, err := strconv.Atoi(callback.Data)
				if err != nil {
					continue main
				}
				thgift := gifts[giftnum]
				btns.Reset([]int16{1})
				if thgift.Sender == endusersession.GetUser().TgID {
					btns.AddBtcommon("cancel gift")
				}
				btns.AddBack(false)
				callback, err = calls.Callbackreciver(fmt.Sprintf(" bandwidthh: %s",  thgift.Bandwidth.BToString()), btns)
				if err != nil {
					return err
				}
				if callback.Data != "cancel gift" {
					continue main
				}
				if gifts[giftnum].Sender != endusersession.GetUser().TgID {
					calls.Alertsender("cannot cancel recived gift")
					continue main
				}
				err = a.ctrl.CancelGift(gifts[giftnum], endusersession.GetUser())
				if err != nil {
					calls.Alertsender("gift cancel failed: " + err.Error())
					continue main
				}
			}
		}
	}


	return nil
}

// after change of config, it should restart program
func (a *Adminsrv) activeUserStatus(upx *update.Updatectx, Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {
	alluserconfigs := a.ctrl.Boxapi.GetAllUserStatus()
	activeusr := C.SliceToMap(alluserconfigs, func(u opts.UserStatus) int  {
		return u.UserID
	})
	btns := botapi.NewButtons([]int16{2})
	
	var (
		callback *tgbotapi.CallbackQuery
		err error
		online map[int]opts.UserStatus
		onlineIDS []int
	)
	m:
	for {
		btns.Reset([]int16{2})
		btns.AddBtcommon("Online")
		btns.AddBtcommon("All")
		btns.AddClose(false)

		callback, err = calls.Callbackreciver("select option", btns)
		if err != nil {
			break
		}

		switch callback.Data {
		case "Online":
			online = make(map[int]opts.UserStatus, len(activeusr))
			if onlineIDS == nil {
				onlineIDS = make([]int, len(activeusr))
			}
			onlineIDS = onlineIDS[:0]
			for id, user := range activeusr {
				if len(user.Ip) > 0 {
					online[id] = user
					onlineIDS = append(onlineIDS, id)
				}
			}
		case "All":
			online = activeusr
			onlineIDS = C.MapToSliceKey(activeusr)
		case C.BtnClose:
			break m
		}

		if len(online) == 0 {
			calls.Alertsender("no active conf can be found")
			continue
		}

	
		pagecount := len(onlineIDS)/btnpereach
		currentPage := 1
		configs:
		for {
			btns.Reset([]int16{2})
			to := currentPage*btnpereach
			if to > len(onlineIDS) {
				to = (currentPage-1)*btnpereach + len(onlineIDS)%btnpereach
			}
			for i := (currentPage-1)*btnpereach; i < to; i++ {
				btns.AddBtcommon(strconv.Itoa(onlineIDS[i]))
			}
			
			if currentPage < pagecount || (currentPage == pagecount &&  len(onlineIDS)%btnpereach > 0) {
				btns.AddBtcommon(C.BtnNext)
			}
			btns.AddCloseBack()
			callback, err = calls.Callbackreciver("select config to see info", btns)
			if err != nil {
				return err
			}
			switch callback.Data {
			case C.BtnNext:
				currentPage++
			case C.BtnClose:
				break configs
			case C.BtnBack:
				if currentPage == 1 {
					break configs
				}
				currentPage--
			default:
				id, err := strconv.Atoi(callback.Data)
				if err != nil {
					continue
				}
				user, err := a.ctrl.GetUserByConfID(int64(id))
				if err != nil {
					calls.Alertsender("user loading failed " + err.Error())
					continue
				}
				err = a.loaduserinfo(upx, Messagesession, calls, strconv.Itoa(int(user.TgID)))
				if err != nil {
					if errors.Is(err, C.ErrContextDead) {
						return err
					}
					a.logger.Error("failed loaduser ", zap.Error(err))
					calls.Alertsender("user loading failed " + err.Error())
				}
				Messagesession.ResetState()
			}
		}
	}
	Messagesession.DeleteAllMsg()
	return err


}
func (a *Adminsrv) manage(Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {
	callbackreciver := calls.Callbackreciver
	alertsender := calls.Alertsender
	btns := botapi.NewButtons([]int16{2})
	var (
		callback *tgbotapi.CallbackQuery
		err error
	)
	mainloop:
	for {

		btns.Reset([]int16{2})
		btns.Addbutton("🔴 Reset Usage", "reset-usage", "")
		btns.Addbutton("🔴 Restart", "Restart", "")
		btns.Addbutton("🔴 Remove MonthLimitations", "remlimit", "")
		btns.Addbutton("🔴 Remove All Restriction", "remrestriction", "")
		btns.Addbutton("Reset Lang Codes", "langchg", "")
		btns.Addbutton("Refresh Config", "refconf", "")
		btns.AddClose(true)
		
		if callback, err = callbackreciver("select", btns); err != nil {
			break mainloop
		}

		switch callback.Data {
		case "reset-usage":
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
				a.ctrl.Addquemg(controller.ForceResetUsage(1))
				break mainloop
		case "Restart":
				Messagesession.DeleteAllMsg()
				err = common.SendSIGHUP()
				if err != nil {
					Messagesession.SendAlert("Restart Signal Sending Failed "+ err.Error(), nil)
				}
				break mainloop
		case "remlimit":
				
				btns.Reset([]int16{2})
				calls.Alertsender("🔴 Please be cautious! These are critical changes and should be performed with utmost care. 🔴 do thease if you realy want")
				btns.Addcancle()
				btns.AddBtcommon("proceed")
				if callback, err = callbackreciver("this will remove everyone's monthlimitations do you want to continue ?", btns); err != nil {
					break mainloop
				}

				if callback.Data != "proceed" {
					continue
				}
				Messagesession.DeleteAllMsg()
				err = a.ctrl.RemoveAllLimits()
				if err != nil {
					Messagesession.SendAlert("db update err "+ err.Error(), nil)
					continue
				}
				a.ctrl.Addquemg( &botapi.Msgcommon{
					Infocontext: &botapi.Infocontext{
						User_id: a.ctrl.SudoAdmin,
						ChatId: a.ctrl.SudoAdmin,
					},
					Text: "Month Limitation Remove And Db Refreshed",
				})
				return nil
		case "langchg":
								
				btns.Reset([]int16{2})
				btns.Addcancle()
				btns.AddBtcommon("proceed")
				if callback, err = callbackreciver("this will reset everyone's lang code to default lang code", btns); err != nil {
					break mainloop
				}

				if callback.Data != "proceed" {
					continue
				}
				Messagesession.DeleteAllMsg()
				err = a.ctrl.ResetLangCode()
				if err != nil {
					Messagesession.SendAlert("db update err "+ err.Error(), nil)
					continue
				}
				a.ctrl.Addquemg( &botapi.Msgcommon{
					Infocontext: &botapi.Infocontext{
						User_id: a.ctrl.SudoAdmin,
						ChatId: a.ctrl.SudoAdmin,
					},
					Text: "lang code changed",
				})
				return nil
		case "refconf":
			Messagesession.DeleteAllMsg()
			if err = a.ctrl.RefreshAllConfig(); err != nil {
				Messagesession.SendAlert("config refresh err: " +  err.Error(), nil)
			}
			return err
		case "remrestriction":
				btns.Reset([]int16{2})
				calls.Alertsender("🔴 Please be cautious! These are critical changes and should be performed with utmost care. 🔴 do thease if you realy want")
				btns.Addcancle()
				btns.AddBtcommon("proceed")
				if callback, err = callbackreciver("this will remove all user restriction do you want to continue ?", btns); err != nil {
					break mainloop
				}

				if callback.Data != "proceed" {
					continue
				}
				Messagesession.DeleteAllMsg()
				err = a.ctrl.RemoveAllRestriction()
				if err != nil {
					Messagesession.SendAlert("db update err "+ err.Error(), nil)
					continue
				}
				a.ctrl.Addquemg( &botapi.Msgcommon{
					Infocontext: &botapi.Infocontext{
						User_id: a.ctrl.SudoAdmin,
						ChatId: a.ctrl.SudoAdmin,
					},
					Text: "Restrictiones Removed And Db Refreshed",
				})
				return nil
		case C.BtnClose:
				Messagesession.DeleteAllMsg()
				alertsender("manager closed")
				break mainloop
		default:
				Messagesession.DeleteAllMsg()
				alertsender("not Available yet")
				break mainloop
		}
	}
	return err

}
//FIXME: 
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
	path := a.msgstore.GetPath()


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
		rep, err := calls.Sendreciver("you'r current template will save,  do you want to continue, then send ok")
		if err != nil {
			return
		}
		if rep.Text != "ok" {
			calls.Alertsender("All changes you made will not be saved.")
			Messagesession.DeleteAllMsg()
			return
		}
		calls.Alertsender("Editor Closed, Saving Template..")
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
		Messagesession.DeleteAllMsg()
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
			btns.AddClose(true)
	
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
			if err != nil {
				calls.Alertsender("fileName Recive Error")
				return err
			}
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
			
			btns.Reset([]int16{2})
			btns.AddButtonSlice(a.ctrl.Langs)
			se, err := calls.Callbackreciver("select lang tag", btns)
			if err != nil {
				break selecttmpl
			}
			treplymg, err := calls.Sendreciver("Send MSG template content as string")
			if err != nil {
				return err
			}


			btns.Reset([]int16{2})
			btns.AddBtcommon("As Help Page")
			btns.AddBtcommon("As Inline Post")
			btns.AddBtcommon("Default")
			callback, err  = calls.Callbackreciver("select option", btns)
			if err != nil {
				return err
			}
			var templatename string
			switch callback.Data {
			case "As Help Page":
				btns.Reset([]int16{2})
				
				btns.AddBtcommon(C.TmpHelpInfoPage)
				btns.AddBtcommon(C.TmplHelpBuilderHelp)
				btns.AddBtcommon(C.TmplHelpTuto)
				btns.AddBtcommon(C.TmpHelpCmPage)

				callback, err = calls.Callbackreciver("select help page type ", btns) 
				if err != nil {
					return err
				}
				help := *a.ctrl.GetHelepCmdInfo()
				
				switch callback.Data {
				case C.TmpHelpInfoPage:
					help.InfoPageCount++
					templatename = C.TmpHelpInfoPage+ strconv.Itoa(int(help.InfoPageCount))
				case C.TmplHelpBuilderHelp:
					help.BuilderHelp++
					templatename = C.TmplHelpBuilderHelp+ strconv.Itoa(int(help.BuilderHelp))
				case C.TmplHelpTuto:
					help.TutorialPageCount++
					templatename = C.TmplHelpTuto+ strconv.Itoa(int(help.TutorialPageCount))
				case C.TmpHelpCmPage:
					help.CommandPageCount++
					templatename = C.TmpHelpCmPage+ strconv.Itoa(int(help.CommandPageCount))
				}

				Templates[templatename] = map[string]*botapi.MgItem{
					se.Data: {
							Msgtmpl: treplymg.Text,
						},
				}

				file, err := os.OpenFile("./config.json", os.O_RDWR, 0644)
				if err != nil {
					calls.Alertsender("config.json file open error " + err.Error())
					break
				}
				cont, err := io.ReadAll(file)
				if err != nil {
					calls.Alertsender("config.json file read err: " + err.Error())
					file.Close()
					break
				}
				var cf C.Botoptions
				err = json.Unmarshal(cont, &cf)
				if err != nil {
					calls.Alertsender("config.json file unmarshal json err: " + err.Error())
					file.Close()
					break
				}
				if cf.Metadata != nil {
					cf.Metadata.HelperInfo = help
				}

				bt, err := json.Marshal(cf)
				if err == nil {
					file.Seek(0, 0)
					n, err := file.Write(bt)
					if err != nil {
						calls.Alertsender("config.json file write err: " + err.Error())
						file.Close()
						break
					}
					file.Truncate(int64(n))
				}
				file.Close()
				calls.Alertsender("✅ New Help template created successfully! (need restart)")
			case "Default":
				replymg, err = calls.Sendreciver("Send Name for New template")
				if err != nil {
					return err
				}
				templatename = replymg.Text
				if _, ok := Templates[replymg.Text]; ok {
					calls.Alertsender("there is alredy template with this name try again")
					continue selecttmpl
				}
				btns.Reset([]int16{2})
				btns.AddButtonSlice(a.ctrl.Langs)
				Templates[replymg.Text] = map[string]*botapi.MgItem{
					se.Data: {
						Msgtmpl: treplymg.Text,
					},
				}
				if callback.Data == "Default" {
					break
				}
				fallthrough
			case "As Inline Post":
				file, err := os.OpenFile("./config.json", os.O_RDWR, 0644)
				if err != nil {
					calls.Alertsender("config.json file open error " + err.Error())
					break
				}
				cont, err := io.ReadAll(file)
				if err != nil {
					calls.Alertsender("config.json file read err: " + err.Error())
					file.Close()
					break
				}
				var cf C.Botoptions
				err = json.Unmarshal(cont, &cf)
				if err != nil {
					calls.Alertsender("config.json file unmarshal json err: " + err.Error())
					file.Close()
					break
				}
				var post C.InlinePost
				post.Name = templatename
				replymg, err = calls.Sendreciver("Send Discription for the inline post")
				if err != nil {
					return err
				}
				post.Dis = replymg.Text
				replymg, err = calls.Sendreciver("Send Title for the inline post")
				if err != nil {
					return err
				}
				post.Title = replymg.Text
				if cf.Metadata != nil {
					cf.Metadata.InlinePost = append(cf.Metadata.InlinePost, post)
				}
				bt, err := json.Marshal(cf)
				if err == nil {
					file.Seek(0, 0)
					n, err := file.Write(bt)
					if err != nil {
						calls.Alertsender("config.json file write err: " + err.Error())
						file.Close()
						break
					}
					file.Truncate(int64(n))
				}
				file.Close()
				calls.Alertsender("✅ New template created successfully! (need restart)")
			}
			state = 0

			nameslice = append(nameslice, templatename)
			maxpages = len(nameslice)/btnpereach		
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
	
	
			var mode = "prv"
			
			buf := &botapi.PreBuuf{
				Buffer: &bytes.Buffer{},
			}
			var msghook = func (original *botapi.MgItem) any {
				switch mode {
				case "dt":
					buf.Reset()
					dc := json.NewEncoder(buf)
					dc.SetIndent("", " ")
					err = dc.Encode(original)
					if err != nil {
						buf = nil
						return "Encoding Error"		
					}
					return buf
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
				case "btnconf":
					calls.Alertsender("example button := btn1[https://t.me]" )
					Messagesession.SendExtranal(botapi.Htmlstring(`Example Btn JSON Config : {\n\"btns\": [[\"line1btn1[https://t.me]\", \"line1btn1[https://t.me]\"],\n[\"line2btn1[https://t.me]\"],\n[\"line3btn1[https://t.me]\"\n]]}`), nil, "", true)
					fallthrough
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
	return nil
}
func (a *Adminsrv) vpnConfig(Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {
	calls.Alertsender("⚠️ Please proceed with caution! Even a small mistake can disrupt the system. Double-check before making changes.")
	build := &confBuilder{
		Tgcalls: calls,
		ctx: a.ctx,
		Messagesession: Messagesession,
		btns: botapi.NewButtons([]int16{2}),
		wiz: a.xraywiz,
		store: a.xraywiz.bulderstore,
		
		disableTimeout: true,
		singlemode: true,
		disableAutoSave: true,
		disableFieldSkip: true,
	}

	var err error
	build.Builder, err = builder.NewBuilderFromFile(build, a.ctrl.SboxConfPath())
	if err != nil {
		Messagesession.SendAlert("builder creare failed "+ err.Error(), nil)
		return nil
	}
	build.Builder.CallBack = func() builder.Buf {
		return &botapi.PreBuuf{
			Buffer: &bytes.Buffer{},
		}
	}
	build.Builder.Disableautosave = true
	calls.Alertsender("after doing all edits close the builder then you will be asked to save or cancel")
	time.Sleep(3 * time.Second)
	build.run()
	return nil
}
func (a *Adminsrv) botconfig(Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {
	calls.Alertsender("⚠️ Please proceed with caution! Even a small mistake can disrupt the system. Double-check before making changes.")
	
	conf := C.Botoptions{}
	f, err := os.ReadFile("config.json")
	if err != nil {
		Messagesession.SendAlert("file open error: " + err.Error(), nil)
		return err
	}
	err = json.Unmarshal(f, &conf)
	if err != nil {
		Messagesession.SendAlert("json unmarshal err: " + err.Error(), nil)
		return err
	}

	a.ctrl.IncCriticalOp()
	defer a.ctrl.DecCriticalOp()

	pre := botapi.PreBuuf{
		Buffer: &bytes.Buffer{},
	}
	conec := builder.NewConnector(calls)
	wlkr, _ := walker.NewWalker(&conf)
	wlkr.SetValue = setvaluefunc(conec, func(curval reflect.Value, path string, wlkr *walker.Walker, item any) (reflect.Value, bool) {
		if curval.Type().String() == "constbot.InlinePost" {
			return reflect.ValueOf(C.InlinePost{
				//TODO: fill here
			}), true
		}
		return reflect.Value{}, false
	})
	wlkr.CanSetCheck = cansetfunc(conec)

	err = builder.AnyFieldChange(wlkr, &conf, conec, func(item any) any {
		itm, ok := wlkr.CurrentPtrIface()
		if ok {
			pre.Reset()
			enc := json.NewEncoder(&pre)
			enc.SetIndent("", " ")
			err = enc.Encode(itm)
			if err != nil {
				return "json Encode Err " + err.Error()
			}
			return &pre
		} else {
			return "current item cannot be marshal"
		}
	}, a.logger)

	if err != nil {
		Messagesession.SendAlert("field change got err: " + err.Error(), nil)
		return err
	}

	rep, err := calls.Sendreciver("you'r current template will save,  do you want to continue, then send ok")
	if err != nil {
		Messagesession.SendAlert("field change got err: " + err.Error(), nil)
		return err
	}
	if rep.Text != "ok" {
		calls.Alertsender("All changes you made will not be saved.")
		return nil
	}
	file, err := os.OpenFile("config.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	m, err := json.MarshalIndent(&conf, "", " ")
	if err != nil {
		calls.Alertsender( "Encode Err " + err.Error() + "All changes you made will not be saved.")
		return nil
	}
	n, err := file.Write(m)
	if err != nil {
		calls.Alertsender( "Write Err " + err.Error() + "All changes you made will not be saved.")
		return nil
	}
	file.Truncate(int64(n))
	return nil
}

func (a *Adminsrv) userLists(upx *update.Updatectx, Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {
	btns := botapi.NewButtons([]int16{2})
	for _, btname := range a.ctrl.AvailableUserList() {
		btns.AddBtcommon(btname)
	}

	btns.AddClose(true)
	Messagesession.Edit("select target user type", btns, "")
	callback, err := a.callback.GetcallbackContext(upx.Ctx, btns.ID())
	if err != nil {
		return err
	}
	if callback.Data == C.BtnClose {
		Messagesession.Edit("Canceled", nil, "")
		return nil
	}

	var userlist = []int64{}
	err = a.ctrl.GetUserList(callback.Data, &userlist) 
	if err != nil {
		Messagesession.EditText("fetching user list failed try again", nil)
		return err
	}

	pagecount := len(userlist)/btnpereach
	currentPage := 1
	configs:
	for {
			btns.Reset([]int16{2})
			to := currentPage*btnpereach
			if to > len(userlist) {
				to = (currentPage-1)*btnpereach + len(userlist)%btnpereach
			}
			for i := (currentPage-1)*btnpereach; i < to; i++ {
				btns.AddBtcommon(strconv.Itoa(int(userlist[i])))
			}
			
			if currentPage < pagecount || (currentPage == pagecount &&  len(userlist)%btnpereach > 0) {
				btns.AddBtcommon(C.BtnNext)
			}
			if pagecount > 1 {
				btns.AddCloseBack()
			} else {
				btns.Addbutton(C.BtnClose, C.BtnBack, "")
			}
			callback, err = calls.Callbackreciver("select config to see info", btns)
			if err != nil {
				return err
			}
			switch callback.Data {
			case C.BtnNext:
				currentPage++
			case C.BtnClose:
				break configs
			case C.BtnBack:
				if currentPage == 1 {
					break configs
				}
				currentPage--
			default:
				id, err := strconv.Atoi(callback.Data)
				if err != nil {
					continue
				}
				user, err := a.ctrl.GetUserById(int64(id))
				if err != nil {
					calls.Alertsender("user loading failed " + err.Error())
					continue
				}
				err = a.loaduserinfo(upx, Messagesession, calls, strconv.Itoa(int(user.TgID)))
				if err != nil {
					if errors.Is(err, C.ErrContextDead) {
						return err
					}
					a.logger.Error("failed loaduser ", zap.Error(err))
					calls.Alertsender("user loading failed " + err.Error())
				}
				Messagesession.ResetState()
			}
	}
	Messagesession.DeleteAllMsg()
	return nil
	
}

func (a *Adminsrv) fclose(Messagesession *botapi.Msgsession,  calls common.Tgcalls) error {
	userID, err := calls.Sendreciver("Please enter the user ID of the user whose active sessions you wish to terminate:")
	if err != nil {
		return  err
	}
	id, err := strconv.Atoi(userID.Text)
	if err != nil {
		calls.Alertsender("send correct id")
	}
	if forcecloser, loaded := a.ctrl.Checksession(int64(id)); loaded {
			if closer, ok := forcecloser.(controller.ForceCloser); ok {
				if err := closer.ForceClose(); err != nil {
					Messagesession.SendError(err, "something went wrong while closing old session")
				}
				Messagesession.SendAlert("success", nil)
			}
			return nil
	}
	Messagesession.SendAlert("no sessions", nil)
	return  nil
}

func setvaluefunc(conec builder.Connector, custom walker.SetValue) walker.SetValue {
	return func(curval reflect.Value, path string, wlkr *walker.Walker, item any)  (reflect.Value, bool) {
		switch curval.Kind() {
		case reflect.String:
			s, err := conec.ReciveVal("send new string value")
			if err != nil {
				break
			}
			return reflect.ValueOf(s), true
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint:
			s, err := conec.ReciveVal("send new int value")
			if err != nil {
				break
			}
			d, err := strconv.Atoi(s)
			if err != nil {
				break
			}
			switch curval.Kind() {
			case reflect.Int16:
				return reflect.ValueOf(int16(d)), true
			case reflect.Int32:
				return reflect.ValueOf(int32(d)), true
			case reflect.Int64:
				return reflect.ValueOf(int64(d)), true
			case reflect.Int:
				return reflect.ValueOf(d), true
			case reflect.Uint:
				return reflect.ValueOf(uint(d)), true

			}	
		case reflect.Float32, reflect.Float64:
			s, err := conec.ReciveVal("send new float value")
			if err != nil {
				break
			}
			d, err := strconv.ParseFloat(s, 64)
			if err != nil {
				break
			}
			switch curval.Kind() {
			case reflect.Float64:
				return reflect.ValueOf(d), true 
			case reflect.Float32:
				return reflect.ValueOf(float32(d)), true 
			}
		case reflect.Bool:
			se, err := conec.Select([]string{"true", "false"}, "select true/false")
			if err != nil {
				break
			}
			var bval bool
			if se == "true" {
				bval = true
			}
			return reflect.ValueOf(bval), true
		default:
			if custom != nil {
				return custom(curval, path, wlkr, item )
			}
		}
		return reflect.Value{}, false
	}
}

func cansetfunc(_ builder.Connector) walker.CansetCheck {
	return func(val reflect.Value, nextItemPath string, wlkr *walker.Walker) bool {
		switch val.Type().String() {
		case "constbot.InlinePost":
			return true
		}
		return false
	}	
}