package service

import (
	"context"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)

type Service interface {
	Exec(*update.Updatectx) error
	Init() error
	Name() string
	Canhandle(*update.Updatectx) (bool, error)
}

func GetallService(ctx context.Context, logger *zap.Logger, ctrl *controller.Controller, btapi botapi.BotAPI, msgstore *botapi.MessageStore) ([]Service, error) {
	if err := C.LoadUserMsg(); err != nil {
		return nil, err
	}
	
	// if err = msgstore.Init(btapi, ctrl.SudoAdmin, logger); err != nil {
	// 	return nil, err
	// }

	callbacksrv := NewCallback(ctx, logger, ctrl, btapi)
	defaultsrv := NewDefaulsrv(ctx, callbacksrv, logger)
	xraysrv := NewXraywiz(ctx, callbacksrv, logger, ctrl, defaultsrv, btapi, msgstore)
	adminsrv := NewAdminsrv(ctx, logger, callbacksrv, defaultsrv, xraysrv, btapi, ctrl, msgstore )
	usersrv := NewuserService(ctx, callbacksrv, logger, adminsrv, ctrl, defaultsrv, btapi, msgstore)
	inline := NewInline(ctx, logger, btapi, ctrl)

	allservice := []Service{
		callbacksrv, adminsrv, defaultsrv, usersrv, xraysrv, inline,
	}

	return allservice, nil
}
