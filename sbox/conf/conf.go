package conf

import (
	"errors"
	"net/netip"
	"sync/atomic"

	"github.com/sagernet/sing-box/connectedbot/opts"
	"github.com/sagernet/sing-box/option"

	C "github.com/sadeepa24/connected_bot/constbot"
)

type Inboud struct {
	Id              int16 //ID from json file
	Name            string
	Tag             string
	Type            string
	Support         []string
	ListenAddres    string
	Listenport      int
	Tlsenabled      bool
	TransPortType   string
	TransPortPath string
	Custom_info     string
	Domain          string
	PublicIp        string
}

type ExportInfo struct {
	Sni string
	Host string
	Server string
	
}

func (in *Inboud) setTransportinfo(tra *option.V2RayTransportOptions) {
	if tra == nil {
		return
	}
	in.TransPortType = tra.Type
	switch tra.Type {
	case "ws":
		in.TransPortPath = tra.WebsocketOptions.Path
	case "http":
		in.TransPortPath = tra.WebsocketOptions.Path
	case "grpc":
		in.TransPortPath = tra.GRPCOptions.ServiceName
	case "quic":
		in.TransPortPath = "transport quic does not have path"
	case "httpupgrade":
		in.TransPortPath = tra.HTTPUpgradeOptions.Path
	}

} 

func (in *Inboud) setListenInfo(ls option.ListenOptions) {
	if ls.Listen != nil {
		in.ListenAddres = ls.Listen.Build(netip.Addr{}).String()
	}
	in.Listenport= int(ls.ListenPort)
}

func (in *Inboud) setTLSInfo(tls *option.InboundTLSOptions) {
	if tls == nil {
		return
	}
	in.Tlsenabled = tls.Enabled
	//TODO: add more later
}


func (in *Inboud) AddOption(opts option.Inbound) error {
	in.Name = opts.Tag
	in.Type = opts.Type
	in.Id = *opts.Id
	in.Tag = opts.Tag
	in.Domain = opts.Domain
	in.PublicIp = opts.Public_Ip
	in.Custom_info = opts.Custom_info
	in.Support = opts.SupportInfo

	switch opts.Type {
	case C.Vless:
		var (
			ok      bool
			vlessin *option.VLESSInboundOptions
		)
		if vlessin, ok = opts.Options.(*option.VLESSInboundOptions); !ok {
			return errors.New("something went wrong when proc inbound " + opts.Tag)
		}
		in.setListenInfo(vlessin.ListenOptions)
		in.setTLSInfo(vlessin.TLS)
		in.setTransportinfo(vlessin.Transport)
	case "trojan":
		var (
			ok      bool
			trojanin *option.TrojanInboundOptions
		)
		if trojanin, ok = opts.Options.(*option.TrojanInboundOptions); !ok {
			return errors.New("something went wrong when proc inbound " + opts.Tag)
		}
		in.setListenInfo(trojanin.ListenOptions)
		in.setTLSInfo(trojanin.TLS)
		in.setTransportinfo(trojanin.Transport)
	case "vmess":
		var (
			ok      bool
			vmessin *option.VMessInboundOptions
		)
		if vmessin, ok = opts.Options.(*option.VMessInboundOptions); !ok {
			return errors.New("something went wrong when proc inbound " + opts.Tag)
		}
		in.setListenInfo(vmessin.ListenOptions)
		in.setTLSInfo(vmessin.TLS)
		in.setTransportinfo(vmessin.Transport)
	case "tuic":
		var (
			ok      bool
			tuic *option.TUICInboundOptions
		)
		if tuic, ok = opts.Options.(*option.TUICInboundOptions); !ok {
			return errors.New("something went wrong when proc inbound " + opts.Tag)
		}
		in.setListenInfo(tuic.ListenOptions)
		in.setTLSInfo(tuic.TLS)
	//TODO: add proto

	default:
		return C.ErrNotsupported

	}
	return nil
}

func (in *Inboud) Port() int {
	return in.Listenport
}

func (in *Inboud) Laddr() string {
	return in.ListenAddres
}

func (in *Inboud) TransortType() string {
	return in.TransPortType
}
func (in *Inboud) TransportPath() string {
	return in.TransPortPath
}

func (in *Inboud) TlsIsEnabled() bool {
	return in.Tlsenabled
}


type Outbound struct {
	Id   int16
	Name string
	Tag  string
	Type string
	//Option      *option.Outbound
	Custom_info string
	Latency     *atomic.Int32
}

func (out *Outbound) AddOption(opts option.Outbound) error {
	out.Custom_info = opts.Custom_info
	out.Id = *opts.Id
	out.Name = opts.Tag
	out.Tag = opts.Tag
	out.Type = opts.Type
	if out.Latency == nil {
		out.Latency = new(atomic.Int32)
	}
	return nil //may need in future
}

func (out *Outbound) AddOptionEndpoint(opts option.Endpoint) error {
	out.Custom_info = opts.Custom_info
	out.Id = *opts.Id
	out.Name = opts.Tag
	out.Tag = opts.Tag
	out.Type = opts.Type
	if out.Latency == nil {
		out.Latency = new(atomic.Int32)
	}
	return nil //may need in future
}



type Sboxstatus struct {
	Download  C.Bwidth
	Upload    C.Bwidth
	Online_ip map[string]int16
	Disabled  bool
}

func (s *Sboxstatus) Fill(st opts.UserStatus) {
	s.Download = C.Bwidth(st.Download)
	s.Upload = C.Bwidth(st.Upload)
	s.Online_ip = st.Ip
	s.Disabled = st.Disabled
}
func (s Sboxstatus) FullUsage() C.Bwidth {
	return s.Download + s.Upload
}