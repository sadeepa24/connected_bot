package inbound

import (
	"context"
	"net"
	"os"
	"sync"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/mux"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/common/uot"
	"github.com/sagernet/sing-box/connectedbot"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/transport/v2ray"
	vmess "github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing-vmess/packetaddr"
	"github.com/sagernet/sing-vmess/vless"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

var (
	_ adapter.Inbound           = (*VLESS)(nil)
	_ adapter.InjectableInbound = (*VLESS)(nil)
)

type VLESS struct {
	myInboundAdapter
	ctx       context.Context
	users     []option.VLESSUser
	userss    *sync.Map
	service   *vless.Service[int]
	tlsConfig tls.ServerConfig
	transport adapter.V2RayServerTransport

	// connlist []io.Closer

}

func NewVLESS(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.VLESSInboundOptions) (*VLESS, error) {
	inbound := &VLESS{
		myInboundAdapter: myInboundAdapter{
			protocol:      C.TypeVLESS,
			network:       []string{N.NetworkTCP},
			ctx:           ctx,
			router:        uot.NewRouter(router, logger),
			logger:        logger,
			tag:           tag,
			listenOptions: options.ListenOptions,
		},
		ctx:   ctx,
		users: options.Users,
		userss: &sync.Map{},
	}


	var err error
	inbound.router, err = mux.NewRouterWithOptions(inbound.router, logger, common.PtrValueOrDefault(options.Multiplex))
	if err != nil {
		return nil, err
	}
	service := vless.NewService[int](logger, adapter.NewUpstreamContextHandler(inbound.newConnection, inbound.newPacketConnection, inbound))
	service.UpdateUsers(common.MapIndexed(inbound.users, func(index int, _ option.VLESSUser) int {
		return index
	}), common.Map(inbound.users, func(it option.VLESSUser) string {
		return it.UUID
	}), common.Map(inbound.users, func(it option.VLESSUser) string {
		return it.Flow
	}), common.Map(inbound.users, func(it option.VLESSUser) int {
		return it.Maxlogin
	}),
)
	inbound.service = service
	if options.TLS != nil {
		inbound.tlsConfig, err = tls.NewServer(ctx, logger, common.PtrValueOrDefault(options.TLS))
		if err != nil {
			return nil, err
		}
	}
	if options.Transport != nil {
		inbound.transport, err = v2ray.NewServerTransport(ctx, common.PtrValueOrDefault(options.Transport), inbound.tlsConfig, (*vlessTransportHandler)(inbound))
		if err != nil {
			return nil, E.Cause(err, "create server transport: ", options.Transport.Type)
		}
	}
	inbound.connHandler = inbound
	return inbound, nil
}

func (h *VLESS) Start() error {
	if h.tlsConfig != nil {
		err := h.tlsConfig.Start()
		if err != nil {
			return err
		}
	}
	if h.transport == nil {
		return h.myInboundAdapter.Start()
	}
	if common.Contains(h.transport.Network(), N.NetworkTCP) {
		tcpListener, err := h.myInboundAdapter.ListenTCP()
		if err != nil {
			return err
		}
		go func() {
			sErr := h.transport.Serve(tcpListener)
			if sErr != nil && !E.IsClosed(sErr) {
				h.logger.Error("transport serve error: ", sErr)
			}
		}()
	}
	if common.Contains(h.transport.Network(), N.NetworkUDP) {
		udpConn, err := h.myInboundAdapter.ListenUDP()
		if err != nil {
			return err
		}
		go func() {
			sErr := h.transport.ServePacket(udpConn)
			if sErr != nil && !E.IsClosed(sErr) {
				h.logger.Error("transport serve error: ", sErr)
			}
		}()
	}
	return nil
}

func (h *VLESS) Close() error {
	return common.Close(
		h.service,
		&h.myInboundAdapter,
		h.tlsConfig,
		h.transport,
	)
}

func (h *VLESS) newTransportConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	h.injectTCP(conn, metadata)
	return nil
}

func (h *VLESS) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	var err error
	if h.tlsConfig != nil && h.transport == nil {
		conn, err = tls.ServerHandshake(ctx, conn, h.tlsConfig)
		if err != nil {
			return err
		}
	}
	return h.service.NewConnection(adapter.WithContext(log.ContextWithNewID(ctx), &metadata), conn, adapter.UpstreamMetadata(metadata))
}

func (h *VLESS) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	return os.ErrInvalid
}

func (h *VLESS) newConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	userIndex, loaded := auth.UserFromContext[int](ctx)
	if !loaded {
		return os.ErrInvalid
	}
	//user := h.users[userIndex].Name

	user := ""
	useerr, ok := h.userss.Load(userIndex)

	if ok {
		userob, ok := useerr.(string)
		if ok {
			user = userob
		}
	}

	if user == "" {
		user = F.ToString(userIndex)
	} else {
		metadata.User = user
	}
	h.logger.InfoContext(ctx, "[", user, "] inbound connection to ", metadata.Destination)
	return h.router.RouteConnection(ctx, conn, metadata)
}

func (h *VLESS) newPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	userIndex, loaded := auth.UserFromContext[int](ctx)
	if !loaded {
		return os.ErrInvalid
	}
	//user := h.users[userIndex].Name

	user := ""
	useerr, ok := h.userss.Load(userIndex)

	if ok {
		userob, ok := useerr.(string)
		if ok {
			user = userob
		}
	}

	if user == "" {
		user = F.ToString(userIndex)
	} else {
		metadata.User = user
	}
	if metadata.Destination.Fqdn == packetaddr.SeqPacketMagicAddress {
		metadata.Destination = M.Socksaddr{}
		conn = packetaddr.NewConn(conn.(vmess.PacketConn), metadata.Destination)
		h.logger.InfoContext(ctx, "[", user, "] inbound packet addr connection")
	} else {
		h.logger.InfoContext(ctx, "[", user, "] inbound packet connection to ", metadata.Destination)
	}
	return h.router.RoutePacketConnection(ctx, conn, metadata)
}

