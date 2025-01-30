package db

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	C "github.com/sadeepa24/connected_bot/constbot"
)

type User struct {
	CheckID  uint
	TgID     int64          `gorm:"primaryKey;column:tg_id"`
	Name     string         `gorm:"type:varchar(100)"`
	Username sql.NullString `gorm:"type:varchar(100);column:username"`
	DeviceID sql.NullString `gorm:"type:varchar(300);column:device_id"`
	Lang     string         `gorm:"type:varchar(50);column:lang"`

	IsTgPremium       bool `gorm:"column:is_tg_premium"`
	IsInChannel       bool `gorm:"column:is_in_channel"`
	IsInGroup         bool `gorm:"column:is_in_group"`
	IsRemoved         bool `gorm:"column:is_removed"` //common for group and channel
	Restricted 		  bool `gorm:"column:restricted"` // admin can restrict users
	GroupBanned       bool `gorm:"column:group_banned"`
	ChannelBanned     bool `gorm:"column:channel_banned"`
	IsVipUser         bool `gorm:"column:is_vip_user"`
	IsBotStarted      bool `gorm:"column:is_bot_started"`
	//IsAdmin           bool `gorm:"column:is_admin"`
	IsDistributedUser bool `gorm:"column:is_dis_user"`
	IsCapped          bool `gorm:"column:is_capped"`
	IsMonthLimited    bool `gorm:"column:is_month_limited"`
	RecheckVerificity bool `gorm:"column:recheck_verificity"`

	Points int64

	CalculatedQuota C.Bwidth // This value includes Main User quota which is calculated on watchman + Giftquota
	AdditionalQuota C.Bwidth `gorm:"column:additional_quota"` // this is static does not reset, value always in byte
	GiftQuota       C.Bwidth // this value can be +,-
	CappedQuota     C.Bwidth `gorm:"column:capped_quota"`
	UsedQuota       C.Bwidth // current total quota used by the user
	SavedQuota      C.Bwidth //this value used for when a user over use month usage this value store next months savings from him    (his quota - fake usage)

	MonthUsage       C.Bwidth `gorm:"column:month_usage"` //Usage of current Month will reset with end of month
	AlltimeUsage     C.Bwidth `gorm:"column:all_time_usage"`
	AddtionalConfig  int16    `gorm:"column:max_config_count"`
	ConfigCount      int16    `gorm:"column:config_count"`
	DeletedConfCount int16    `gorm:"column:deleted_conf_count"`

	WebToken sql.NullString `gorm:"type:varchar(200);column:web_token"`
	Configs  []Config       `gorm:"foreignKey:UserID"`
	//Gifts 		[]Gift 			`gorm:"foreignKey:UserID"`

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

// retuns sum of addtional + gift +
func (u *User) GenaralQuotSum() C.Bwidth {
	if u.IsCapped {
		return u.CappedQuota
	}
	return u.CalculatedQuota
}

func (u *User) Iscaptimeover() bool {
	return u.Captime.AddDate(0, 0, 30).Compare(time.Now()) <= 0
}

type Config struct {
	Id         int64 `gorm:"primaryKey"`
	Name       string
	UUID       uuid.UUID
	Type       string
	Password   string
	Active     bool
	UserID     int64 `gorm:"not null;column:user_id"`
	InboundID  int16 `gorm:"not null"`
	OutboundID int16 `gorm:"not null"`

	Inbound  Inbound  `gorm:"foreignKey:ID"`
	Outbound Outbound `gorm:"foreignKey:ID"`

	Usage    C.Bwidth // total usage for this month as byte
	Download C.Bwidth // total download for this month as byte
	Upload   C.Bwidth // total uploads for this month as byte
	Quota    C.Bwidth // changes every day when according to groups user

	LoginLimit int16
	//DeletedAt 		gorm.DeletedAt `gorm:"index"`

}

type UsageHistory struct {
	ID       int64 `gorm:"primaryKey"`
	Name     string
	Download C.Bwidth
	Upload   C.Bwidth
	UserID   int64
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
	ID          int64
	UserID      int64
	ElpsedPoint int64
	Resong      string
}

type Gift struct {
	ID      int64 `gorm:"primaryKey"`
	Sender  int64
	Reciver int64

	SendValid   bool // used by watchman
	ReciveValid bool //used by watchman when prosessing batch records in preprosess
	Bandwidth   C.Bwidth
	Date        time.Time
	Valid       bool // used by watchman

	ComQuota C.Bwidth //common quota exited when creating the record
	//DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type SboxConfigs struct {
	ID       int64 `gorm:"primaryKey"`
	Name     string
	UserID   int64
	ConfPath string
}

func (u *Gift) Isgifttimeover() bool {
	return u.Date.AddDate(0, 0, 30).Compare(time.Now()) <= 0
}

// All thinsgs Downthere will store in ram until program kill
type Admin struct {
	Id int64
	//DeletedAt gorm.DeletedAt `gorm:"index"`

}

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

type Adminchat struct {
	Id       int64 `gorm:"primaryKey"`
	Name     string
	Type     string
	UserName string
	//DeletedAt gorm.DeletedAt `gorm:"index"`

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
