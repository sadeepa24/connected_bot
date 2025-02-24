//go:build ignore

package service

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"time"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"github.com/sagernet/sing-vmess/vless"
)

// Deprecated: use commandCreateV2 instead
func (u *Xraywiz) commandCreate(upx *update.Updatectx) error {

	Messagesession := botapi.NewMsgsession(upx.Ctx, u.botapi, upx.User.Id, upx.User.Id, upx.User.Lang)
	Usersession, err := controller.NewctrlSession(u.ctrl, upx, false)
	if err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionExcist), nil)
		} else {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionFail), nil)
		}
		return nil
	}
	defer Usersession.Close()

	if upx.User.IsDistributedUser {
		Messagesession.SendAlert(C.MsgCrdisuser, nil)
		return nil
	}

	if upx.User.ConfigCount >= (u.ctrl.Maxconfigcount - upx.User.AddtionalConfig) {
		Messagesession.Edit(struct {
			Count int16
		}{
			Count: upx.User.ConfigCount,
		}, nil, C.TmpCrAlreadyHave)
		u.logger.Info(upx.User.Name + " user already have maximum config count")
		return nil
	}

	if upx.User.MonthUsage > upx.User.GenaralQuotSum() {
		Messagesession.SendAlert(C.MsgUsageExceed, nil)
	}

	var (
		callback      *tgbotapi.CallbackQuery
		replymeassage *tgbotapi.Message
		inID          int
		outID         int
	)

	// send Avelable Inbounds inline mode

	btns := botapi.NewButtons([]int16{2})

	for _, inbound := range u.ctrl.Getinbounds() {
		btns.Addbutton(inbound.Type, strconv.Itoa(int(inbound.Id)), "")
	}

	btns.AddClose(true)
	if _, err = Messagesession.EditText(C.MsgselectIn, btns); err != nil {
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}

	if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
		return err
	}

	if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, func() error { return nil }); ok {
		return err
	}

	if inID, err = strconv.Atoi(callback.Data); err != nil {
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}

	sboxin, loaded := u.ctrl.Getinbound(inID)

	if !loaded {
		Messagesession.SendAlert(C.MsgCrInerr, nil)
		return nil
	}
	btns.Reset([]int16{2})
	btns.AddBtcommon(C.BtnConform)
	btns.Addcancle()
	if _, err = Messagesession.Edit(struct {
		InName         string
		InType         string
		InPort         int
		InAddr         string
		InInfo         string
		Domain         string
		PublicIp       string
		TranstPortType string
		TlsEnabled     bool
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
	}, btns, C.TmpCrInInfo); err != nil {
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}

	if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
		return err
	}

	if err = checkconform(callback.Data, Messagesession); err != nil {
		return err
	}

	btns.Reset([]int16{2})

	for _, outbound := range u.ctrl.Getoutbounds() {
		btns.Addbutton(outbound.Type, strconv.Itoa(int(outbound.Id)), "")
	}

	btns.AddClose(true)

	if _, err = Messagesession.EditText(C.MsgselectOut, btns); err != nil {
		return err
	}

	if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
		return err
	}

	if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, func() error {
		return nil
	}); ok {
		return err
	}

	if outID, err = strconv.Atoi(callback.Data); err != nil {
		return err
	}

	sboxout, loaded := u.ctrl.Getoutbound(outID)

	if !loaded {
		Messagesession.SendAlert(C.MsgCrOuterr, nil)
		return nil
	}

	btns.Reset([]int16{2})
	btns.AddBtcommon(C.BtnConform)
	btns.Addcancle()
	if _, err = Messagesession.Edit(struct {
		OutName string
		OutType string
		OutInfo string
		Latency int32
	}{
		OutName: sboxout.Name,
		OutType: sboxout.Type,
		OutInfo: sboxout.Custom_info,
		Latency: sboxout.Latency.Load(),
	}, btns, C.TmpCrOutInfo); err != nil {
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}

	if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
		return err
	}

	if err = checkconform(callback.Data, Messagesession); err != nil {
		return err
	}

	// Selecting Quota for New creating config

	if Usersession.LeftQuota() <= 0 {
		Messagesession.SendAlert(C.MsgnoQuota, nil)
		return nil
	}

	fusage := Usersession.GetFullUsage()
	var reduce C.Bwidth
	if upx.User.MonthUsage+fusage.Downloadtd+fusage.Uploadtd != fusage.Download+fusage.Upload {
		Messagesession.SendAlert(C.MsgCrQuotaNote, nil)
		reduce = upx.User.MonthUsage + fusage.Downloadtd + fusage.Uploadtd - fusage.Download + fusage.Upload
	}

	if _, err = Messagesession.Edit(struct {
		Quota string
	}{
		Quota: (Usersession.LeftQuota() - reduce).BToString(),
	}, nil, C.TmpCrAvblQuota); err != nil {
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}

	var (
		quotacalc      func() error
		quotafroconfig int
		retry          int16
	)

	quotacalc = func() error {
		retry++
		if upx.Ctx.Err() != nil {
			Messagesession.SendAlert(C.GetMsg(C.MsgContextDead), nil)
			return C.ErrContextDead
		}

		if retry > 5 {
			Messagesession.EditText(C.GetMsg(C.Msgretryfail), nil)
			return nil
		}
		if replymeassage, err = u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.FromChat().ID); err != nil {
			return err
		}
		Messagesession.Addreply(replymeassage.MessageID)

		quotafroconfig, err = strconv.Atoi(replymeassage.Text)
		if err != nil {
			Messagesession.SendNew(C.GetMsg(C.MsgValidInt), nil, "")
			return quotacalc()
		}

		if C.Bwidth(quotafroconfig).GbtoByte() > (Usersession.LeftQuota() - reduce) {
			Messagesession.SendAlert("you can't add more quota than your limit "+(Usersession.LeftQuota()-reduce).BToString(), nil)
			return quotacalc()
		}
		return nil
	}

	if err = quotacalc(); err != nil {
		return err
	}

	if _, err = Messagesession.SendNew(C.MsgGetName, nil, ""); err != nil {
		return err
	}

	var getName func() error
	var confName string

	getName = func() error {
		if replymeassage, err = u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.Id, upx.FromChat().ID); err != nil {
			u.botapi.SendError(err, upx.User.Id)
			return err
		}
		Messagesession.Addreply(replymeassage.MessageID)
		if replymeassage.IsCommand() {
			Messagesession.SendNew(C.GetMsg(C.MsgValidName), nil, "")
			return getName()
		}
		confName = replymeassage.Text
		if replymeassage.Text == "" {
			confName = "noname"
		}
		return nil

	}

	if err = getName(); err != nil {
		Messagesession.DeleteAllMsg()
		if !errors.Is(err, C.ErrContextDead) {
			Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		}
		return err
	}

	var LoginLimit int
	retry = 0
	if _, err := Messagesession.EditText(C.MsgCrLogin, nil); err != nil {
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}

	for {
		if upx.Ctx.Err() != nil {
			return C.ErrContextDead
		}

		if retry > 5 {
			Messagesession.SendAlert(C.GetMsg(C.Msgretryfail), nil)
			return nil
		}

		if replymeassage, err = u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID); err != nil {
			return err
		}

		Messagesession.Addreply(replymeassage.MessageID)

		if replymeassage == nil {
			continue
		}

		if LoginLimit, err = strconv.Atoi(replymeassage.Text); err != nil {
			Messagesession.SendAlert(C.GetMsg(C.MsgValidInt), nil)
			continue
		}

		if LoginLimit > C.MaxLoginLimit || LoginLimit <= 0 {
			Messagesession.SendAlert(C.MsgCrLoginwarn, nil)
			continue
		}

		break

	}

	config, err := Usersession.AddNewConfig(int16(inID), int16(outID), C.Bwidth(quotafroconfig).GbtoByte(), int16(LoginLimit), confName)

	if err != nil {
		switch {
		case errors.Is(err, C.ErrInboundNotFound), errors.Is(err, C.ErrDatabaseCreate), errors.Is(err, C.ErrTypeMissmatch), errors.Is(err, C.ErrContextDead):
			Messagesession.SendAlert(C.MsgCrFailed, nil)

		default:
			Messagesession.SendAlert(C.MsgInternalErr, nil)
		}

		return err

	}

	// Messagesession.SendNew(C.MsgGetSni, nil)

	// if replymeassage, err = u.defaultsrv.ExcpectMsg(upx.User.TgID, upx.FromChat().ID);  err != nil {
	// 	Messagesession.SendAlert(C.MsgSnifail, nil)
	// }

	Messagesession.DeleteAllMsg()
	Messagesession.SendAlert(C.MsgCrsuccsess, nil)

	Messagesession.SendExtranal(struct {
		UUID       string
		Domain     string
		Transport  string
		ConfigName string
		TlsEnabled bool
		Port       int
	}{
		Domain:     sboxin.Domain,
		ConfigName: sboxin.Name,
		Port:       sboxin.Port(),
		Transport:  sboxin.Transporttype,
		TlsEnabled: sboxin.Tlsenabled,
		UUID:       config.UUID.String(),
	}, nil, C.TmpCrSendUID, true)

	Messagesession.SendAlert(C.MsgCrConfigIn, nil)

	return u.defaultsrv.Droper(upx)
}

