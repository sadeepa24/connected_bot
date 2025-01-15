package controller

import (
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
)

type Allconfigs []*db.Config

func (a Allconfigs) RemoveAll() {

}

func (a Allconfigs) Appendconf(conf *sbox.Userconfig, userid int64) {

}
