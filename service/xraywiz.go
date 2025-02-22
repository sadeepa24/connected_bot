package service

import (
	"context"
	"errors"
	"net/netip"
	"strconv"
	"sync"
	"time"

	// tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/builder/v1"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)

type Xraywiz struct {
	ctx        context.Context
	callback   *Callback
	logger     *zap.Logger
	//admin      *Adminsrv
	defaultsrv *Defaultsrv

	ctrl   *controller.Controller
	botapi botapi.BotAPI

	MessageStore *botapi.MessageStore

	confstore *builder.ConfigStore

	builds *sync.Map // current building session
}

func NewXraywiz(
	ctx context.Context,
	callback *Callback,
	logger *zap.Logger,
	//admin *Adminsrv,
	ctrl *controller.Controller,
	defaultsrv *Defaultsrv,
	botapi botapi.BotAPI,
	msgstore *botapi.MessageStore,
	//confstore *ConfigStore,

) *Xraywiz {
	return &Xraywiz{
		ctx:          ctx,
		callback:     callback,
		logger:       logger,
		//admin:        admin,
		ctrl:         ctrl,
		botapi:       botapi,
		defaultsrv:   defaultsrv,
		MessageStore: msgstore,
		builds:       &sync.Map{},
		//confstore: confstore,
	}

}

func (x *Xraywiz) Exec(upx *update.Updatectx) error {
	switch {
	case upx.Update.Message != nil:
		if upx.Update.Message.IsCommand() {
			return x.Commandhandler(upx.Update.Message.Command(), upx)
		}
	default:
		return x.defaultsrv.FromserviceExec(upx)
	}

	return nil
}

func (x *Xraywiz) Name() string {
	return C.Xraywizservicename
}

func (x *Xraywiz) Init() error {
	var err error
	if x.confstore, err = builder.NewConfStore(x.ctrl.StorePath()); err != nil {
		return err
	}
	return nil
}

func (x *Xraywiz) Canhandle(upx *update.Updatectx) (bool, error) {
	if upx == nil {
		return false, errors.New("required giving nil upx")
	}
	return upx.Service == C.Xraywizservicename, nil
}

func (u *Xraywiz) Commandhandler(cmd string, upx *update.Updatectx) error {

	if !upx.FromChat().IsPrivate() {
		return nil
	}
	//upx.Configs, _ = u.ctrl.Getconfigs(upx.User.Id)
	//upx.User.Dbuser.Configs = upx.Configs

	Messagesession := botapi.NewMsgsession(upx.Ctx, u.botapi, upx.User.Id, upx.User.Id, upx.User.Lang)
	switch cmd {
	case C.CmdCreate:
		return u.commandCreateV2(upx, Messagesession)
	case C.CmdStatus:
		return u.commandStatus(upx, Messagesession)
	case C.CmdConfigure:
		return u.commandConfigureV2(upx, Messagesession)
	case C.CmdInfo:
		return u.commandInfoV2(upx, Messagesession)
	case C.CmdBuild:
		return u.commandBuildV2(upx, Messagesession)
	default:
		u.logger.Warn("unknown CMD Recived" + upx.Update.Info())
		upx.Cancle()
		upx = nil //drop
		return nil
	}
}

func (u *Xraywiz) commandCreateV2(upx *update.Updatectx, Messagesession *botapi.Msgsession) error {
	if upx.User.IsDistributedUser {
		Messagesession.SendAlert(C.GetMsg(C.MsgCrdisuser), nil)
		return nil
	}

	Usersession, err := controller.NewctrlSession(u.ctrl, upx, false)
	if err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionExcist), nil)
		} else {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionFail), nil)
		}
		upx = nil
		return nil
	}
	defer Usersession.Close()
	opts := common.OptionExcutors{
		Upx:            upx,
		Ctrl:           u.ctrl,
		Usersession:    Usersession,
		MessageSession: Messagesession,
		Btns:           botapi.NewButtons([]int16{2}),
		Tgcalls: common.Tgcalls{
			
			Callbackreciver: func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error) {
				_, err := Messagesession.Edit(msg, btns, "")
				if err != nil {
					return nil, err
				}
				return u.callback.GetcallbackContext(upx.Ctx, btns.ID())
			},
			Alertsender: func(msg string) { Messagesession.SendAlert(msg, nil) },
			Sendreciver: func(msg any) (*tgbotapi.Message, error) {
				if msg != nil {
					if _, err := Messagesession.Edit(msg, nil, ""); err != nil {
						return nil, err
					}
				}
				mg, err := u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)
				if err == nil {
					Messagesession.Addreply(mg.MessageID)
				}
				return mg, err
			},
		},
		Logger: u.logger,
		
	

	}

	opts.Btns.Reset([]int16{2})
	cretors := allcreators()
	for _, creator := range cretors {
		opts.Btns.AddBtcommon(creator.Name())
	}

	callback, err := opts.Callbackreciver(botapi.UpMessage{
		Template: struct {
			*botapi.CommonUser
			CreaterCount int
		}{
			CommonUser: &botapi.CommonUser{
				Name:     upx.User.Name,
				Username: upx.FromChat().UserName,
				TgId:     upx.User.TgID,
			},
			CreaterCount: len(cretors),
		},
		TemplateName: C.TmplCrSelect,
	}, opts.Btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnClose:
		return nil
	}
	for _, creator := range allcreators() {
		if creator.Name() == callback.Data {
			return creator.Excute(opts)
		}
	}

	return nil
}

