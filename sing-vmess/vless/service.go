package vless

import (
	//"container/list"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/netip"
	"sync"
	"time"

	vmess "github.com/sagernet/sing-vmess"
	sAtomic "github.com/sagernet/sing/common/atomic"
	"github.com/sagernet/sing/common/auth"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"

	"github.com/gofrs/uuid/v5"
)

type Service[T comparable] struct {
	userMap  map[[16]byte]T
	userFlow map[T]string
	//poolMap map[[16]byte]*poolUnit
	logger   logger.Logger
	handler  Handler

	userMapsync *sync.Map
}

type UserStatus struct {
	Isuser bool
	Download int64
	Upload int64
	Disabled bool
	Ipmap map[netip.Addr]int64

}

type UserUnit[T comparable] struct {
	user T
	pool *poolUnit
	Flowstring string

}

type poolUnit struct {
	//ippmap sync.Map
	ipmap map[netip.Addr]*ipunit //TODO: use sync.Map instead of mutex
	maxlogin int
	poolaccsess *sync.RWMutex
	downlink *sAtomic.Int64
	uplink *sAtomic.Int64

	bandwidthlimit int
	disabled bool

	connections map[int32]JustCloser

	bandwidthlimiterdisabled bool

}


// custom errors for bot

var ErrUserNotFound = errors.New("user not found")
var ErrInboundNotFound = errors.New("inbound not found")
var ErrVlessService = errors.New("error occured from service when adding user")
var ErrInvalidInbound = errors.New("inbound missmatch")

var _ N.TCPConnectionHandler = (*Service[int])(nil)


func (s *Service[T]) Adduser(uuid uuid.UUID, loginlimit int, bandwidth int, user T, flow string) error {
	
	if loginlimit <= 0 {
		return ErrVlessService
	}
	
	s.userMapsync.Store(uuid, UserUnit[T]{
		user: user,
		pool: &poolUnit{
			ipmap: make(map[netip.Addr]*ipunit, loginlimit),
			maxlogin: loginlimit,
			poolaccsess: &sync.RWMutex{},
			downlink: new(sAtomic.Int64),
			uplink: new(sAtomic.Int64),
			bandwidthlimit: bandwidth,
			disabled: false,
			bandwidthlimiterdisabled: false,
			connections: map[int32]JustCloser{},
		},
		Flowstring: flow,
	})
	return nil
}

func (s *Service[T]) CheckUser(uuid uuid.UUID) (any, bool) {
	return s.userMapsync.Load(uuid)
}

func (s *Service[T]) Getstatus(uuid uuid.UUID) (UserStatus, error) {
	user, ok := s.userMapsync.Load(uuid)
	if !ok {
		return UserStatus{
			Isuser: false,
		}, ErrUserNotFound
	}
	userunit, ok := user.(UserUnit[T])
	if !ok {
		return UserStatus{}, E.New("user conversion ")
	}
	return UserStatus{
		Isuser: true,
		Disabled: userunit.pool.disabled,
		Download: userunit.pool.downlink.Load(),
		Upload: userunit.pool.uplink.Load(),
		Ipmap: s.getipmap(userunit),
	}, nil

}

//
func (s *Service[T]) CloseAll(uuid uuid.UUID) {
	unit, loaded := s.userMapsync.Load(uuid)
	if !loaded {
		return
	}
	userunit, ok := unit.(UserUnit[T])

	if !ok {
		return
	}

	userunit.pool.poolaccsess.RLock()
	for _, cls := range userunit.pool.connections {
		cls.JustClose()
	}
	userunit.pool.poolaccsess.RUnlock()	

}

func (s *Service[T]) getipmap(userunit UserUnit[T]) map[netip.Addr]int64 {

	tmpmap := map[netip.Addr]int64{}

	userunit.pool.poolaccsess.RLock()
	for ip, ipunit := range userunit.pool.ipmap {
		tmpmap[ip] = ipunit.count.Load()
	}
	userunit.pool.poolaccsess.RUnlock()
	return tmpmap
}

