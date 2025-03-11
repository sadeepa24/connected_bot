package service

import (
	"context"
	"errors"
	"sync"

	//
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)

type Callback struct {
	ctx         context.Context
	currentcall sync.Map
	logger      *zap.Logger
	ctrl        *controller.Controller
	botapi      botapi.BotAPI

	chanPool	sync.Pool
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
		chanPool: sync.Pool{
			New: func() any {
				cbackchan := make(chan *tgbotapi.CallbackQuery)
				return cbackchan
			},
		},
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

		select {
		case val <- upx.Update.CallbackQuery:
		case <-c.ctx.Done():
			return errors.New("context canceled while sending callback query")
		default:
			return errors.New("channel is closed or full")
		}

	}
	return errors.New("callback quary not found error")
}

func (c *Callback) Getcallback(uniqid int64) (*tgbotapi.CallbackQuery, error) {
	return c.GetcallbackContext(context.Background(), uniqid)
}

func (c *Callback) GetcallbackContext(ctx context.Context, uniqid int64) (*tgbotapi.CallbackQuery, error) {
	cbackchan := c.chanPool.Get().(chan *tgbotapi.CallbackQuery)
	c.currentcall.Store(uniqid, cbackchan)
	if len(cbackchan) > 0 {
		<- cbackchan
	}
	var (
		err error
		val *tgbotapi.CallbackQuery
	)
	select {
	case <-ctx.Done():
		err = C.ErrContextDead
		close(cbackchan) // more safe
	case val = <-cbackchan:
		c.chanPool.Put(cbackchan)
	}
	c.currentcall.Delete(uniqid)

	return val, err

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
	return (upctx.Update != nil) && (upctx.Update.CallbackQuery != nil), nil
}
