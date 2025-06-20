package db

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	C "github.com/sadeepa24/connected_bot/constbot"
	sbConf "github.com/sadeepa24/connected_bot/sbox/conf"
)

type User struct {
	CheckID  uint
	TgID     int64          `gorm:"primaryKey;column:tg_id"`
	Name     string         `gorm:"type:varchar(100)"`
	Username string 		`gorm:"type:varchar(40);column:username"`
	Lang     string         `gorm:"type:varchar(50);column:lang"`

	IsTgPremium       bool `gorm:"column:is_tg_premium"`
	IsInChannel       bool `gorm:"column:is_in_channel"`
	IsInGroup         bool `gorm:"column:is_in_group"`
	IsRemoved         bool `gorm:"column:is_removed"` //common for group and channel
	//IsPaused		  bool	`gorm:"column:is_paused"` //FIXME: add this feture later and remove monthlimitation
	Restricted 		  bool `gorm:"column:restricted"` // admin can restrict users
	GroupBanned       bool `gorm:"column:group_banned"`
	ChannelBanned     bool `gorm:"column:channel_banned"`
	IsPaused		  bool	`gorm:"column:paused"`
	//IsVipUser         bool `gorm:"column:is_vip_user"`
	IsBotStarted      bool `gorm:"column:is_bot_started"`
	//IsAdmin           bool `gorm:"column:is_admin"`
	IsDistributedUser bool `gorm:"column:is_dis_user"`
	IsCapped          bool `gorm:"index;column:is_capped"`
	IsMonthLimited    bool `gorm:"column:is_month_limited"`
	RecheckVerificity bool `gorm:"column:recheck_verificity"`
	CapDays 		  int32 `gorm:"column:cap_days"`

	Points int64

	CalculatedQuota C.Bwidth `gorm:"index"`// This value includes Main User quota which is calculated on watchman + Giftquota
	AdditionalQuota C.Bwidth `gorm:"index;column:additional_quota"` // this is static does not reset, value always in byte (this value does not use yet in codebase may be future)
	GiftQuota       C.Bwidth `gorm:"index"` // this value can be +,-
	CappedQuota     C.Bwidth `gorm:"index;column:capped_quota"`
	UsedQuota       C.Bwidth // current total quota used by the user
	//SavedQuota      C.Bwidth //this value used for when a user over use month usage this value store next months savings from him    (his quota - fake usage)

	MonthUsage       C.Bwidth `gorm:"index;column:month_usage"` //Usage of current Month will reset with end of month
	AlltimeUsage     C.Bwidth `gorm:"index;column:all_time_usage"`
	//AddtionalConfig  int16    `gorm:"column:max_config_count"`
	ConfigCount      int16    `gorm:"column:config_count"`
	DeletedConfCount int16    `gorm:"column:deleted_conf_count"`
	EmptyCycle		 int16    `gorm:"column:empty_cycle"`
	Templimited 	 bool 	  `gorm:"column:temp_limited"`
	WarnRatio 		 int16    `gorm:"column:warn_ratio"`
				

	//WebToken sql.NullString `gorm:"type:varchar(200);column:web_token"`
	Configs  []Config       `gorm:"foreignKey:UserID"`
	SentGifts     []Gift `gorm:"foreignKey:Sender"`
	ReceivedGifts []Gift `gorm:"foreignKey:Reciver"`
	
	// Gifts 		[]Gift 			`gorm:"foreignKey:UserID"`

	Captime time.Time
	//Gifttime  time.Time
	Joined    time.Time
	LeaveTime time.Time `gorm:"column:leave_time"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) AddPoint(n int64) {
	u.Points = u.Points + n
}

// returns
func (u *User) GetCalculatedQuota() C.Bwidth {
	if u.IsCapped {
		return u.CappedQuota
	}
	return u.CalculatedQuota
}

func (u User) String() string {
	return fmt.Sprintf(
		"User{TgID: %d, Name: %s, Username: %s, Points: %d, Lang: %s, "+
			"CalculatedQuota: %s, AdditionalQuota: %s, GiftQuota: %s, CappedQuota: %s, UsedQuota: %s, "+
			"MonthUsage: %s, AlltimeUsage: %s}",
		u.TgID, u.Name, u.Username, u.Points, u.Lang,
		u.CalculatedQuota.BToString(), u.AdditionalQuota.BToString(), u.GiftQuota.BToString(),
		u.CappedQuota.BToString(), u.UsedQuota.BToString(),
		u.MonthUsage.BToString(), u.AlltimeUsage.BToString(),
	)
}


func (u *User) Iscaptimeover(days int) bool {
	return u.Captime.AddDate(0, 0, days).Compare(time.Now()) <= 0
}
func (u *User) Verified() bool {
	return u.IsInChannel && u.IsInGroup
}

func (u *User) CanUse() bool {
	return !(u.Restricted || u.IsDistributedUser || u.IsMonthLimited  || u.IsPaused || u.Templimited) 
}


type Config struct {
	Id         int64 `gorm:"primaryKey"`
	Name       string
	UUID       string `gorm:"not null;uniqueIndex"` //common for all vless & vmess inbound
	Password   string // common for all trojan, tuic, shadowsocks (everything which use a password)
	Active     bool
	UserID     int64 `gorm:"index;not null;column:user_id"`
	
	//InboundID  int16 `gorm:"not null"`
	
	OutboundID int16 `gorm:"not null"`
	InboundIds InIds `gorm:"type:blob"`
	Usage    C.Bwidth // total usage for this month as byte
	Download C.Bwidth // total download for this month as byte
	Upload   C.Bwidth // total uploads for this month as byte
	Quota    C.Bwidth // changes every day when according to groups user

	LoginLimit int16
	CreatedAt time.Time
	//DeletedAt 		gorm.DeletedAt `gorm:"index"`

}

type InIds []int16 

func (s InIds) Value() (driver.Value, error) {
	val := make([]byte, len(s)*2)
	for i, id := range s {
		binary.LittleEndian.PutUint16(val[i*2:], uint16(id))
	}
	return val, nil
}

func (s *InIds) Scan(value interface{}) error {
	if value == nil {
		*s = []int16{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan InIds: expected []byte, got something else")
	}
	if len(bytes)%2 != 0 {
		return errors.New("invalid byte slice length: must be a multiple of 2")
	}
	var ids []int16
	for i := 0; i < len(bytes); i += 2 {
		id := int16(binary.LittleEndian.Uint16(bytes[i:]))
		ids = append(ids, id)
	}
	*s = ids
	return nil
}




func (c *Config) LeftQuota() C.Bwidth {
	return (c.Quota - c.Usage)
}

func (c *Config) ExportUrlLink(in sbConf.Inboud, info *sbConf.ExportInfo) string {
	if info == nil {
		info = &sbConf.ExportInfo{
			Host: "connectedbot",
			Sni: "connectedbot",
			Server: in.Domain,
		}
	}

	if info.Server == "" {
		info.Server = in.Domain
	}
	switch in.Type {
	case "vless", "trojan":
		passoruuid := c.UUID
		if in.Type == "trojan" {
			passoruuid = c.Password
		}
		url :=  in.Type+"://"+ passoruuid+ "@" + info.Server + ":" + strconv.Itoa(in.Port()) + "?encryption=none&"
		if in.TransPortType != "" {
			switch in.TransPortType {
			case "ws", "http", "httpupgrade":
				url += ("path="+in.TransPortPath+ "&host=" + info.Host+ "&")

			case "grpc":
				url += ("serviceName="+in.TransPortPath+ "&authority=" + info.Host+ "&")
			}
			url += ("type=" + in.TransPortType +  "&")
		}
		if in.Tlsenabled {
			url += ("security=tls&sni="+ info.Sni +"&")
		}
		url,_ = strings.CutSuffix(url, "&")
		url += ("#"+ c.Name)
		return url
	
	// This won't be used because this VPN protocol is not added as a supported inbound 
	// due to increased time complexity associated with user count.
	case "vmess":
		tlsinfo := "none"
		if in.Tlsenabled {
			tlsinfo = "tls"
		}
		url := fmt.Sprintf(
			`
		{
		  "add": "%s",
		  "aid": "%d",
		  "scy": "auto",
		  "host": "%s",
		  "id": "%s",
		  "net": "%s",
		  "path": "%s",
		  "port": "%d",
		  "ps": "%s",
		  "tls": "%s",
		  "sni": "%s",
		  "type": "none",
		  "v": "2"
		}
			`, info.Server, 0, info.Host, c.UUID, in.TransPortType, in.TransPortPath, in.Port(), c.Name, tlsinfo, info.Sni,
		)
		return "vmess://"+ base64.RawStdEncoding.EncodeToString([]byte(url))

	default:
		return "currently proto " + in.Type + "isn't support for exporting"
	}
}

func (u *Config) GetuniqName() string {
	return strconv.Itoa(int(u.Id)) + strings.TrimSpace(u.Name) + strconv.Itoa(int(u.UserID))
}

func (c *Config) UpdateUsages(status sbConf.Sboxstatus) {
	c.Usage += status.Download + status.Upload
	c.Download += status.Download
	c.Upload += status.Upload
}

func (c *Config) GetUUID() uuid.UUID {
	return uuid.FromStringOrNil(c.UUID) //this won't return nil because db's uuid verified before store them in db
}

type UsageHistory struct {
	ID       int64 `gorm:"primaryKey"`
	Name     string
	Download C.Bwidth
	Upload   C.Bwidth
	UserID   int64 `gorm:"index"`
	Usage    C.Bwidth
	Date     time.Time
	ConfigID int64
}

type GiftLog struct {
	ID        int64 `gorm:"primaryKey"`
	SendID    int64
	RecivedID int64
	Bandwidth C.Bwidth
	Date      time.Time
}

type PointLog struct {
	ID          int64 `gorm:"primaryKey"`
	UserID      int64 `gorm:"index"`
	ElpsedPoint int64
	Resong      string
}

type Gift struct {
	ID          int64    `gorm:"primaryKey"`
	Sender      int64    `gorm:"index:idx_sender_reciver"`
	Reciver     int64    `gorm:"index:idx_sender_reciver"`
	// SendValid   bool     `gorm:"index:idx_valid_flags"`
	// ReciveValid bool     `gorm:"index:idx_valid_flags"`
	Bandwidth   C.Bwidth
	Date        time.Time
	//ComQuota    C.Bwidth // Main common quota which was exist when gift was created
	//DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type SboxConfigs struct {
	ID       int64 `gorm:"primaryKey"`
	Name     string
	UserID   int64 `gorm:"index"`
	ConfPath string
}

type RestrictUser struct {
	ID int64 `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100)"`
	Reason string `gorm:"type:varchar(250)"`
	Username string `gorm:"type:varchar(40);column:username"`
}