func (s *Service[T]) RemoveUser(uuid uuid.UUID) (UserStatus, error) {
	user, ok := s.userMapsync.LoadAndDelete(uuid)

	if !ok {
		
		return UserStatus{
			Isuser: false,
		}, ErrUserNotFound
	}

	userunit, ok := user.(UserUnit[T])
	if !ok {
		return UserStatus{}, E.New("user conversion ")
	}
	return UserStatus{
		Isuser: true,
		Disabled: userunit.pool.disabled,
		Download: userunit.pool.downlink.Load(),
		Upload: userunit.pool.uplink.Load(),
		Ipmap: s.getipmap(userunit),
	}, nil
}

type JustCloser interface {
	JustClose() error
}



// Does not Use anymore
func (s *Service[T]) StartcheckerV2() {
	ticker := time.NewTicker(5 * time.Second) // Increase sleep duration
    defer ticker.Stop()
	
	for {
		<-ticker.C 
		s.userMapsync.Range(func(key, value any) bool {
			realval, ok := value.(UserUnit[T])
			if !ok {
				return true
			}
			punit := realval.pool
			//fmt.Println(punit.downlink.Load())
			if (punit.downlink.Load())>= int64(punit.bandwidthlimit) && !punit.disabled && !punit.bandwidthlimiterdisabled{
				//fmt.Println("here")
				punit.poolaccsess.Lock()
				punit.disabled = true	
				punit.poolaccsess.Unlock()
			}
			if len(punit.ipmap) == 0 {
				return true
			}
			//fmt.Println(punit.downlink.Load()/ (1024*1024))
			for ip, unit := range punit.ipmap {
				//fmt.Printf("%v ", unit.count.Load())
				if unit.count.Load() == 0 {
					punit.poolaccsess.Lock()
					delete(punit.ipmap, ip)
					punit.poolaccsess.Unlock()
				}
			}
			return true

		})
	}
}


type ipunit struct {
	count *sAtomic.Int64
}

type Handler interface {
	N.TCPConnectionHandler
	N.UDPConnectionHandler
	E.Handler
}

func NewService[T comparable](logger logger.Logger, handler Handler) *Service[T] {
	return &Service[T]{
		logger:  logger,
		handler: handler,
	}
}

func (s *Service[T]) UpdateUsers(userList []T, userUUIDList []string, userFlowList []string, maxloginList []int) {
	userMap := make(map[[16]byte]T)
	userFlowMap := make(map[T]string)
	poolmap := make(map[[16]byte]*poolUnit)

	s.userMapsync = &sync.Map{}

	for i, userName := range userList {
		
		var mxlogin int
		userID := uuid.FromStringOrNil(userUUIDList[i])
		if userID == uuid.Nil {
			userID = uuid.NewV5(uuid.Nil, userUUIDList[i])
		}
		mxlogin = maxloginList[i]
		if mxlogin <= 0 {
			mxlogin = 1
		}
		poolunt := &poolUnit{
			ipmap: make(map[netip.Addr]*ipunit, mxlogin),
			maxlogin: 4,
			poolaccsess: &sync.RWMutex{},
			downlink: new(sAtomic.Int64),
			uplink: new(sAtomic.Int64),
			bandwidthlimit: 10 * 1024 * 1024 * 1024,
			disabled: false,
			bandwidthlimiterdisabled: false,
			connections: map[int32]JustCloser{},
		}

		poolmap[userID] = poolunt
		s.userMapsync.Store(userID, UserUnit[T]{
			user: userName,
			pool: poolunt,
			Flowstring: userFlowList[i],
		})


		//go poolmap[userID].startchecker()

		userMap[userID] = userName
		userFlowMap[userName] = userFlowList[i]
	}

	
	s.userMap = userMap
	s.userFlow = userFlowMap
	// s.poolMap = poolmap

	//go s.StartcheckerV2()

}

