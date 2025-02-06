package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"go.uber.org/zap"
)

type configState struct {
	ctx            context.Context
	State          int
	Messagesession *botapi.Msgsession
	//upx            *update.Updatectx
	userId int64  //tgid
	dbuser *db.User
	btns           *botapi.Buttons
	wiz            *Xraywiz

	Usersession *controller.CtrlSession

	common.Tgcalls

	lastconfig *db.Config

	//TODO: remove later
	conform func(msg any, name string) (bool, error)
}

func (c *configState) run() error {
	var err error
	main:
	for {
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
		switch c.State {

		case stconfhome:
			err = c.home()
		case stconfaction:
			err = c.action()
		case stconfchangein:
			err = c.changeIn()
		case stconfchangeout:
			err = c.changeOut()
		default:
			break main
		}
		if err != nil {
			return err
		}

	}
	
	return nil
}


func (c *configState) home() error {

	var (
		err      error
		callback *tgbotapi.CallbackQuery
	)

	c.btns.Reset([]int16{2})

	for _, config := range c.Usersession.GetUser().Configs {
		c.btns.Addbutton(config.Name, strconv.Itoa(int(config.Id)), "")
	}
	c.btns.AddClose(true)

	if callback, err = c.Callbackreciver(botapi.UpMessage{
		Template: struct {
			*botapi.CommonUser
			ConfCount int16
		}{
			CommonUser: &botapi.CommonUser{
				Name:     c.dbuser.Name,
				TgId:     c.userId,
				Username: c.dbuser.Username.String,
			},
			ConfCount: c.Usersession.GetUser().ConfigCount,
		},
		TemplateName: C.TmplConfigureHome,
	}, c.btns); err != nil {
		return err
	}

	if callback.Data == C.BtnClose {
		c.State = stconfclosed
		return nil
	}
	
	confID, err := strconv.Atoi(callback.Data)
	if err != nil {
		c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
		return nil
	}
	if c.lastconfig, err = c.Usersession.GetConfig(int64(confID)); err != nil {
		if errors.Is(err, C.ErrConfigNotFound) {
			c.Alertsender(C.GetMsg(C.MsgConfUnfoun))
		}
		return nil
	}
	c.State = stconfaction
	return nil
}

