package service

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/sbox"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"github.com/sagernet/sing-vmess/vless"
)

func (u *Xraywiz) commandInfoV2(upx *update.Updatectx,  Messagesession *botapi.Msgsession) error {
	Messagesession.AddreplyNoDelete(upx.Update.Message.MessageID)
	var (
		Usersession *controller.CtrlSession
		err         error
	)

	if Usersession, err = controller.NewctrlSession(u.ctrl, upx, false); err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.EditText(C.GetMsg(C.MsgSessionExcist), nil)
		}
		upx = nil
		Messagesession = nil
		Usersession = nil
		return nil
	}
	defer Usersession.Close()


	var state int
	var callback *tgbotapi.CallbackQuery
	var selectedconf int
	btns := botapi.NewButtons([]int16{2, 1})
	info:
	for {
		if upx.Ctx.Err() != nil {
			return C.ErrContextDead
		}

		switch state {
		// 0 home
		// 1 userinfo
		// 2 configs
		// 3 check outbounds
		// 4 config info
	
		case 0:

			btns.Reset([]int16{2, 2, 1})
			btns.Addbutton(C.BtnUserInfo, C.BtnUserInfo, "")
			btns.Addbutton(C.BtnConfigs, C.BtnConfigs, "")
			btns.AddBtcommon(C.BtnCheckOutbounds)
			btns.AddBtcommon(C.BtnCheckInbounds)
			btns.AddClose(false)
	
			Messagesession.Edit(botapi.UpMessage{
				Template:     struct{}{},
				TemplateName: C.TmplGetinfoHome,
			}, btns, "")
	
			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}
			
			switch callback.Data {
			case C.BtnClose:
				Messagesession.DeleteAllMsg()
				break info
			case C.BtnUserInfo:
				state = 1
			case C.BtnConfigs:
				state = 2
			case C.BtnCheckOutbounds:
				state = 3
			case C.BtnCheckInbounds:
				state = 5
			}
			
		case 1:
			btns.Reset([]int16{2})
			btns.AddCloseBack()

			tusage := Usersession.TotalUsage()

			if _, err = Messagesession.Edit(userinfo{

				CommonUser: &botapi.CommonUser{
					Name:     upx.User.Name,
					TgId:     upx.User.TgID,
					Username: upx.FromChat().UserName,
				},
				CappedQuota: upx.User.CappedQuota.BToString(),
				IsTemplimited: upx.User.Templimited,
				TempLimitRate: upx.User.WarnRatio,
				IsVerified: upx.User.Verified(),


				UsagePercentage: ((tusage * 100)/(Usersession.GetUser().CalculatedQuota + upx.User.AdditionalQuota)).String(),
				GiftQuota: upx.User.GiftQuota.BToString(),
				Joined:    upx.User.Joined.Format("2006-01-02 15:04:05"),
				Dedicated: C.Bwidth(u.ctrl.CommonQuota.Load()).BToString(),
				TQuota:    (Usersession.GetUser().CalculatedQuota + upx.User.AdditionalQuota).BToString(),
				LeftQuota: Usersession.LeftQuota().BToString(),
				TUsage:    tusage.BToString(),
				AlltimeUsage: (upx.User.AlltimeUsage+tusage).BToString(),
				ConfCount: Usersession.GetUser().ConfigCount,
				CapEndin:  upx.User.Captime.AddDate(0, 0, 30).String(),

				Disendin:     ((u.ctrl.ResetCount - u.ctrl.CheckCount.Load()) * u.ctrl.RefreshRate) / 24,
				UsageResetIn: ((u.ctrl.ResetCount - u.ctrl.CheckCount.Load()) * u.ctrl.RefreshRate) / 24,

				Iscapped:       upx.User.IsCapped,
				IsMonthLimited: upx.User.IsMonthLimited,
				Isdisuser:      upx.User.IsDistributedUser,

				JoinedPlace: upx.User.CheckID,
			}, btns, C.TmpUserInfo); err != nil {
				u.logger.Error(err.Error())
				state = 0
				continue info
			}

			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}
			if callback.Data == C.BtnClose {
				break info
			}
			state = 0

		case 2:
			if Usersession.GetUser().ConfigCount == 0 {
				Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgInfoNoconfigs), true)
				state = 0
				continue info

			}

			btns.Reset([]int16{2})
			
			for _, config := range Usersession.GetUser().Configs {
				btns.Addbutton(config.Name, strconv.Itoa(int(config.Id)), "")
			}
			btns.AddCloseBack()
			if _, err = Messagesession.EditText(C.GetMsg(C.MsgInfoSelectConfig), btns); err != nil {
				continue info
			}

			callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID())
			if err != nil {
				return err
			}

			switch callback.Data {
			case C.BtnBack:
				state = 0
				continue info
			case C.BtnClose:
				break info
			}
		
			selectedconf, err = strconv.Atoi(callback.Data)
			if err != nil {
				continue info
			}

			state = 4


		case 3:
			allouts := u.ctrl.Getoutbounds()
			if len(allouts) == 0 {
				Messagesession.Callbackanswere(callback.ID, "no any outbound found", true)
				state = 0
				continue info
			}
			btns.Reset([]int16{2})
			for _, out:= range allouts {
				btns.Addbutton(out.Tag, strconv.Itoa(int(out.Id)), "")
			}
			btns.AddCloseBack()
	

			if _, err = Messagesession.Edit("select outbound to see details", btns, ""); err != nil {
				continue info
			}
			
			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}

			switch callback.Data{
			case C.BtnBack:
				state = 0
				continue
			case C.BtnClose:
				Messagesession.DeleteAllMsg()
				return nil
			default:
				id, err := strconv.Atoi(callback.Data)
				if err != nil {
					continue info
				}

				out, ok := u.ctrl.Getoutbound(id)
				if !ok {
					Messagesession.Callbackanswere(callback.ID, "outbound not found", true)
					continue info 
				}
				btns.Reset([]int16{2})
				btns.AddBtcommon("Check Latency")
				btns.AddCloseBack()
				if _, err = Messagesession.Edit(struct {
					OutName string
					Info string
					Latency int32
					Type string

				}{
					OutName: out.Name,
					Info: out.Custom_info,
					Latency: out.Latency.Load(),
					Type: out.Type,

				}, btns, C.TmplOutInfo); err != nil {
					u.logger.Error(err.Error())
					continue info
	
				}

				

				latency:
				for {
					if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
						return err
					}
					switch callback.Data {
					case C.BtnBack:
						continue info
					case C.BtnClose:
						break info
					case"Check Latency":
						ping, err := u.ctrl.UrlTestOut(out.Tag)
						
						if err != nil {
							Messagesession.Callbackanswere(callback.ID, "Latency checking error, outbound timeout", true)
							continue latency
						}
						out.Latency.Swap(int32(ping))
						Messagesession.Callbackanswere(callback.ID, fmt.Sprintf("⚡ outbound latency %v", ping), true)
					}

				}

			
			}


		case 4:

			confid := selectedconf

			selectedconfig, err := Usersession.GetConfig(int64(confid))

			if err != nil {
				Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msgconfcannotfind), true)
				continue info
			}

			btns.Reset([]int16{2})
			//btns.Addbutton(C.BtnFullUsage, C.BtnFullUsage, "")
			btns.AddBtcommon(C.BtnCloseConn)
			btns.AddBtcommon("Refresh")
			btns.AddCloseBack()

			status, err := Usersession.Getstatus(int64(confid))

			if err != nil {
				if errors.Is(err, C.ErrContextDead) {
					return err
				} else if errors.Is(err, vless.ErrUserNotFound) {
					status = sbox.Sboxstatus{
						Download:  0,
						Upload:    0,
						Online_ip: map[netip.Addr]int64{},
					}
				} else {
					Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.GetMsg(C.Msgconfcannotfind)), true)
					Messagesession.DeleteAllMsg()
					return err
				}

			}

			sboxin, _ := u.ctrl.Getinbound(int(selectedconfig.InboundID))
			sboxout, _ := u.ctrl.Getoutbound(int(selectedconfig.OutboundID))

			if _, err = Messagesession.Edit(configinfo{
				CommonUser: &botapi.CommonUser{
					Name:     upx.User.Name,
					Username: upx.FromChat().UserName,
					TgId:     upx.User.TgID,
				},

				TotalQuota:     selectedconfig.Quota.BToString(),
				ConfigName:     selectedconfig.Name,
				ConfigType:     selectedconfig.Type,
				ConfigUUID:     selectedconfig.UUID,
				Loginlimit: selectedconfig.LoginLimit,
				UsedPresenTage: float64(int(((selectedconfig.Usage+status.FullUsage()).Float64()/selectedconfig.Quota.Float64())*100*1000)) / 1000,
				//UsedPresenTage: (((selectedconfig.Usage + status.FullUsage()).Float64()/selectedconfig.Quota.Float64()))*100,

				ResetDays: ((u.ctrl.ResetCount - u.ctrl.CheckCount.Load()) * u.ctrl.RefreshRate) / 24,

				ConfigDownload: (selectedconfig.Download + status.Download).BToString(),
				ConfigUpload:   (selectedconfig.Upload + status.Upload).BToString(),

				ConfigDownloadtd: (status.Download).BToString(),
				ConfigUploadtd:   (status.Upload).BToString(),

				ConfigUsagetd: (status.Download + status.Upload).BToString(),
				ConfigUsage:   (status.Download + status.Upload + selectedconfig.Usage).BToString(),

				InName:         sboxin.Name,
				InType:         sboxin.Type,

				PublicIp: sboxin.PublicIp,
				PublicDomain: sboxin.Domain,
				
				InPort:         sboxin.Port(),
				InAddr:         u.ctrl.DefaultPubip,
				InInfo:         sboxin.Custom_info,
				TranstPortType: sboxin.TransortType(),
				TransPortPath:  sboxin.TransportPath(),
				TlsEnabled:     sboxin.TlsIsEnabled(),
				UsageDuration:  time.Since(u.ctrl.GetLastRefreshtime()).Round(1 * time.Second).String(),
				SupportInfo:    sboxin.Support,

				OutName: sboxout.Name,
				OutType: sboxout.Type,
				OutInfo: sboxout.Custom_info,
				Latency: sboxout.Latency.Load(),

				Online: len(status.Online_ip),
				IpMap:  status.Online_ip,
				//TODO: fill here
			}, btns, C.TmpConfigInfo); err != nil {
				u.logger.Error(err.Error())
				state = 2
				continue info

			}
			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}
			
			switch callback.Data {
			case C.BtnClose:
				break info
			case C.BtnCloseConn:
				Usersession.ConfigCloseConn(int64(confid))
			case C.BtnBack:
				state = 2
			case "Refresh":
				continue info
			}

			if callback.Data == C.BtnFullUsage {
				//TODO: code here

				Messagesession.SendAlert(C.GetMsg(C.Msgconfcannotfind), nil)
				Messagesession.SendAlert("usage history function is not avalable yet", nil)
			}

		case 5:
			allins := u.ctrl.Getinbounds()
			if len(allins) == 0 {
				Messagesession.Callbackanswere(callback.ID, "no any outbound found", true)
				state = 0
				continue info
			}
			btns.Reset([]int16{2})
			for _, in:= range allins {
				btns.Addbutton(in.Tag, strconv.Itoa(int(in.Id)), "")
			}
			btns.AddCloseBack()

			if _, err = Messagesession.Edit("select inbound to see details", btns, ""); err != nil {
				continue info
			}
			
			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}
			switch callback.Data{
			case C.BtnBack:
				state = 0
				continue
			case C.BtnClose:
				Messagesession.DeleteAllMsg()
				return nil
			default:
				id, err := strconv.Atoi(callback.Data)
				if err != nil {
					continue info
				}

				in, ok := u.ctrl.Getinbound(id)
				if !ok {
					Messagesession.Callbackanswere(callback.ID, "inbound not found", true)
					continue info 
				}
				btns.Reset([]int16{2})
				btns.AddCloseBack()
				if _, err = Messagesession.Edit(struct {
					InName         string
					InType         string
					InPort         int
					InAddr         string
					InInfo         string
					TranstPortType string
					TlsEnabled     bool
					SupportInfo    []string
					Domain         string
					PublicIp       string
					Support        []string



				}{
					InName: in.Name,
					InType: in.Type,
					InPort: in.Port(),
					TlsEnabled: in.Tlsenabled,
					TranstPortType: in.Transporttype,
					InAddr: u.ctrl.DefaultPubip,
					InInfo: in.Custom_info,
					SupportInfo: in.Support,
					Domain: in.Domain,
					PublicIp: in.PublicIp,
					
				}, btns, C.TmplInInfo); err != nil {
					u.logger.Error(err.Error())
					continue info
	
				}
				if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
					return err
				}
				switch callback.Data {
				case C.BtnClose:
					break info
				}
			}

		}

	}
	Messagesession.DeleteAllMsg()
	return err

	//return err
}