func (s *Service[T]) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	
	request, err := ReadRequest(conn)
	if err != nil {
		return err
	}
	userunit, is := s.userMapsync.Load(uuid.UUID(request.UUID))
	
	if !is {
		return E.New("unknown UUID: ", uuid.FromBytesOrNil(request.UUID[:]))
	}
	unit, ok := userunit.(UserUnit[T])
	if !ok { return E.New("type convertion error") }
	
	user := unit.user
	poolun := unit.pool

	poolun.poolaccsess.RLock()
	if poolun.disabled {
		poolun.poolaccsess.RUnlock()
		s.logger.Info("bancdwidth limit hit ", "user ")
		return E.New("Bandwidth limititation hit closing connections")
	}
	ipunitt, loaded := poolun.ipmap[metadata.Source.Addr]
	poolun.poolaccsess.RUnlock()
	

	
	if !loaded && len(poolun.ipmap) >= poolun.maxlogin {
		return E.New("ip pool already filled new connection from diffrent ips rejected ", metadata.Source.Addr.String(), string(request.UUID[:16]), )
	} else if !loaded {
		ipunitt = &ipunit{
			count: new(sAtomic.Int64),
		}
		poolun.poolaccsess.Lock()
		poolun.ipmap[metadata.Source.Addr] = ipunitt
		poolun.poolaccsess.Unlock()		
	}
	// s.botlog <- "user conn from " + string(request.UUID[:16])

	ctx = auth.ContextWithUser(ctx, user)
	metadata.Destination = request.Destination

	userFlow := s.userFlow[user]
	if request.Flow == FlowVision && request.Command == vmess.NetworkUDP {
		return E.New(FlowVision, " flow does not support UDP")
	} else if request.Flow != userFlow {
		return E.New("flow mismatch: expected ", flowName(userFlow), ", but got ", flowName(request.Flow))
	}

	if request.Command == vmess.CommandUDP {
		return s.handler.NewPacketConnection(ctx, &serverPacketConn{
			ExtendedConn: bufio.NewCounterConn(bufio.NewExtendedConn(conn), []N.CountFunc{
				func(n int64) {
					poolun.uplink.Add(n)
				},
			}, []N.CountFunc{
				func(n int64) {
					poolun.downlink.Add(n)
				},
			}), 
			destination: request.Destination}, metadata)
	}
	conid := rand.Int31()
	responseConn := &serverConn{
		ExtendedConn: bufio.NewCounterConn(bufio.NewExtendedConn(conn), []N.CountFunc{
			func(n int64) {
				poolun.uplink.Add(n)
			},
		}, []N.CountFunc{
			func(n int64) {
				poolun.downlink.Add(n)
			},
		}),
		conid: conid, 
		writer: bufio.NewVectorisedWriter(conn),
		ct: &closeconn{
			closed: new(sAtomic.Bool),

		},
		onClose: func ()  {

			if (poolun.downlink.Load())>= int64(poolun.bandwidthlimit) && !poolun.bandwidthlimiterdisabled{
				poolun.poolaccsess.Lock()
				for _, close := range poolun.connections {
					close.JustClose()
				}
				poolun.disabled = true
				poolun.ipmap = map[netip.Addr]*ipunit{}
				poolun.poolaccsess.Unlock()
				return
			}
			
			if ipunitt.count.Add(-1) <= 2 { // remove the ip addr from map only when 2 or less connection left
				poolun.poolaccsess.Lock()
				delete(poolun.ipmap, metadata.Source.Addr)
				delete(poolun.connections, conid)
				poolun.poolaccsess.Unlock()
				return
			}
			
		},
	}
	poolun.poolaccsess.Lock()
	poolun.connections[conid] = responseConn
	poolun.poolaccsess.Unlock()
	

	switch userFlow {
	case FlowVision:
		conn, err = NewVisionConn(responseConn, conn, request.UUID, s.logger)
		if err != nil {
			return E.Cause(err, "initialize vision")
		}
	case "":
		conn = responseConn
	default:
		return E.New("unknown flow: ", userFlow)
	}
	
	switch request.Command {
	case vmess.CommandTCP:
		ipunitt.count.Add(1)
		return s.handler.NewConnection(ctx, conn, metadata)
		
	case vmess.CommandMux:
		ipunitt.count.Add(1)
		err = vmess.HandleMuxConnection(ctx, conn, s.handler)
		ipunitt.count.Add(-1)
		return err
		
	default:
		return E.New("unknown command: ", request.Command)
	}
}

func flowName(value string) string {
	if value == "" {
		return "none"
	}
	return value
}

var _ N.VectorisedWriter = (*serverConn)(nil)

type serverConn struct {
	conid int32
	N.ExtendedConn
	writer          N.VectorisedWriter
	responseWritten bool
	onClose func() //TODO:
	ct *closeconn
}

type closeconn struct {
	closed *sAtomic.Bool
}

func (c *serverConn) Close() error {
	if !c.ct.closed.Swap(true) {
		c.onClose()
	}
	return c.ExtendedConn.Close()
}

