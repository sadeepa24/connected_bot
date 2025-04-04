package update

import (
	"sync"

	C "github.com/sadeepa24/connected_bot/constbot"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
)

type UpdatePool struct {
	pool sync.Pool

}

func NewupdatePool() *UpdatePool {
	return &UpdatePool{
		pool: sync.Pool{
			New: func() any {
				return &Updatectx{}
			},
		},
	}
}


func (u *UpdatePool) Newupdate(origin *tgbotapi.Update) *Updatectx {
	
	if origin.Message != nil {}

	upx := u.pool.Get().(*Updatectx)

	upx.Update = origin
	//upx.iscallback = false
	upx.Ctx = nil
	upx.Command = C.CmdNull
	upx.Serviceset = false
	return upx
}

func (u *UpdatePool) Put(upx *Updatectx) {
	u.pool.Put(upx)
}