package singapi

import (
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sagernet/sing-box/connectedbot/opts"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)
type Comconf struct {
	dbconf *db.Config
	inboundtypes map[string]string
}

var _ opts.ComProto = (*Comconf)(nil)

func (c *Comconf) Vless() (option.VLESSUser, bool) {
	if _, ok := c.inboundtypes[C.TypeVLESS]; ok {
		return option.VLESSUser{
			Name: c.dbconf.GetuniqName(),
			UUID: c.dbconf.UUID,
			// Flow: "", //TODO: Add Flow later
		}, true
	}
	return option.VLESSUser{}, false
}

func (c *Comconf) Vmess() (option.VMessUser, bool) {
	if _, ok := c.inboundtypes[C.TypeVMess]; ok {
		return option.VMessUser{
			Name: c.dbconf.GetuniqName(),
			UUID: c.dbconf.UUID,
			//AlterId: ,//TODO: Add AlterID later
		}, true
	}
	return option.VMessUser{}, false
}

func (c *Comconf) Trojan() (option.TrojanUser, bool) {
	if _, ok := c.inboundtypes[C.TypeTrojan]; ok {
		return option.TrojanUser{
			Name: c.dbconf.GetuniqName(),
			Password: c.dbconf.Password,
		}, true
	}
	return option.TrojanUser{}, false
}

func (c *Comconf) Tuic() (option.TUICUser, bool) {
	return option.TUICUser{}, false
}

func (c *Comconf) UserStr() string {
	return c.dbconf.GetuniqName()
}
func (c *Comconf) Password() string {
	return c.dbconf.Password
}

func (c *Comconf) UUID() string {
	return c.dbconf.UUID
}
func (c *Comconf) UserName() string {
	return c.dbconf.GetuniqName()
}

func (c *Comconf) Uid() int {
	return int(c.dbconf.Id)
}