func (c *configState) action() error {
	var (
		err      error
		callback *tgbotapi.CallbackQuery
		ok       bool
	)
	c.btns.Reset([]int16{2})
	c.btns.Addbutton(C.BtnChangeIn, C.BtnChangeIn, "")
	c.btns.Addbutton(C.BtnChangeOut, C.BtnChangeOut, "")
	c.btns.Addbutton(C.BtnChangeName, C.BtnChangeName, "")
	c.btns.Addbutton(C.BtnChangeQuota, C.BtnChangeQuota, "")
	c.btns.Addbutton(C.BtnChangeLogin, C.BtnChangeLogin, "")
	c.btns.Addbutton(C.BtnDelete, C.BtnDelete, "")

	c.btns.AddCloseBack()

	if _, err = c.Messagesession.Edit(struct{ ConfName string }{ConfName: c.lastconfig.Name}, c.btns, C.TmpConfiConfigure); err != nil {
		return err
	}
	if callback, err = c.wiz.callback.GetcallbackContext(c.ctx, c.btns.ID()); err != nil {
		return err
	}

	switch callback.Data {

	case C.BtnBack:
		c.State = stconfhome
	case C.BtnClose:
		c.State = stconfclosed
	case C.BtnChangeIn:
		c.State = stconfchangein
	case C.BtnChangeOut:
		c.State = stconfchangeout

	case C.BtnChangeName:

		if _, err = c.Messagesession.EditText(C.GetMsg(C.MsgNewName), nil); err != nil {
			return nil
		}
		
		name, err := common.ReciveString(c.Tgcalls)
		if err != nil {
			return err
		}

		ok, err = c.conform(struct {
			*botapi.CommonUser
			NewName string
		}{
			CommonUser: &botapi.CommonUser{
				Name:     c.dbuser.Name,
				Username: c.dbuser.Username.String,
				TgId:     c.userId,
			},
			NewName: name,
		}, C.TmpNameChange)

		if err != nil {
			return nil
		}
		if ok {
			c.lastconfig.Name = name
			err = c.Usersession.Save()
			if err != nil {
				c.Alertsender(C.GetMsg(C.MsgNameChangeFailed))
			}
			c.Alertsender(C.GetMsg(C.MsgNamechangeSuc))
		}

		return nil

	case C.BtnDelete:
		c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgdelConnWarn), true)
		if ok, err = c.conform(C.GetMsg(C.MsgSure), ""); err != nil {
			return err
		}
		if ok {
			err = c.Usersession.DeleteConfig(c.lastconfig.Id)
			if err != nil {
				c.Alertsender(C.GetMsg(C.MsgdelFail))
				if errors.Is(err, C.ErrOnDeactivation) || errors.Is(err, C.ErrOnDb) {
					c.Usersession.ActivateConfig(c.lastconfig.Id)
					return err
				}

			} else {
				c.Messagesession.SendAlert(C.GetMsg(C.MsgdelSuccses), nil)
				c.Usersession.Save()
			}
		}
		c.State = stconfhome
		return nil

	case C.BtnChangeQuota:
		c.Messagesession.Edit(struct {
			AvblQuota string
			ConfName string
		}{
			AvblQuota: (c.lastconfig.Quota + c.Usersession.LeftQuota()).BToString(),
			ConfName: c.lastconfig.Name,
		}, nil, C.TmpConQuota)

		newquota, err := common.ReciveBandwidth(c.Tgcalls, (c.lastconfig.Quota + c.Usersession.LeftQuota()), c.Usersession.GetconfigUsageTotal(c.lastconfig.Id)  )

		if err != nil {
			return err
		}

		c.lastconfig.Quota = newquota.GbtoByte()
		c.Usersession.DeactivateConfig(c.lastconfig.Id)
		c.Usersession.ActivateConfig(c.lastconfig.Id)

		if err = c.Usersession.SaveConfigs(); err != nil {
			if errors.Is(err, C.ErrContextDead) {
				return err
			}
			c.Alertsender(C.GetMsg(C.Msgwrong))
			return nil
		}

		c.Alertsender(C.GetMsg(C.MsgCoQuota))
		return nil
	
	case C.BtnChangeLogin:
		c.Alertsender("send new login limit count (0 < x <= 5)") 
		limit, err := common.ReciveInt(c.Tgcalls, 0, 5)
		if err != nil {
			return nil
		}
		if limit <= 0 || limit > 5 {
			c.Alertsender("login should be between 0 and 5")
			return nil
		}

		_, err = c.Usersession.ChangeLoginLimit(c.lastconfig.Id, int32(limit))
		if err != nil {
			c.Alertsender("failed")
		}

	}

	return nil
}

func (c *configState) changeIn() error {
	var err error
	var callback *tgbotapi.CallbackQuery
	var ok bool

	c.btns.Reset([]int16{2})
	for _, in := range c.wiz.ctrl.Getinbounds() {
		if c.lastconfig.InboundID == int16(in.Id) {
			c.btns.Addbutton(in.Type+"_"+in.Tag+" "+C.GetMsg(C.ButtonSelectEmjoi), strconv.Itoa(int(in.Id)), "")
			continue
		}
		c.btns.Addbutton(in.Type+"_"+in.Tag+" ", strconv.Itoa(int(in.Id)), "")
	}
	c.btns.AddCloseBack()

	callback, err = c.Callbackreciver(C.GetMsg(C.MsgInsel), c.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnClose:
		c.State = stconfclosed
	case C.BtnBack:
		c.State = stconfaction
	default:
		var inid int
		if inid, err = strconv.Atoi(callback.Data); err != nil {
			return c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
		}
		if inid == int(c.lastconfig.InboundID) {
			c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgInAlredSelected), true)
			return nil
		}
		sboxin, loader := c.wiz.ctrl.Getinbound(inid)
		if !loader {
			return c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
		}
		ok, err = c.conform(struct {
			InName         string
			InType         string
			InPort         int
			InAddr         string
			InInfo         string
			Domain         string
			PublicIp       string
			TranstPortType string
			TlsEnabled     bool
			Support        []string
		}{
			InName:         sboxin.Name,
			InType:         sboxin.Type,
			InPort:         sboxin.Port(),
			InAddr:         sboxin.Laddr(),
			PublicIp:       sboxin.PublicIp,
			Domain:         sboxin.Domain,
			InInfo:         sboxin.Custom_info,
			TranstPortType: sboxin.TransortType(),
			TlsEnabled:     sboxin.TlsIsEnabled(),
			Support:        sboxin.Support,
		}, C.TmpInchange)

		if ok {
			err = c.Usersession.ChangeInbound(c.lastconfig.Id, int64(inid))
			if err != nil {
				switch {
				case errors.Is(err, C.ErrInboundNotFound):
					c.Messagesession.DeleteAllMsg()
					c.Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
					return nil
				}
			}
			c.Usersession.Save()
			c.Messagesession.SendAlert(C.GetMsg(C.MsgInchangesucses), nil)
			c.State = stconfaction
			return nil
		}
		if err != nil {
			c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
		}

	}
	return nil
}

