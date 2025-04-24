package builder

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/netip"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	//option "github.com/sadeepa24/connected_bot/builder/sbox_option/v2"
	//"github.com/sagernet/sing-box/option"

	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox/conf"
	"github.com/sadeepa24/walker"
	singC "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	json "github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badoption"
)


type ErrorExit struct {
	Err error
}

var (
	ErrInboundNotFound = errors.New("inbound not found")
	ErrInboundUnknown = errors.New("inbound unknown")
	ErrOutboundNotFound = errors.New("outbound not found")
	ErrEndpointNotFound = errors.New("endpoint not found")
	ErrTagRecive = errors.New("something went wrong while reciving tag")
	ErrTransportTyUnknown = errors.New("transport type unknown")
	ErrOutboundTyUnknown = errors.New("outbound unknown")
	ErrNilInput = errors.New("nil object as input")

)


var allintyp = []string{
	"vless",
	"vmess",
	"trojan",
	"tuic",
	"tun",
	singC.TypeShadowTLS,
	singC.TypeHysteria2,
	singC.TypeHysteria,
	singC.TypeMixed,
	singC.TypeSOCKS,

}
var allouttyp = []string{
	"vless",
	"vmess",
	"trojan",
	singC.TypeHysteria,
	singC.TypeHysteria2,
	singC.TypeShadowTLS,
	singC.TypeSelector,
	singC.TypeShadowsocks,
	singC.TypeDirect,
	singC.TypeTor,
	singC.TypeSSH,
	singC.TypeHTTP,
	singC.TypeURLTest,
}

type Create uint8

const (
	DONE  = "⚙ done ⚙"
	BACK = "⚙ back ⚙"
	SET  = "⚙ set ⚙"
	APPEND = "⚙ append ⚙"
	REMOVE = "⚙ remove ⚙"
	RESET = "⚙ reset ⚙"
	PARSE = "⚙ parse ⚙"
	Pass = "passsssssss"

)

type Buf interface {
	io.Reader
	io.Writer
	Reset()
	Len() int
}


type Builder struct {
	
	outselector *option.Outbound //special outbound
	ctx context.Context
	inbounds map[string]*option.Inbound
	outbounds map[string]*option.Outbound
	endpoints map[string]*option.Endpoint
	ruleset map[string]*option.RuleSet
	// Route *option.RouteOptions
	// Dns *option.DNSOptions
	
	conec Connector
	opts *option.Options

	// wlkr *walker.Walker

	CallBack func() (Buf)

	fromfile bool
	filepath string
	closed bool
	
	Disableautosave bool //if true must call Save before close

	buf Buf
	// io.ReadWriter

}


func NewBuilder(conec Connector) (*Builder, error) {
	if conec == nil {
		return nil, errors.New("nil connector")
	}
	return &Builder{
		conec: conec,
		ctx: globalCtx,
		outbounds: map[string]*option.Outbound{},
		endpoints: map[string]*option.Endpoint{},
		opts: &option.Options{
			Route: &option.RouteOptions{},
			DNS: &option.DNSOptions{},
			
		},
	}, nil
}
func NewBuilderFromFile(conec Connector, path string) (*Builder, error) {
	builder := &Builder{
		conec: conec,
		ctx: globalCtx,
		fromfile: true,
		filepath: path,
	}
	var err error
	builder.opts, err  = optsfromfile(path)
	if err != nil {
		return nil, err
	}
	return builder, builder.Reset(builder.opts)
}
func optsfromfile(path string) (*option.Options, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {return nil, err}
	if stat.Size() == 0 {
		file.WriteString("{}")
		return &option.Options{}, nil
	}
	filect, _ := io.ReadAll(file)
	opts, err := json.UnmarshalExtendedContext[option.Options](globalCtx, filect)
	if err != nil {
		return nil, err
	}
	return &opts, nil
}




func (b *Builder) ResetAny(opts any) error {
	switch ot := opts.(type) {
	case *option.Options:
		return b.Reset(ot)
	case option.Options:
		return b.Reset(&ot)
	case string:
		var opt option.Options
		err := opt.UnmarshalJSONContext(globalCtx, []byte(ot))
		if err != nil {
			return err
		}
		return b.Reset(&opt)
	case []byte:
		var opt option.Options
		err := opt.UnmarshalJSONContext(globalCtx, []byte(ot))
		if err != nil {
			return err
		}
		return b.Reset(&opt)
	default:
		return errors.New("unsupported type")
	}
}
func (b *Builder) Reset(opts *option.Options) error {
	b.opts = opts
	b.inbounds = C.SliceToMapPtr(opts.Inbounds, func(i option.Inbound) string {
		return i.Tag
	})
	b.endpoints = C.SliceToMapPtr(opts.Endpoints, func(i option.Endpoint) string {
		return i.Tag
	})
	b.outbounds = C.SliceToMapPtr(opts.Outbounds, func(i option.Outbound) string {
		return i.Tag
	})
	C.ExcuteSliceTill(opts.Outbounds, func(t *option.Outbound) bool {
		if t != nil && t.Type == singC.TypeSelector {
			b.outselector = t
			if t.Options == nil {
				t.Options = &option.SelectorOutboundOptions{}
			}
			return false
		}
		return true
	})
	b.ruleset = map[string]*option.RuleSet{}
	if opts.Route != nil {
		b.ruleset = C.SliceToMapPtr(opts.Route.RuleSet, func(i option.RuleSet) string { 
			return i.Tag
		})
	}
	return nil
}
func (b *Builder) ResetConf(path string) error {
	var err error
	b.opts, err  = optsfromfile(path)
	if err != nil {
		return err
	}
	b.filepath = path
	b.fromfile = true
	return b.Reset(b.opts)
}


