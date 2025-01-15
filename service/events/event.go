package events

import (
	"context"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/common"
	"github.com/sadeepa24/connected_bot/controller"

	"github.com/sadeepa24/connected_bot/update"
)

//all events should implemet this interface after creating a event add it to GetallAvblkEvent
type Event interface {
	Name() string
	Expired() bool
	Price() int64
	Excute(eventctx Eventctx) error
	ExcuteComplete(eventctx Eventctx) error
	Info() string
}
type Eventctx struct {
	Ctx             context.Context        // required
	Upx             *update.Updatectx      // required
	Btns            *botapi.Buttons        // required
	Alertsender     common.Alertsender     // required
	Sendreciver     common.Sendreciver     // required
	Callbackreciver common.Callbackreciver // required
	Ctrl            *controller.Controller  // required by aplusconf
	//ConfStore *builder.ConfigStore  // required by aplusconf
}

const (
	CurrentEvents = 1
)

func GetallAvblkEvent(ctrl *controller.Controller) map[string]Event {
	eve := make(map[string]Event, CurrentEvents)
	aplus := &Aplusconf{
		ctrl:     ctrl,
		duration: 30,
		price:    20,
		makeDate: "2025-01-02 12:01:01",
	}
	eve[aplus.Name()] = aplus
	return eve
}
