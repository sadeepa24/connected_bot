package builder

import (
	"encoding/base64"
	"errors"
	"reflect"
	"strings"

	//option "github.com/sadeepa24/connected_bot/builder/sbox_option/v2"
	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/walker"
	singC "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	json "github.com/sagernet/sing/common/json"
	"go.uber.org/zap"
)

type Connector interface {
	Select(options []string, msg any) (string, error)
	ReciveVal(msg string) (string, error)
	AlertSend(msg string) (error)
}


type SimpleConnector struct {
	common.Tgcalls
	btns *botapi.Buttons
}

func (sc *SimpleConnector) Select(options []string, msg any) (string, error) {
	sc.btns.Reset([]int16{2})
	for _, opo := range options {
		if opo == Pass {
			sc.btns.Passline()
			continue
		}
		sc.btns.AddBtcommon(opo)
	}
	clbk, err := sc.Callbackreciver(msg, sc.btns)
	if err != nil {
		return "", err
	}
	return clbk.Data, nil
}

func (sc *SimpleConnector) ReciveVal(msg string) (string, error) {
	mg, err := sc.Sendreciver(msg)
	if err != nil {
		return "", err
	}
	return mg.Text, err
}

func (sc *SimpleConnector) AlertSend(msg string) error { 
	sc.Alertsender(msg)
	return nil
}



func NewConnector(calls common.Tgcalls) Connector {
	return &SimpleConnector{
		btns: botapi.NewButtons([]int16{2}),
		Tgcalls: calls,
	}
}




//for vmess parsing
type VmessLink struct {
	Ps   string      `json:"ps"`
	Add  string      `json:"add"`
	Port string 	 `json:"port"`
	ID   string      `json:"id"`
	Aid  string 	 `json:"aid"`
	Net  string      `json:"net"`
	Type string      `json:"type"`
	Host string      `json:"host"`
	Path string      `json:"path"`
	TLS  string      `json:"tls"`
	SNI  string 	 `json:"sni"`
	SCY  string 	 `json:"scy"`
	Fp  string 		 `json:"fp"`
	Alpn string	`json:"alpn"`
	alpnSlice []string `json:"-"`
}

func ParseVmessUrl(url string) (*VmessLink, error) {
	jsonout := make([]byte, len(url))
	decodedInput := url[8:]
	n, err := base64.StdEncoding.Decode(jsonout, []byte(decodedInput))
	if err != nil {
		return nil, err
	}
	vmess :=  &VmessLink{}
	err = json.Unmarshal(jsonout[:n], &vmess)
	if err != nil {
		return vmess, err
	}
	vmess.alpnSlice = strings.Split(vmess.Alpn, ",")
	return vmess, nil
}

type Error struct {
	error
}







//builder
//T shouldn't be a pointer
func createopts[T any](b *Builder, name string, c *walker.CommonFuncs) (T, error) {
	
	var new T
	s, err := b.conec.Select([]string{"yes", "no"}, "do you want to configure "+ name + "(select no if you want just empty transport)")
	if err != nil {
		return new, err
	}
	if s == "no" {
		return new, nil
	}
	lswlkr, _ := walker.NewWalker(&new)
	b.conec.AlertSend("press done after configuring " + name)
	lswlkr.SetValue = b.commonSetValue
	lswlkr.CanSetCheck = b.commonCheck
	lswlkr.ApendCt = b.commonApnd

	if c != nil {
		if c.SetValue != nil {
			lswlkr.SetValue = c.SetValue
		}
		if c.CanSetCheck != nil {
			lswlkr.CanSetCheck = c.CanSetCheck
		}
		if c.SliceTagHook != nil {
			lswlkr.SliceTagHook = c.SliceTagHook
		}
		if c.ApendCt != nil {
			lswlkr.ApendCt = c.ApendCt
		}
	}

	err = b.anyFieldChange(lswlkr, &new)
	if err != nil {
		return new, err
	}
	return new, nil
}
//IO common
func (b *Builder) createV2rayTransport() (*option.V2RayTransportOptions, error){
	s, err := b.conec.Select([]string{"yes", "no"}, "do you want to configure inbound Transport(select no if you want just empty transport)") 
	if err != nil {
		return nil, err
	}
	if s == "no" {
		return nil, nil	
	}
	tra := &option.V2RayTransportOptions{}
	
	b.conec.AlertSend("press done after configuring transport options")
	tra.Type, err = b.conec.Select([]string{"ws", "http", "quic", "grpc", "httpupgrade"}, "select transport type")
	if err != nil {
		return nil, err
	}
	var wlkr *walker.Walker
	switch tra.Type {
	case "ws":
		tra.WebsocketOptions = option.V2RayWebsocketOptions{}
		wlkr, _ = walker.NewWalker(&tra.WebsocketOptions)
	case "http":
		tra.HTTPOptions = option.V2RayHTTPOptions{}
		wlkr, _ = walker.NewWalker(&tra.HTTPOptions)
	case "quic":
		tra.QUICOptions = option.V2RayQUICOptions{}
		wlkr, _ = walker.NewWalker(&tra.QUICOptions)
	case "grpc":
		tra.GRPCOptions = option.V2RayGRPCOptions{}
		wlkr, _ = walker.NewWalker(&tra.GRPCOptions)
	case "httpupgrade":
		tra.HTTPUpgradeOptions = option.V2RayHTTPUpgradeOptions{}
		wlkr, _ = walker.NewWalker(&tra.HTTPUpgradeOptions)
	default:
		return nil, errors.New("unknown transport type: " + tra.Type)
	}
	wlkr.SetValue = b.commonSetValue
	wlkr.CanSetCheck = b.commonCheck
	wlkr.ApendCt = b.commonApnd

	err = b.anyFieldChange(wlkr, tra)
	if err != nil {
		return nil, err
	}
	return tra, nil
}




