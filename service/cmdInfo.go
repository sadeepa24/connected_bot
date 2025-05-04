package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	sbConf "github.com/sadeepa24/connected_bot/sbox/conf"
	"github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
)

type getinfo struct {
	ctx context.Context
	state int
	callback *Callback
	Messagesession *botapi.Msgsession
	dbuser *db.User
	//upx *update.Updatectx
	calls common.Tgcalls

	btns *botapi.Buttons
	ctrl *controller.Controller

	Usersession *controller.CtrlSession

	lastselectConf *db.Config
	lastselectInbn sbConf.Inboud


}

const (
	infostatehome = 0
	infostateuser = 1
	infostateconfs = 2
	infostateinbn = 3
	infostateoutbn = 4
	infostateconf = 5
	infostategift = 12

	infostateconfins = 6
	infostateconfin = 7

	infostateclosed = 10
)


func (g *getinfo) home() error {
	
	g.btns.Reset([]int16{2})
	g.btns.Addbutton(C.BtnUserInfo, C.BtnUserInfo, "")
	g.btns.Addbutton(C.BtnConfigs, C.BtnConfigs, "")
	g.btns.AddBtcommon(C.BtnGifts)
	g.btns.Passline()
	g.btns.AddBtcommon(C.BtnCheckOutbounds)
	g.btns.AddBtcommon(C.BtnCheckInbounds)
	g.btns.AddClose(false)


	callback, err := g.calls.Callbackreciver(botapi.UpMessage{
		Template:     struct{}{},
		TemplateName: C.TmplGetinfoHome,
	}, g.btns)

	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnClose:
		g.state = infostateclosed
	case C.BtnUserInfo:
		g.state = infostateuser
	case C.BtnConfigs:
		g.state = infostateconfs
	case C.BtnCheckOutbounds:
		g.state = infostateoutbn
	case C.BtnCheckInbounds:
		g.state = infostateinbn
	case C.BtnGifts:
		g.state = infostategift
	}
	
	return nil
}