func (u *Xraywiz) commandStatus(upx *update.Updatectx,  Messagesession *botapi.Msgsession) error {
	Messagesession.Addreply(upx.Update.Message.MessageID)

	Usersession, err := controller.NewctrlSession(u.ctrl, upx, false)
	if err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionExcist), nil)
		} else {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionFail), nil)
		}

		return u.defaultsrv.Droper(upx)

	}
	defer Usersession.Close()

	if upx.User.ConfigCount == 0 {
		Messagesession.SendAlert(C.GetMsg(C.MsgStNoconfig), nil)
		return nil
	}

	usage := Usersession.GetFullUsage()

	if len(Usersession.GetUser().Configs) > 0 {
		btns := botapi.NewButtons([]int16{2})
		for _, config := range Usersession.GetUser().Configs {
			btns.Addbutton(config.Name, strconv.Itoa(int(config.Id)), "")
		}
		btns.AddClose(true)
		Messagesession.SendNew(struct {
			TDownload     string
			TUpload       string
			MDownload     string
			MUpload       string
			MonthAll      string
			Alltime       string
			UsageDuration string
		}{
			TDownload:     usage.Downloadtd.BToString(),
			TUpload:       usage.Uploadtd.BToString(),
			MDownload:     usage.Download.BToString(),
			MUpload:       usage.Upload.BToString(),
			MonthAll:      usage.Full().BToString(),
			UsageDuration: time.Since(u.ctrl.GetLastRefreshtime()).Round(1 * time.Second).String(),
			Alltime:       (upx.User.AlltimeUsage + usage.Full()).BToString(),
		}, btns, C.TmpStTotal)

		for {

			var callback *tgbotapi.CallbackQuery
			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}

			if upx.Ctx.Err() != nil {
				ctx, cancle := context.WithTimeout(context.Background(), 1*time.Minute)
				Messagesession.SetNewcontext(ctx)
				Messagesession.DeleteAllMsg()
				cancle()
				return err
			}

			if callback.Data == C.BtnClose {
				Messagesession.DeleteAllMsg()
				return nil
			}

			confid, err := strconv.Atoi(callback.Data)
			if err != nil {
				continue
			}
			usage, rawstatus := Usersession.GetConfigFullUsage(int64(confid))

			Messagesession.Callbackanswere(callback.ID, u.MessageStore.MsgWithouerro(C.TmpStcallback, upx.User.Lang, struct {
				TDownload     string
				TUpload       string
				MDownload     string
				MUpload       string
				Online        int
				Ip            []netip.Addr
				ConnCount     []int64
				IpMap         map[netip.Addr]int64
				UsageDuration string
			}{
				TDownload:     usage.Downloadtd.BToString(),
				TUpload:       usage.Uploadtd.BToString(),
				MDownload:     usage.Download.BToString(),
				MUpload:       usage.Upload.BToString(),
				Online:        len(rawstatus.Online_ip),
				UsageDuration: time.Since(u.ctrl.GetLastRefreshtime()).Round(1 * time.Second).String(),
				Ip:            C.MapToSliceKey(rawstatus.Online_ip),
				ConnCount:     C.MapToSlice(rawstatus.Online_ip),
				IpMap:         rawstatus.Online_ip,
			}), true)
		}

	}

	//Usersession = nil
	//upx = nil
	Messagesession = nil
	return u.defaultsrv.Droper(upx)
}