func (b *Builder) Save() error {
	if  b.fromfile {
		file, err := os.OpenFile(b.filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		jd := json.NewEncoder(file)
		return jd.Encode(b.opts)
	}
	return nil
}
func (b *Builder) Close() error {
	if b.closed {
		return nil
	}
	b.closed = true
	b.buf.Reset()
	if b.fromfile  && !b.Disableautosave {
		return b.Save()
	}
	return nil
}



func (b *Builder) AllIn() map[string]*option.Inbound {
	return b.inbounds
}
func (b *Builder) AllOut() map[string]*option.Outbound {
	return b.outbounds
}
func (b *Builder) AllEnd() map[string]*option.Endpoint {
	return b.endpoints
}
func (b *Builder) reciveTag(io string) string {
	for {
		tag, err := b.conec.ReciveVal("send tag name")	
		if err != nil {
			return ""
		}
		switch io {
		case "in":
			_, alread := b.outbounds[tag]
			if !alread {
				return tag
			}
		case "end":
			_, alread := b.endpoints[tag]
			if !alread {
				return tag
			}
		case "out":
			_, alread := b.endpoints[tag]
			if !alread {
				return tag
			}
		default:
			return ""
		}
		b.conec.AlertSend("send another valid tag")
	}
}



//TODO: remove thease funcs after creating the builder
func (b *Builder) SetInbound(ins map[string]*option.Inbound) {
	b.inbounds = ins
}


//common for everything
func (b *Builder) anyFieldChange(wlkr *walker.Walker, item any) error {
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
		se, err := b.conec.Select(items, b.exportany(item))
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
			sr, err := b.conec.Select(wlkr.Items(), "select the item to remove")
			if err != nil {
				return err
			}
			wlkr.RemoveFromSLice(sr)
			continue
		case RESET:
			err = wlkr.Construct()
			if err != nil {
				b.conec.AlertSend("reset failed "+ err.Error())
			}
		case PARSE:
			ptr, ok := wlkr.CurrentPtr()
			if !ok {
				b.conec.AlertSend("cannot parse to current object")
				continue
			}
			ob, err := b.conec.ReciveVal("send you'r object")
			if err != nil {
				return err
			}
			if ob == ""{
				b.conec.AlertSend("cannot parse empty value object")
				continue
			}
			err = json.UnmarshalContext(b.ctx, []byte(ob), ptr.Interface())
			if err != nil {
				b.conec.AlertSend("current object parse failed " + err.Error())
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


func(b *Builder) commonSetValue(curval reflect.Value, path string, wlkr *walker.Walker, item any)  (reflect.Value, bool)   {
	switch curval.Type().String() {
	case "string":
		str, err := b.conec.ReciveVal("send string value")
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(str), true
	case "bool":
		str, err := b.conec.Select([]string{"true", "false"}, "select bool value")
		if err != nil {
			return reflect.Value{}, false
		}
		if str == "true" {
			return reflect.ValueOf(true), true
		}
		return reflect.ValueOf(false), true
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		str, err := b.conec.ReciveVal("send int value")
		if err != nil {
			return reflect.Value{}, false
		}
		val, err := strconv.Atoi(str)
		if err != nil {
			return reflect.Value{}, false
		}
		switch  curval.Type().String() {
		case "int8":
			return reflect.ValueOf(int8(val)), true
		case "int16":
			return reflect.ValueOf(int16(val)), true
		case "int32":
			return reflect.ValueOf(int32(val)), true
		case "int64":
			return reflect.ValueOf(int64(val)), true
		case "uint8":
			return reflect.ValueOf(uint8(val)), true
		case "uint16":
			return reflect.ValueOf(uint16(val)), true
		case "uint32":
			return reflect.ValueOf(uint32(val)), true
		case "uint64":
			return reflect.ValueOf(uint64(val)), true
		}
		return reflect.ValueOf(val), true
	case "badoption.Duration":
		str, err := b.conec.ReciveVal("send duration value ex - 100ms")
		if err != nil {return reflect.Value{}, false}
		tdr, err := time.ParseDuration(str)
		if err != nil {
			b.conec.AlertSend("duration parse error " + err.Error())
			return reflect.Value{}, false
		}
		return reflect.ValueOf(badoption.Duration(tdr)),true
	case "badoption.Addr", "*badoption.Addr":
		str, err := b.conec.ReciveVal("send address")
		if err != nil {
			return reflect.Value{}, false
		}
		addr, err := netip.ParseAddr(str)
		if err != nil {
			b.conec.AlertSend("duration parse error " + err.Error())
			return reflect.Value{}, false
		}
		adr := badoption.Addr(addr)
		switch curval.Type().String() {
		case "badoption.Addr":
			return reflect.ValueOf(adr), true
		case "*badoption.Addr":
			return reflect.ValueOf(&adr), true
		}
	case "*badoption.Prefixable", "badoption.Prefixable", "*netip.Prefix":
		str, err := b.conec.ReciveVal("send prefix")
		if err != nil {
			return reflect.Value{}, false
		}
		prefix, err := netip.ParsePrefix(str)
		if err != nil {
			return reflect.Value{}, false
		}
		if curval.Type().String()  == "*netip.Prefix" {
			return reflect.ValueOf(&prefix), true
		}
		bdr := badoption.Prefixable(prefix)
		if curval.Type().String() ==  "*badoption.Prefixable" {
			return reflect.ValueOf(&bdr), true
		}
		return reflect.ValueOf(bdr), true

	case "badoption.Listable[string]":
		str, err := b.conec.ReciveVal("send comma seprated list")
		if err != nil {
			return reflect.Value{}, false
		}
		all := strings.Split(str, ",")
		return reflect.ValueOf(badoption.Listable[string](all)), true
	case "badoption.Listable[uint16]":
		str, err := b.conec.ReciveVal("send comma seprated list")
		if err != nil {
			return reflect.Value{}, false
		}
		all := strings.Split(str, ",")
		uints := []uint16{}
		for _, s := range all {
			i, err := strconv.Atoi(s)
			if err == nil {
				uints = append(uints, uint16(i))
			}
		}
		return reflect.ValueOf(badoption.Listable[uint16](uints)), true
	case "badoption.HTTPHeader":
		hdr, err := b.conec.ReciveVal("send comma seprated header list ex : key1:value1,key2:value2")
		if err != nil {
			return reflect.Value{}, false
		}
		all := strings.Split(hdr, ",")
		hdrs := badoption.HTTPHeader{}
		for _, h := range all {
			kv := strings.Split(h, ":")
			if len(kv) != 2 {
				return reflect.Value{}, false
			}
			hdrs[kv[0]] = badoption.Listable[string]{kv[1]}
		}
		return reflect.ValueOf(hdrs), true
	case "option.DomainStrategy":
		si, err := b.conec.Select([]string{"as_is", "prefer_ipv4", "prefer_ipv6", "ipv4_only", "ipv6_only"}, "select domain strategy")
		if err != nil {
			return reflect.Value{}, false
		}
		var gy option.DomainStrategy
		switch si {
			case "as_is":
				gy = option.DomainStrategy(0)
			case "prefer_ipv4":
				gy = option.DomainStrategy(1)
			case "prefer_ipv6":
				gy = option.DomainStrategy(2)
			case "ipv4_only":
				gy = option.DomainStrategy(2)
			case "ipv6_only":
				gy = option.DomainStrategy(4)
			default:
				return reflect.Value{}, false
		}
		return reflect.ValueOf(gy), true
	case "option.InterfaceType":
		typ, err := b.conec.Select(C.MapToSliceKey(singC.StringToInterfaceType), "select string to interface type")
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(singC.StringToInterfaceType[typ]), true
	case "badoption.Listable[option.InterfaceType]":
		typ, err := b.conec.Select(C.MapToSliceKey(singC.StringToInterfaceType), "select string to interface type")
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(singC.StringToInterfaceType[typ]), true
	}

	return reflect.Value{}, false
}
func (b *Builder) commonCheck(curval reflect.Value,  nextItemPath string,  wlkr *walker.Walker) bool {
	//FIXME: change to a map
	switch curval.Type().String() {
	case "badoption.Addr", "*badoption.Prefixable",  "*netip.Prefix", "netip.Prefix", "badoption.Prefixable", "*badoption.Addr", "badoption.Listable[uint16]", "badoption.Listable[string]", "option.DomainStrategy", "badoption.HTTPHeader", "badoption.Duration":
		return true
	}
	return false
}
func (b *Builder) commonApnd(slice reflect.Value, path string) bool {
	switch slice.Type().String() {

	case "badoption.Listable[string]":
		switch {
			case strings.HasSuffix(path, "Inbound"):
				if len(b.inbounds) == 0 {
					b.conec.AlertSend("no inbounds found")
					return false
				}
				in := []string{}
				
				for _, inn := range b.inbounds {
					in = append(in, inn.Tag)
				}
				sin, err := b.conec.Select(in, "select inbound tag")
				if err != nil {
					return false
				}
				slice.Set(reflect.Append(slice, reflect.ValueOf(sin)))
			case strings.HasSuffix(path, "Ruleset"):
				if len(b.ruleset) == 0 {
					b.conec.AlertSend("no ruleset found")
					return false
				}
				in := []string{}
				
				for _, inn := range b.ruleset {
					in = append(in, inn.Tag)
				}
				sin, err := b.conec.Select(in, "select ruleset tag")
				if err != nil {
					return false
				}
				slice.Set(reflect.Append(slice, reflect.ValueOf(sin)))
			default:
				s, err := b.conec.ReciveVal("send value to add")
				if err != nil {
					return false
				}
				slice.Set(reflect.Append(slice, reflect.ValueOf(s)))
				return true
		}
	case "badoption.Listable[github.com/sagernet/sing-box/option.InterfaceType]":
		typ, err := b.conec.Select(C.MapToSliceKey(singC.StringToInterfaceType), "select interface type to append")
		if err != nil {
			return false
		}
		slice.Set(reflect.Append(slice, reflect.ValueOf(option.InterfaceType(singC.StringToInterfaceType[typ]))))
		return true
	default:

	}
	return false
}





func (b *Builder) inbounsetval(curval reflect.Value, path string,  wlkr *walker.Walker, item any,)  (reflect.Value, bool)  {
	switch {
		case strings.HasSuffix(path, ".Options"):
			wlkr.WalkBack()
			v, _ := wlkr.Child("Type")
			opts, err := b.createInopts(v.String())
			if err != nil {
				return reflect.Value{}, false
			}
			return reflect.ValueOf(opts), true
		case strings.HasSuffix(path, ".Options.Transport"):
			if wlkr.Current().IsNil() {
				tratype, err := b.conec.Select([]string{"ws", "http", "quic", "grpc", "httpupgrade"}, "select transport type")
				if err != nil {
					return reflect.Value{}, false
				}
				return reflect.ValueOf(&option.V2RayTransportOptions{
					Type: tratype,
				}), true
			}
	}
	return b.commonSetValue(curval, path, wlkr, item)
}
func (b *Builder) inbounsetcheck(curval reflect.Value, nextItemPath string, wlkr *walker.Walker) bool {
	switch {
	case strings.HasSuffix(nextItemPath, ".Options"):
		return true
	}
	return b.commonCheck(curval, nextItemPath, wlkr)
}
func (b *Builder) inbounapnd(slice reflect.Value, path string) bool {

	endtyp := slice.Type().Elem().String()
	switch endtyp {
	case "option.VLESSUser",  "option.VMessUser", "option.TUICUser":
		uname, err := b.conec.ReciveVal("send user name")
		if err != nil {
			return false
		}
		uuid, err := b.conec.ReciveVal("send user uuid")
		if err != nil {
			return false
		}
		switch endtyp {
		case  "option.VMessUser":
			aleter, err := b.conec.ReciveVal("send alter id to skip send .")
			if err != nil {
				return false
			}
			altid := 0
			if aleter != "" {
				altid, _ = strconv.Atoi(aleter)
			}
			slice.Set(reflect.Append(slice, reflect.ValueOf(option.VMessUser{
				Name: uname,
				UUID: uuid,
				AlterId: altid,
			})))
			return true
		case  "option.VLESSUser":
			flow, err := b.conec.ReciveVal("send user flow")
			if err != nil {
				return false
			}
			slice.Set(reflect.Append(slice, reflect.ValueOf(option.VLESSUser{
				Name: uname,
				UUID: uuid,
				Flow:  flow,
			})))
			return true
		case  "option.TUICUser":
			flow, err := b.conec.ReciveVal("send user password")
			if err != nil {
				return false
			}
			slice.Set(reflect.Append(slice, reflect.ValueOf(option.TUICUser{
				Name: uname,
				UUID: uuid,
				Password:  flow,
			})))
			return true
		}
	case "option.TrojanUser", "option.ShadowTLSUser", "option.ShadowsocksUser", "option.Hysteria2User":
		uname, err := b.conec.ReciveVal("send user name")
		if err != nil {
			return false
		}
		passwrd, err := b.conec.ReciveVal("send user password")
		if err != nil {
			return false
		}

		switch endtyp {
		case "option.TrojanUser":
			slice.Set(reflect.Append(slice, reflect.ValueOf(option.TrojanUser{
				Name: uname,
				Password: passwrd,
			})))
		case "option.ShadowTLSUser":
			slice.Set(reflect.Append(slice, reflect.ValueOf(option.ShadowTLSUser{
				Name: uname,
				Password: passwrd,
			})))
		case "option.ShadowsocksUser":
			slice.Set(reflect.Append(slice, reflect.ValueOf(option.ShadowsocksUser{
				Name: uname,
				Password: passwrd,
			})))
		case"option.Hysteria2User":
			slice.Set(reflect.Append(slice, reflect.ValueOf(option.Hysteria2User{
				Name: uname,
				Password: passwrd,
			})))

		}
	case "option.HysteriaUser":
		return false
		// uname, err := b.conec.ReciveVal("send user name")
		// if err != nil {
		// 	return false
		// }
		
		// slice.Set(reflect.Append(slice, reflect.ValueOf(option.HysteriaUser{
		// 	Name: uname,
		// 	A
		// })))
	}
	return b.commonApnd(slice, path)
}

func (b *Builder) AddInbound(item any) error {
	//item maybe json, vless urlencoded, or option.Inbound, 
	finalin := option.Inbound{}
	
	switch it := item.(type) {
	
	case Create:
		var err error
		finalin, err = b.createFullInbound()
		if err != nil {
			return 	err
		}
	case string:
		switch {
		case strings.HasPrefix("{", it):
			err := json.Unmarshal([]byte(it), &finalin)
			if err != nil {		
				b.conec.AlertSend("error unmarshaling json " + err.Error())
				return err
			}
		}
	case option.Inbound:
		finalin = it
	case *option.Inbound:
		finalin = *it
	}
	err := b.validatein(finalin)
	if err != nil {
		return err
	}
	b.opts.Inbounds = append(b.opts.Inbounds, finalin)
	b.inbounds[finalin.Tag] = &b.opts.Inbounds[len(b.opts.Inbounds) - 1]
	return nil

}
func (b *Builder) RemoveInbound(tag string) error {
	delete(b.inbounds, tag)
	return nil
}
func (b *Builder) InboundFieldEditor(tag string) error {
	in, ok := b.inbounds[tag]
	if !ok {
		return ErrInboundNotFound
	}
	return b.inFieldsEditor(in)
} 
func (b *Builder) validatein(in option.Inbound) error {
	_, already := b.inbounds[in.Tag]
	if already {
		return errors.New("inbound with same tag " + in.Tag + " is already in inbound list")
	}
	for n := range allintyp {
		if allintyp[n] == in.Type {
			break
		}
		if n == len(allintyp) - 1 {
			return errors.New("unknown inbound type " + in.Type)
		}
	}
	return nil
}
func (b *Builder) inFieldsEditor(in *option.Inbound) error {
	inboundwlk, err := walker.NewWalker(in)
	if err != nil {
		return err
	}
	inboundwlk.SetValue = b.inbounsetval
	inboundwlk.CanSetCheck = b.inbounsetcheck
	inboundwlk.ApendCt = b.inbounapnd
	inboundwlk.SliceTagHook = func(val reflect.Value, path string, i int) string {
		switch {
			case val.Kind() == reflect.String:
				return val.String()
			case strings.HasSuffix(val.Type().String(), "VLESSUser"):
				if val.FieldByName("Name").String() != "" {
					return val.FieldByName("Name").String()
				}
		}

		return strconv.Itoa(i)
	}
	return b.anyFieldChange(inboundwlk, in)
	
}






//Outbound
func (b *Builder) outSetValue(curval reflect.Value, path string, wlkr *walker.Walker, item any,)  (reflect.Value, bool)  {
	return b.commonSetValue(curval, path, wlkr, item)
}
func (b *Builder) outCheckVal(curval reflect.Value, nextItemPath string, wlkr *walker.Walker) bool {
	return b.commonCheck(curval, nextItemPath, wlkr)
}
func (b *Builder) outApnd(slice reflect.Value, path string) bool {
	return b.commonApnd(slice, path)
}

func (b *Builder) outTransport(typ, host, path string) (*option.V2RayTransportOptions, error) {
	
	switch typ {
	case "tcp":
		return nil, nil
	case "ws":
		return &option.V2RayTransportOptions{
			Type: "ws",
			WebsocketOptions: option.V2RayWebsocketOptions{
				Headers: badoption.HTTPHeader(map[string]badoption.Listable[string]{
					"host":{host},
				}),
				Path: path,
			},
		}, nil
	case "http":
		return &option.V2RayTransportOptions{
			Type: "http",
			HTTPOptions: option.V2RayHTTPOptions{
				Host: badoption.Listable[string]{host},
				Path: path,
				Method: "GET",
			},
		}, nil
	case "httpupgrade":
		return &option.V2RayTransportOptions{
			Type: "httpupgrade",
			HTTPUpgradeOptions: option.V2RayHTTPUpgradeOptions{
				Host: host,
				Path: path,
			},
		}, nil
	case "grpc":
		return &option.V2RayTransportOptions{
			Type: "grpc",
			GRPCOptions: option.V2RayGRPCOptions{
				ServiceName: host,
			},
		}, nil
	case "quic":
			return &option.V2RayTransportOptions{
				Type: "quic",
				QUICOptions: option.V2RayQUICOptions{},
			}, nil
	}
	return nil, ErrTransportTyUnknown
}
func (b *Builder) outTransportUrl(url *url.URL) (*option.V2RayTransportOptions, error) {
	
	switch url.Query().Get("type") {
	case "tcp":
		return nil, nil
	case "ws":
		return &option.V2RayTransportOptions{
			Type: "ws",
			WebsocketOptions: option.V2RayWebsocketOptions{
				Headers: badoption.HTTPHeader(map[string]badoption.Listable[string]{
					"host":{url.Query().Get("host")},
				}),
				Path: url.Query().Get("path"),
			},
		}, nil
	case "http":
		return &option.V2RayTransportOptions{
			Type: "http",
			HTTPOptions: option.V2RayHTTPOptions{
				Host: badoption.Listable[string]{url.Query().Get("host")},
				Path: url.Query().Get("path"),
				Method: "GET",
			},
		}, nil
	case "httpupgrade":
		return &option.V2RayTransportOptions{
			Type: "httpupgrade",
			HTTPUpgradeOptions: option.V2RayHTTPUpgradeOptions{
				Host: url.Query().Get("host"),
				Path: url.Query().Get("path"),
			},
		}, nil
	case "grpc":
		return &option.V2RayTransportOptions{
			Type: "grpc",
			GRPCOptions: option.V2RayGRPCOptions{
				ServiceName: url.Query().Get("service"),
			},
		}, nil
	case "quic":
			return &option.V2RayTransportOptions{
				Type: "quic",
				QUICOptions: option.V2RayQUICOptions{},
			}, nil
	}
	return nil, ErrTransportTyUnknown
}
func (b *Builder) outTlsURL(url *url.URL) (*option.OutboundTLSOptions) {
	
	if url.Query().Get("security") == "tls" {
		out :=  &option.OutboundTLSOptions{
			Enabled: true,
			ServerName: url.Query().Get("sni"),
			Insecure: true,
		}
		if url.Query().Get("fp") != "" {
			out.UTLS = &option.OutboundUTLSOptions{
				Enabled: true,
				Fingerprint: url.Query().Get("fp"),
			}
		}
		if url.Query().Get("alpn") != "" {
			out.ALPN = strings.Split(url.Query().Get("alpn"), ",")
		}
		return out
	}
	return nil
}


func (b *Builder) AddOutbound(out any) error { 
	finaot := option.Outbound{}
	var err error
	switch ot := out.(type) {
		case string:
			finaot.Tag = b.reciveTag("out")
			if finaot.Tag == "" {
				return ErrTagRecive
			}
			switch {
			case strings.HasPrefix(ot, "vless://"), strings.HasPrefix(ot, "trojan://"):
				url, err := url.Parse(ot)
				if err != nil {
					return err
				}
				prt, err := strconv.Atoi(url.Port())
				if err != nil {
					return err
				}
				sropts :=  option.ServerOptions{
					Server: url.Hostname(),
					ServerPort: uint16(prt),
				}
				transport, err := b.outTransportUrl(url)
				if err != nil {
					return err
				}
				switch url.Scheme {
				case "vless":
					finaot.Type = "vless"
					finaot.Options = &option.VLESSOutboundOptions{
						ServerOptions: sropts,
						UUID: url.User.String(),
						OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
							TLS: b.outTlsURL(url),
						},
						Transport: transport,
					}
				case "trojan":
					finaot.Type = "trojan"
					finaot.Options = &option.TrojanOutboundOptions{
						Password: url.User.String(),
						ServerOptions: sropts,
						OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
							TLS: b.outTlsURL(url),
						},
						Transport: transport,
					}
				}
			case strings.HasPrefix(ot, "vmess://"):
				vmess, err := ParseVmessUrl(ot)
				if err != nil {
					return err
				}
				finaot.Type = "vmess"
				transort, err := b.outTransport(vmess.Net, vmess.Host, vmess.Path)
				if err != nil {
					return err
				}
				aid, _  :=  strconv.Atoi(vmess.Aid)
				port, _ := strconv.Atoi(vmess.Port)
				finaot.Options = &option.VMessOutboundOptions{
					DialerOptions: option.DialerOptions{},
					ServerOptions: option.ServerOptions{
						Server: vmess.Add,
						ServerPort: uint16(port),
					},
					AlterId: aid,
					Transport: transort,
					UUID: vmess.ID,
					Security: vmess.SCY,
				}
				if vmess.TLS != "" {
					finaot.Options.(*option.VMessOutboundOptions).TLS  = &option.OutboundTLSOptions{
						Enabled: true,
						ServerName: vmess.SNI,
						ALPN: vmess.Alpn,
					}
					if vmess.Fp != "" {
						finaot.Options.(*option.VMessOutboundOptions).TLS.UTLS = &option.OutboundUTLSOptions{
							Enabled: true,
							Fingerprint: vmess.Fp,
						}
					}
				}
			case strings.HasPrefix(ot, "{"):
				err := finaot.UnmarshalJSONContext(b.ctx, []byte(ot))
				if err != nil {
					return err
				}
			default:
				b.conec.AlertSend("looks like you sending unsupported string if you have custom protcol try sending json formated one")
				return ErrOutboundTyUnknown
			}
		case option.Outbound:
			finaot = ot
		case *option.Outbound:
			finaot = *ot
		case Create:
			finaot, err = b.createFullOutbound()
			if err != nil {
				return err
			}
	}
	err = b.validateout(finaot)
	if err != nil {
		return err
	}
	b.opts.Outbounds = append(b.opts.Outbounds, finaot)
	b.outbounds[finaot.Tag] = &b.opts.Outbounds[len(b.opts.Outbounds) - 1]
	if b.outselector != nil {
		se, err := b.conec.Select([]string{"yes", "no"}, "do you want to add this outbound's tag into selector")
		if err == nil && se == "yes" {
			outse, ok := b.outselector.Options.(*option.SelectorOutboundOptions)
			if ok {
				outse.Outbounds = append(outse.Outbounds, finaot.Tag)
			}
		}
	}
	return nil
}
//TODO: this method should be change after making distributed system
func (b *Builder) LoadOutFromDbconf(dbconf db.Config, cin conf.Inboud, expinfo conf.ExportInfo) error { 
	finaot := option.Outbound{}
	finaot.Tag = b.reciveTag("out")
	if finaot.Tag == "" {
		return ErrTagRecive
	}
	var err error
	switch cin.Type {
	case "vless":
		finaot.Type = "vless"
		opts := &option.VLESSOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server: cin.PublicIp,
				ServerPort: uint16(cin.Port()),
			},
			UUID: dbconf.UUID,
		}
		if cin.Tlsenabled {
			opts.TLS = &option.OutboundTLSOptions{
				Enabled: true,
				ServerName: expinfo.Sni,
				Insecure: true,
			}
		}
		opts.Transport, err = b.outTransport(cin.TransPortType, expinfo.Host, cin.TransPortPath)
		if err != nil {
			return err
		}
		finaot.Options = opts
	case "tuic":
		return errors.New("currently tuic loading not supported")
	case "trojan":
		finaot.Type = "trojan"
		opts := &option.TrojanOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server: cin.PublicIp,
				ServerPort: uint16(cin.Port()),
			},
			Password: dbconf.Password,
		}
		if cin.Tlsenabled {
			opts.TLS = &option.OutboundTLSOptions{
				Enabled: true,
				ServerName: expinfo.Sni,
				Insecure: true,
			}
		}
		opts.Transport, err = b.outTransport(cin.TransPortType, expinfo.Host, cin.TransPortPath)
		if err != nil {
			return err
		}
		finaot.Options = opts
	}
	err = b.validateout(finaot)
	if err != nil {
		return err
	}
	b.opts.Outbounds = append(b.opts.Outbounds, finaot)
	b.outbounds[finaot.Tag] = &b.opts.Outbounds[len(b.opts.Outbounds) - 1]
	return nil
}
func (b *Builder) RemoveOutbound(tag string) error { 
	delete(b.outbounds, tag)
	return nil
}
func (b *Builder) OutboundFieldsChange(tag string) error {
	out, ok := b.outbounds[tag]
	if !ok {
		return ErrInboundNotFound
	}
	return b.outFieldEdit(out)
}
func (b *Builder) outFieldEdit(out *option.Outbound) error {
	outboundwlk, err := walker.NewWalker(out)
	if err != nil {
		return err
	}
	outboundwlk.SetValue = b.outSetValue
	outboundwlk.CanSetCheck = b.outCheckVal
	outboundwlk.ApendCt = b.outApnd
	outboundwlk.SliceTagHook = func(val reflect.Value, path string, i int) string {
		switch {
			case val.Kind() == reflect.String:
				return val.String()
		}
		return strconv.Itoa(i)
	}
	return b.anyFieldChange(outboundwlk, out)
	
}
func (b *Builder) validateout(out option.Outbound) error {
	_, already := b.outbounds[out.Tag]
	if already {
		return errors.New("outbound with same tag " + out.Tag + " is already in outbound list")
	}
	_, already = b.endpoints[out.Tag]
	if already {
		return errors.New("endpoint with same tag " + out.Tag + " is already in endpoint list")
	}
	return nil
}





