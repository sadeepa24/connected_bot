package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/parser"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"go.uber.org/zap"
)

type ServerOption struct {
	HttpPath          string   `json:"http_path"`
	AllowedUpdates    []string `json:"allowed_updates,omitempty"`
	FullUrl           string   `json:"full_url"`
	Secret            string   `json:"secret"`
	DisableWebhookSet bool     `json:"disable_setwebhook"`
	Custom_Message    string   `json:"req_reject_message"`
	ListenOption	 ListenOption `json:"listen_option"`
}

type ListenOption struct {
	AllowdIPCidr		  []string `json:"allowd_cidr"`
	ConnRejectMessage     string   `json:"reject_message"`
	ServerName        string   `json:"server_name"`
	Cert              string   `json:"cert"`
	Key               string   `json:"key"`
	Addr              string   `json:"addr"`
}

type Webhookserver struct {
	listner net.Listener
	http.Server
	Handler http.Handler
	ctx     context.Context
	addr    net.Addr
	logger  *zap.Logger

	secret string

	AllowdObs         []string
	FullUrl           string
	DIsableWebhookSet bool
}

type BotHandler struct {
	logger *zap.Logger
	path   string
	parser parser.Parserwrap
	secreatToken string
	CustomMessage []byte
}

func New(ctx context.Context, srvopt *ServerOption, parser parser.Parserwrap, logger *zap.Logger) *Webhookserver {
	if srvopt == nil {
		return nil
	}
	ls, err := newwhls(srvopt.ListenOption)
	if err != nil {
		logger.Error("Webhook Server Listner Creating Failed ", zap.Error(err))
		return nil
	}
	if srvopt.Custom_Message == "" {
		srvopt.Custom_Message = "server cannot prosess this request"
	}
	wsh := &Webhookserver{
		listner: ls,
		logger:  logger,
		ctx:     ctx,
		Server: http.Server{
			Addr: ls.Addr().String(),
			//ErrorLog: ,
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
			Handler: &BotHandler{
				logger: logger,
				path:   srvopt.HttpPath,
				parser: parser,
				secreatToken: srvopt.Secret,
				CustomMessage: []byte(srvopt.Custom_Message),
			},
		},
		FullUrl:           srvopt.FullUrl,
		AllowdObs:         srvopt.AllowedUpdates,
		secret:            srvopt.Secret,
		addr:              ls.Addr(),
		DIsableWebhookSet: srvopt.DisableWebhookSet,
	}
	return wsh
}

func (w *Webhookserver) Start(botapi botapi.BotAPI, errchan chan error) error {
	if !w.DIsableWebhookSet {
		if err := botapi.SetWebhook(w.FullUrl, w.secret, "", w.AllowdObs); err != nil {
			errchan <- err
			return err
		}
		w.logger.Debug("webhook setting succsess")
	}
	errchan <- nil
	w.logger.Info("webhook listener started on " + w.Addr)
	return w.Server.Serve(w.listner)

}

func (w *Webhookserver) Close() error {
	return w.Server.Close()
}

func (w *BotHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	w.logger.Debug("request recived from " + req.RemoteAddr)
	if req.Method != "POST" {
		w.logger.Warn("Unsupported Http Method " + req.Method + " Recived from " + req.RemoteAddr)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write(w.CustomMessage)
		return
	}
	if req.URL.Path != w.path {
		w.logger.Warn("MissMatched Requests Path " + req.URL.Path + " from " + req.RemoteAddr)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write(w.CustomMessage)
		return
	}
	if req.Header.Get("X-Telegram-Bot-Api-Secret-Token") != w.secreatToken {
		w.logger.Warn("Auth Token Not Found Req from " + req.RemoteAddr)
		writer.WriteHeader(http.StatusForbidden)
		writer.Write(w.CustomMessage)
		return
	}
	var update tgbotapi.Update
    decoder := json.NewDecoder(req.Body)
    if err := decoder.Decode(&update); err != nil {
		w.logger.Error("Failed to decode JSON - ", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
        return
    }
	go w.parsertimetgmsg(&update)
	//go w.parser.Parse(msg)
	writer.WriteHeader(http.StatusOK)
}

// remove this after testings
func (w *BotHandler) parsertimetgmsg(tg *tgbotapi.Update) {
	st := time.Now()
	if err := w.parser.Parse(tg); err != nil {
		w.logger.Error("updated errored",  zap.Error(err))
	}
	w.logger.Info("elpsed time for processing request ⏳ " + time.Since(st).String())
}

type webhookls struct {
	net.Listener
	allowdip RangeCheck
	tlsconfig *tls.Config
	rejectMessage []byte
}

func newwhls(lsopts ListenOption) (*webhookls, error) {
	tcpAddr,err := net.ResolveTCPAddr("tcp", lsopts.Addr)
	if err != nil {
		return nil, errors.New("listen address fault")
	}
	ls, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	var rangeChecker RangeCheck
	if len(lsopts.AllowdIPCidr) > 0 {
		rangeChecker, err = NewCIDRRange(lsopts.AllowdIPCidr)
	}
	var tlsconf *tls.Config
	if lsopts.Cert != "" && lsopts.Key != "" {
		cert, err := tls.LoadX509KeyPair(lsopts.Cert, lsopts.Key)
		if err != nil {
			return nil, err
		}
		tlsconf = &tls.Config{
			ServerName:   lsopts.ServerName,
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS10,
			InsecureSkipVerify: true,
		}
	}

	if lsopts.ConnRejectMessage == "" {
		lsopts.ConnRejectMessage = "connection reject due to security reason"
	}

	return &webhookls{
		Listener: ls,
		rejectMessage: []byte(lsopts.ConnRejectMessage),
		tlsconfig: tlsconf,
		allowdip: rangeChecker,
	}, err
}

func (w *webhookls) Accept() (net.Conn, error) {
	conn, err := w.Listener.Accept()
	if err != nil {
		return nil, err
	}
	if w.allowdip != nil {
		if !w.allowdip.Contains(net.ParseIP(GetIP(conn))) {
			conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
			conn.Write(w.rejectMessage)
			conn.Close()
			return nil, net.UnknownNetworkError("unknown remote addr " + conn.RemoteAddr().Network(), )
		}
	}
	if w.tlsconfig != nil {
		conn = tls.Server(conn, w.tlsconfig)
	}
	return conn, err
}

func GetIP(conn net.Conn) string {
	//all connections are tcp so it's allright
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		return addr.IP.String()
	}
	return ""
}