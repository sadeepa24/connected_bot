package bottype

import (
	"fmt"

	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
)

// Standar Userfor all aplication
type User struct {
	Id     int64
	Tguser *tgbotapi.User
	*db.User
	Newuser bool

	chinfo string
}

// type User *db.User
//type User = *db.User

func Newuser(tguser *tgbotapi.User, dbuser *db.User) *User {
	MUser := User{
		Id:      tguser.ID,
		Tguser:  tguser,
		Newuser: false,
		User:    dbuser,
		chinfo: "",
	}
	return &MUser
}

// check if user is admin of bot not groups

func (u *User) Info() string { 
	if u == nil {
		return ""
	}
	if u.chinfo != "" {
		return u.chinfo
	}
	if u.User != nil {
		u.chinfo = fmt.Sprintf("user [%s] tg_id [%d] username [%s]", u.User.Name, u.User.TgID, u.User.Username.String )
	}
	return u.chinfo

 }

func (u *User) Getdbuser() *db.User { return u.User }

func (u *User) Isverified() bool { return u.IsInChannel && u.IsInGroup }

func (u *User) IsremovedUser() bool { return u.IsRemoved }

func (u *User) IsnewUser() bool { return u.Newuser }

func (u *User) Isbotstarted() bool { return u.IsBotStarted }

func (u *User) IsBannedAny() bool { return u.GroupBanned || u.ChannelBanned }

type UsageHistory struct {
	Total        int
	Thissession  int
	Usagehistory []db.UsageHistory
	UserId       int64
}

// Upload and Download should include today values
type FullUsage struct {
	Uploadtd   C.Bwidth // Upload Until Last Refresh
	Downloadtd C.Bwidth 
	Download   C.Bwidth
	Upload     C.Bwidth
}

func (f FullUsage) Full() C.Bwidth {
	return f.Upload + f.Download
}

func (f FullUsage) Today() C.Bwidth {
	return f.Uploadtd + f.Downloadtd
}

type HelpCommandInfo struct {
	InfoPageCount     int16 `json:"info_pages"`
	TutorialPageCount int16 `json:"tuto_pages"`
	CommandPageCount  int16 `json:"cmd_pages"`
	BuilderHelp       int16 `json:"builder_pages"`
}