func (b *Builder) AddEndpoint(out any) error { 
	finaot := option.Endpoint{}
	switch ot := out.(type) {
		case string:
			err := finaot.UnmarshalJSONContext(b.ctx, []byte(ot))
			if err != nil {
				return err
			}
		case option.Endpoint:
			finaot = ot
		case *option.Endpoint:
			finaot = *ot
		case Create:
			var err error
			finaot, err = b.createFullEndPoint()
			if err != nil {
				return err
			}
	}
	err := b.validateendpoint(finaot)
	if err != nil {
		return err
	}
	b.opts.Endpoints = append(b.opts.Endpoints, finaot)
	b.endpoints[finaot.Tag] = &b.opts.Endpoints[len(b.opts.Endpoints) - 1]
	return nil
}
func (b *Builder) RemoveEndpoint(tag string) error { 
	delete(b.endpoints, tag)
	return nil
}
func (b *Builder) EndpointFieldEdit(tag string) error {
	end, ok := b.endpoints[tag]
	if !ok {
		return ErrEndpointNotFound
	}
	return b.endpointFieldEdit(end)
}
func (b *Builder) endpointFieldEdit(end *option.Endpoint) error {
	endpointwlk, err := walker.NewWalker(end)
	if err != nil {
		return err
	}
	endpointwlk.SetValue = b.commonSetValue
	endpointwlk.CanSetCheck = b.commonCheck
	endpointwlk.ApendCt = b.commonApnd
	endpointwlk.SliceTagHook = func(val reflect.Value, path string, i int) string {
		return strconv.Itoa(i)
	}
	return b.anyFieldChange(endpointwlk, end)
	
}
func (b *Builder) validateendpoint(out option.Endpoint) error {
	_, already := b.endpoints[out.Tag]
	if already {
		return errors.New("endpoint with same tag " + out.Tag + " is already in endpoint list")
	}
	_, already = b.outbounds[out.Tag]
	if already {
		return errors.New("outbound with same tag " + out.Tag + " is already in outbound list")
	}
	return nil
}

