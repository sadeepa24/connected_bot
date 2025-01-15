package service

import (
	"errors"
	"strconv"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
)

type creator interface {
	name
	Excute(opts common.OptionExcutors) error
}

type vlessCreator struct{}

var _ creator = (*vlessCreator)(nil)

func (v *vlessCreator) Name() string { return C.Vless }

// needed options
// sendreciver should be compitable with only for reciving
// All option in opts needed for this Method
func (v *vlessCreator) Excute(opts common.OptionExcutors) error {
	btns := opts.Btns
	Messagesession := opts.MessageSession
	upx := opts.Upx
	Usersession := opts.Usersession
	var (
		err      error
		callback *tgbotapi.CallbackQuery
	)
	btns.Reset([]int16{})

	for _, inbound := range opts.Ctrl.Getinbounds() {
		btns.Addbutton(inbound.Type +"_"+ strconv.Itoa(inbound.Port()), strconv.Itoa(int(inbound.Id)), "")
	}

	btns.AddClose(true)

	if callback, err = opts.Callbackreciver(C.GetMsg(C.MsgselectIn), btns); err != nil {
		Messagesession.SendAlert(C.GetMsg(C.MsgSessionOver), nil)
		return err
	}

	if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, func() error { return nil }); ok {
		return err
	}

	var inID int

	if inID, err = strconv.Atoi(callback.Data); err != nil {
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}

	sboxin, loaded := opts.Ctrl.Getinbound(inID)

	if !loaded {
		Messagesession.SendAlert(C.GetMsg(C.MsgCrInerr), nil)
		return nil
	}
	btns.Reset([]int16{2})
	btns.AddBtcommon(C.BtnConform)
	btns.Addcancle()

	if callback, err = opts.Callbackreciver(botapi.UpMessage{
		Template: struct {
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
		},
		TemplateName: C.TmpCrInInfo,
	}, btns); err != nil {
		return err
	}

	if err = checkconform(callback.Data, Messagesession); err != nil {
		return err
	}

	btns.Reset([]int16{2})

	for _, outbound := range opts.Ctrl.Getoutbounds() {
		btns.Addbutton(outbound.Type+"_"+outbound.Tag, strconv.Itoa(int(outbound.Id)), "")
	}

	btns.AddClose(true)

	if callback, err = opts.Callbackreciver(C.GetMsg(C.MsgselectOut), btns); err != nil {
		return err
	}

	if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, func() error {
		return nil
	}); ok {
		return err
	}

	var outID int

	if outID, err = strconv.Atoi(callback.Data); err != nil {
		return err
	}

	sboxout, loaded := opts.Ctrl.Getoutbound(outID)

	if !loaded {
		Messagesession.SendAlert(C.GetMsg(C.MsgselectOut), nil)
		return nil
	}

	btns.Reset([]int16{2})
	btns.AddBtcommon(C.BtnConform)
	btns.Addcancle()

	if callback, err = opts.Callbackreciver(botapi.UpMessage{
		Template: struct {
			OutName string
			OutType string
			OutInfo string
			Latency int32
		}{
			OutName: sboxout.Name,
			OutType: sboxout.Type,
			OutInfo: sboxout.Custom_info,
			Latency: sboxout.Latency.Load(),
		},
		TemplateName: C.TmpCrOutInfo,
	}, btns); err != nil {
		return err
	}

	if err = checkconform(callback.Data, Messagesession); err != nil {
		return err
	}

	// Selecting Quota for New creating config

	if Usersession.LeftQuota() <= 0 {
		Messagesession.SendAlert(C.GetMsg(C.MsgnoQuota), nil)
		return nil
	}

	fusage := Usersession.GetFullUsage()
	var reduce C.Bwidth
	if upx.User.MonthUsage+fusage.Downloadtd+fusage.Uploadtd != fusage.Download+fusage.Upload {
		Messagesession.SendAlert(C.GetMsg(C.MsgCrQuotaNote), nil)
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
		replymeassage  *tgbotapi.Message
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

		if replymeassage, err = opts.Sendreciver(nil); err != nil {
			return err
		}
		Messagesession.Addreply(replymeassage.MessageID)

		quotafroconfig, err = strconv.Atoi(replymeassage.Text)
		if err != nil {
			Messagesession.SendNew(C.GetMsg(C.MsgValidInt), nil, "")
			return quotacalc()
		}

		if quotafroconfig <= 0 {
			Messagesession.SendAlert("cannot be 0 or minus ", nil)
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

	if _, err = Messagesession.SendNew(C.GetMsg(C.MsgGetName), nil, ""); err != nil {
		return err
	}

	var getName func() error
	var confName string

	getName = func() error {
		if replymeassage, err = opts.Sendreciver(nil); err != nil {
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
	if _, err := Messagesession.EditText(C.GetMsg(C.MsgCrLogin), nil); err != nil {
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

		if replymeassage, err = opts.Sendreciver(nil); err != nil {
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
			Messagesession.SendAlert(C.GetMsg(C.MsgCrLoginwarn), nil)
			continue
		}

		break

	}

	config, err := Usersession.AddNewConfig(int16(inID), int16(outID), C.Bwidth(quotafroconfig).GbtoByte(), int16(LoginLimit), confName)

	if err != nil {
		opts.Logger.Error("Error When Config Create - " +  err.Error())
		switch {
		case errors.Is(err, C.ErrInboundNotFound), errors.Is(err, C.ErrDatabaseCreate), errors.Is(err, C.ErrTypeMissmatch), errors.Is(err, C.ErrContextDead):
			Messagesession.SendAlert(C.GetMsg(C.MsgCrFailed), nil)

		default:
			Messagesession.SendAlert(C.GetMsg(C.MsgInternalErr), nil)
		}

		return err

	}

	Messagesession.DeleteAllMsg()
	Messagesession.SendAlert(C.GetMsg(C.MsgCrsuccsess), nil)

	Messagesession.SendExtranal(struct {
		UUID          string
		Domain        string
		Transport     string
		ConfigName    string
		TlsEnabled    bool
		Port          int
		Path          string
		TransportType string
		*botapi.CommonUser
	}{
		CommonUser: &botapi.CommonUser{
			Name:     opts.Upx.User.Name,
			Username: opts.Upx.Chat.UserName,
			TgId:     opts.Upx.User.TgID,
		},
		Domain:        sboxin.Domain,
		ConfigName:    sboxin.Name,
		Port:          sboxin.Port(),
		Transport:     sboxin.Transporttype,
		TlsEnabled:    sboxin.Tlsenabled,
		UUID:          config.UUID.String(),
		Path:          sboxin.Option.VLESSOptions.GetPath(),
		TransportType: sboxin.Option.VLESSOptions.TransportType(),
	}, nil, C.TmpCrSendUID, true)

	//Messagesession.SendExtranal(fmt.Sprintf("vless://%v@%v:%v?path=%v&security=%v&type=", config.UUID.String(), sboxin.Domain, sboxin.Port(), sboxin.Option.VLESSOptions.GetPath()     ), nil, "", true)

	opts.Alertsender(C.GetMsg(C.MsgCrConfigIn))

	return nil
}

func allcreators() []creator {
	creators := []creator{}

	creators = append(creators, &vlessCreator{})
	return creators
}
