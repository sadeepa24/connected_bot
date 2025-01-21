package connected

import (
	"context"

	"github.com/sadeepa24/connected_bot/controller"
	option "github.com/sadeepa24/connected_bot/sbox_option/v1"
	"github.com/sadeepa24/connected_bot/server"
	"github.com/sadeepa24/connected_bot/watchman"
	"go.uber.org/zap"
)

type Botoptions struct {
	Watchman *watchman.Watchmanconfig `json:"watchman,omitempty"`

	Dbpath              string                  `json:"dbpath,omitempty"`
	TemplatesPath       string                  `json:"template_path"`
	Bottoken            string                  `json:"bot_token,omitempty"`
	Botmainurl          string                  `json:"bot_mainurl,omitempty"`
	Metadata            *controller.MetadataConf `json:"metadata,omitempty"`
	WebHookServerOption *server.ServerOption    `json:"webhook_server,omitempty"`
	MGdelBuffer         int                     `json:"delete_buffer,omitempty"`
	SboxConfPath        string                  `json:"sbox_path,omitempty"`
	LoggerOption 		LoggerOptions   		`json:"log,omitempty"`
	//MessageTempPath string                   `json:"template_path,omitempty"`

	Logger     *zap.Logger                         `json:"-"`
	Ctx        context.Context                     `json:"-"`
	//Templates  map[string]map[string]botapi.MgItem `json:"-"`
	Sboxoption option.Options                      `json:"-"`
}

type LoggerOptions struct {
	Paths []string `json:"path,omitempty"`
	Level string `json:"level,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}