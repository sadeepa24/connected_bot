package singapi

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/sbox"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/connectedbot"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/deprecated"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-vmess/vless"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/json"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/service"
	"go.uber.org/zap"
)

type SingAPI struct {
	sbox.Sboxcontroller
	box *box.Box

	urltests *sync.Map
	outtags  []string
	logger   *zap.Logger

	logchan chan any
}

var _ sbox.Sboxcontroller = &SingAPI{}

func NewsingAPI(ctx context.Context, optpath string, logger *zap.Logger) (*SingAPI, option.Options, error) {
	filecont, err := os.ReadFile(optpath)
	if err!= nil {
		return nil, option.Options{}, err
	}

	var opts option.Options
	globalCtx := service.ContextWith(ctx, deprecated.NewStderrManager(log.StdLogger()))
	globalCtx = box.Context(globalCtx, include.InboundRegistry(), include.OutboundRegistry(), include.EndpointRegistry(), include.DNSTransportRegistry())

	opts, err = json.UnmarshalExtendedContext[option.Options](globalCtx,  filecont)
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
	outmap := []string{}
	for _, out := range opts.Outbounds {
		outmap = append(outmap, out.Tag)
	}
	return &SingAPI{
		box:      instance,
		urltests: &sync.Map{},
		logger:   logger,
		outtags:  outmap,
	}, opts, nil

}

func (s *SingAPI) Start() error {
	return s.box.Start()
}

func (s *SingAPI) Close() error {
	return s.box.Close()
}

func (s *SingAPI) AddUser(suser *sbox.Userconfig) (sbox.Sboxstatus, error) {
	return s.common(suser, func(botuser connectedbot.BotUser, intag string) (connectedbot.StatusOutput, error) {
		s.box.RemoveAllRule(suser.GetuniqName())
		s.box.Addoutbounduser(suser.GetuniqName(), suser.Outboundtag)
		return s.box.AddUser(botuser, intag)
	})
}

func (s *SingAPI) RemoveUser(suser *sbox.Userconfig) (sbox.Sboxstatus, error) {
	s.box.RemoveAllRule(suser.GetuniqName())
	return s.common(suser, s.box.RemoveUser)
}

func (s *SingAPI) GetstatusUser(suser *sbox.Userconfig) (sbox.Sboxstatus, error) {
	return s.common(suser, s.box.GetastatusUser)

}

func (s *SingAPI) AddUserReset(suser *sbox.Userconfig) (sbox.Sboxstatus, error) {
	return s.common(suser, func(botuser connectedbot.BotUser, intag string) (connectedbot.StatusOutput, error) {
		s.box.RemoveAllRule(suser.GetuniqName())
		s.box.Addoutbounduser(suser.GetuniqName(), suser.Outboundtag)
		return s.box.AddUserReset(botuser, intag)
	})

}

func (s *SingAPI) RemoveAllRuleForuser(user string) {
	s.box.RemoveAllRule(user)
}

func (s *SingAPI) CloseConns(suser *sbox.Userconfig) error {
	
	var upconfig connectedbot.BotUser
	switch suser.Type {

	case C.TypeVLESS:
		upconfig = connectedbot.BotUser{
			Intype: C.TypeVLESS,
			User: connectedbot.Vlessuser{
				User:      int(suser.DbID),
				Bandwidth: (suser.Quota - suser.Usage).Int(),
				VlessUser: option.VLESSUser{
					UUID:     suser.UUID.String(),
					Name:     suser.GetuniqName(),
					Maxlogin: int(suser.LoginLimit),
				},
			},
		}
	default:
		return constbot.ErrTypeMissmatch //only support vless for now
	}
	return s.box.CloseAll(upconfig, suser.Inboundtag)

 }



func (s *SingAPI) common(suser *sbox.Userconfig, chfunc func(connectedbot.BotUser, string) (connectedbot.StatusOutput, error)) (sbox.Sboxstatus, error) {
	realbandwidth := suser.Quota - suser.Usage
	var upconfig connectedbot.BotUser
	switch suser.Type {

	case C.TypeVLESS:
		upconfig = connectedbot.BotUser{
			Intype: C.TypeVLESS,
			User: connectedbot.Vlessuser{
				User:      int(suser.DbID),
				Bandwidth: int(realbandwidth),
				VlessUser: option.VLESSUser{
					UUID:     suser.UUID.String(),
					Name:     suser.GetuniqName(),
					Maxlogin: int(suser.LoginLimit),
				},
			},
		}
	default:
		return sbox.Sboxstatus{}, constbot.ErrTypeMissmatch //only support vless for now
	}
	var status connectedbot.StatusOutput
	var err error

	//fmt.Println(*suser)
	if realbandwidth <= 0 {
		status, err = s.box.RemoveUser(upconfig, suser.Inboundtag)
		if err != nil {
			if errors.Is(err, vless.ErrUserNotFound) {
				return sbox.Sboxstatus{
					Download:  0,
					Upload:    0,
					Disabled:  true,
					Online_ip: map[netip.Addr]int64{},
				}, nil
			}

		}

	} else {
		status, err = chfunc(upconfig, suser.Inboundtag)

	}

	if err != nil {
		return sbox.Sboxstatus{}, err
	}

	result, ok := status.Status.(connectedbot.VlessStatus)
	if !ok {
		return sbox.Sboxstatus{}, constbot.ErrResultMalformed
	}

	return sbox.Sboxstatus{
		Download:  constbot.Bwidth(result.Download),
		Upload:    constbot.Bwidth(result.Upload),
		Disabled:  result.Disabled,
		Online_ip: result.Online_ip,
	}, nil

}

func (s *SingAPI) SetLogChan(logchan chan any) {
	s.logchan = logchan
}

func (s *SingAPI) GetLogChan() chan any {
	return s.logchan
}
func (s *SingAPI) UrlTest(outtag string) (int16, error) {
	outbound, err := s.box.GetOutBound(outtag)
	if err != nil {
		return 0, err
	}
	timeoutctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()
	return URLTest(timeoutctx, "", outbound)

	// t, ok := s.urltests.Load(outtag)
	// if !ok {
	// 	return 10000, errors.New("not found")
	// }
	// return t.(uint16), nil

}

func (s *SingAPI) RefreshUrlTest() {
	for _, tag := range s.outtags {
		outbound, err := s.box.GetOutBound(tag)
		if err != nil {
			continue
		}
		tmpcontext, cancle := context.WithTimeout(context.Background(), 50*time.Second)
		t, err := URLTest(tmpcontext, "", outbound)

		cancle()
		if err != nil {
			s.logger.Error("url testing err outbound " + tag + " ", zap.Error(err))
			continue
		}
		s.urltests.Swap(tag, t)

	}
}

// Copied Function from singbox
func URLTest(ctx context.Context, link string, detour N.Dialer) (t int16, err error) {
	if link == "" {
		link = "https://1.1.1.1/"
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
		Timeout: C.TCPTimeout,
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