func (c *configState) changeOut() error {
	var err error
	var callback *tgbotapi.CallbackQuery
	var ok bool

	c.btns.Reset([]int16{2})
	for _, out := range c.wiz.ctrl.Getoutbounds() {
		if c.lastconfig.OutboundID == int16(out.Id) {
			c.btns.Addbutton(out.Type+"_"+out.Tag+" "+C.GetMsg(C.ButtonSelectEmjoi), strconv.Itoa(int(out.Id)), "")
			continue
		}
		c.btns.Addbutton(out.Type+"_"+out.Tag+" ", strconv.Itoa(int(out.Id)), "")
	}
	c.btns.AddCloseBack()

	callback, err = c.Callbackreciver(C.GetMsg(C.Msgoutsel), c.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnClose:
		c.State = stconfclosed
	case C.BtnBack:
		c.State = stconfaction
	default:
		var outid int
		if outid, err = strconv.Atoi(callback.Data); err != nil {
			return c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
		}
		if outid == int(c.lastconfig.OutboundID) {
			c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgInAlredSelected), true)
			return nil
		}

		sboxout, loaded := c.wiz.ctrl.Getoutbound(outid)
		if !loaded {
			return c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
		}
		ok, err = c.conform(struct {
			OutName string
			OutType string
			OutInfo string
			Latency int32
		}{
			OutName: sboxout.Name,
			OutType: sboxout.Type,
			OutInfo: sboxout.Custom_info,
			Latency: sboxout.Latency.Load(),
		}, C.TmpOutchange)
		if ok {
			err = c.Usersession.ChangeOutbound(c.lastconfig.Id, int64(outid))
			if err != nil {
				if !errors.Is(err, C.ErrContextDead) {
					c.Messagesession.DeleteAllMsg()
					c.Messagesession.SendAlert("outbound change failed", nil)
					c.Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)

					return nil

				}
				return err

			}
			c.Usersession.Save()
			c.Messagesession.SendAlert(C.GetMsg(C.MsOutchangesucses), nil)
			return nil
		}
		if err != nil {
			c.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
		}

	}

	return nil
}

const (
	stconfhome      = 1
	stconfaction    = 3
	stconfchangein  = 4
	stconfchangeout = 5
	stconfclosed    = 6
)

func (u *Xraywiz) commandConfigureV2(upx *update.Updatectx,  Messagesession *botapi.Msgsession) error {
	Messagesession.Addreply(upx.Update.Message.MessageID)
	var (
		Usersession *controller.CtrlSession
		err         error
	)
	if upx.User.ConfigCount <= 0 {
		Messagesession.SendAlert(C.GetMsg(C.MsgNoconfigstochange), nil)
		return nil
	}
	if Usersession, err = controller.NewctrlSession(u.ctrl, upx, false); err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionExcist), nil)
		} else {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionFail), nil)
		}
		upx = nil
		Messagesession = nil
		Usersession = nil
		return nil
	}
	defer Usersession.Close()

	configState := &configState{
		ctx:            upx.Ctx,
		State:          stconfhome,
		userId: upx.User.TgID,
		dbuser: Usersession.GetUser(),
		btns:           botapi.NewButtons([]int16{2}),
		Usersession:    Usersession,
		wiz:            u,
		Messagesession: Messagesession,

		Tgcalls: common.Tgcalls{
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

	conformbtns := botapi.NewButtons([]int16{1, 1})
	conformbtns.Addbutton(C.BtnConform, C.BtnConform, "")
	conformbtns.Addbutton(C.BtnCancle, C.BtnCancle, "")

	configState.conform = func(msg any, name string) (bool, error) {
		if _, err = Messagesession.Edit(msg, conformbtns, name); err != nil {
			return false, err
		}
		var callback *tgbotapi.CallbackQuery
		if callback, err = u.callback.GetcallbackContext(upx.Ctx, conformbtns.ID()); err != nil {
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

	if err = configState.run(); err !=nil {
		u.logger.Error("configuration state errored", zap.Error(err))
	}
	if upx.Ctx.Err() != nil {
		tempctx, closetemp := context.WithTimeout(u.ctx, 15*time.Second)
		defer closetemp()
		Messagesession.SetNewcontext(tempctx)
	}
	Messagesession.DeleteAllMsg()
	return nil

}