// Deprecated: use commandConfigureV2
func (u *Xraywiz) commandConfigure(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(upx.Ctx, u.botapi, upx.User.Id, upx.User.Id, upx.User.Lang)
	Messagesession.Addreply(upx.Update.Message.MessageID)
	var (
		Usersession *controller.CtrlSession
		err         error
	)

	if Usersession, err = controller.NewctrlSession(u.ctrl, upx, false); err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionExcist), nil)
		} else {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionFail), nil)
		}
		Messagesession = nil
		Usersession = nil
		return nil
	}
	defer Usersession.Close()

	if len(Usersession.GetUser().Configs) <= 0 {
		Messagesession.SendAlert(C.GetMsg(C.MsgNoconfigstochange), nil)
		return nil
	}

	var (
		home    func() error
		conform func(any, string) (bool, error)
	)

	conformbtns := botapi.NewButtons([]int16{1, 1})
	conformbtns.Addbutton(C.BtnConform, C.BtnConform, "")
	conformbtns.Addbutton(C.BtnCancle, C.BtnCancle, "")

	var totalrecursive int

	conform = func(msg any, name string) (bool, error) {

		if upx.Ctx.Err() != nil {
			return true, err
		}
		totalrecursive++

		if _, err = Messagesession.Edit(msg, conformbtns, name); err != nil {

			if errors.Is(err, C.ErrClientRequestFail) {
				return conform(msg, name)
			}

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

	btns := botapi.NewButtons([]int16{2})
	home = func() error {

		if upx.Ctx.Err() != nil {
			return err
		}

		totalrecursive++
		if totalrecursive > u.ctrl.MaxRecurtion {
			return C.ErrRecurtionExceed
		}
		var (
			config   *db.Config
			ok       bool
			callback *tgbotapi.CallbackQuery
			action   func() error
		)

		btns.Reset([]int16{2})

		for _, config := range Usersession.GetUser().Configs {
			btns.Addbutton(config.Name, strconv.Itoa(int(config.Id)), "")
		}
		btns.AddClose(true)

		Messagesession.EditText("no mg", btns)
		//Messagesession.SetPrimeLast()

		if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
			return err
		}
		if ok, err = closeback(callback.Data, Messagesession.DeleteAllMsg, home); ok {
			return err
		}

		confID, err := strconv.Atoi(callback.Data)

		if err != nil {
			Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
			return home()
		}

		if config, err = Usersession.GetConfig(int64(confID)); err != nil {
			if errors.Is(err, C.ErrConfigNotFound) {
				Messagesession.SendAlert(C.GetMsg(C.MsgConfUnfoun), nil)
				return err
			}
			return home()
		}

		action = func() error {

			if upx.Ctx.Err() != nil {
				return err
			}
			totalrecursive++

			if totalrecursive > u.ctrl.MaxRecurtion {
				return C.ErrRecurtionExceed
			}

			btns.Reset([]int16{2})
			btns.Addbutton(C.BtnChangeIn, C.BtnChangeIn, "")
			btns.Addbutton(C.BtnChangeOut, C.BtnChangeOut, "")
			btns.Addbutton(C.BtnChangeName, C.BtnChangeName, "")
			btns.Addbutton(C.BtnDelete, C.BtnDelete, "")
			btns.Addbutton(C.BtnChangeQuota, C.BtnChangeQuota, "")

			//btns.Addbutton(C.BtnFullInfo,C.BtnFullInfo, "")
			btns.AddCloseBack()

			if _, err = Messagesession.Edit(struct{ ConfName string }{ConfName: config.Name}, btns, C.TmpConfiConfigure); err != nil {
				return err
			}

			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				if errors.Is(err, C.ErrContextDead) {
					return err
				}
				return action()
			}

			if ok, err = closeback(callback.Data, Messagesession.DeleteAllMsg, home); ok {
				return err
			}

			switch callback.Data {

			case C.BtnChangeName:

				if _, err = Messagesession.EditText(C.GetMsg(C.MsgNewName), nil); err != nil {
					return action()
				}
				var getName func() (string, error)

				getName = func() (string, error) {

					if upx.Ctx.Err() != nil {
						return "", err
					}

					totalrecursive++
					if totalrecursive > u.ctrl.MaxRecurtion {
						return "", C.ErrRecurtionExceed
					}

					var (
						replymeassage *tgbotapi.Message
						confName      string
					)
					if replymeassage, err = u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.Id, upx.FromChat().ID); err != nil {
						if errors.Is(err, C.ErrContextDead) {
							return "", C.ErrContextDead
						}
						return getName()
					}
					Messagesession.Addreply(replymeassage.MessageID)
					if replymeassage.IsCommand() {
						_, err = Messagesession.EditText(C.GetMsg(C.MsgValidName), nil)
						return getName()
					}
					confName = replymeassage.Text

					if replymeassage.Text == "" {
						confName = "noname"
					}

					return confName, nil

				}

				name, err := getName()
				if errors.Is(err, C.ErrContextDead) || errors.Is(err, C.ErrRecurtionExceed) {
					return err

				} else if err != nil {
					return action()
				}

				ok, err = conform(struct {
					*botapi.CommonUser
					NewName string
				}{
					CommonUser: &botapi.CommonUser{
						Name:     upx.User.Name,
						Username: upx.FromChat().UserName,
						TgId:     upx.User.TgID,
					},
					NewName: name,
				}, C.TmpNameChange)

				if err != nil {
					return action()
				}
				if ok {
					config.Name = name
					err = Usersession.Save()
					if err != nil {
						Messagesession.SendAlert(C.GetMsg(C.MsgNameChangeFailed), nil)
					}
					Messagesession.SendAlert(C.GetMsg(C.MsgNamechangeSuc), nil)
				}

				return action()

			case C.BtnChangeIn:

				var inboundchange func() error

				inboundchange = func() error {

					if upx.Ctx.Err() != nil {
						return err
					}
					totalrecursive++
					if totalrecursive > u.ctrl.MaxRecurtion {
						return C.ErrRecurtionExceed
					}

					btns.Reset([]int16{2})
					for _, in := range u.ctrl.Getinbounds() {
						if config.InboundID == int16(in.Id) {
							btns.Addbutton(in.Type+" "+C.GetMsg(C.ButtonSelectEmjoi), strconv.Itoa(int(in.Id)), "")
							continue
						}

						btns.Addbutton(in.Type+" ", strconv.Itoa(int(in.Id)), "")
					}

					btns.AddCloseBack()

					if _, err = Messagesession.EditText(C.GetMsg(C.MsgInsel), btns); err != nil {
						return inboundchange()
					}

					callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID())
					if err != nil {
						return err
					}

					ok, err = closeback(callback.Data, Messagesession.DeleteAllMsg, action)
					if ok {
						return err
					}

					var inid int
					if inid, err = strconv.Atoi(callback.Data); err != nil {
						return Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
					}

					sboxin, loader := u.ctrl.Getinbound(inid)

					if !loader {
						return Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
					}

					ok, err = conform(struct {
						InName         string
						InType         string
						InPort         int
						InAddr         string
						InInfo         string
						Domain         string
						PublicIp       string
						TranstPortType string
						TlsEnabled     bool
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
					}, C.TmpInchange)

					if ok {
						err = Usersession.ChangeInbound(config.Id, int64(inid))
						if err != nil {
							switch {
							case errors.Is(err, C.ErrInboundNotFound):
								Messagesession.DeleteAllMsg()
								Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
								return nil
							}
						}
						Usersession.Save()
						Messagesession.SendAlert(C.GetMsg(C.MsgInchangesucses), nil)

						return action()
					}
					if err != nil {
						Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
					}

					return inboundchange()
				}

				return inboundchange()

			case C.BtnChangeOut:

				var outboundchange func() error

				outboundchange = func() error {

					if upx.Ctx.Err() != nil {
						return err
					}

					totalrecursive++
					if totalrecursive > u.ctrl.MaxRecurtion {
						return C.ErrRecurtionExceed
					}

					btns.Reset([]int16{2})
					for _, in := range u.ctrl.Getoutbounds() {
						if config.OutboundID == int16(in.Id) {
							btns.Addbutton(in.Type+" "+C.GetMsg(C.ButtonSelectEmjoi), strconv.Itoa(int(in.Id)), "")
							continue
						}

						btns.Addbutton(in.Type+" ", strconv.Itoa(int(in.Id)), "")
					}

					btns.AddCloseBack()

					if _, err = Messagesession.EditText(C.GetMsg(C.Msgoutsel), btns); err != nil {
						return outboundchange()
					}

					callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID())
					if err != nil {
						return err
					}

					ok, err = closeback(callback.Data, Messagesession.DeleteAllMsg, action)
					if ok {
						return err
					}

					var outid int
					if outid, err = strconv.Atoi(callback.Data); err != nil {
						return Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
					}

					sboxout, loaded := u.ctrl.Getoutbound(outid)
					if !loaded {
						return Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
					}
					ok, err = conform(struct {
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
						err = Usersession.ChangeOutbound(config.Id, int64(outid))
						if err != nil {
							if !errors.Is(err, C.ErrContextDead) {
								Messagesession.DeleteAllMsg()
								Messagesession.SendAlert("outbound change failed", nil)
								Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)

								return nil

							}
							return err

						}
						Usersession.Save()
						Messagesession.SendAlert(C.GetMsg(C.MsOutchangesucses), nil)
						return action()
					}
					if err != nil {
						Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
					}

					return outboundchange()
				}

				return outboundchange()

			case C.BtnDelete:

				if upx.Ctx.Err() != nil {
					return err
				}
				Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgdelConnWarn), true)
				if ok, err = conform(C.GetMsg(C.MsgSure), ""); err != nil {
					return err
				}

				if ok {
					err = Usersession.DeleteConfig(config.Id)
					if err != nil {
						Messagesession.SendAlert(C.GetMsg(C.MsgdelFail), nil)
						if errors.Is(err, C.ErrOnDeactivation) || errors.Is(err, C.ErrOnDb) {
							Usersession.ActivateConfig(config.Id)
							return err
						}

					} else {
						Messagesession.SendAlert(C.GetMsg(C.MsgdelSuccses), nil)
						Usersession.Save()
					}
				}
				return home()

			case C.BtnChangeQuota:

				var newquota int
				var retry = 0

				Messagesession.Edit(struct {
					AvblQuota string
				}{
					AvblQuota: (config.Quota + Usersession.LeftQuota()).BToString(),
				}, nil, C.TmpConQuota)

				for {
					if upx.Ctx.Err() != nil {
						return err
					}
					if retry > 5 {
						Messagesession.SendAlert(C.GetMsg(C.Msgretryfail), nil)
						return action()
					}

					replaymg, err := u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)
					if err != nil {
						return err
					}
					if replaymg == nil {
						continue
					}
					retry++
					Messagesession.Addreply(replaymg.MessageID)

					if newquota, err = strconv.Atoi(replaymg.Text); err != nil {
						Messagesession.SendAlert(C.GetMsg(C.MsgValidInt), nil)
						continue
					}

					if C.Bwidth(newquota).GbtoByte() > (config.Quota + Usersession.LeftQuota()) {
						Messagesession.SendAlert(C.GetMsg(C.MsgQuotawarn), nil)
						continue
					}

					break

				}

				config.Quota = C.Bwidth(newquota).GbtoByte()
				Usersession.DeactivateConfig(config.Id)
				Usersession.ActivateConfig(config.Id)

				if err = Usersession.SaveConfigs(); err != nil {
					if errors.Is(err, C.ErrContextDead) {
						return err
					}
					Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
					return action()
				}

				Messagesession.SendAlert(C.GetMsg(C.MsgCoQuota), nil)
				return action()

				/* case C.BtnFullInfo:

				btns.Reset([]int16{1,1,2})

				btns.Addbutton(C.BtnUsageHistory, C.BtnUsageHistory, "")
				btns.Addbutton(C.BtnHome, C.BtnHome, "")
				btns.AddCloseBack("")

				// today, alltime, err := Usersession.GetconfigUsage(config.Id)
				// if err != nil {
				// }

				status, err := Usersession.Getstatus(config.Id)
				if err != nil {
					Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgwrong), true)
					if errors.Is(err, C.ErrContextDead) {
						return err
					}
					return action()
				}


				inbn, loaded := u.ctrl.Getinbound(int(config.InboundID))
				if !loaded {
					inbn,_ = u.ctrl.DefaultInboud()
				}
				outbn, loaded := u.ctrl.Getoutbound(int(config.OutboundID))
				if !loaded {
					outbn,_ = u.ctrl.Defaultoutboud()
				}


				Messagesession.Edit(`

				Full Info About Config

				Usage:-
				today:-
				down - `+ status.Download.BToString()  + `  / up -`+ status.Upload.BToString()  + `

				month:-
				down -`+ status.Download.BToString()  + ` / up -`+ status.Download.BToString()  + `



				Config:-
				name - `+ config.Name + `
				uuid - `+ config.UUID.String() + `

				Inbound:-

				type -  `+ inbn.Type + `
				port -  `+ strconv.Itoa(inbn.Listenport) + `
				info - `+ inbn.Custom_info + `
				transport - `+ inbn.Transporttype + `

				Outbound:-

				type -  `+ outbn.Type+ `
				info - `+ outbn.Custom_info + `

				`, btns)

				if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
					if errors.Is(err, C.ErrContextDead) {
						return err
					}
					Messagesession.SendNew(C.GetMsg(C.Msgwrong), nil)
					return action()
				}

				if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, action); ok {
					return err
				}

				switch callback.Data {
				case C.BtnHome:
					return home()
				case C.BtnUsageHistory:
					Messagesession.SendAlert("developing function", nil)
				}
				*/

			}

			return nil
		}

		return action()

	}

	err = home()
	if err != nil {
		if upx.Ctx.Err() != nil {
			tempctx, closetemp := context.WithTimeout(u.ctx, 15*time.Second)
			defer closetemp()
			Messagesession.EditNewcontext(tempctx, C.GetMsg(C.MsgSessionOver), nil, "")
			return nil
		}

		if errors.Is(err, C.ErrRecurtionExceed) {
			Messagesession.SendAlert(C.GetMsg(C.MsgRecursionExceed), nil)
			return err
		}

		return err
	}

	return nil
}