func (g *getinfo) allininfo() error {
	allins := g.ctrl.Getinbounds()
	if len(allins) == 0 {
		g.Messagesession.SendAlert("no inbound found",nil)
		g.state = infostatehome
		return nil
	}
	for _, in:= range allins {
		g.btns.Addbutton(in.Tag, strconv.Itoa(int(in.Id)), "")
	}
	g.btns.AddCloseBack()

	callback, err := g.calls.Callbackreciver("select inbound to see details", g.btns); 
	if err != nil {
		return err
	}
	switch callback.Data{
	case C.BtnBack:
		g.state = infostatehome
		return nil
	case C.BtnClose:
		g.state = infostateclosed
		return nil
	default:
		id, err := strconv.Atoi(callback.Data)
		if err != nil {
			return nil
		}
		in, ok := g.ctrl.Getinbound(int16(id))
		if !ok {
			g.Messagesession.Callbackanswere(callback.ID, "inbound not found", true)
			return nil 
		}
		g.btns.Reset([]int16{2})
		g.btns.AddCloseBack()
		if _, err = g.Messagesession.Edit(in, g.btns, C.TmplInInfo); err != nil {
			return nil
		}
		if callback, err = g.callback.GetcallbackContext(g.ctx, g.btns.ID()); err != nil {
			return err
		}
		switch callback.Data {
		case C.BtnClose:
			g.state = infostateclosed
		}
	}
	return nil
}
func (g *getinfo) alloutinfo() error {
	allouts := g.ctrl.Getoutbounds()
	if len(allouts) == 0 {
		g.Messagesession.SendAlert("no outbound found", nil)
		g.state = infostatehome
		return nil
	}
	for _, out:= range allouts {
		g.btns.Addbutton(out.Tag, strconv.Itoa(int(out.Id)), "")
	}
	g.btns.AddCloseBack()

	callback, err := g.calls.Callbackreciver("select outbound to see details", g.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnBack:
		g.state = infostatehome
	case C.BtnClose:
		g.state = infostateclosed
		return nil
	default:
		id, err := strconv.Atoi(callback.Data)
		if err != nil {
			return nil
		}
		out, ok := g.ctrl.Getoutbound(int16(id))
		if !ok {
			g.Messagesession.Callbackanswere(callback.ID, "outbound not found", true)
			return nil
		}
		g.btns.Reset([]int16{2})
		g.btns.AddBtcommon(C.BtnoutLatancy)
		g.btns.AddCloseBack()
		if _, err = g.Messagesession.Edit(commonout{
			OutName: out.Name,
			OutType: out.Type,
			OutInfo: out.Custom_info,
			Latency: out.Latency.Load(),
		}, g.btns, C.TmplOutInfo); err != nil {
			return nil
		}
		latency:
		for {
			if callback, err = g.callback.GetcallbackContext(g.ctx, g.btns.ID()); err != nil {
				return err
			}
			switch callback.Data {
			case C.BtnBack:
				return nil
			case C.BtnClose:
				g.state = infostateclosed
				return nil
			case C.BtnoutLatancy:
				ping, err := g.ctrl.Boxapi.UrlTest(out.Tag)
				
				if err != nil {
					g.Messagesession.Callbackanswere(callback.ID, "Latency checking error, outbound timeout", true)
					continue latency
				}
				out.Latency.Swap(int32(ping))
				g.Messagesession.Callbackanswere(callback.ID, fmt.Sprintf("⚡ outbound latency %v", ping), true)
			}
		}
	}
	return nil
}
func (g *getinfo) userinfo() error {
	
	
	if len(g.ctrl.Metadata.Langs) > 1 {
		g.btns.AddBtcommon(C.BtnLangChange)
	}
	g.btns.AddCloseBack()
	tusage := g.Usersession.TotalUsage()

	if _, err := g.Messagesession.Edit(userinfo{

		CommonUser: &botapi.CommonUser{
			Name:     g.dbuser.Name,
			TgId:     g.dbuser.TgID,
			Username: g.dbuser.Username,
		},
		Paused: g.dbuser.IsPaused,
		CappedQuota: g.dbuser.CappedQuota.BToString(),
		IsTemplimited: g.dbuser.Templimited,
		TempLimitRate: g.dbuser.WarnRatio,
		IsVerified: g.dbuser.Verified(),
		NonUseCycle: g.dbuser.EmptyCycle,
		UsagePercentage: float64(int(((tusage * 100)/(g.Usersession.GetUser().CalculatedQuota + g.dbuser.AdditionalQuota)) * 1000))/1000,
		GiftQuota: g.dbuser.GiftQuota.BToString(),
		Joined:    g.dbuser.Joined.Format("2006-01-02 15:04:05"),
		Dedicated: C.Bwidth(g.ctrl.CommonQuota.Load()).BToString(),
		TQuota:    (g.Usersession.GetUser().CalculatedQuota + g.dbuser.AdditionalQuota).BToString(),
		LeftQuota: g.Usersession.LeftQuota().BToString(),
		TUsage:    tusage.BToString(),
		AlltimeUsage: (g.dbuser.AlltimeUsage+tusage).BToString(),
		ConfCount: g.Usersession.GetUser().ConfigCount,
		CapEndin:  g.dbuser.Captime.AddDate(0, 0, int(g.dbuser.CapDays)).String(),
		CapDays: g.dbuser.CapDays,
		Points: g.dbuser.Points,

		Disendin:     ((g.ctrl.ResetCount - g.ctrl.CheckCount.Load()) * g.ctrl.RefreshRate) / 24,
		UsageResetIn: ((g.ctrl.ResetCount - g.ctrl.CheckCount.Load()) * g.ctrl.RefreshRate) / 24,
		
		Iscapped:       g.dbuser.IsCapped,
		IsMonthLimited: g.dbuser.IsMonthLimited,
		Isdisuser:      g.dbuser.IsDistributedUser,
		JoinedPlace: g.dbuser.CheckID,

	}, g.btns, C.TmpUserInfo); err != nil {
		g.state = infostatehome
	}

	callbackqr, err := g.callback.GetcallbackContext(g.ctx, g.btns.ID())
	if err != nil {
		return err
	}
	switch callbackqr.Data {
	case C.BtnBack:
		g.state = infostatehome
	case C.BtnClose:
		g.state = infostateclosed
	case C.BtnLangChange:
		g.btns.Reset([]int16{2})
		for _, ln := range g.ctrl.Metadata.Langs {
			if ln == g.dbuser.Lang {
				g.btns.Addbutton(ln +" " + C.GetMsg(C.ButtonSelectEmjoi), ln, "")
				continue
			}
			g.btns.AddBtcommon(ln)
		}
		g.btns.AddCloseBack()
	
		callbackqr, err := g.calls.Callbackreciver("select lang code", g.btns)
		if err != nil {
			return err
		}


		switch callbackqr.Data {
		case C.BtnClose:
			g.state = infostateclosed
			return nil
		case C.BtnBack:
			return nil
		default:
			g.Usersession.ChangeLang(callbackqr.Data)
			g.calls.Alertsender("Lang Changed To "  + callbackqr.Data)
			g.Messagesession.ChangeLang(callbackqr.Data)
		}
	}
	return nil	
}
func (g *getinfo) gifts() error {
	gifts, err := g.Usersession.AllGifts(false)
	if err != nil {
		g.calls.Alertsender("fetching gift failed try again later")
		g.callback.logger.Error("fetching gifts err" + err.Error())
		g.state = infostatehome
		return nil
	}
	if len(gifts) == 0 {
		g.state = infostatehome
		g.calls.Alertsender("no gifts (sent or recived)")
		return nil
	}

	for {
		g.btns.Reset([]int16{2})
		for i := range gifts {
			g.btns.AddBtcommon(strconv.Itoa(i))
		}
		g.btns.AddBack(true)
		giftnum, err := g.calls.Callbackreciver("select gift number to see info", g.btns)
		if err != nil {
			return err
		}

		if giftnum.Data == C.BtnBack {
			g.state = infostatehome
			break
		}
		giftp, err := strconv.Atoi(giftnum.Data)
		if err  != nil {
			continue
		}
		gift := gifts[giftp]
		gifttyp := "recived" 
		if gift.Sender == g.Usersession.GetUser().TgID {
			gifttyp = "sent"
		}
		err = g.Messagesession.Callbackanswere(giftnum.ID, `Type `+ gifttyp+`\nBandwidth `+ gift.Bandwidth.BToString(), true)

		if err != nil {
			g.callback.logger.Error("callback answere err: " + err.Error())
		}
	}

	return nil
}