//Route Rule
func (b *Builder) AddRrule(r any) error {
	if r == nil {
		return ErrNilInput
	}
	if b.opts.Route == nil {
		b.opts.Route = &option.RouteOptions{}
	}
	var (
		finrule option.Rule
		err error
	)
	switch rule := r.(type) {
	case *option.Rule:
		finrule = *rule
	case option.Rule:
		finrule = rule
	case string:
		err = finrule.UnmarshalJSON([]byte(rule))
	case []byte:
		err = finrule.UnmarshalJSON(rule)
	}
	if err != nil {
		return Error{
			error: err,
		}
	}
	if finrule.Type == singC.RuleTypeLogical && finrule.LogicalOptions.Rules != nil {
		for lRule := range finrule.LogicalOptions.Rules {
			err = b.validateRrule(finrule.LogicalOptions.Rules[lRule])
			if err != nil {
				return err
			}
		}
	}
	err = b.validateRrule(finrule)

	if err != nil {return err}
	b.opts.Route.Rules = append(b.opts.Route.Rules, finrule)
	return nil
}
func (b *Builder) validateRrule(r option.Rule) error {
	
	if r.DefaultOptions.RuleSet != nil {
		for _, rset := range r.DefaultOptions.RuleSet {
			_, ok := b.ruleset[rset]
			if !ok {
				return Error{
					error: errors.New("unknown ruleset name ["+rset+"] in rule which isn't in current config, remove ruleset name or change it"),
				}
			}
		}
	}
	switch r.DefaultOptions.Action {
	case singC.RuleActionTypeRoute:
		_, ok := b.outbounds[r.DefaultOptions.RouteOptions.Outbound]
		if !ok {
			return Error{
				error: errors.New("unknown outbound ["+r.DefaultOptions.RouteOptions.Outbound+"] in rule which isn't in current config, change rule outbound name "),
			}
		}
	}
	return nil
}

