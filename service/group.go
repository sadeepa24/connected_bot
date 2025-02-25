//go:build ignore

package service

import (
	"context"
	"fmt"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)


type Groupsrv struct {
	ctx      context.Context
	callback *Callback
	logger   zap.Logger
	db       *db.Database
	admin    *Adminsrv
	botapi   botapi.BotAPI
}

func (g *Groupsrv) Init() error {
	return nil
}

func (g *Groupsrv) Exec(upx *update.Updatectx) error {
	return fmt.Errorf("not implemented")
}

func (g *Groupsrv) Name() string {
	return "grousrv"
}

func (g *Groupsrv) Canhandle(upx *update.Updatectx) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
