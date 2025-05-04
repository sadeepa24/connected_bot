package controller

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox/conf"
	"github.com/sagernet/sing-box/option"
	"go.uber.org/zap"
)

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

	Inbounds  []conf.Inboud
	Outbounds []conf.Outbound

	inboundasMap  map[int16]conf.Inboud
	outboundasMap map[int16]conf.Outbound

	rawoptions option.Options

	defaultinbound  conf.Inboud
	defaultoutbound conf.Outbound

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
	MaxGiftCount int64
	MaxBuildConf int
	BackupCycle int

	SudoAdmin    int64
	ConfigFolder string

	HelperInfo C.HelpCommandInfo

	inlineposts []C.InlinePost

	CommonWarnRate int16 

	Langs []string

	boxpath string



}


func (m *Metadata) Init(metaconf C.MetadataConf, logger *zap.Logger) error {

	if metaconf.StorePath == "" {
		return errors.New("configs store path not found")
	}
	if metaconf.ConfigFolder == "" {
		return errors.New("config folder path not found")
	}
	m.inlineposts = metaconf.InlinePost
	if !strings.HasSuffix(metaconf.ConfigFolder, "/") {
		metaconf.ConfigFolder = metaconf.ConfigFolder + "/"
	}
	m.ConfigFolder = metaconf.ConfigFolder

	m.storePath = metaconf.StorePath

	m.MaxGiftCount = metaconf.MaxGiftCount
	m.MaxBuildConf = metaconf.MaxBuildConf
	m.CommonQuota = new(atomic.Int64)
	m.VerifiedUserCount = new(atomic.Int32)
	m.Dbusercount = new(atomic.Int32)
	m.Maxconfigcount = metaconf.Maxconfigcount
	m.CheckCount = new(atomic.Int32)
	m.BackupCycle = metaconf.BackupRate
	if m.BackupCycle == 0 {
		m.BackupCycle = 1
	}

	m.HelperInfo = metaconf.HelperInfo
	if metaconf.DefaultLang == "" {
		return errors.New("default lang should be select")
	}

	if len(metaconf.Langs) == 0 {
		metaconf.Langs = append(metaconf.Langs, metaconf.DefaultLang)
	}
	m.Langs = metaconf.Langs

	if metaconf.SudoAdmin == 0 {
		return errors.New("sudo admin not found")
	}
	m.RefreshRate = metaconf.RefreshRate
	if metaconf.CommonWarnRatio == 0 {
		metaconf.CommonWarnRatio = 24 //default for 2 days
	}
	
	m.CommonWarnRate = metaconf.CommonWarnRatio
	m.SudoAdmin = metaconf.SudoAdmin
	return nil
}
func (mn *Metadata) GetWarnRate() int16 {
	return mn.CommonWarnRate/int16(mn.RefreshRate)
}
func (m *Metadata) SboxConfPath() string {
	return m.boxpath
}

func (m *Metadata) DefaultInboud() (conf.Inboud, db.Inbound) {
	return m.defaultinbound, db.Inbound{
		ID:   int16(m.defaultinbound.Id),
		Tag:  m.defaultinbound.Tag,
		Type: m.defaultinbound.Type,
		Name: m.defaultinbound.Name,
		Info: "",
	}
}

func (m *Metadata) Defaultoutboud() (conf.Outbound, db.Outbound) {
	return m.defaultoutbound, db.Outbound{
		ID:   int16(m.defaultoutbound.Id),
		Tag:  m.defaultoutbound.Tag,
		Type: m.defaultoutbound.Type,
		Name: m.defaultoutbound.Name,
		Info: "",
	}
}

func (m *Metadata) Getinbounds() []conf.Inboud {
	return m.Inbounds
}

func (m *Metadata) StorePath() string {
	return m.storePath
}
func (m *Metadata) ConfFolder() string {
	return m.ConfigFolder
}

func (m *Metadata) Getoutbounds() []conf.Outbound {
	return m.Outbounds
}

func (m *Metadata) Getinbound(id int16) (conf.Inboud, bool) {
	in, ok := m.inboundasMap[id]
	return in, ok
}
func (m *Metadata) GetAllinbound() map[int16]conf.Inboud {
	return m.inboundasMap
}


func (m *Metadata) GetinboundList(ids []int16) (map[int16]conf.Inboud) {
	inlist := make(map[int16]conf.Inboud, len(ids))
	for _, id := range ids {
		ins, ok := m.Getinbound(id)
		if !ok {
			continue
		}
		inlist[id] = ins
	}
	return inlist
}


func (m *Metadata) Getoutbound(id int16) (conf.Outbound, bool) {

	in, ok := m.outboundasMap[id]
	return in, ok
}

func (m *Metadata) GetdbInbound(id int16) (db.Inbound, error) {
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

func (m *Metadata) GetdbOutbound(id int16) (db.Outbound, error) {
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

func (m *Metadata) GetInlinePost() []C.InlinePost {	
	return m.inlineposts
}

// func (m *Metadata) UpdateInlinePost(newposts []C.InlinePost) {	
// 	m.postmu.Lock()
// 	defer m.postmu.Unlock()
// 	m.inlineposts = newposts
// }
type Overview struct {
	Mu *sync.RWMutex

	BandwidthAvailable C.Bwidth
	MonthTotal C.Bwidth
	AllTime C.Bwidth


	


	VerifiedUserCount int64
	TotalUser int32
	CUser int64 
	CappedUser int64
	DistributedUser int64
	QuotaForEach C.Bwidth
	Restricte int64
	TempLimitedUser int64
	TotalConfCount int64
	ActiveConfCount int64
	TotalUpdates int64
	MonthLimitedUser int64


	DaysToReset int32
	LastRefresh time.Time

}


func (o *Overview) String() string {
	o.Mu.RLock()
	defer o.Mu.RUnlock()
	return fmt.Sprintf(
		"Overview:\n"+
			"Server Bandwidth: %s\n"+
			"Month Total Usage: %s\n"+
			"All Time Usage: %s\n"+
			"Quota For Each: %s\n\n"+
			"User Who Can Acctualy Use The Config: %d\n"+
			"Verified User Count: %d\n"+
			"Total User: %d\n"+
			"Capped User: %d\n"+
			"Distributed User: %d\n"+
			"Restricted: %d\n"+
			"MonthLimited: %d\n"+
			"Temp Limited User: %d\n\n"+
			"Total Conf Count: %d\n"+
			"Active Conf Count: %d\n\n"+
			"Last Refresh: %s\n" + 
			"Days To Reset: %d\n\n"+
			"Total Update Recived (until last refresh): %d\n\n",
			
		o.BandwidthAvailable.BToString(),
		o.MonthTotal.BToString(),
		o.AllTime.BToString(),
		o.QuotaForEach.BToString(),
		o.CUser,
		o.VerifiedUserCount,
		o.TotalUser,
		o.CappedUser,
		o.DistributedUser,
		o.Restricte,
		o.MonthLimitedUser,
		o.TempLimitedUser,
		o.TotalConfCount,
		o.ActiveConfCount,
		o.LastRefresh.Format(time.RFC3339),
		o.DaysToReset,
		o.TotalUpdates,
	)
}


type DbError struct {
	error
	exit bool
	msg string
}

func (c DbError) UserMsg() string { return c.msg }
func (c DbError) Exit() bool { return c.exit }

var _ C.Error = (*DbError)(nil)