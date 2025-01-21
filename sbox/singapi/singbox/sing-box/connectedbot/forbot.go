package connectedbot

import (
	"net/netip"

	"github.com/gofrs/uuid/v5"
	"github.com/sagernet/sing-box/option"
)

type VlessStatus struct {
	Download int64
	Upload int64
	Disabled bool
	Online_ip map[netip.Addr]int64
}

type Vlessuser struct {
	VlessUser option.VLESSUser
	User int
	Bandwidth int
}

type StatusOutput struct {
	Type string
	Status any
}

type BotUser struct {
	Intype string
	User   any
}


func (v *Vlessuser) Getuid() (uuid.UUID, error) {
	return uuid.FromString(v.VlessUser.UUID)
}