// Deprecated: Use commandHelpV2
func (u *Usersrv) commandHelp(upx *update.Updatectx) error {
	u.logger.Info("help comma excuted  by " + upx.User.Info())
	Messagesession := botapi.NewMsgsession(upx.Ctx, u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	Messagesession.Addreply(upx.Update.Message.MessageID)
	btns := botapi.NewButtons([]int16{2, 1, 1})

	var home func() error

	home = func() error {

		btns.Reset([]int16{2, 2, 1})
		btns.AddBtcommon(C.Btncommand)
		btns.AddBtcommon(C.BtnBtinfo)
		btns.AddBtcommon(C.BtnFaq)
		btns.AddBtcommon(C.BtnAbout)
		btns.AddClose(false)

		Messagesession.Edit(struct {
			Name     string
			Username string
			TgId     int64
		}{
			Name:     upx.User.Name,
			Username: upx.User.Tguser.UserName,
			TgId:     upx.User.TgID,
		}, btns, C.TmpHelpHome)

		var (
			callback *tgbotapi.CallbackQuery
			err      error
			gotopage func(int, int, string) error
		)

		gotopage = func(page, max int, originpage string) error {
			if upx.Ctx.Err() != nil {
				return C.ErrContextDead
			}

			btns.Reset([]int16{2})
			btns.AddBack(false)
			if page != max {
				btns.AddBtcommon(C.BtnNext)
			}

			btns.AddClose(false)

			Messagesession.Edit(struct {
				*botapi.CommonUser
			}{
				&botapi.CommonUser{
					Name:     upx.User.Name,
					Username: upx.Chat.UserName,
					TgId:     upx.User.TgID,
				},
			}, btns, originpage+strconv.Itoa(page))

			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}

			switch callback.Data {
			case C.BtnBack:
				if page == 1 {
					return home()
				}
				return gotopage(page-1, max, originpage)
			case C.BtnNext:
				return gotopage(page+1, max, originpage)

			case C.BtnClose:
				//Messagesession.DeleteAllMsg()
				Messagesession.Callbackanswere(callback.ID, C.MsgHeloClosed, false)
				return nil
			}
			return nil
		}

		if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
			return err
		}

		switch callback.Data {

		case C.BtnClose:
			Messagesession.Callbackanswere(callback.ID, C.MsgHeloClosed, false)
			return nil

		case C.BtnFaq:
			Messagesession.Callbackanswere(callback.ID, C.MsgCallbackFaq, true)
			return home()

		case C.Btncommand:
			if !upx.User.Isverified() {
				Messagesession.Callbackanswere(callback.ID, C.Msghelpnoverify, true)
				return home()
			}

			return gotopage(1, int(C.HelpPags), C.TmpHelpCmPage)

		case C.BtnBtinfo:
			if !upx.User.Isverified() {
				Messagesession.Callbackanswere(callback.ID, C.Msghelpnoverify, true)
				return home()
			}

			return gotopage(1, int(C.InfoPage), C.TmpHelpInfoPage)

		case C.BtnAbout:
			btns.Reset([]int16{2})
			btns.AddCloseBack()

			Messagesession.Edit(struct {
				*botapi.CommonUser
			}{
				&botapi.CommonUser{
					Name:     upx.User.Name,
					Username: upx.Chat.UserName,
					TgId:     upx.User.TgID,
				},
			}, btns, C.TmpAbout)

			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}

			if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, home); ok {
				return err
			}

			return nil

		}

		return nil

	}

	return home()
}

