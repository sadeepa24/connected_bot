package service

import (

	//

	"strconv"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/tg/tgbotapi"
)

func CreateConfig(opts common.OptionExcutors) error {
	btns := opts.Btns
	Messagesession := opts.MessageSession
	Usersession := opts.Usersession
	var (
		err      error
		callback *tgbotapi.CallbackQuery
	)

	cuurentIns := make(map[int16]string, len(opts.Ctrl.Inbounds))
	//select inbound for new config
	inselectloop:
	for {
		btns.Reset([]int16{})
		for _, inbound := range opts.Ctrl.Getinbounds() {
			if _, ok := cuurentIns[int16(inbound.Id)]; ok {
				btns.Addbutton(inbound.Type +"_"+ strconv.Itoa(inbound.Port()) + C.GetMsg(C.ButtonSelectEmjoi), strconv.Itoa(int(inbound.Id)), "")
				continue
			}
			btns.Addbutton(inbound.Type +"_"+ inbound.Tag, strconv.Itoa(int(inbound.Id)), "")
		}
		btns.Passline()
		btns.AddBtcommon(C.BtnDone)
		btns.AddClose(true)
		if callback, err = opts.Callbackreciver(C.GetMsg(C.MsgselectIn), btns); err != nil {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionOver), nil)
			return err
		}
		if callback.Data == C.BtnDone {
			if len(cuurentIns) == 0 {
				Messagesession.Callbackanswere(callback.ID, "you have to select at least 1 inbound before presing done", true)
				continue
			}
			break inselectloop
		}
		if ok, err := closeback(callback.Data, Messagesession.DeleteAllMsg, func() error { return nil }); ok {
			return err
		}
		var inID int
		if inID, err = strconv.Atoi(callback.Data); err != nil {
			Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
			return err
		}

		sboxin, loaded := opts.Ctrl.Getinbound(int16(inID))
		if !loaded {
			Messagesession.SendAlert(C.GetMsg(C.MsgCrInerr), nil)
			return nil
		}
		btns.Reset([]int16{2})
		btns.AddBtcommon(C.BtnConform)
		btns.Addcancle()
	
		if callback, err = opts.Callbackreciver(botapi.UpMessage{
			Template: sboxin,
			TemplateName: C.TmplInInfo,
		}, btns); err != nil {
			return err
		}

		if callback.Data == C.BtnCancle {
			continue
		}
		if _, ok := cuurentIns[int16(inID)]; ok {
			delete(cuurentIns, int16(inID))
			continue
		}
		cuurentIns[int16(inID)] = sboxin.Tag 
	}
	if len(cuurentIns) == 0 {
		Messagesession.SendExtranal("you have to select at least 1 inbound before presing done", nil, "", true)
		return C.ErrNoAnyInbound
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
	sboxout, loaded := opts.Ctrl.Getoutbound(int16(outID))
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

	if Usersession.LeftQuota() <= 0 {
		Messagesession.SendAlert(C.GetMsg(C.MsgnoQuota), nil)
		return nil
	}
	var reduce C.Bwidth

	// checks whether user has usage of deleted configs
	// reduce will be that deleted config's usage
	if 	Usersession.TotalUsage() != Usersession.GetFullUsage().Full()  {
		Messagesession.SendAlert(C.GetMsg(C.MsgCrQuotaNote), nil)
		reduce = Usersession.TotalUsage() - Usersession.GetFullUsage().Full()
	}

	if Usersession.LeftQuota() - reduce <= 0 {
		Messagesession.SendAlert(C.GetMsg(C.MsgNoQuota), nil)
		return nil
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

	quotafroconfig, err := common.ReciveBandwidth(opts.Tgcalls, (Usersession.LeftQuota() - reduce), 0 )
	if err != nil {
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert("Bandwidth Recive Failed", nil)
		Messagesession.SendError(err, "")
		return err
	}
	if _, err = Messagesession.Edit(C.GetMsg(C.MsgGetName), nil, ""); err != nil {
		return err
	}
	confName, err := common.ReciveName(opts.Tgcalls)
	if err != nil {
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert("Name Recive Failed", nil)
		Messagesession.SendError(err, "")
		return err
	}
	if _, err := Messagesession.EditText(C.GetMsg(C.MsgCrLogin), nil); err != nil {
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return err
	}
	LoginLimit, err := common.ReciveInt(opts.Tgcalls, int(opts.Ctrl.LoginLimit), 0)
	if err != nil {
		Messagesession.SendAlert("LoginLimit Recive Failed", nil)
		Messagesession.SendError(err, "")
		return err
	}
	opts.Ctrl.IncCriticalOp()
	defer opts.Ctrl.DecCriticalOp()
	newconfig, err := Usersession.AddNewConfig(C.MapToSliceKey(cuurentIns), int16(outID), C.Bwidth(quotafroconfig).GbtoByte(), int16(LoginLimit), confName)
	if err != nil {
		Messagesession.SendError(err, C.GetMsg(C.MsgCrFailed))
		return err
	}

	Messagesession.DeleteAllMsg()
	Messagesession.SendAlert(C.GetMsg(C.MsgCrsuccsess), nil)
	Messagesession.SendExtranal(*newconfig, nil, C.TmpSendConf, true)
	opts.Alertsender(C.GetMsg(C.MsgCrConfigIn))
	return nil
}