func (b *Builder) createFullInbound() (option.Inbound, error) {
	in := option.Inbound{}
	typ, err := b.conec.Select(allintyp, "select inbound type")
	if err != nil {
		return in, err
	}
	in.Tag = b.reciveTag("in")
	if in.Tag == "" {
		return in, ErrTagRecive
	}
	in.Type = typ
	in.Options, err = b.createInopts(in.Type)
	if err != nil {
		return in, err
	}
	return in, nil
}
func (b *Builder) createInopts(typ string) (any, error) {

	var (
		lsopts         option.ListenOptions
		v2rayTransport *option.V2RayTransportOptions
		multiplex      *option.InboundMultiplexOptions
		tls            option.InboundTLSOptionsContainer
		err            error
	)

	switch typ {
	case singC.TypeTun:
		break
	case singC.TypeShadowTLS,singC.TypeSOCKS:
		lsopts, err = createopts[option.ListenOptions](b, "Listenoptions", nil)
		if err != nil {
			return nil, err
		}
	case singC.TypeVLESS, singC.TypeVMess, singC.TypeTrojan:
		v2rayTransport, err = b.createV2rayTransport()
		if err != nil {
			return nil, err
		}
		s, err := createopts[option.InboundMultiplexOptions](b, "InboundMultiplex", nil)
		if err != nil {
			return nil, err
		}
		multiplex = &s
		fallthrough
	default:
		lsopts, err = createopts[option.ListenOptions](b, "Listenoptions", nil)
		if err != nil {
			return nil, err
		}
		tls, err = createopts[option.InboundTLSOptionsContainer](b, "InboundTlsoptions", nil)
		if err != nil {
			return nil, err
		}
	}

	switch typ {
	case singC.TypeVLESS:
		return &option.VLESSInboundOptions{
			ListenOptions:              lsopts,
			Multiplex:                  multiplex,
			Transport:                  v2rayTransport,
			InboundTLSOptionsContainer: tls,
			Users:                      []option.VLESSUser{},
		}, nil
	case singC.TypeVMess:
		return &option.VMessInboundOptions{
			ListenOptions:              lsopts,
			Multiplex:                  multiplex,
			Transport:                  v2rayTransport,
			InboundTLSOptionsContainer: tls,
			Users:                      []option.VMessUser{},
		}, nil
	case singC.TypeTrojan:
		return &option.TrojanInboundOptions{
			ListenOptions:              lsopts,
			Multiplex:                  multiplex,
			Transport:                  v2rayTransport,
			InboundTLSOptionsContainer: tls,
			Users:                      []option.TrojanUser{},
		}, nil
	case singC.TypeTUIC:
		return &option.TUICInboundOptions{
			ListenOptions: lsopts,
			InboundTLSOptionsContainer: tls,
		}, nil
	case singC.TypeShadowTLS:
		return &option.ShadowTLSInboundOptions{
			ListenOptions: lsopts,
		}, nil
	case singC.TypeSOCKS:
		return &option.SocksInboundOptions{
			ListenOptions: lsopts,
		}, nil
	case singC.TypeMixed:
		return &option.HTTPMixedInboundOptions{
			ListenOptions: lsopts,
			InboundTLSOptionsContainer: tls,
		}, nil
	case singC.TypeHysteria2:
		return &option.Hysteria2InboundOptions{
			ListenOptions: lsopts,
			InboundTLSOptionsContainer: tls,
		}, nil
	case singC.TypeHysteria:
		return &option.HysteriaInboundOptions{
			ListenOptions: lsopts,
			InboundTLSOptionsContainer: tls,
		}, nil
	case singC.TypeTun:
		return &option.TunInboundOptions{}, nil
	}
	return nil, ErrInboundUnknown
}


