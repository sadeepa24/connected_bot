package service

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/sbox/conf"
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
	Active bool
	ConfName string
	ConfigName string
	ConfigUUID string
	ConfigPassword string


	ConfigUpload     string
	ConfigDownload   string
	ConfigUploadtd   string
	ConfigDownloadtd string
	ConfigUsage      string
	ConfigUsagetd    string
	UsedPresenTage   float64

	ResetDays int32

	// PublicIp string
	// PublicDomain string

	// InName         string
	// InType         string
	// InPort         int
	// InAddr         string
	// InInfo         string
	// TranstPortType string
	// TransPortPath string
	Loginlimit int16
	// TlsEnabled     bool
	// SupportInfo    []string

	//conf.Outbound

	commonout

	UsageDuration string

	Online int
	IpMap  map[string]int16
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
	TempLimitRate int16
	AlltimeUsage string
	UsagePercentage float64
	NonUseCycle int16
	CapDays int32
	Points  int64
	Paused 	bool

	CappedQuota string

	IsVerified 	   bool
	Iscapped       bool
	Isgifted       bool
	Isdisuser      bool
	IsMonthLimited bool
	IsTemplimited bool
	
	JoinedPlace uint
}

type exportConfig struct {
	exportin
	conf.Inboud
}


type exportin struct {
	ProtoUrl string
	conf.ExportInfo
}

type commonout struct {
	OutName string
	OutType	string
	OutInfo	string
	Latency int32
}

func timeout(timebreaks time.Duration, counter *atomic.Int32, ctx context.Context, cancel context.CancelFunc, MsgS *botapi.Msgsession) {
	ticker := time.NewTicker(timebreaks)
	for {
		select {
		case <-ticker.C:
			if counter.Add(-1) <= 0 {
				if MsgS != nil {
					MsgS.SendAlert("timeout", nil)
				}
				cancel()
				ticker.Stop()
				return
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}