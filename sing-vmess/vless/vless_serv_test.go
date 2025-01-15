package vless_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sagernet/sing-vmess/vless"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)


type HandlerForTest struct {
	N.TCPConnectionHandler
	N.UDPConnectionHandler
	E.Handler
}

var AddressSerializer = M.NewSerializer(
	M.AddressFamilyByte(0x01, M.AddressFamilyIPv4),
	M.AddressFamilyByte(0x03, M.AddressFamilyIPv6),
	M.AddressFamilyByte(0x02, M.AddressFamilyFqdn),
	M.PortThenAddress(),
)


var uuidList [][16]byte
var srcIpList []M.Metadata


func (h *HandlerForTest) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	//fmt.Println("new conn called")
	//fmt.Println(metadata)
	err := conn.Close()
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func (h *HandlerForTest) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata M.Metadata) error {
	//fmt.Println("new packet conn called")
	return nil
}

func (h *HandlerForTest) NewError(ctx context.Context, err error) {
	fmt.Println("new error called")
}



// this is special testing for connected bot
func TestVlessService(t *testing.T) {
	logchan := make(chan any, 1000)

	go func() {
		for val := range logchan {
			if inttt, ok := val.(int); ok {
				fmt.Println(inttt)
			}
			
		}
	}()

	testctx := context.Background()
	
	
	service := vless.NewService[int](&testlogger{}, &HandlerForTest{})
	service.UpdateUsers([]int{}, []string{}, []string{}, []int{})
	service.Adduser([16]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, 3, 222222, 1, "")
	

	// adding uuid to benchmark
	for i := 2; i < 10000; i++ {
		sss, _ := uuid.NewV4()
		service.Adduser(
			sss,
			2,
			100,
			i,
			"",
		)
	}

	
	createuuidList()
	createIpList()

	for _, id := range uuidList {
		service.Adduser(id, 10, 100, 3, "")
	}

	// go func ()  {
	// 	// ticker := time.NewTicker(10 * time.Millisecond)
	// 	for {
		
	// 		time.Sleep(10 * time.Millisecond)
	// 		fmt.Println("all closed")
	// 		service.CloseAll(uuidList[rand.Int31n(49)])
	// 	}	
	// }()
	
	for i := 0; i < 1000000; i++ {
		st := time.Now()
		err := service.NewConnection(testctx, &exampleconn{uid: uuidList[rand.Int31n(49)]}, srcIpList[rand.Int31n(99)])
		fmt.Println(time.Since(st).Microseconds())
		if err != nil{
			fmt.Println(err)
		}
	}


}

func createuuidList() {
	for i := 0; i < 50; i++ {
		uid, err := uuid.NewV4()
		if err != nil {
			continue
		} 
		uuidList = append(uuidList, uid)
	}

	// randomNumber := rand.Intn(50) 
}

func createIpList() {
	startipd := "172.10.19."
	port := ":443"
	dst := M.ParseSocksaddr("1.1.1.1:443")
	
	mtdata := M.Metadata{
		Protocol: "tcp",
		Destination: dst,

	}


	for i := 0; i < 100; i++ {
		ip := startipd + strconv.Itoa(i) + port
		mtdata.Source = M.ParseSocksaddr(ip)
		srcIpList = append(srcIpList, mtdata)
	}
}






type exampleconn struct{
	buf *bytes.Buffer
	bufset bool
	uid [16]byte
}
func (e *exampleconn) Read(b []byte) (n int, err error) {

	if !e.bufset {
		e.buf = makereq(e.uid)
		e.bufset = true
	}
	if e.buf == nil {
		return 0, err
	}
	return e.buf.Read(b)

}
func (e *exampleconn) Write(b []byte) (n int, err error) {
	return 0, nil
}
func (e *exampleconn) Close() error {
	return nil
}
func (e *exampleconn) LocalAddr() net.Addr {
	return nil
}
func (e *exampleconn) RemoteAddr() net.Addr {
	return nil
}
func (e *exampleconn) SetDeadline(t time.Time) error {
	return nil
}
func (e *exampleconn) SetReadDeadline(t time.Time) error {
	return nil
}
func (e *exampleconn) SetWriteDeadline(t time.Time) error {
	return nil
}

const (
	Version = 0
)

func makereq(uid [16]byte ) *bytes.Buffer {
	var buf bytes.Buffer
	
	request := vless.Request{
		UUID:     uid,
		Flow:     "",
		Command:  1, 
	}
	

	// Write version
	err := binary.Write(&buf, binary.BigEndian, uint8(Version))
	if err != nil {
			fmt.Println("Error writing version:", err)
			return nil
	}

	// Write UUID
	_, err = buf.Write(request.UUID[:])
	if err != nil {
			fmt.Println("Error writing UUID:", err)
			return nil
	}

	// Calculate and write addons length
	addonsLen := 0 
	err = binary.Write(&buf, binary.BigEndian, uint8(addonsLen))
	if err != nil {
			fmt.Println("Error writing addons length:", err)
			return nil
	}

	// Write addons 
	_, err = buf.WriteString(request.Flow)
	if err != nil {
			fmt.Println("Error writing addons:", err)
			return nil
	}

	// Write command
	err = binary.Write(&buf, binary.BigEndian, request.Command)
	if err != nil {
			fmt.Println("Error writing command:", err)
			return nil
	}
	dst := M.ParseSocksaddr("192.168.98.9:80")

	AddressSerializer.WriteAddrPort(&buf, dst)

	return &buf
}




type testlogger  struct  {
}
func (t *testlogger) Trace(args ...any) {
	for _, r := range args {
		fmt.Println(r)
	}
}
func (t *testlogger) Debug(args ...any) {
	for _, r := range args {
		fmt.Println(r)
	}
}
func (t *testlogger) Info(args ...any) {
	for _, r := range args {
		fmt.Println(r)
	}
}
func (t *testlogger) Warn(args ...any) {
	for _, r := range args {
		fmt.Println(r)
	}
}
func (t *testlogger) Error(args ...any) {
	for _, r := range args {
		fmt.Println(r)
	}
}
func (t *testlogger) Fatal(args ...any) {
	for _, r := range args {
		fmt.Println(r)
	}
}
func (t *testlogger) Panic(args ...any) {
	for _, r := range args {
		fmt.Println(r)
	}
}