func (b *Builder) createFullOutbound() (option.Outbound, error) {
	out := option.Outbound{}
	typ, err := b.conec.Select(allouttyp, "select outbound type")
	if err != nil {
		return out, err
	}
	out.Tag = b.reciveTag("out")
	if out.Tag == "" {
		return out, ErrTagRecive
	}
	out.Type = typ
	out.Options, err = b.createOutopts(out.Type)
	if err != nil {
		return out, err
	}
	return out, nil
}
func (b *Builder) createOutopts(typ string) (any, error) {
	var (
		dialer 			option.DialerOptions
		v2rayTransport *option.V2RayTransportOptions
		multiplex      *option.OutboundMultiplexOptions
		tls            	option.OutboundTLSOptionsContainer
		srvopts 	 	option.ServerOptions
		err            	error
		allcurrentouts map[string]bool
	)

	switch typ {
	case singC.TypeSelector, singC.TypeURLTest:
		btns := []string{"done"}
		allcurrentouts = map[string]bool{}
		for {
			for key := range b.outbounds {
				if allcurrentouts[key] {
					btns = append(btns, C.GetMsg(C.ButtonSelectEmjoi) + key)
					continue
				}
				btns = append(btns, key)
			}
			defout, err := b.conec.Select(btns, "select outbounds")
			if err != nil {
				return nil, err
			}
			btns = []string{"done"}
			if defout == "done" {
				break
			}
			if strings.HasPrefix(defout, C.GetMsg(C.ButtonSelectEmjoi)) {
				delete(allcurrentouts, defout[len(C.GetMsg(C.ButtonSelectEmjoi)):])
			} else {
				allcurrentouts[defout] = true
			}
		}
	case singC.TypeDirect:
		break
	case singC.TypeSOCKS, singC.TypeSSH:
		srvopts, err = createopts[option.ServerOptions](b, "Server options", nil)
		if err != nil {
			return nil, err
		}
		fallthrough
	case singC.TypeTor:
		dialer, err = createopts[option.DialerOptions](b, "Dialer Options", nil)
		if err != nil {
			return nil, err
		}
	case singC.TypeVLESS, singC.TypeVMess, singC.TypeTrojan:
		v2rayTransport, err = b.createV2rayTransport()
		if err != nil {
			return nil, err
		}
		s, err := createopts[option.OutboundMultiplexOptions](b, "Outbound Multiplex", nil)
		if err != nil {
			return nil, err
		}
		multiplex = &s
		fallthrough
	default:
		tls, err = createopts[option.OutboundTLSOptionsContainer](b, "Outbound Tls Options", nil)
		if err != nil {
			return nil, err
		}
		dialer, err = createopts[option.DialerOptions](b, "Dialer Options", nil)
		if err != nil {
			return nil, err
		}
		srvopts, err = createopts[option.ServerOptions](b, "Server options", nil)
		if err != nil {
			return nil, err
		}
	}

	switch typ {
	case singC.TypeVLESS:
		return &option.VLESSOutboundOptions{
			Multiplex:                  multiplex,
			Transport:                  v2rayTransport,
			OutboundTLSOptionsContainer: tls,
			ServerOptions: srvopts,
			DialerOptions: 		   dialer,
		}, nil
	case singC.TypeVMess:
		return &option.VMessOutboundOptions{
			ServerOptions: srvopts,
			DialerOptions: 		   dialer,
			OutboundTLSOptionsContainer: tls,
			Multiplex:                  multiplex,
			Transport:                  v2rayTransport,
		}, nil
	case singC.TypeTrojan:
		return &option.TrojanOutboundOptions{
			ServerOptions: srvopts,
			DialerOptions: 		   dialer,
			OutboundTLSOptionsContainer: tls,
			Multiplex:                  multiplex,
			Transport:                  v2rayTransport,

		}, nil
	case singC.TypeTUIC:
		return &option.TUICOutboundOptions{
			ServerOptions: srvopts,
			DialerOptions: dialer,
			OutboundTLSOptionsContainer: tls,
		}, nil
	case singC.TypeShadowTLS:
		return &option.ShadowTLSOutboundOptions{
			DialerOptions: dialer,
			ServerOptions: srvopts,
			OutboundTLSOptionsContainer: tls,
		}, nil
	case singC.TypeShadowsocks:
		return &option.ShadowsocksOutboundOptions{
			DialerOptions: dialer,
			ServerOptions: srvopts,
			Multiplex: multiplex,
		}, nil
	case singC.TypeSOCKS:
		return &option.SOCKSOutboundOptions{
			DialerOptions: dialer,
			ServerOptions: srvopts,
		}, nil
	case singC.TypeHTTP:
		return &option.HTTPOutboundOptions{
			ServerOptions: srvopts,
			DialerOptions: dialer,
		}, nil
	case singC.TypeHysteria2:
		return &option.Hysteria2OutboundOptions{
			DialerOptions: dialer,
			ServerOptions: srvopts,
		}, nil
	case singC.TypeHysteria:
		return &option.HysteriaOutboundOptions{
			DialerOptions: dialer,
			ServerOptions: srvopts,
		}, nil
	case singC.TypeSelector:
		defaultout, err := b.conec.Select(C.MapToSliceKey(b.outbounds), "select default outbound")
		if err != nil {
			return nil, err
		}
		return &option.SelectorOutboundOptions{
			Default: defaultout,
			Outbounds: C.MapToSliceKey(allcurrentouts),
		}, nil
	case singC.TypeTor:
		return &option.TorOutboundOptions{
			DialerOptions: dialer,
		}, nil
	case singC.TypeSSH:
		return &option.SSHOutboundOptions{
			DialerOptions: dialer,
			ServerOptions: srvopts,
		}, nil
	case singC.TypeDirect:
		return &option.DirectOutboundOptions{

		}, nil
	case singC.TypeURLTest:
		testurl, err := b.conec.Select(C.MapToSliceKey(b.outbounds), "send test url")
		if err != nil {
			return nil, err
		}
		return &option.URLTestOutboundOptions{
			Outbounds: C.MapToSliceKey(allcurrentouts),
			URL: testurl,

		}, nil
	}

	return nil, ErrOutboundTyUnknown
}