func (u *Gift) Isgifttimeover() bool {
	return u.Date.AddDate(0, 0, 30).Compare(time.Now()) <= 0
}

// All thinsgs Downthere will store in ram until program kill

type Inbound struct {
	ID   int16  `gorm:"primaryKey"`
	Tag  string `gorm:"type:varchar(100)"`
	Name string
	Type string
	Info string

	//DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Outbound struct {
	ID   int16  `gorm:"primaryKey"`
	Tag  string `gorm:"type:varchar(100)"`
	Id   int64
	Name string
	Type string
	Info string
}


type Metadata struct {
	Id                int32
	CommonQuota       C.Bwidth //common quota for all use which is changing over many condition like verified user count capped user count capped total addtional, how ever actual user quota calculated based on this quota
	Maxconfigcount    int16
	ChannelId         int64
	GroupID           int64
	Channelusercount  int64
	Groupusercount    int64
	VerifiedUserCount int64

	Dbusercount       int64
	LoginLimit        int32
	BandwidthAvelable C.Bwidth

	CheckCount  int32 // current check count
	ResetCount  int32 //neded ChecCounts for reset db
	RefreshRate int32 //rate of refreshing in hours

	PublicDomain string
	PublicIp string
	TotalUpdates int64

	CommonWarnRatio int16
	TotalConfigCount int64
}

type Reffral struct {
	UserId    int64 `gorm:"primaryKey"`
	OwnerID   int64
	CreatedAt time.Time
	Expired   bool
}

type Event struct {
	ID     int64 `gorm:"primaryKey"`
	Name   string
	UserId int64
}


type Overview struct {
	Mu *sync.RWMutex `gorm:"-"`

	BandwidthAvailable C.Bwidth
	MonthTotal C.Bwidth
	AllTime C.Bwidth
	BandwidthAddtional C.Bwidth

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

	DaysToReset int32 `gorm:"-"`
	LastRefresh time.Time

	UpdateTime time.Time

}


func (o *Overview) String() string {
	o.Mu.RLock()
	defer o.Mu.RUnlock()
	return fmt.Sprintf(
		"Overview:\n"+
			"Server Bandwidth: %s\n"+
			"Month Total Usage: %s\n"+
			"All Time Usage: %s\n"+
			"For Addtional Quota: %s\n"+
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
		o.BandwidthAddtional.BToString(),
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