//DNS rule
func (b *Builder) AddDrule(r any) error {
	if r == nil {
		return ErrNilInput
	}
	if b.opts.DNS == nil {
		b.opts.DNS = &option.DNSOptions{}
	}
	var (
		finrule option.DNSRule
		err error
	)
	switch rule := r.(type) {
	case *option.DNSRule:
		finrule = *rule
	case option.DNSRule:
		finrule = rule
	case string:
		err = finrule.UnmarshalJSONContext(globalCtx, []byte(rule))
	case []byte:
		err = finrule.UnmarshalJSONContext(globalCtx, rule)
	}
	if err != nil {
		return Error{
			error: err,
		}
	}
	if finrule.Type == singC.RuleTypeLogical && finrule.LogicalOptions.Rules != nil {
		for lRule := range finrule.LogicalOptions.Rules {
			err = b.validateDrule(finrule.LogicalOptions.Rules[lRule])
			if err != nil {
				return err
			}
		}
	}
	err = b.validateDrule(finrule)
	if err != nil {return err}
	b.opts.DNS.Rules = append(b.opts.DNS.Rules, finrule)
	return nil
}
func (b *Builder) validateDrule(r option.DNSRule) error {

	if r.DefaultOptions.RuleSet != nil {
		for _, rset := range r.DefaultOptions.RuleSet {
			_, ok := b.ruleset[rset]
			if !ok {
				return Error{
					error: errors.New("unknown ruleset name ["+rset+"] in rule which isn't in current config, remove ruleset name or change it"),
				}
			}
		}
	}
	switch r.DefaultOptions.Action {
	case singC.RuleActionTypeRoute:
	}
	return nil
}



