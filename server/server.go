package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/parser"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"go.uber.org/zap"
)

type ServerOption struct {
	Addr              string   `json:"addr"`
	Cert              string   `json:"cert"`
	Key               string   `json:"key"`
	ServerName        string   `json:"server_name"`
	HttpPath          string   `json:"http_path"`
	AllowedUpdates    []string `json:"allowed_updates,omitempty"`
	FullUrl           string   `json:"full_url"`
	Secret            string   `json:"secret"`
	DisableWebhookSet bool     `json:"disable_setwebhook"`
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
}

func New(ctx context.Context, srvopt *ServerOption, parser parser.Parserwrap, logger *zap.Logger) *Webhookserver {
	if srvopt == nil {
		return nil
	}

	var tlsconf *tls.Config

	if srvopt.Cert != "" && srvopt.Key != "" {
		cert, err := tls.LoadX509KeyPair(srvopt.Cert, srvopt.Key)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		tlsconf = &tls.Config{
			ServerName:   srvopt.ServerName,
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS10,
			InsecureSkipVerify: true,
		}
	}
	ls, err := newwhls(srvopt.Addr, tlsconf)

	if err != nil {
		return nil
	}

	wsh := &Webhookserver{
		listner: ls,
		logger:  logger,
		ctx:     ctx,
		Server: http.Server{
			Addr: ls.Addr().String(),
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
			Handler: &BotHandler{
				logger: logger,
				path:   srvopt.HttpPath,
				parser: parser,
				secreatToken: srvopt.Secret,
				
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
	w.logger.Info("webhook listener started on " + w.Addr)
	if !w.DIsableWebhookSet {
		if err := botapi.SetWebhook(w.FullUrl, w.secret, "", w.AllowdObs); err != nil {
			errchan <- err
			return err
		}
		w.logger.Debug("webhook setting succsess")
	}
	errchan <- nil
	return w.Server.Serve(w.listner)

}

func (w *Webhookserver) Close() error {
	return w.Server.Close()

}

func (w *BotHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	w.logger.Debug("request recived from " + req.RemoteAddr)
	if req.Method != "POST" {
		w.logger.Warn("Unsupported Http Method Recived " + req.Method + "from " + req.RemoteAddr)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte("පලයන් හුත්තෝ යන්න. එනවා මෙතන රෙද්දක් කරන්න, පකයා"))
		return
	}

	if req.URL.Path != w.path {
		w.logger.Warn("MissMatched Requests Path " + req.URL.Path + "from " + req.RemoteAddr)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte("පලයන් හුත්තෝ යන්න. එනවා මෙතන රෙද්දක් කරන්න, පකයා"))
		return
	}
	if req.Header.Get("X-Telegram-Bot-Api-Secret-Token") != w.secreatToken {
		w.logger.Warn("Auth Token Not Found Req from " + req.RemoteAddr)
		writer.WriteHeader(http.StatusForbidden)
		writer.Write([]byte("පලයන් හුත්තෝ යන්න. එනවා මෙතන රෙද්දක් කරන්න, පකයා"))
		return
	}

	data := make([]byte, req.ContentLength)
	var read int
	
	for read < len(data) {
		n, err := req.Body.Read(data[read:])
		read += n
		if err != nil {
			if err == io.EOF {
				break
			}
			w.logger.Error("Error reading request body: " +  err.Error())
			break
		}
		
	}
	
	if read < len(data) {
		w.logger.Debug("Incomplete read, expected:", zap.Int("data len", len(data)), zap.Int("but got", read))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	w.logger.Debug("Read: ", zap.Int("read", read), zap.Int("data len", len(data)))
	req.Body.Close()
	
	
	
	var msg = &tgbotapi.Update{}
	if err := json.Unmarshal(data, msg); err != nil {
		w.logger.Error("request json body unmarshal failed err - " + err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	//TODO: //

	go w.parsertimetgmsg(msg)
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
	ls net.Listener
}

func newwhls(addres string, tlsconfig *tls.Config) (*webhookls, error) {
	ls, err := net.Listen("tcp", addres)
	if err != nil {
		return nil, err
	}
	if tlsconfig != nil {
		ls = tls.NewListener(ls, tlsconfig)
	}
	return &webhookls{
		ls: ls,
	}, nil
}

func (w *webhookls) Accept() (net.Conn, error) {
	return w.ls.Accept()
}

func (w *webhookls) Close() error {
	return w.ls.Close()
}

func (w *webhookls) Addr() net.Addr {
	return w.ls.Addr()
}
