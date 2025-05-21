package service

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	//
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
)

const (
	sthomehelp   = 0
	sthelpclosed = 1
	sthelpabout  = 3
	stgototpage  = 4
)

type HelpState struct {
	ctx            context.Context
	cancel context.CancelFunc
	
	counter *atomic.Int32

	State          int
	Messagesession *botapi.Msgsession
	btns           *botapi.Buttons
	wiz            *Usersrv
	upx            *update.Updatectx
	Page           int
	MaxPages       int
	PageName       string

	helperinfo *C.HelpCommandInfo
}

func (h *HelpState) home() error {
	h.btns.Reset([]int16{2, 2, 1, 1})
	h.btns.AddBtcommon(C.Btncommand)
	h.btns.AddBtcommon(C.BtnBtinfo)
	h.btns.AddBtcommon(C.BtnBuilderHelp)
	//h.btns.AddBtcommon(C.BtnFaq)
	h.btns.AddBtcommon(C.BtnTutorial)
	h.btns.AddBtcommon(C.BtnAbout)

	h.btns.AddClose(false)

	h.Messagesession.Edit(struct {
		botapi.CommonUser
	}{
		CommonUser: botapi.CommonUser{
			Name: h.upx.User.Name,
			Username: h.upx.User.Tguser.UserName,
			TgId:     h.upx.User.TgID,
		},
		
	}, h.btns, C.TmpHelpHome)

	var (
		callback *tgbotapi.CallbackQuery
		err      error
	)

	if callback, err = h.wiz.callback.GetcallbackContext(h.ctx, h.btns.ID()); err != nil {
		return err
	}
	h.counter.Add(1)
	switch callback.Data {
	case C.BtnClose:
		h.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgHeloClosed), false)
		h.State = sthelpclosed
		return nil

	case C.BtnFaq:
		h.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgCallbackFaq), true)
		return nil

	case C.Btncommand, C.BtnBtinfo, C.BtnBuilderHelp, C.BtnTutorial:
		if !h.upx.User.Isverified() {
			h.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msghelpnoverify), true)
			return nil
		}

		switch callback.Data {
		case C.Btncommand:
			h.PageName = C.TmpHelpCmPage
			h.MaxPages = int(h.helperinfo.CommandPageCount)
		case C.BtnBtinfo:
			h.PageName = C.TmpHelpInfoPage
			h.MaxPages = int(h.helperinfo.InfoPageCount)
		case C.BtnBuilderHelp:
			h.PageName = C.TmplHelpBuilderHelp
			h.MaxPages = int(h.helperinfo.BuilderHelp)
		case C.BtnTutorial:
			h.PageName = C.TmplHelpTuto
			h.MaxPages = int(h.helperinfo.TutorialPageCount)
		}

		h.Page = 1
		h.State = stgototpage

	case C.BtnAbout:
		h.State = sthelpabout

	}
	return nil
}

func (h *HelpState) about() error {
	h.btns.Reset([]int16{2})
	h.btns.AddCloseBack()
	h.Messagesession.Edit(struct {
		*botapi.CommonUser
	}{
		&botapi.CommonUser{
			Name:     h.upx.User.Name,
			Username: h.upx.Chat.UserName,
			TgId:     h.upx.User.TgID,
		},
	}, h.btns, C.TmpAbout)

	var callback *tgbotapi.CallbackQuery
	var err error
	if callback, err = h.wiz.callback.GetcallbackContext(h.ctx, h.btns.ID()); err != nil {
		return err
	}
	h.counter.Add(1)
	switch callback.Data {
	case C.BtnClose:
		h.Messagesession.RemoveBtns()
		h.State = sthelpclosed
	case C.BtnBack:
		h.State = sthomehelp
	}
	return nil
}

func (h *HelpState) gotopage() error {
	h.btns.Reset([]int16{2})
	h.btns.AddBack(false)
	if h.Page != h.MaxPages {
		h.btns.AddBtcommon(C.BtnNext)
	}
	h.btns.AddClose(false)

	_, err := h.Messagesession.Edit(struct {
		*botapi.CommonUser
	}{
		&botapi.CommonUser{
			Name:     h.upx.User.Name,
			Username: h.upx.Chat.UserName,
			TgId:     h.upx.User.TgID,
		},
	}, h.btns, h.PageName+strconv.Itoa(h.Page))
	if err != nil {
		return err
	}

	var (
		callback *tgbotapi.CallbackQuery
	)

	if callback, err = h.wiz.callback.GetcallbackContext(h.ctx, h.btns.ID()); err != nil {
		return err
	}
	h.counter.Add(1)

	switch callback.Data {
	case C.BtnBack:
		if h.Page == 1 {
			h.State = sthomehelp
			return nil
		}
		h.Page--
	case C.BtnNext:
		h.Page++
	case C.BtnClose:
		//Messagesession.DeleteAllMsg()
		h.Messagesession.RemoveBtns()
		h.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgHeloClosed), false)
		h.State = sthelpclosed
		return nil
	}
	return nil
}

func (h *HelpState) timeout() {
	timeout(30 * time.Second, h.counter, h.ctx, h.cancel, h.Messagesession)
}

func (u *Usersrv) commandHelpV2(upx *update.Updatectx, Messagesession *botapi.Msgsession) error {
	btns := botapi.NewButtons([]int16{1})
	ctx, cancelfunc := context.WithTimeout(u.ctrl.GetBaseContext(), 7 * time.Minute)
	state := HelpState{
		ctx: ctx,
		cancel: cancelfunc,
		Messagesession: Messagesession,
		btns:           btns,
		wiz:            u,
		upx:            upx,
		helperinfo:     u.ctrl.GetHelepCmdInfo(),
		counter: new(atomic.Int32),
	}
	state.counter.Add(2)
	Messagesession.SetNewcontext(ctx)
	go state.timeout()
	go func ()  {
		var err error
		defer state.cancel()
		help:
		for {
			switch state.State {
			case sthomehelp:
				err = state.home()
			case stgototpage:
				err = state.gotopage()
			case sthelpabout:
				err = state.about()
			case sthelpclosed:
				return
			default:
				break help
			}
			if err != nil || ctx.Err() != nil {
				return
			}
		}
	}()
	return nil
}