//Route
func (b *Builder) createRouteRule() (option.Rule, error) {
	r := option.Rule{}
	s, err := b.conec.Select([]string{"default", "logical"}, "select rule type")
	if err != nil {
		return r, err
	}
	switch s {
	case "default":
		r.Type = "default"
		r.DefaultOptions = option.DefaultRule{}
		optswlkr, _  := walker.NewWalker(&r.DefaultOptions)
		optswlkr.SetValue = b.commonSetValue
		optswlkr.CanSetCheck = b.commonCheck
		optswlkr.ApendCt = b.commonApnd
		optswlkr.SliceTagHook = func(val reflect.Value, path string, i int) string {
			switch val.Kind() {
			case reflect.String:
				return val.String()
			}
			return strconv.Itoa(i)
		}
		err  = b.anyFieldChange(optswlkr, &r.DefaultOptions)
		return r, err
	case "logical":
		r.Type = "logical"
		r.LogicalOptions = option.LogicalRule{}
		optswlkr, _  := walker.NewWalker(&r.LogicalOptions)
		optswlkr.SetValue = b.commonSetValue
		optswlkr.CanSetCheck = b.commonCheck
		optswlkr.ApendCt = b.routeruleApnd
		optswlkr.SliceTagHook = func(val reflect.Value, path string, i int) string {
			switch val.Kind() {
			case reflect.String:
				return val.String()
			}
			return strconv.Itoa(i)
		}
		err  = b.anyFieldChange(optswlkr, &r.DefaultOptions)
		return r, err
	}
	return option.Rule{}, errors.New("not supported or unknown rule type")
}
func (b *Builder) routeruleApnd(slice reflect.Value, path string) bool {
	switch slice.Type().String() {
	case "[]option.Rule":
		newr, err := b.createRouteRule()
		if err != nil {
			return false
		}
		slice.Set(reflect.Append(slice, reflect.ValueOf(newr)))
	case "[]option.RuleSet":
		tag, err := b.conec.ReciveVal("send tag name for ruleset")
		_, ok := b.ruleset[tag]
		if ok {
			b.conec.AlertSend("there is already rule set with tag " + tag)
			return false
		}
		if err != nil {
			return false
		}
		typ, err := b.conec.Select([]string{singC.RuleSetTypeInline, singC.RuleSetTypeRemote, singC.RuleSetTypeLocal}, "select rule set type")
		if err != nil {
			return false
		}
		newr := option.RuleSet{
			Tag: tag,
			Type: typ,
		}
		b.ruleset[newr.Tag] = &newr
		slice.Set(reflect.Append(slice, reflect.ValueOf(newr)))
	}
	return b.commonApnd(slice, path)
}
func (b *Builder) RouteFieldChange() error {
	if b.opts.Route == nil {
		b.opts.Route = &option.RouteOptions{}
	} 
	rwlkr, err :=  walker.NewWalker(b.opts.Route)
	if err != nil {
		return err
	}
	rwlkr.SetValue = b.commonSetValue
	rwlkr.ApendCt = b.routeruleApnd
	rwlkr.CanSetCheck = b.commonCheck
	return b.anyFieldChange(rwlkr, b.opts.Route)
}



