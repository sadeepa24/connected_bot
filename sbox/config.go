package sbox

import (
	"net/netip"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gofrs/uuid"
	C "github.com/sadeepa24/connected_bot/constbot"
	option "github.com/sagernet/sing-box/option"
)

// Standar config For Whole Aplication
type Userconfig struct {
	*Vlessgroup
	*Trojangroup
	*Commongroup
	DbID        int64
	UsercheckId int
	Name        string
	Usage       C.Bwidth
	Quota       C.Bwidth
	Inboundtag  string
	Outboundtag string
	LoginLimit  int32
	InboundId   int16
	OutboundID  int16
	TgId int64 //user telegram id

	Type string

	Password string //optional other protocole like trojan

}

func (u *Userconfig) GetuniqName() string {
	return strconv.Itoa(int(u.DbID)) + strings.TrimSpace(u.Name) + strconv.Itoa(int(u.TgId))
}

type Vlessgroup struct {
	UUID uuid.UUID
}

type Trojangroup struct {
	Password string
}

type Commongroup struct {
	User string
	Pass string
}

// Standar Inbound For Whole Aplication
type Inboud struct {
	Id              int64 //ID from json file
	Name            string
	Tag             string
	Type            string
	Option          *option.Inbound
	Support         []string
	ListenAddres    string
	Listenport      int
	Tlsenabled      bool
	Transporttype   string
	Transportoption option.V2RayTransportOptions
	Custom_info     string
	Domain          string
	PublicIp        string
}

func (in *Inboud) Port() int {
	if in.Option == nil {
		return 0
	}
	switch in.Type {
	case C.Vless:
		return in.Listenport
	}
	return 0
}

func (in *Inboud) Laddr() string {
	if in.Option == nil {
		return "noaddr"
	}
	switch in.Type {
	case C.Vless:
		return in.Vlessopts().Listen.Build(netip.IPv4Unspecified()).String()
	}
	return "noaddr"
}

func (in *Inboud) TransortType() string {
	if in.Option == nil {
		return "notype"
	}
	switch in.Type {
	case C.Vless:
		if in.Vlessopts().Transport == nil {
			return "notype"
		}
		return in.Vlessopts().Transport.Type
	}
	return "notype"
}
func (in *Inboud) TransportPath() string {
	if in.Option == nil {
		return "/"
	}
	switch in.Type {
	case C.Vless:
		if in.Vlessopts().Transport == nil {
			return "notype"
		}
		switch in.Vlessopts().Transport.Type {
		case "ws":
			return in.Vlessopts().Transport.WebsocketOptions.Path
		case "http":
			return in.Vlessopts().Transport.HTTPOptions.Path
		default:
			return ""
		}

	}
	return "nopath"
}

func (in *Inboud) TlsIsEnabled() bool {
	if in.Option == nil {
		return false
	}
	switch in.Type {
	case C.Vless:
		if in.Vlessopts().TLS == nil {
			return false
		}
		return in.Vlessopts().TLS.Enabled
	}
	return false
}

func (in *Inboud) Vlessopts() *option.VLESSInboundOptions {
	if in.Option == nil {
		return &option.VLESSInboundOptions{}
	}
	switch in.Type {
	case C.Vless:
		return in.Option.Options.(*option.VLESSInboundOptions)
	}
	return &option.VLESSInboundOptions{}
}

type Outbound struct {
	Id          int64
	Name        string
	Tag         string
	Type        string
	//Option      *option.Outbound
	Custom_info string
	Latency     *atomic.Int32
}

type Sboxstatus struct {
	Download  C.Bwidth
	Upload    C.Bwidth
	Online_ip map[netip.Addr]int64
	Disabled  bool
}

func (s Sboxstatus) FullUsage() C.Bwidth {
	return s.Download + s.Upload
}