var _ adapter.V2RayServerTransportHandler = (*vlessTransportHandler)(nil)

type vlessTransportHandler VLESS

func (t *vlessTransportHandler) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	return (*VLESS)(t).newTransportConnection(ctx, conn, adapter.InboundContext{
		Source:      metadata.Source,
		Destination: metadata.Destination,
	})
}

func (h *VLESS) AddUser(user connectedbot.BotUser) (connectedbot.StatusOutput, error) {
	
	var status connectedbot.StatusOutput
	
	if user.Intype != C.TypeVLESS {
		return status, E.New("Invalid Inbound Type")
	}
	newuser, ok := user.User.(connectedbot.Vlessuser)
	if !ok {
		return status, E.New("Cannot convert user into ")
	}
	uid, err := newuser.Getuid()
	if err != nil {
		return status, err
	}

	_, Avlbl := h.service.CheckUser(uid)
	if Avlbl {
		usrstatus, err := h.service.Getstatus(uid)
		if err != nil {
			return status, err
		}
		return connectedbot.StatusOutput{
			Type: C.TypeVLESS,
			Status: connectedbot.VlessStatus{
				Disabled: usrstatus.Disabled,
				Download: usrstatus.Download,
				Upload: usrstatus.Upload,
			},
		}, nil
	}
	h.userss.Store(newuser.User, newuser.VlessUser.Name)
	err = h.service.Adduser(uid, newuser.VlessUser.Maxlogin, newuser.Bandwidth, newuser.User, newuser.VlessUser.Flow)
	return connectedbot.StatusOutput{
		Type: C.TypeVLESS,
		Status: connectedbot.VlessStatus{
			Disabled: false,
			Download: 0,
			Upload: 0,
		},
	}, err
}


func (h *VLESS) AddUserReset(user connectedbot.BotUser) (connectedbot.StatusOutput, error) {
	var status connectedbot.StatusOutput
	if user.Intype != C.TypeVLESS {
		return status, E.New("Invalid Inbound Type")
	}
	newuser, ok := user.User.(connectedbot.Vlessuser)
	if !ok {
		return status, E.New("Cannot convert user into ")
	}
	uid, err := newuser.Getuid()
	if err != nil {
		return status, err
	}
	res := connectedbot.StatusOutput{
		Type: C.TypeVLESS,
		Status: connectedbot.VlessStatus{
			Disabled: false,
			Download: 0,
			Upload: 0,
		},
	}
	_, Avlbl := h.service.CheckUser(uid)
	if Avlbl {
		usrstatus, err := h.service.RemoveUser(uid)
		if err != nil {
			return status, err
		}
		h.userss.Delete(newuser.User)
		res.Status = connectedbot.VlessStatus{
			Disabled: usrstatus.Disabled,
			Download: (usrstatus.Download),
			Upload: (usrstatus.Upload),
		}
	}
	h.userss.Store(newuser.User, newuser.VlessUser.Name)
	err = h.service.Adduser(uid, newuser.VlessUser.Maxlogin, newuser.Bandwidth, newuser.User, newuser.VlessUser.Flow)
	
	return res, err
}

func (h *VLESS) RemoveUser(user connectedbot.BotUser) (connectedbot.StatusOutput, error) {
	var status connectedbot.StatusOutput
	
	
	if user.Intype != C.TypeVLESS {
		return status, E.New("Invalid Inbound Type")
	}
	newuser, ok := user.User.(connectedbot.Vlessuser)
	if !ok {
		return status, E.New("Cannot convert user into connectedbot.Vlessuser ")
	}
	uid, err := newuser.Getuid()
	if err != nil {
		return status, err
	}
	
	
	usrstatus, err := h.service.RemoveUser(uid)
	h.userss.Delete(newuser.User)

	if err != nil {
		return status, err
	}
	return connectedbot.StatusOutput{
		Type: C.TypeVLESS,
		Status: connectedbot.VlessStatus{
			Disabled: usrstatus.Disabled,
			Download: usrstatus.Download,
			Upload: usrstatus.Upload,
			Online_ip: usrstatus.Ipmap,
		},
	}, nil

}

func (h *VLESS) GetastatusUser(user connectedbot.BotUser) (connectedbot.StatusOutput, error) {
	
	var status connectedbot.StatusOutput


	if user.Intype != C.TypeVLESS {
		return status, E.New("Invalid Inbound Type")
	}
	newuser, ok := user.User.(connectedbot.Vlessuser)
	if !ok {
		return status, E.New("Cannot convert user into ")
	}
	uid, err := newuser.Getuid()
	if err != nil {
		return status, err
	}

	usrstatus, err := h.service.Getstatus(uid)

	if err != nil {
		return status, err
	}
	return connectedbot.StatusOutput{
		Type: C.TypeVLESS,
		Status: connectedbot.VlessStatus{
			Disabled: usrstatus.Disabled,
			Download: usrstatus.Download,
			Upload: usrstatus.Upload,
			Online_ip: usrstatus.Ipmap,
		},
	}, nil
	
}

func (h *VLESS) CloseAll(user connectedbot.BotUser) error {

	if user.Intype != C.TypeVLESS {
		return E.New("Invalid Inbound Type")
	}
	newuser, ok := user.User.(connectedbot.Vlessuser)
	if !ok {
		return E.New("Cannot convert user into ")
	}
	uid, err := newuser.Getuid()
	if err != nil {
		return err
	}
	h.service.CloseAll(uid)
	return nil
}