//DNS
func (b *Builder) createDnsRule() (option.DNSRule, error) {
	r := option.DNSRule{}
	s, err := b.conec.Select([]string{"default", "logical"}, "select rule type")
	if err != nil {
		return r, err
	}
	switch s {
	case "default":
		r.Type = "default"
		r.DefaultOptions = option.DefaultDNSRule{}
		optswlkr, _  := walker.NewWalker(&r.DefaultOptions)
		optswlkr.SetValue = b.commonSetValue
		optswlkr.CanSetCheck = b.commonCheck
		optswlkr.ApendCt = b.dnsApnd
		optswlkr.SliceTagHook = func(val reflect.Value, path string, i int) string {
			switch val.Kind() {
			case reflect.String:
				return val.String()
			}
			return strconv.Itoa(i)
		}
		err  = b.anyFieldChange(optswlkr, &r.DefaultOptions)
		return r, err
	}
	return r, errors.New("not supported or unknown rule type")
}
func (b *Builder) dnsApnd(slice reflect.Value, path string) bool {
	switch slice.Type().String() {
	case "[]option.DNSServerOptions":
		//s, err := b.conec.Select([]string{"tcp", "udp"}, "select dns server type") //TODO: add more later
		//if err != nil {return false}
		var err error
		srv := option.DNSServerOptions{}
		srv.Tag, err = b.conec.ReciveVal("send dns server tag")
		if err != nil {return false}
		dnswlkr, err :=  walker.NewWalker(&srv)
		if err != nil {
			return false
		}
		dnswlkr.SetValue = b.commonSetValue
		dnswlkr.CanSetCheck = b.commonCheck
		dnswlkr.ApendCt = b.dnsApnd
		err = b.anyFieldChange(dnswlkr, &srv)
		if err != nil {return false}
		slice.Set(reflect.Append(slice, reflect.ValueOf(srv)))
		return true
		// switch s {
		// 	case "tcp":
		// 		//srv.Type = "tcp"
		// 		// srv.Options = &option.LocalDNSServerOptions{
		// 		// 	DialerOptions: option.DialerOptions{}, //TODO: after creating dialer option add it here
		// 		// }
		// 		dnswlkr, err :=  walker.NewWalker(&srv)
		// 		if err != nil {
		// 			return false
		// 		}
		// 		dnswlkr.SetValue = b.commonSetValue
		// 		dnswlkr.CanSetCheck = b.commonCheck
		// 		dnswlkr.ApendCt = b.dnsApnd
		// 		err = b.anyFieldChange(dnswlkr, &srv)
		// 		if err != nil {return false}
		// 		slice.Set(reflect.Append(slice, reflect.ValueOf(srv)))
		// 		return true
		// 		//
		// 		// srv := sboxoption.DNSServerOptions{}
		// 	case "udp":
		//}
	case "[]option.DNSRule":
		newr, err := b.createDnsRule()
		if err != nil {
			return false
		}
		slice.Set(reflect.Append(slice, reflect.ValueOf(newr)))
	}
	return b.commonApnd(slice, path)
}
func (b *Builder) DNSfieldChange() error {
	if b.opts.DNS == nil {
		b.opts.DNS = &option.DNSOptions{}
	} 
	dnswlkr, err :=  walker.NewWalker(b.opts.DNS)
	if err != nil {
		return err
	}
	dnswlkr.SetValue = b.commonSetValue
	dnswlkr.CanSetCheck = b.commonCheck
	dnswlkr.ApendCt = b.dnsApnd
	return b.anyFieldChange(dnswlkr, b.opts.DNS)
}