// Deprecated: Use commandInfoV2
func (u *Xraywiz) commandInfo(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession( upx.Ctx, u.botapi, upx.User.Id, upx.User.Id, upx.User.Lang)
	Messagesession.AddreplyNoDelete(upx.Update.Message.MessageID)
	var (
		Usersession *controller.CtrlSession
		err         error
	)

	if Usersession, err = controller.NewctrlSession(u.ctrl, upx, false); err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.EditText(C.GetMsg(C.MsgSessionExcist), nil)
		}
		upx = nil
		Messagesession = nil
		Usersession = nil
		return nil
	}
	defer Usersession.Close()

	var (
		home           func() error
		totalrecursive int
	)

	home = func() error {
		if upx.Ctx.Err() != nil {
			return C.ErrContextDead
		}
		totalrecursive++
		if totalrecursive > u.ctrl.MaxRecurtion {
			return C.ErrRecurtionExceed
		}

		btns := botapi.NewButtons([]int16{2, 1})
		btns.Addbutton(C.BtnUserInfo, C.BtnUserInfo, "")
		btns.Addbutton(C.BtnConfigs, C.BtnConfigs, "")
		btns.AddClose(false)

		Messagesession.Edit(botapi.UpMessage{
			Template:     struct{}{},
			TemplateName: C.TmplGetinfoHome,
		}, btns, "")

		callback, err := u.callback.GetcallbackContext(upx.Ctx, btns.ID())
		if err != nil {
			return err
		}
		if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, home); ok {
			return err
		}

		switch callback.Data {

		case C.BtnUserInfo:
			btns.Reset([]int16{2})
			btns.AddCloseBack()

			if _, err = Messagesession.Edit(struct {
				*botapi.CommonUser
				Dedicated string

				TQuota       string
				LeftQuota    string
				ConfCount    int16
				TUsage       string
				GiftQuota    string
				Joined       string
				CapEndin     string
				Disendin     int32
				UsageResetIn int32

				Iscapped       bool
				Isgifted       bool
				Isdisuser      bool
				IsMonthLimited bool

				JoinedPlace uint
			}{

				CommonUser: &botapi.CommonUser{
					Name:     upx.User.Name,
					TgId:     upx.User.TgID,
					Username: upx.FromChat().UserName,
				},

				GiftQuota: upx.User.GiftQuota.BToString(),
				Joined:    upx.User.Joined.Format("2006-01-02 15:04:05"),
				Dedicated: C.Bwidth(u.ctrl.CommonQuota.Load()).BToString(),
				TQuota:    (Usersession.GetUser().CalculatedQuota + upx.User.AdditionalQuota).BToString(),
				LeftQuota: Usersession.LeftQuota().BToString(),
				TUsage:    Usersession.TotalUsage().BToString(),
				ConfCount: Usersession.GetUser().ConfigCount,
				CapEndin:  upx.User.Captime.AddDate(0, 0, 30).String(),

				Disendin:     ((u.ctrl.ResetCount - u.ctrl.CheckCount.Load()) * u.ctrl.RefreshRate) / 24,
				UsageResetIn: ((u.ctrl.ResetCount - u.ctrl.CheckCount.Load()) * u.ctrl.RefreshRate) / 24,

				Iscapped:       upx.User.IsCapped,
				IsMonthLimited: upx.User.IsMonthLimited,
				Isdisuser:      upx.User.IsDistributedUser,

				JoinedPlace: upx.User.CheckID,
			}, btns, C.TmpUserInfo); err != nil {
				u.logger.Error(err.Error())
				return home()
			}

			callback, err := u.callback.GetcallbackContext(upx.Ctx, btns.ID())
			if err != nil {
				if errors.Is(err, C.ErrContextDead) {
					return C.ErrContextDead
				}
				return home()
			}

			if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, home); ok {
				return err
			}

			return home()

		case C.BtnConfigs:

			if Usersession.GetUser().ConfigCount == 0 {
				Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgInfoNoconfigs), true)
				return home()
			}

			var confinfo func() error

			confinfo = func() error {
				if upx.Ctx.Err() != nil {
					return C.ErrContextDead
				}

				totalrecursive++
				if totalrecursive > u.ctrl.MaxRecurtion {
					return C.ErrRecurtionExceed
				}

				btns.Reset([]int16{2})

				for _, config := range Usersession.GetUser().Configs {
					btns.Addbutton(config.Name, strconv.Itoa(int(config.Id)), "")
				}
				btns.AddCloseBack()

				if _, err = Messagesession.EditText(C.GetMsg(C.MsgInfoSelectConfig), btns); err != nil {
					if errors.Is(err, C.ErrContextDead) {
						return err
					}
					return confinfo()
				}
				callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID())
				if err != nil {
					return err
				}
				if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, home); ok {
					return err
				}

				confid, err := strconv.Atoi(callback.Data)
				if err != nil {
					return confinfo()
				}

				selectedconfig, err := Usersession.GetConfig(int64(confid))

				if err != nil {
					Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgconfcannotfind), true)
					return confinfo()
				}

				btns.Reset([]int16{1, 2})
				//btns.Addbutton(C.BtnFullUsage, C.BtnFullUsage, "")
				btns.AddCloseBack()

				status, err := Usersession.Getstatus(int64(confid))

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

				sboxin, _ := u.ctrl.Getinbound(int(selectedconfig.InboundID))
				sboxout, _ := u.ctrl.Getoutbound(int(selectedconfig.OutboundID))

				if _, err = Messagesession.Edit(struct {
					*botapi.CommonUser
					//*botapi.CommonUsage

					TotalQuota string

					ConfigName string
					ConfigType string
					ConfigUUID fmt.Stringer

					ConfigUpload     string
					ConfigDownload   string
					ConfigUploadtd   string
					ConfigDownloadtd string
					ConfigUsage      string
					ConfigUsagetd    string
					UsedPresenTage   float64

					ResetDays int32

					InName         string
					InType         string
					InPort         int
					InAddr         string
					InInfo         string
					TranstPortType string
					TlsEnabled     bool
					SupportInfo    []string

					OutName string
					OutType string
					OutInfo string
					Latency int32

					UsageDuration string

					Online int
					IpMap  map[netip.Addr]int64
				}{
					CommonUser: &botapi.CommonUser{
						Name:     upx.User.Name,
						Username: upx.FromChat().UserName,
						TgId:     upx.User.TgID,
					},

					TotalQuota:     selectedconfig.Quota.BToString(),
					ConfigName:     selectedconfig.Name,
					ConfigType:     selectedconfig.Type,
					ConfigUUID:     selectedconfig.UUID,
					UsedPresenTage: float64(int(((selectedconfig.Usage+status.FullUsage()).Float64()/selectedconfig.Quota.Float64())*100*1000)) / 1000,
					//UsedPresenTage: (((selectedconfig.Usage + status.FullUsage()).Float64()/selectedconfig.Quota.Float64()))*100,

					ResetDays: ((u.ctrl.ResetCount - u.ctrl.CheckCount.Load()) * u.ctrl.RefreshRate) / 24,

					ConfigDownload: (selectedconfig.Download + status.Download).BToString(),
					ConfigUpload:   (selectedconfig.Upload + status.Upload).BToString(),

					ConfigDownloadtd: (status.Download).BToString(),
					ConfigUploadtd:   (status.Upload).BToString(),

					ConfigUsagetd: (status.Download + status.Upload).BToString(),
					ConfigUsage:   (status.Download + status.Upload + selectedconfig.Usage).BToString(),

					InName:         sboxin.Name,
					InType:         sboxin.Type,
					InPort:         sboxin.Port(),
					InAddr:         sboxin.Laddr(),
					InInfo:         sboxin.Custom_info,
					TranstPortType: sboxin.TransortType(),
					TlsEnabled:     sboxin.TlsIsEnabled(),
					UsageDuration:  time.Since(u.ctrl.GetLastRefreshtime()).Round(1 * time.Second).String(),
					SupportInfo:    sboxin.Support,

					OutName: sboxout.Name,
					OutType: sboxout.Type,
					OutInfo: sboxout.Custom_info,
					Latency: sboxout.Latency.Load(),

					Online: len(status.Online_ip),
					IpMap:  status.Online_ip,
					//TODO: fill here
				}, btns, C.TmpConfigInfo); err != nil {

					u.logger.Error(err.Error())
					switch {
					case errors.Is(err, C.ErrClientRequestFail):
						time.Sleep(100 * time.Millisecond)
						return confinfo()
					case errors.Is(err, C.ErrContextDead), errors.Is(err, C.ErrTmplRender):
						return err

					}

				}
				callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID())

				if err != nil {
					return err
				}
				if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, confinfo); ok {
					return err
				}

				if callback.Data == C.BtnFullUsage {
					//TODO: cde here

					Messagesession.SendAlert(C.GetMsg(C.Msgconfcannotfind), nil)
					Messagesession.SendAlert("usage history function is not avalable yet", nil)
				}

				return confinfo()
			}

			return confinfo()
		
		case C.BtnCheckOutbounds:
			if upx.Ctx.Err() != nil {
				return err
			}
		
		}
		return nil
	}

	if err = home(); err != nil {
		if errors.Is(err, C.ErrContextDead) {
			tempctx, cancle := context.WithTimeout(u.ctx, 1*time.Minute)
			Messagesession.SetNewcontext(tempctx)
			Messagesession.DeleteAllMsg()
			Messagesession.EditText(C.GetMsg(C.MsgSessionOver), nil)
			cancle()
			err = nil
		} else if errors.Is(err, C.ErrRecurtionExceed) {
			Messagesession.SendAlert(C.GetMsg(C.MsgRecursionExceed), nil)
			err = nil
		}
	}

	return err
}
