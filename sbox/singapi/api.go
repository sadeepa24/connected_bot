package singapi

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sadeepa24/connected_bot/sbox/conf"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/connectedbot/opts"
	sC "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/deprecated"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/json"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/service"
	"go.uber.org/zap"
)

type BoxApi struct {
	box *box.Box
	inbounds map[int16]option.Inbound
	outbounds map[int16]string
	logger *zap.Logger
	urltests sync.Map

	clback sbox.Callback
}

var _ sbox.Controller = (*BoxApi)(nil)

func NewsingAPI(ctx context.Context, optpath string, logger *zap.Logger) (*BoxApi, option.Options, error) {
	filecont, err := os.ReadFile(optpath)
	if err!= nil {
		return nil, option.Options{}, err
	}	
	globalCtx := service.ContextWith(ctx, deprecated.NewStderrManager(log.StdLogger()))
	globalCtx = box.Context(globalCtx, include.InboundRegistry(), include.OutboundRegistry(), include.EndpointRegistry())

	opts, err := json.UnmarshalExtendedContext[option.Options](globalCtx,  filecont)
	if err != nil {
		return nil, opts, errors.New("sing box option unmarshelling error "+ err.Error())
	}
	instance, err := box.New(box.Options{
		Context: globalCtx,
		Options: opts,
	})
	if err != nil {
		return nil, opts, errors.Join(err, errors.New( "sing box insrance creation failed "))
	}
	logger.Debug("sing box instance created successfully")
	
	outbounds := make(map[int16]string, len(opts.Outbounds))
	for _, out := range opts.Outbounds {
		outbounds[*out.Id] = out.Tag
	}

	inbounds := make(map[int16]option.Inbound, len(opts.Inbounds))
	
	for _, in := range opts.Inbounds {
		switch in.Type {
		case "vless", "trojan":
		case "vmess":
			return nil, opts, errors.New("unsupported inbound type: " + in.Type + ". Protocols like vmess are not suitable for large user bases due to increased time complexity with user count.")
		default:
			return nil, opts, errors.New("unsupported inbound type: " + in.Type )
		}
		inbounds[*in.Id] = in
	}

	if len(inbounds) == 0 || len(outbounds) == 0 {
		return nil, opts, errors.New("inbound and outbound count cannot be zero")
	}

	return &BoxApi{
		box:      instance,
		logger: logger,
		inbounds: inbounds,
		outbounds: outbounds,
	}, opts, nil

}


func (s *BoxApi) Start() error {
	return s.box.Start()
}
func (s *BoxApi) Close() error {
	return s.box.Close()
}
func (b *BoxApi) SetCallBack(callback sbox.Callback) {
	b.box.SetCallback(b.ReciveCallback)
	b.clback = callback
}

func (b *BoxApi) AddConfig(dbconf *db.Config) (conf.Sboxstatus, error) {
	if dbconf.LeftQuota() <= 0 {
		return conf.Sboxstatus{}, ErrConfigNoQuota
	}
	return b.common(dbconf, b.box.AddUser)
}

func (b *BoxApi) AddConfigReset(dbconf *db.Config) (conf.Sboxstatus, error) {
	if dbconf.LeftQuota() <= 0 {
		return conf.Sboxstatus{}, ErrConfigNoQuota
	}
	return b.common(dbconf, b.box.AddUserReset)
}

func (b *BoxApi) GetStatusConfig(dbconf *db.Config) (conf.Sboxstatus, error) {
	return b.common(dbconf, b.box.GetStatusUser)
}

func (b *BoxApi) RemoveConfig(dbconf *db.Config) (conf.Sboxstatus, error) {
	dbconf.Active = false
	return b.common(dbconf, b.box.RemoveUser)
}

func (b *BoxApi) common(dbconf *db.Config , exec func(u opts.User) (opts.UserStatus, error) )(conf.Sboxstatus, error) {
	var stbox conf.Sboxstatus

	// if !dbconf.Canuse() {
	// 	return 
	// }
	out, ok  := b.outbounds[dbconf.OutboundID]
	if !ok {
		return stbox, ErrOutboundNotFound
	}
	ins, mp := b.getinlist(dbconf.InboundIds)
	if len(ins) == 0 {
		return stbox, ErrInboundNotFound
	}
	status, err := exec(opts.User{
		MaxLogin: dbconf.LoginLimit,
		Bandwidth:  dbconf.LeftQuota().Int64(),
		UserStr: dbconf.GetuniqName(),
		Outbound: out,
		Uid: int(dbconf.Id),
		InboundList: ins,
		Proto: &Comconf{
			dbconf: dbconf,
			inboundtypes: mp,
		},
	})
	stbox.Fill(status)
	return stbox, boxerror(err)
}

