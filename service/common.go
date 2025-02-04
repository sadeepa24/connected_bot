package service

import (
	"errors"
	"net/netip"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
)

func closeback(callbackdata string, deletemsg, backfunc func() error) (bool, error) {
	switch callbackdata {
	case C.BtnBack:
		return true, backfunc()
	case C.BtnClose:
		return true, deletemsg()
	}
	return false, nil
}

func checkconform(callbackData string, mgsessn *botapi.Msgsession) error {
	switch callbackData {
	case C.BtnConform:
		return nil
	case C.BtnCancle:
		mgsessn.DeleteAllMsg()
		mgsessn.SendAlert("you cancled creating config", nil)
		return errors.New("user cancled")
	}
	return errors.New("condition unmatched conformation")
}


//common types 

type configinfo struct {
	*botapi.CommonUser
	//*botapi.CommonUsage

	TotalQuota string

	ConfigName string
	ConfigType string
	ConfigUUID string

	ConfigUpload     string
	ConfigDownload   string
	ConfigUploadtd   string
	ConfigDownloadtd string
	ConfigUsage      string
	ConfigUsagetd    string
	UsedPresenTage   float64

	ResetDays int32

	PublicIp string
	PublicDomain string

	InName         string
	InType         string
	InPort         int
	InAddr         string
	InInfo         string
	TranstPortType string
	Loginlimit int16
	TlsEnabled     bool
	SupportInfo    []string

	OutName string
	OutType string
	OutInfo string
	Latency int32

	UsageDuration string

	Online int
	IpMap  map[netip.Addr]int64
}

type userinfo struct {
	*botapi.CommonUser
	Dedicated string

	TQuota       string
	LeftQuota    string
	ConfCount    int16
	TUsage       string
	GiftQuota    string
	Joined       string
	CapEndin     string
	Disendin     int32
	UsageResetIn int32

	Iscapped       bool
	Isgifted       bool
	Isdisuser      bool
	IsMonthLimited bool

	JoinedPlace uint
}