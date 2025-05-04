package constbot

import (
	"context"

	"go.uber.org/zap"
)


type Botoptions struct {
	Watchman *Watchmanconfig `json:"watchman,omitempty"`

	Dbpath              string                   `json:"db_path,omitempty"`
	UsageDbpath         string                   `json:"usagedb_path,omitempty"`
	TemplatesPath       string                   `json:"templates_path"`
	Bottoken            string                   `json:"bot_token,omitempty"`
	Botmainurl          string                   `json:"bot_mainurl,omitempty"`
	Metadata            *MetadataConf 			 `json:"metadata,omitempty"`
	WebHookServerOption *ServerOption     		 `json:"webhook_server,omitempty"`
	SboxConfPath        string                   `json:"sbox_path,omitempty"`
	LoggerOption 		LoggerOptions   		 `json:"log,omitempty"`

	Logger     *zap.Logger                         `json:"-"`
	Ctx        context.Context                     `json:"-"`
}
type Watchmanconfig struct {
	Delbuffer  int   `json:"del_buffer"` //msg count to buffer before delete
}

type MetadataConf struct {
	//ForceAdd          bool   `json:"forceAdd,omitempty"`
	ChannelID         int64  `json:"channel_id,omitempty"`
	GroupID           int64  `json:"groupd_id,omitempty"`
	BandwidthAvelable string `json:"bandwidth,omitempty"`
	LoginLimit        int16  `json:"login_limit,omitempty"`
	MaxGiftCount 	  int64  `json:"max_gift,omitempty"`
	MaxBuildConf  	  int	 `json:"max_build_conf,omitempty"`
	//Userquota         int32  `json:"userquota,omitempty"`
	//Verifiedcount     int32  `json:"verifiedcount,omitempty"`
	Maxconfigcount    int16  `json:"max_config_count,omitempty"`
	//CheckCount        int32  `json:"checkcount,omitempty"`  // database checked count for exting period
	RefreshRate       int32  `json:"refresh_rate,omitempty"` //rate of db refresh in hours
	BackupRate 		  int     `json:"backup_rate,omitempty"`

	GroupLink  string `json:"group_link,omitempty"`
	Channelink string `json:"channel_link,omitempty"`
	Botlink    string `json:"bot_link,omitempty"`

	GroupName   string `json:"group_name,omitempty"`
	ChannelName string `json:"channel_name,omitempty"`
	BotName     string `json:"bot_name,omitempty"`

	//SudoAdminId int64 `json:"adminId,omitempty"`
	//AllAdmin  []int64 `json:"alladmin,omitempty"`
	SudoAdmin int64   `json:"admin,omitempty"`

	WatchMgbuf int	  `json:"group_maxmg,omitempty"`

	DefaultDomain   string `json:"default_domain,omitempty"`
	DefaultPublicIp string `json:"default_publicip,omitempty"`

	StorePath    string `json:"store_path,omitempty"`
	ConfigFolder string `json:"config_folder,omitempty"`

	HelperInfo HelpCommandInfo `json:"help_cmd,omitempty"`

	InlinePost []InlinePost `json:"inline_posts,omitempty"`
	Langs []string  `json:"allowed_langs,omitempty"`
	DefaultLang string	`json:"default_lang,omitempty"`

	CommonWarnRatio int16  `json:"warn_rate,omitempty"`

	
}

type InlinePost struct {
	Dis   string `json:"description,omitempty"`
	Title string `json:"title,omitempty"`
	Name  string `json:"template_name,omitempty"`
}

type LoggerOptions struct {
	Paths []string `json:"paths,omitempty"`
	Level zap.AtomicLevel `json:"level,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}




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

type HelpCommandInfo struct {
	InfoPageCount     int16 `json:"info_pages"`
	TutorialPageCount int16 `json:"tuto_pages"`
	CommandPageCount  int16 `json:"cmd_pages"`
	BuilderHelp       int16 `json:"builder_pages"`
}