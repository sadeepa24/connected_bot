package service

import (
	"context"
	"errors"
	"sync"

	//
	C "github.com/sadeepa24/connected_bot/constbot"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)

type Defaultsrv struct {
	ctx      context.Context
	callback *Callback
	logger   zap.Logger
	chanPool sync.Pool
	//admin    *Adminsrv
	//ctrl *controller.Controller
	//botapi botapi.BotAPI

	msgpool *sync.Map
}

// type store struct {
// 	msg *tgbotapi.Message
// 	chatid int64
// 	userid int64
// }

func NewDefaulsrv(
	ctx context.Context,
	callback *Callback,
	logger *zap.Logger,

) *Defaultsrv {
	return &Defaultsrv{
		msgpool:  &sync.Map{},
		logger:   *logger,
		callback: callback,
		ctx:      ctx,
		chanPool: sync.Pool{
			New: func() any {
				cbackchan := make(chan *tgbotapi.Message)
				return cbackchan
			},
		},
	}
}

func (d *Defaultsrv) Init() error {
	return nil
}

func (d *Defaultsrv) Exec(upx *update.Updatectx) error {
	if upx.FromChat() == nil || upx.FromUser() == nil {
		// prosess this later
		return nil
	}

	var (
		val any
		ok  bool
	)

	if val, ok = d.msgpool.Load(upx.FromChat().ID + upx.FromUser().ID); !ok {
		return nil
	}
	var sendchan chan *tgbotapi.Message
	if sendchan, ok = val.(chan *tgbotapi.Message); !ok {
		return nil
	}
	select {
	case sendchan <- upx.Update.Message:
	case <-d.ctx.Done():
		return errors.New("context canceled while sending callback query")
	default:
		return errors.New("channel is closed or full")
	}
	return nil
}

func (d *Defaultsrv) ExcpectMsg(userId int64, chatId int64) (*tgbotapi.Message, error) {
	return d.ExcpectMsgContext(context.Background(), userId, chatId)
}

func (d *Defaultsrv) ExcpectMsgContext(ctx context.Context, userID int64, chatId int64) (*tgbotapi.Message, error) {
	comebkchan := d.chanPool.Get().(chan *tgbotapi.Message)
	d.msgpool.Store(chatId+userID, comebkchan)

	select {
	case <-ctx.Done():
		d.msgpool.Delete(chatId + userID)
		close(comebkchan)
		return nil, C.ErrContextDead
	case val := <-comebkchan:
		d.msgpool.Delete(chatId + userID)
		d.chanPool.Put(comebkchan)
		return val, nil
	}

}

func (d *Defaultsrv) Ismsgrequired(userId int64, Chatid int64) bool {
	var ok bool
	_, ok = d.msgpool.Load(userId + Chatid)
	return ok
}

func (d *Defaultsrv) Name() string {
	return C.Defaultservicename
}

func (d *Defaultsrv) Canhandle(upx *update.Updatectx) (bool, error) {
	return upx.Service == C.Defaultservicename, nil
}

func (d *Defaultsrv) FromserviceExec(upx *update.Updatectx) error {
	upx.Cancle()
	return nil
}

func (d *Defaultsrv) Droper(upx *update.Updatectx) error {
	upx.Cancle()
	if upx.User == nil {
		return nil
	}
	upx.User.User = nil
	upx.User = nil
	upx.Update = nil

	return nil
}