func (g *getinfo) configs() error {
	if g.Usersession.GetUser().ConfigCount == 0 {
		g.Messagesession.SendAlert(C.GetMsg(C.MsgInfoNoconfigs), nil)
		g.state = infostatehome
		return nil
	}

	for _, config := range g.Usersession.GetUser().Configs {
		g.btns.Addbutton(config.Name, strconv.Itoa(int(config.Id)), "")
	}
	g.btns.AddCloseBack()

	callback, err := g.calls.Callbackreciver(C.GetMsg(C.MsgInfoSelectConfig), g.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnBack:
		g.state = infostatehome
		return nil
	case C.BtnClose:
		g.state = infostateclosed
		return nil
	}

	selectedconf, err := strconv.Atoi(callback.Data)
	if err != nil {
		return nil
	}
	g.lastselectConf, err = g.Usersession.GetConfig(int64(selectedconf))
	if err != nil {
		g.Messagesession.SendError(err, C.GetMsg(C.Msgwrong))
		return nil
	}
	g.state = infostateconf
	return nil
}

func (g *getinfo) configinfo() error {
	confid := g.lastselectConf.Id
	
	//btns.Addbutton(C.BtnFullUsage, C.BtnFullUsage, "")
	g.btns.AddBtcommon(C.BtnCInbounds)
	g.btns.AddBtcommon(C.BtnCloseConn)
	g.btns.AddBtcommon(C.BtnRefresh)
	g.btns.AddCloseBack()

	status, err := g.Usersession.Getstatus(int64(confid))

	if err != nil &&  errors.Is(err, C.ErrConfigNotFound)  {
		g.Messagesession.SendError(err, C.GetMsg(C.Msgconfcannotfind))
		return nil
	}
	sboxout, ok := g.ctrl.Getoutbound(g.lastselectConf.OutboundID)

	if !ok {
		g.Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		g.state = infostateconfs
		return nil
	}
	var perctage float64

	if g.lastselectConf.Quota > 0 {
		perctage = float64(int(((g.lastselectConf.Usage+status.FullUsage()).Float64()/g.lastselectConf.Quota.Float64())*100*1000)) / 1000
	}

	if _, err = g.Messagesession.Edit(configinfo{
		CommonUser: &botapi.CommonUser{
			Name:     g.dbuser.Name,
			Username: g.dbuser.Username,
			TgId:     g.dbuser.TgID,
		},
		Active: g.lastselectConf.Active,

		TotalQuota:     g.lastselectConf.Quota.BToString(),
		ConfigName:     g.lastselectConf.Name,
		ConfigUUID:     g.lastselectConf.UUID,
		ConfigPassword: g.lastselectConf.Password,
		Loginlimit: g.lastselectConf.LoginLimit,
		UsedPresenTage: perctage,
		ResetDays: ((g.ctrl.ResetCount - g.ctrl.CheckCount.Load()) * g.ctrl.RefreshRate) / 24,

		ConfigDownload: (g.lastselectConf.Download + status.Download).BToString(),
		ConfigUpload:   (g.lastselectConf.Upload + status.Upload).BToString(),

		ConfigDownloadtd: (status.Download).BToString(),
		ConfigUploadtd:   (status.Upload).BToString(),

		ConfigUsagetd: (status.Download + status.Upload).BToString(),
		ConfigUsage:   (status.Download + status.Upload + g.lastselectConf.Usage).BToString(),

		UsageDuration:  time.Since(g.ctrl.GetLastRefreshtime()).Round(1 * time.Second).String(),
		
		commonout: commonout{
			OutName: sboxout.Name,
			OutType: sboxout.Type,
			OutInfo: sboxout.Custom_info,
			Latency: sboxout.Latency.Load(),
		},


		Online: len(status.Online_ip),
		IpMap:  status.Online_ip,


	}, g.btns, C.TmpConfigInfo); err != nil {
		g.state = infostateconfs
		return nil
	}
	callback, err := g.callback.GetcallbackContext(g.ctx, g.btns.ID()); 
	if err != nil {
		return err
	}
	
	switch callback.Data {
	case C.BtnClose:
		g.state = infostateclosed
		return nil
	case C.BtnCloseConn:
		g.Usersession.ConfigCloseConn(int64(confid))
	case C.BtnBack:
		g.state = infostateconfs
		return nil
	case C.BtnCInbounds:
		g.state = infostateconfins
	case C.BtnRefresh:
		return nil
	}

	if callback.Data == C.BtnFullUsage {
		//TODO: maybe later
		g.Messagesession.SendAlert(C.GetMsg(C.Msgconfcannotfind), nil)
		g.Messagesession.SendAlert("usage history function is not avalable yet", nil)
	}


	return nil

}