func (c *serverConn) JustClose() error {
	return c.ExtendedConn.Close()
}

func (c *serverConn) Read(b []byte) (n int, err error) {
	return c.ExtendedConn.Read(b)
}

func (c *serverConn) Write(b []byte) (n int, err error) {
	if !c.responseWritten {
		_, err = bufio.WriteVectorised(c.writer, [][]byte{{Version, 0}, b})
		if err == nil {
			n = len(b)
		}
		c.responseWritten = true
		return
	}
	return c.ExtendedConn.Write(b)
}

func (c *serverConn) WriteBuffer(buffer *buf.Buffer) error {
	if !c.responseWritten {
		header := buffer.ExtendHeader(2)
		header[0] = Version
		header[1] = 0
		c.responseWritten = true
	}
	return c.ExtendedConn.WriteBuffer(buffer)
}

func (c *serverConn) WriteVectorised(buffers []*buf.Buffer) error {
	if !c.responseWritten {
		err := c.writer.WriteVectorised(append([]*buf.Buffer{buf.As([]byte{Version, 0})}, buffers...))
		c.responseWritten = true
		return err
	}
	return c.writer.WriteVectorised(buffers)
}

func (c *serverConn) FrontHeadroom() int {
	if c.responseWritten {
		return 0
	}
	return 2
}

func (c *serverConn) ReaderReplaceable() bool {
	return true
}

func (c *serverConn) WriterReplaceable() bool {
	return c.responseWritten
}

func (c *serverConn) NeedAdditionalReadDeadline() bool {
	return true
}

func (c *serverConn) Upstream() any {
	return c.ExtendedConn
}

type serverPacketConn struct {
	N.ExtendedConn
	responseWriter  io.Writer
	responseWritten bool
	destination     M.Socksaddr
}

func (c *serverPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = c.ExtendedConn.Read(p)
	if err != nil {
		return
	}
	if c.destination.IsFqdn() {
		addr = c.destination
	} else {
		addr = c.destination.UDPAddr()
	}
	return
}

func (c *serverPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if !c.responseWritten {
		if c.responseWriter == nil {
			var packetLen [2]byte
			binary.BigEndian.PutUint16(packetLen[:], uint16(len(p)))
			_, err = bufio.WriteVectorised(bufio.NewVectorisedWriter(c.ExtendedConn), [][]byte{{Version, 0}, packetLen[:], p})
			if err == nil {
				n = len(p)
			}
			c.responseWritten = true
			return
		} else {
			_, err = c.responseWriter.Write([]byte{Version, 0})
			if err != nil {
				return
			}
			c.responseWritten = true
		}
	}
	return c.ExtendedConn.Write(p)
}

func (c *serverPacketConn) ReadPacket(buffer *buf.Buffer) (destination M.Socksaddr, err error) {
	var packetLen uint16
	err = binary.Read(c.ExtendedConn, binary.BigEndian, &packetLen)
	if err != nil {
		return
	}

	_, err = buffer.ReadFullFrom(c.ExtendedConn, int(packetLen))
	if err != nil {
		return
	}

	destination = c.destination
	return
}

func (c *serverPacketConn) WritePacket(buffer *buf.Buffer, destination M.Socksaddr) error {
	if !c.responseWritten {
		if c.responseWriter == nil {
			var packetLen [2]byte
			binary.BigEndian.PutUint16(packetLen[:], uint16(buffer.Len()))
			err := bufio.NewVectorisedWriter(c.ExtendedConn).WriteVectorised([]*buf.Buffer{buf.As([]byte{Version, 0}), buf.As(packetLen[:]), buffer})
			c.responseWritten = true
			return err
		} else {
			_, err := c.responseWriter.Write([]byte{Version, 0})
			if err != nil {
				return err
			}
			c.responseWritten = true
		}
	}
	packetLen := buffer.Len()
	binary.BigEndian.PutUint16(buffer.ExtendHeader(2), uint16(packetLen))
	return c.ExtendedConn.WriteBuffer(buffer)
}

func (c *serverPacketConn) FrontHeadroom() int {
	return 2
}

func (c *serverPacketConn) NeedAdditionalReadDeadline() bool {
	return true
}

func (c *serverPacketConn) Upstream() any {
	return c.ExtendedConn
}