func (b *BoxApi) ResetInbounds(dbconf *db.Config) error {
	return b.commonNoError(dbconf, b.box.ResetInbound)
}

func (b *BoxApi) ChangeOutbound(dbconf *db.Config) error {
	opts, err := b.createopts(dbconf)
	if err != nil {
		return err
	}
	return boxerror(b.box.ChangeOutbound(opts))
}

func (b *BoxApi) CloseConns(dbconf *db.Config) error {
	return b.commonNoError(dbconf, b.box.CloseAllConn) 
}

func (b *BoxApi) UrlTest(outtag string) (int16, error) {
	outbound, ok := b.box.Outbound().Outbound(outtag)
	if !ok {
		return -1, errors.New("not found")
	}
	timeoutctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()
	return URLTest(timeoutctx, "", outbound)
}

func (b *BoxApi) RefreshUrlTest() {
	for _, tag := range b.outbounds {
		outbound, ok := b.box.Outbound().Outbound(tag)
		if !ok {
			continue
		}
		tmpcontext, cancle := context.WithTimeout(context.Background(), 50*time.Second)
		t, err := URLTest(tmpcontext, "", outbound)

		cancle()
		if err != nil {
			b.logger.Error("url testing err outbound " + tag + " ", zap.Error(err))
			continue
		}
		b.urltests.Swap(tag, t)

	}
}

func (b *BoxApi) ReciveCallback(code int16, stts *opts.CallBackResult) {
	if b.clback != nil {
		b.clback(code, stts)
	}
}



func (b *BoxApi) getinlist(inids []int16) ([]string, map[string]string) {
	var ins []string
	inimap := make(map[string]string, len(inids))
	for _, inid := range inids {
		ini, ok  := b.inbounds[inid]
		if !ok {
			continue
		}
		ins = append(ins, ini.Tag)
		inimap[ini.Type] = ini.Type
	}
	return ins, inimap
}

func (b *BoxApi) commonNoError(dbconf *db.Config , exec func(u opts.User) ) (error) {
	out, ok  := b.outbounds[dbconf.OutboundID]
	if !ok {
		return ErrOutboundNotFound
	}
	ins, mp := b.getinlist(dbconf.InboundIds)
	if len(ins) == 0 {
		return ErrInboundNotFound
	}
	exec(opts.User{
		MaxLogin: dbconf.LoginLimit,
		Bandwidth:  dbconf.LeftQuota().Int64(),
		UserStr: dbconf.GetuniqName(),
		Outbound: out,
		Uid: int(dbconf.Id),
		InboundList: ins,
		Proto: &Comconf{
			dbconf: dbconf,
			inboundtypes: mp,
		},
	})
	return nil
}

func (b *BoxApi) createopts(dbconf *db.Config) (opts.User, error)  {
	out, ok  := b.outbounds[dbconf.OutboundID]
	if !ok {
		return opts.User{}, ErrOutboundNotFound
	}
	ins, mp := b.getinlist(dbconf.InboundIds)
	if len(ins) == 0 {
		return opts.User{}, ErrInboundNotFound
	}
	return opts.User{
		MaxLogin: dbconf.LoginLimit,
		Bandwidth:  dbconf.LeftQuota().Int64(),
		UserStr: dbconf.GetuniqName(),
		Outbound: out,
		Uid: int(dbconf.Id),
		InboundList: ins,
		Proto: &Comconf{
			dbconf: dbconf,
			inboundtypes: mp,
		},
	}, nil
}

//optional
func (b *BoxApi) GetAllUserStatus() []opts.UserStatus {
	return b.box.AllUserStatus()
}



// Copied Function from singbox
func URLTest(ctx context.Context, link string, detour N.Dialer) (t int16, err error) {
	if link == "" {
		link = "https://google.com/"
	}
	linkURL, err := url.Parse(link)
	if err != nil {
		return
	}
	hostname := linkURL.Hostname()
	port := linkURL.Port()
	if port == "" {
		switch linkURL.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}

	start := time.Now()
	instance, err := detour.DialContext(ctx, "tcp", M.ParseSocksaddrHostPortStr(hostname, port))
	if err != nil {
		return
	}
	defer instance.Close()
	if earlyConn, isEarlyConn := common.Cast[N.EarlyConn](instance); isEarlyConn && earlyConn.NeedHandshake() {
		start = time.Now()
	}
	req, err := http.NewRequest(http.MethodHead, link, nil)
	if err != nil {
		return
	}
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return instance, nil
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: sC.TCPTimeout,
	}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return
	}
	resp.Body.Close()
	t = int16(time.Since(start) / time.Millisecond)
	return
}