func (g *getinfo) confallin() error {
	allins := g.ctrl.GetinboundList(g.lastselectConf.InboundIds)
	
	all := []exportConfig{}
	for _, isns := range allins {
		g.btns.Addbutton(isns.Tag + "_" + strconv.Itoa(isns.Port()), strconv.Itoa(int(isns.Id)), "") 
		all = append(all, exportConfig{
			exportin: exportin{
				ProtoUrl: "use export button",
			},
			Inboud: isns,
		})
	}
	g.btns.Passline()
	g.btns.AddBtcommon(C.BtnExportIn)
	g.btns.AddCloseBack()

	callback, err := g.calls.Callbackreciver(botapi.UpMessage{
		Template: struct {
			AllIn []exportConfig
			InboundCount int
		}{
			AllIn: all,
			InboundCount: len(all),
		},
		TemplateName: C.TmplAllIn,
		Buttons: g.btns,
	}, g.btns)
	if  err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnBack:
		g.state = infostateconf
		return nil
	case C.BtnClose:
		g.state = infostateclosed
		return nil
	case C.BtnExportIn:
		g.Messagesession.Edit("send info to export config", nil, "")
		expinfo, err := common.ReciveExpInfo(g.calls)
		if err != nil {
			if _, ok := err.(botapi.Error); ok {
				return err
			}
			g.Messagesession.SendError(err, "")
		}
		for i := range all {
			all[i].exportin.ProtoUrl = g.lastselectConf.ExportUrlLink(all[i].Inboud, &expinfo)
			all[i].ExportInfo = expinfo
		}
		g.Messagesession.SendExtranal(struct {
			AllIn []exportConfig
			InboundCount int
		}{
			AllIn: all,
			InboundCount: len(all),
		}, nil, C.TmplAllIn, true)
	}

	sin, err := strconv.Atoi(callback.Data)
	if err != nil {
		return nil
	}
	var ok bool
	g.lastselectInbn, ok = g.ctrl.Getinbound(int16(sin))
	if !ok {
		g.Messagesession.SendAlert("inbound not found", nil)
		return nil
	}
	g.state  = infostateconfin
	return nil
}

