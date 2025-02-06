package controller

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sadeepa24/connected_bot/update/bottype"
)

type MetadataConf struct {
	//ForceAdd          bool   `json:"forceAdd,omitempty"`
	ChannelID         int64  `json:"channel_id,omitempty"`
	GroupID           int64  `json:"groupd_id,omitempty"`
	BandwidthAvelable string `json:"bandwidth,omitempty"`
	LoginLimit        int16  `json:"login_limit,omitempty"`
	//Userquota         int32  `json:"userquota,omitempty"`
	//Verifiedcount     int32  `json:"verifiedcount,omitempty"`
	Maxconfigcount    int16  `json:"max_config_count,omitempty"`
	//CheckCount        int32  `json:"checkcount,omitempty"`  // database checked count for exting period
	RefreshRate       int32  `json:"refresh_rate,omitempty"` //rate of db refresh in hours

	GroupLink  string `json:"group_link,omitempty"`
	Channelink string `json:"channel_link,omitempty"`
	Botlink    string `json:"bot_link,omitempty"`

	GroupName   string `json:"group_name,omitempty"`
	ChannelName string `json:"channel_name,omitempty"`
	BotName     string `json:"bot_name,omitempty"`

	//SudoAdminId int64 `json:"adminId,omitempty"`
	//AllAdmin  []int64 `json:"alladmin,omitempty"`
	SudoAdmin int64   `json:"admin,omitempty"`

	WatchMgbuf int

	DefaultDomain   string `json:"default_domain,omitempty"`
	DefaultPublicIp string `json:"default_publicip,omitempty"`

	StorePath    string `json:"store_path,omitempty"`
	ConfigFolder string `json:"config_folder,omitempty"`

	HelperInfo bottype.HelpCommandInfo `json:"help_cmd,omitempty"`

	InlinePost []string `json:"inline_posts,omitempty"`
}

type Metadata struct {
	ChannelId int64
	GroupID   int64
	//AdminList map[int64]string
	storePath string
	//UserQuota        *atomic.Int64 // Last calculated userquota

	//UserQuota		C.Bwidth // Last calculated userquota should use with rwmutext
	CommonQuota *atomic.Int64 // This is commonquota for all user userquota may vary to their settings
	//Channelusercount *atomic.Int32
	//Groupusercount   *atomic.Int32
	VerifiedUserCount *atomic.Int32
	Maxconfigcount    int16

	Dbusercount       *atomic.Int32
	LoginLimit        int32
	BandwidthAvelable C.Bwidth

	Inbounds  []sbox.Inboud
	Outbounds []sbox.Outbound

	inboundasMap  map[int]sbox.Inboud
	outboundasMap map[int]sbox.Outbound

	defaultinbound  sbox.Inboud
	defaultoutbound sbox.Outbound

	CheckCount  *atomic.Int32
	ResetCount  int32 //static value that db should reset when checkcount eqal to this
	RefreshRate int32

	GroupLink  string
	Channelink string
	Botlink    string

	GroupName   string
	ChannelName string
	BotName     string

	DefaultDomain string
	DefaultPubip  string

	MaxRecurtion int

	SudoAdmin    int64
	ConfigFolder string

	HelperInfo bottype.HelpCommandInfo

	inlineposts []string

	//mu *sync.RWMutex

}

// func (m *Metadata) Lock() {
// 	m.mu.Lock()
// }
// func (m *Metadata) Unlock() {
// 	m.mu.Unlock()
// }

func (m *Metadata) Init(metaconf MetadataConf) error {

	if metaconf.StorePath == "" {
		return errors.New("configs store path not found")
	}
	if metaconf.ConfigFolder == "" {
		return errors.New("config folder path not found")
	}
	m.inlineposts = metaconf.InlinePost
	m.ConfigFolder = metaconf.ConfigFolder

	m.storePath = metaconf.StorePath

	m.CommonQuota = new(atomic.Int64)
	m.VerifiedUserCount = new(atomic.Int32)
	m.Dbusercount = new(atomic.Int32)
	m.Maxconfigcount = metaconf.Maxconfigcount
	m.CheckCount = new(atomic.Int32)

	m.HelperInfo = metaconf.HelperInfo

	m.MaxRecurtion = 20 //TODO: change this


	if metaconf.SudoAdmin == 0 {
		return errors.New("sudo admin not found")
	}
	m.SudoAdmin = metaconf.SudoAdmin
	return nil
}

func (m *Metadata) DefaultInboud() (sbox.Inboud, db.Inbound) {
	return m.defaultinbound, db.Inbound{
		ID:   int16(m.defaultinbound.Id),
		Tag:  m.defaultinbound.Tag,
		Type: m.defaultinbound.Type,
		Name: m.defaultinbound.Name,
		Info: "",
	}
}

func (m *Metadata) Defaultoutboud() (sbox.Outbound, db.Outbound) {
	return m.defaultoutbound, db.Outbound{
		ID:   int16(m.defaultoutbound.Id),
		Tag:  m.defaultoutbound.Tag,
		Type: m.defaultoutbound.Type,
		Name: m.defaultoutbound.Name,
		Info: "",
	}
}

func (m *Metadata) Getinbounds() []sbox.Inboud {
	return m.Inbounds
}

func (m *Metadata) StorePath() string {
	return m.storePath
}
func (m *Metadata) ConfFolder() string {
	return m.ConfigFolder
}

func (m *Metadata) Getoutbounds() []sbox.Outbound {
	return m.Outbounds
}

func (m *Metadata) Getinbound(id int) (sbox.Inboud, bool) {

	in, ok := m.inboundasMap[id]
	return in, ok
}
func (m *Metadata) Getoutbound(id int) (sbox.Outbound, bool) {

	in, ok := m.outboundasMap[id]
	return in, ok
}

func (m *Metadata) GetdbInbound(id int) (db.Inbound, error) {
	inbound, ok := m.inboundasMap[id]
	if !ok {
		return db.Inbound{}, C.ErrInboundNotFound
	}
	return db.Inbound{
		ID:   int16(inbound.Id),
		Tag:  inbound.Tag,
		Type: inbound.Type,
		Name: inbound.Name,
		Info: "",
	}, nil
}

func (m *Metadata) GetdbOutbound(id int) (db.Outbound, error) {
	outbound, ok := m.outboundasMap[id]
	if !ok {
		return db.Outbound{}, C.ErrOutboundNotFound
	}
	return db.Outbound{
		ID:   int16(outbound.Id),
		Tag:  outbound.Tag,
		Type: outbound.Type,
		Name: outbound.Name,
		Info: "",
	}, nil
}

func (m *Metadata) GetInlinePost() []string {
	return m.inlineposts
}


type Overview struct {
	Mu *sync.RWMutex

	BandwidthAvailable C.Bwidth
	//DownLoad C.Bwidth
	//Upload C.Bwidth
	MonthTotal C.Bwidth
	AllTime C.Bwidth


	


	VerifiedUserCount int64
	TotalUser int32
	CappedUser int64
	DistributedUser int64
	QuotaForEach C.Bwidth
	Restricted int64





	LastRefresh time.Time


	Error error
}