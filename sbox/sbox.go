package sbox

import (
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox/conf"
	"github.com/sagernet/sing-box/connectedbot/opts"
)


type Controller interface {
	Start() error
	Close() error

	AddConfig(dbconf *db.Config) (conf.Sboxstatus, error)
	AddConfigReset(dbconf *db.Config) (conf.Sboxstatus, error)
	RemoveConfig(dbconf *db.Config) (conf.Sboxstatus, error)
	GetStatusConfig(dbconf *db.Config) (conf.Sboxstatus, error)

	GetAllUserStatus() map[int]opts.UserStatus
	//GetAllInbound() ([]conf.Inboud, error)
	CloseConns(dbconf *db.Config) error

	UrlTest(outtag string) (int16, error)
	RefreshUrlTest()

	ResetInbounds(dbconf *db.Config) error
	ChangeOutbound(dbconf *db.Config) error
}