func (g *getinfo) confinboundinfo() error {
	g.btns.AddBtcommon(C.BtnExportConf)
	g.btns.AddCloseBack()
	
	tm := exportConfig{
		exportin: exportin{
			ProtoUrl: g.lastselectConf.ExportUrlLink(g.lastselectInbn, nil),
		},
		Inboud: g.lastselectInbn,
	}
	callback, err := g.calls.Callbackreciver(botapi.UpMessage{
		Template: tm,
		Buttons: g.btns,
		TemplateName: C.TmplConfIn,
	}, g.btns);
	if err != nil {
		return err
	}
	
	switch callback.Data{
	case C.BtnBack:
		g.state = infostateconfins
	case C.BtnClose:
		g.state  = infostateclosed
	case C.BtnExportConf:
		g.Messagesession.Edit("send info to export config", nil, "")
		tm.ExportInfo, err = common.ReciveExpInfo(g.calls)
		if err != nil {
			if _, ok := err.(botapi.Error); ok {
				return err
			}
			g.Messagesession.SendError(err, "")
		}
		tm.ProtoUrl = g.lastselectConf.ExportUrlLink(g.lastselectInbn, &tm.ExportInfo)
		g.Messagesession.SendExtranal(tm, nil, C.TmplConfIn, true)
	}
	return nil
}




func (g *getinfo) run() error {
	var err error
	
	info:
	for {
		if g.ctx.Err() != nil {
			break
		}

		g.btns.Reset([]int16{2})

		switch g.state {
		case infostatehome:
			err = g.home()
		case infostateinbn:
			err = g.allininfo()
		case infostateoutbn:
			err = g.alloutinfo()
		case infostateuser:
			err = g.userinfo()
		case infostateconfs:
			err = g.configs()
		case infostateconf:
			err= g.configinfo()
		case infostateconfins:
			err = g.confallin()
		case infostateconfin:
			err =  g.confinboundinfo()
		case infostategift:
			err = g.gifts()
		case infostateclosed:
			break info
		}

		if err != nil {
			break
		}
	}
	return err
}


func (u *Xraywiz) commandInfoV2(upx *update.Updatectx,  Messagesession *botapi.Msgsession) error {
	Messagesession.AddreplyNoDelete(upx.Update.Message.MessageID)
	var (
		Usersession *controller.CtrlSession
		err         error
	)

	if Usersession, err = controller.NewctrlSession(u.ctrl, upx, false); err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.EditText(C.GetMsg(C.MsgSessionExcist), nil)
		}
		return nil
	}
	defer Usersession.Close()

	btns := botapi.NewButtons([]int16{2})
	infoistate := getinfo{
		state: 0,
		callback: u.callback,
		btns: btns,
		Messagesession: Messagesession,
		Usersession: Usersession,
		ctx: upx.Ctx,
		dbuser: upx.Dbuser(),
		ctrl: u.ctrl,
		calls:  common.Tgcalls{
			Alertsender: func(msg string) {
				Messagesession.SendAlert(msg, nil)
			},
			Sendreciver: func(msg any) (*tgbotapi.Message, error) {
				if msg != nil {
					_, err := Messagesession.Edit(msg, nil, "")
					if err != nil {
						return nil, err
					}
				}
				mg, err := u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)
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
				return u.callback.GetcallbackContext(upx.Ctx, btns.ID())
			},
		},
	}

	err = infoistate.run()
	if err != nil {
		Messagesession.SetNewcontext(u.ctx)
		Messagesession.SendError(err, "")
	}
	Messagesession.DeleteAllMsg()
	return err
}