func (b *Builder) createFullEndPoint() (option.Endpoint, error) {
	out := option.Endpoint{}
	typ, err := b.conec.Select([]string{"wireguard"}, "select endpoint type")
	if err != nil {
		return out, err
	}
	tag, err := b.conec.ReciveVal("send endpoint tag")
	if err != nil {
		return out, err
	}
	out.Tag = tag
	out.Type = typ
	out.Options = &option.WireGuardEndpointOptions{}
	return out, nil
}



func AnyFieldChange(wlkr *walker.Walker, item any, conec Connector, itemexport func(item any) any, lgr *zap.Logger) error {
	
	//TODO: remove this after stablizing the walker (almost)
	defer func() {
		if r := recover(); r != nil {
			lgr.Error("panic recovered in botconfig", zap.Any("recover", r))
			conec.AlertSend("An unexpected error occurred. Please try again later. Or send the error to dev")
		}
	}()
	
	var items []string
	inedit:
	for {
		if wlkr.EndValue  {
			wlkr.Change("")
			wlkr.WalkBack()
		}
		items = wlkr.Items()
		items = append(items, Pass)
		if wlkr.CanSet() {
			items = append(items, SET)
		}
		if wlkr.Current().Kind() == reflect.Slice {
			items = append(items, APPEND)
			if len(items) > 1 {
				items = append(items, REMOVE)
			}
		}
		if wlkr.Current().CanAddr() {
			items = append(items, PARSE)
		}
		items = append(items, BACK, DONE, RESET)
		
		se, err := conec.Select(items, itemexport(item))
		
		if err != nil {
			return err
		}
		switch se {
		case DONE:
			break inedit
		case BACK:
			if wlkr.Path == "" {
				break inedit
			}
			wlkr.WalkBack()
			continue
		case SET:
			wlkr.Change("")
		case APPEND:
			wlkr.AppendCustom()
		case REMOVE:
			sr, err := conec.Select(wlkr.Items(), "select the item to remove")
			if err != nil {
				return err
			}
			wlkr.RemoveFromSLice(sr)
			continue
		case RESET:
			err = wlkr.Construct()
			if err != nil {
				conec.AlertSend("reset failed "+ err.Error())
			}
		case PARSE:
			ptr, ok := wlkr.CurrentPtr()
			if !ok {
				conec.AlertSend("cannot parse to current object")
				continue
			}
			ob, err := conec.ReciveVal("send you'r object")
			if err != nil {
				return err
			}
			if ob == "" {
				conec.AlertSend("cannot parse empty value object")
				continue
			}
			err = json.Unmarshal([]byte(ob), ptr.Interface())
			if err != nil {
				conec.AlertSend("current object parse failed " + err.Error())
			}
		default:
			wlkr.WalkInto(se)
		}
		if err != nil {
			log.Error(err)	
		}
	}
	return nil
}