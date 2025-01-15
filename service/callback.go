package service

import (
	"context"
	"errors"
	"sync"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"go.uber.org/zap"
)

type Callback struct {
	ctx         context.Context
	currentcall sync.Map
	logger      *zap.Logger
	ctrl        *controller.Controller
	botapi      botapi.BotAPI
}

func NewCallback(
	ctx context.Context,
	logger *zap.Logger,
	ctrl *controller.Controller,
	botapi botapi.BotAPI,

) *Callback {
	callbk := Callback{
		ctx:         ctx,
		currentcall: sync.Map{},
		logger:      logger,
		ctrl:        ctrl,
		botapi:      botapi,
	}
	return &callbk
}

func (c *Callback) Exec(upx *update.Updatectx) error {
	var (
		upstremer any
		loaded    bool
		ok        bool
		val       chan *tgbotapi.CallbackQuery
		err       error
	)
	if upx.Update.CallbackQuery == nil {
		return nil
	}
	cData := botapi.Callbackdata{}
	if err = cData.FillV2(upx.Update.CallbackData()); err != nil {
		c.botapi.AnswereCallbackCtx(upx.Ctx, &botapi.Callbackanswere{
			Callback_query_id: upx.Update.CallbackQuery.ID,
			Show_alert:        false,
			Text:              C.GetMsg(C.MsgBtnOffline),
		})
		return err
	}
	// if upstremer, loaded = c.currentcall.LoadAndDelete(upx.FromChat().ID + upx.FromUser().ID ); !loaded { // TODO: should change ID
	// 	return nil
	// }

	if upstremer, loaded = c.currentcall.LoadAndDelete(cData.Uniqid); !loaded {
		if cData.Data == C.BtnClose {
			if upx.Update.CallbackQuery.Message != nil {
				if c.botapi.DeleteMsg(upx.Ctx, int64(upx.Update.CallbackQuery.Message.MessageID), upx.FromChat().ID) == nil {
					return nil
				}

			}
		}
		c.botapi.AnswereCallbackCtx(upx.Ctx, &botapi.Callbackanswere{
			Callback_query_id: upx.Update.CallbackQuery.ID,
			Show_alert:        false,
			Text:              C.GetMsg(C.MsgBtnOffline),
			//Text: C.MsgBtnOffline ,
		})
		return nil
	}

	if val, ok = upstremer.(chan *tgbotapi.CallbackQuery); !ok {
		return nil
	}
	if upx.Update.CallbackQuery != nil {
		upx.Update.CallbackQuery.Data = cData.Data

		val <- upx.Update.CallbackQuery

		return nil
	}
	return errors.New("callback quary not found error")
}

func (c *Callback) Getcallback(uniqid int64) (*tgbotapi.CallbackQuery, error) {
	return c.GetcallbackContext(context.Background(), uniqid)
}

func (c *Callback) GetcallbackContext(ctx context.Context, uniqid int64) (*tgbotapi.CallbackQuery, error) {
	cbackchan := make(chan *tgbotapi.CallbackQuery)
	c.currentcall.Store(uniqid, cbackchan)
	select {
	case <-ctx.Done():
		c.currentcall.Delete(uniqid)
		close(cbackchan)
		return nil, C.ErrContextDead
	case val := <-cbackchan:
		c.currentcall.Delete(uniqid)
		close(cbackchan)
		return val, nil
	}

}

func (c *Callback) Sendget(ctx context.Context) {

}

func (c *Callback) Init() error {
	return nil
}

func (c *Callback) Name() string {
	return "callback"
}

func (c *Callback) Canhandle(upctx *update.Updatectx) (bool, error) {
	return upctx.Iscallback()
}