func (b *Builder) ExperimentalField() error {
	if b.opts.Experimental == nil {
		b.opts.Experimental = &option.ExperimentalOptions{}
	}
	wlkr, err :=  walker.NewWalker(b.opts.Experimental)
	if err != nil {
		return err
	}
	wlkr.SetValue = b.commonSetValue
	wlkr.CanSetCheck = b.commonCheck
	wlkr.ApendCt = b.commonApnd
	return b.anyFieldChange(wlkr, b.opts.Experimental)
}
func (b *Builder) LogFieldChange() error {
	if b.opts.Log == nil {
		b.opts.Log = &option.LogOptions{}
	} 
	wlkr, err :=  walker.NewWalker(b.opts.Log)
	if err != nil {
		return err
	}
	wlkr.SetValue = b.commonSetValue
	wlkr.ApendCt = b.commonApnd
	wlkr.CanSetCheck = b.commonCheck
	return b.anyFieldChange(wlkr, b.opts.Log)
}
func (b *Builder) NTPFieldChange() error {
	if b.opts.NTP == nil {
		b.opts.NTP = &option.NTPOptions{}
	} 
	wlkr, err :=  walker.NewWalker(b.opts.NTP)
	if err != nil {
		return err
	}
	wlkr.SetValue = b.commonSetValue
	wlkr.ApendCt = b.commonApnd
	wlkr.CanSetCheck = b.commonCheck
	return b.anyFieldChange(wlkr, b.opts.NTP)
}




func (b *Builder) Export() io.Reader {
	b.opts.Endpoints = C.MapToSlicePtr(b.endpoints)
	b.opts.Inbounds = C.MapToSlicePtr(b.inbounds)
	b.opts.Outbounds = C.MapToSlicePtr(b.outbounds)
	if b.opts.Route != nil {
		b.opts.Route.RuleSet = C.MapToSlicePtr(b.ruleset)
	}
	return b.exportany(b.opts)
}
func (b *Builder) ExportInbound(tag string) (io.Reader, error){
	in, ok := b.inbounds[tag]
	if !ok {
		return nil, ErrInboundNotFound
	}
	return b.exportany(in), nil
}
func (b *Builder) ExportAllIn() io.Reader{
	b.opts.Inbounds = C.MapToSlicePtr(b.inbounds)
	return b.exportany(b.opts.Inbounds)
}
func (b *Builder) ExportAllOut() io.Reader{
	b.opts.Outbounds = C.MapToSlicePtr(b.outbounds)
	return b.exportany(b.opts.Outbounds)
}
func (b *Builder) ExportAllEnd() io.Reader{
	b.opts.Endpoints = C.MapToSlicePtr(b.endpoints)
	return b.exportany(b.opts.Endpoints)
}
func (b *Builder) exportany(item any) io.Reader  {
	if b.buf == nil {
		b.buf = &bytes.Buffer{}
		if b.CallBack != nil {
			b.buf = b.CallBack()
		}
	}
	b.buf.Reset()
	jd := json.NewEncoder(b.buf)
	jd.SetIndent("", " ")
	err := jd.Encode(item)
	if err != nil {
		b.buf.Write([]byte("error encoding json " + err.Error()))
	}
	jd = nil
	return b.buf
}
func(b *Builder)  ExportExtranal(item any) io.Reader {
	return b.exportany(item)
}