package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/service/events"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Usersrv struct {
	ctx      context.Context
	callback *Callback
	logger   *zap.Logger
	//db *db.Database
	admin      *Adminsrv
	ctrl       *controller.Controller
	defaultsrv *Defaultsrv
	adminchat  map[int64]string

	botapicaller botapi.BotAPI
	MessageStore *botapi.MessageStore

	AllEvents map[string]events.Event //all avalable events

}

func NewuserService(ctx context.Context,
	callback *Callback,
	logger *zap.Logger,
	adminsrv *Adminsrv,
	ctrl *controller.Controller,
	defaultsrv *Defaultsrv,
	botapi botapi.BotAPI,
	msgstore *botapi.MessageStore,

) *Usersrv {

	return &Usersrv{
		admin:        adminsrv,
		ctx:          ctx,
		callback:     callback,
		ctrl:         ctrl,
		logger:       logger,
		defaultsrv:   defaultsrv,
		botapicaller: botapi,
		adminchat:    map[int64]string{},
		MessageStore: msgstore,
	}
}

func (u *Usersrv) Exec(upx *update.Updatectx) error {
	u.logger.Info("executing user service " + upx.User.Info() )
	switch {
	case upx.Update.Message != nil:
		if upx.Update.Message.IsCommand() {
			return u.Commandhandler(upx.Update.Message.Command(), upx)
		}
		return nil
	case (upx.Update.ChatMember != nil || upx.Update.MyChatMember != nil):
		return u.ChatmemberUpdate(upx)
	default:
		return u.defaultsrv.FromserviceExec(upx)
	}

}

func (u *Usersrv) Init() error {
	u.logger.Debug("User service inilized")
	var err error

	u.AllEvents = events.GetallAvblkEvent(u.ctrl)
	u.adminchat, err = u.ctrl.Getadminchat()
	if err != nil {
		return err
	}
	return nil
}

func (u *Usersrv) ChatmemberUpdate(upx *update.Updatectx) error {
	u.ctrl.IncCriticalOp()
	defer u.ctrl.DecCriticalOp()

	if upx.Update.ChatMember == nil && upx.Update.MyChatMember == nil {
		return nil
	}

	var (
		updatedchat   string
		ok            bool
		NewchatMember tgbotapi.ChatMember
		err           error
	)

	if updatedchat, ok = u.adminchat[upx.FromChat().ID]; !ok {
		return nil
	}

	switch {
	case upx.Update.ChatMember != nil:
		NewchatMember = upx.Update.ChatMember.NewChatMember
	case upx.Update.MyChatMember != nil:
		NewchatMember = upx.Update.MyChatMember.NewChatMember
	}

	NewUser := upx.User

	//to check is the same user from update or a user add or remove another user
	if NewchatMember.User.ID != NewUser.Id {
		u.logger.Info("user added/baned/removed another " + upx.User.Info() )

		NewUser, ok, err = u.ctrl.GetUser(NewchatMember.User)
		if err != nil {
			return errors.Join(errors.New("chat member parsing failed database fetching"), err)
		}
		if !ok {
			NewUser, err = u.ctrl.Newuser(NewchatMember.User, upx.Chat)
			if err != nil {
				return err
			}
		}

	}
	Messagesession := botapi.NewMsgsession(u.botapicaller, NewUser.Id, NewUser.Id, upx.User.Lang)
	//GroupSession := botapi.NewMsgsession(u.botapicaller, upx.User.Id, u.ctrl.GroupID, upx.User.Lang)
	upx.User = NewUser	// To Create Usersession From Newuser not from the one who made action
	Usersession, err := controller.NewctrlSession(u.ctrl, upx, true) //will force cancle other session is exicist

	if err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			NewUser.RecheckVerificity = true
		}
		return err
	}

	defer Usersession.Close()

	btns := botapi.NewButtons([]int16{2, 1})
	btns.Addbutton(C.BtnChannel, C.BtnChannel, u.ctrl.Channelink)
	btns.Addbutton(C.BtnGroup, C.BtnGroup, u.ctrl.GroupLink)
	btns.Addbutton(C.BtnBot, C.BtnBot, u.ctrl.Botlink)

	switch NewchatMember.Status {
	case C.Statusleft:
		Usersession.Chatupdate(updatedchat, false)
		Usersession.GetUser().IsRemoved = true
		Usersession.DeactivateAll()

		if NewUser.Isbotstarted() {

			Messagesession.Edit(struct {
				*botapi.CommonUser
				LeftQuota string
			}{
				CommonUser: &botapi.CommonUser{
					Name:     NewUser.Name,
					Username: NewUser.Username.String,
					TgId:     upx.User.TgID,
				},
				LeftQuota: Usersession.LeftQuota().BToString(),
			}, btns, C.TmpChatmemLeft)

		}

	case C.Statuskicked:
		Usersession.Banuser(updatedchat)
		Messagesession.SendAlert(C.GetMsg(C.MsgBannedMem), nil)

	case C.Statusmember:
		switch {

		//newly joined group
		
		case updatedchat == C.Group && !NewUser.IsRemoved:
			u.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
				Template: struct {
					*botapi.CommonUser
					Chat         string
					IsInChannel  bool
					IsBotStarted bool
					GroupLink string
					ChanLink string
				}{
					CommonUser: &botapi.CommonUser{
						Name:     NewUser.Name,
						Username: NewUser.Username.String,
						TgId:     NewUser.TgID,
					},
					IsInChannel:  upx.User.IsInChannel,
					IsBotStarted: upx.User.IsBotStarted,
					Chat:         updatedchat,
					GroupLink: u.ctrl.GroupLink,
					ChanLink: u.ctrl.Channelink,
				},
				TemplateName: C.TmpGroupWelcome,
				DestinatioID: u.ctrl.GroupID,
				Lang:         NewUser.Lang,
				Buttons:      btns,
				
			},
			)
			if upx.User.IsInChannel {
				Messagesession.Edit(struct {
					*botapi.CommonUser
				}{
					CommonUser: &botapi.CommonUser{
						Name:     upx.User.Name,
						Username: upx.FromChat().UserName,
						TgId:     upx.User.TgID,
					},
				}, nil, C.TmplInboxVerified)
			}
			Messagesession.Edit(struct {
				*botapi.CommonUser
				IsInChannel  bool
				IsBotStarted bool
			}{
				CommonUser: &botapi.CommonUser{
					Name:     upx.User.Name,
					Username: upx.FromChat().UserName,
					TgId:     upx.User.TgID,
				},
				IsInChannel:  upx.User.IsInChannel,
				IsBotStarted: upx.User.IsBotStarted,
			}, btns, C.TmpWelcomeInbox)

		//newly joined channel
		case updatedchat == C.Channel && !NewUser.IsRemoved:

			if upx.User.IsInGroup {

				u.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
					Template: struct {
						*botapi.CommonUser
						Chat string
						GroupLink string
						ChanLink string
					}{
						CommonUser: &botapi.CommonUser{
							Name:     NewUser.Name,
							Username: NewUser.Username.String,
							TgId:     NewUser.TgID,
						},
						Chat: updatedchat,
					},
					TemplateName: C.TmpChannelWelcome,
					DestinatioID: u.ctrl.GroupID,
					Lang:         NewUser.Lang,
					Buttons:      btns,
				},
				)

				Messagesession.Edit(struct {
					*botapi.CommonUser
				}{
					CommonUser: &botapi.CommonUser{
						Name:     upx.User.Name,
						Username: upx.FromChat().UserName,
						TgId:     upx.User.TgID,
					},
				}, nil, C.TmplInboxVerified)

			}
			Messagesession.Edit(struct {
				*botapi.CommonUser
				IsInChannel  bool
				IsBotStarted bool
				GroupLink string
				ChanLink string
			}{
				CommonUser: &botapi.CommonUser{
					Name:     upx.User.Name,
					Username: upx.FromChat().UserName,
					TgId:     upx.User.TgID,
				},
				IsInChannel:  upx.User.IsInChannel,
				IsBotStarted: upx.User.IsBotStarted,
			}, btns, C.TmpWelcomeInbox)

		// left and joined again channel
		case updatedchat == C.Channel:

			u.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
				Template: struct {
					*botapi.CommonUser
					Chat string
					GroupLink string
					ChanLink string
				}{
					CommonUser: &botapi.CommonUser{
						Name:     NewUser.Name,
						Username: NewUser.Username.String,
						TgId:     NewUser.TgID,
					},
					Chat: updatedchat,
					GroupLink: u.ctrl.GroupLink,
					ChanLink: u.ctrl.Channelink,
				},
				TemplateName: C.TmpChannelWelcome,
				DestinatioID: u.ctrl.GroupID,
				Lang:         NewUser.Lang,
				Buttons:      btns,

			},
			)
			if upx.User.IsInGroup {
				Messagesession.Edit(struct {
					*botapi.CommonUser
				}{
					CommonUser: &botapi.CommonUser{
						Name:     upx.User.Name,
						Username: upx.FromChat().UserName,
						TgId:     upx.User.TgID,
					},
				}, nil, C.TmplInboxVerifiedAgain)
			}

		// left and joined again group
		case updatedchat == C.Group:
			u.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
				Template: struct {
					botapi.CommonUser
					Chat         string
					IsInChannel  bool
					IsBotStarted bool
					GroupLink string
					ChanLink string
				}{
					CommonUser: botapi.CommonUser{
						Name:     NewUser.Name,
						Username: NewUser.Username.String,
						TgId:     NewUser.TgID,
					},
					Chat:        updatedchat,
					IsInChannel: upx.User.IsInChannel,
					GroupLink: u.ctrl.GroupLink,
					ChanLink: u.ctrl.Channelink,
				},
				TemplateName: C.TmpGrpComeback,
				DestinatioID: u.ctrl.GroupID,
				Lang:         NewUser.Lang,
				Buttons:      btns,
			},
			)

			if upx.User.IsInChannel {
				Messagesession.Edit(struct {
					botapi.CommonUser
				}{
					CommonUser: botapi.CommonUser{
						Name:     upx.User.Name,
						Username: upx.FromChat().UserName,
						TgId:     upx.User.TgID,
					},
				}, nil, C.TmplInboxVerifiedAgain)
			}

		}

		Usersession.Chatupdate(updatedchat, true)
		
		if NewUser.Isverified() {
			Usersession.GetUser().IsRemoved = false
			u.ctrl.IncreaseUserCount(1)
		}
		if err = Usersession.ActivateAll(); err != nil {
			return errors.Join(errors.New("chat member parsing config activate failed user " + upx.User.Name), err)
		}

	}

	// NewUser = nil
	// upx = nil
	// Usersession = nil
	return nil
}

func (u *Usersrv) Commandhandler(cmd string, upx *update.Updatectx) error {

	switch cmd {
	case C.CmdStart:
		return u.commandStart(upx)
	case C.CmdHelp:
		return u.commandHelpV2(upx)
	case C.CmdGift:
		if upx.User.IsDistributedUser { break }
		return u.commandGift(upx)
	case C.CmdDistribute:
		return u.commandDistribute(upx)
	case C.CmdCap:
		if upx.User.IsDistributedUser { break }
		return u.commandCap(upx)
	case C.CmdRefer:
		return u.commandReffral(upx)
	case C.CmdSugess:
		return u.commandSuggesion(upx)
	case C.CmdEvents:
		return u.commandEvents(upx)
	case C.CmdPoints:
		return u.commandPoints(upx)
	case C.CmdContact: 
		return u.commandContact(upx)
	case C.CmdRecheck:
		return u.Recheck(upx)
	default:
		u.logger.Warn("unknown cmd recived by userservice - " + cmd)
		return u.defaultsrv.FromserviceExec(upx)

	}

	if upx.User.IsDistributedUser {
		u.ctrl.Addquemg(upx.Ctx, &botapi.Msgcommon{
			Infocontext: &botapi.Infocontext{
				ChatId: upx.User.TgID,
			},
			Text: C.GetMsg(C.MsgNotCmdDIs),

		})
	}
	return nil
}

func (u *Usersrv) commandStart(upx *update.Updatectx) error {
	var err error
	Messagesession := botapi.NewMsgsession(u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	btns := botapi.NewButtons([]int16{1, 1})
	btns.Addbutton(C.BtnChannel, C.BtnChannel, u.ctrl.Channelink)
	btns.Addbutton(C.Group, C.Group, u.ctrl.GroupLink)

	switch {

	case !upx.FromChat().IsPrivate():
		err = errors.New("user send start command group chat " + upx.User.Info())
	case upx.User.IsMonthLimited:

		Messagesession.Edit(struct {
			*botapi.CommonUser
			LimitendIn int32
		}{
			CommonUser: &botapi.CommonUser{
				Name:     upx.User.Name,
				Username: upx.FromChat().UserName,
				TgId:     upx.User.TgID,
			},
			LimitendIn: ((u.ctrl.ResetCount - u.ctrl.CheckCount.Load()) * u.ctrl.RefreshRate) / 24,
		}, nil, C.TmplMonthLimited)

		Messagesession.EditText(C.GetMsg(C.Msgstartmlimited), nil)

	case upx.User.Restricted:

		Messagesession.Edit(struct {
			*botapi.CommonUser
		}{
			CommonUser: &botapi.CommonUser{
				Name:     upx.User.Name,
				Username: upx.FromChat().UserName,
				TgId:     upx.User.TgID,
			},
		}, nil, "restrictedstart")

		//Messagesession.EditText(C.GetMsg(C.Msgstartmlimited), nil)

	case upx.User.IsnewUser():

		Messagesession.Edit(struct {
			*botapi.CommonUser
		}{
			&botapi.CommonUser{
				Name:     upx.User.Name,
				Username: upx.Chat.UserName,
				TgId:     upx.User.TgID,
			},
		}, btns, C.TmpNewUsers)

		upx.User.IsBotStarted = true

		err = u.ctrl.SetIsbotarted(upx.User.Id, true)

	// verified user start the bot first time
	case !upx.User.Isbotstarted() && upx.User.Isverified():
		Messagesession.Edit(struct {
			*botapi.CommonUser
		}{
			&botapi.CommonUser{
				Name:     upx.User.Name,
				Username: upx.Chat.UserName,
				TgId:     upx.User.TgID,
			},
		}, btns, C.TmpNewUsersVerified)

		upx.User.IsBotStarted = true
		err = u.ctrl.SetIsbotarted(upx.User.Id, true)
	//unverified user start the bot first time
	case !upx.User.Isbotstarted():

		// send group links and etc to user
		Messagesession.Edit(struct {
			*botapi.CommonUser
		}{
			&botapi.CommonUser{
				Name:     upx.User.Name,
				Username: upx.Chat.UserName,
				TgId:     upx.User.TgID,
			},
		}, btns, C.TmplUserUnverifiedStart)

		upx.User.IsBotStarted = true
		err = u.ctrl.SetIsbotarted(upx.User.Id, true)

	case upx.User.Isverified():

		btns.Reset([]int16{1})

		Messagesession.Edit(struct {
			*botapi.CommonUser
			*botapi.CommonUsage
		}{
			&botapi.CommonUser{
				Name:     upx.User.Name,
				Username: upx.Chat.UserName,
				TgId:     upx.User.TgID,
			},
			&botapi.CommonUsage{
				AddtionalQuota:  upx.User.AdditionalQuota.BToString(),
				CalculatedQuota: upx.User.CalculatedQuota.BToString(),
				Alltime:         (upx.User.MonthUsage + upx.User.AlltimeUsage).BToString(),
				MUsage:          upx.User.MonthUsage.BToString(),
			},
		}, nil, C.TmpregularVerified)

	case upx.User.IsremovedUser() && !upx.User.IsBannedAny():

		btns.Reset([]int16{1})
		btns.AddUrlbutton(C.BtnChannel, u.ctrl.Channelink)
		btns.AddUrlbutton(C.BtnGroup, u.ctrl.GroupLink)

		Messagesession.Edit(struct {
			*botapi.CommonUser
			IsInChannel bool
			IsinGroup   bool
		}{
			CommonUser: &botapi.CommonUser{
				Name:     upx.User.Name,
				Username: upx.FromChat().UserName,
				TgId:     upx.User.TgID,
			},
			IsInChannel: upx.User.IsInChannel,
			IsinGroup:   upx.User.IsInGroup,
		}, btns, C.TmpRemUserst)

	case upx.User.IsBannedAny():
		Messagesession.SendAlert(C.GetMsg(C.MsgBannedUser), nil)

	case upx.User.IsInChannel:
		btns.Reset([]int16{1})
		btns.AddUrlbutton(C.BtnGroup, u.ctrl.GroupLink)
		Messagesession.EditText(C.GetMsg(C.MsgsttInChan), btns)

	case upx.User.IsInGroup:
		btns.Reset([]int16{1})
		btns.AddUrlbutton(C.BtnChannel, u.ctrl.Channelink)
		Messagesession.EditText(C.GetMsg(C.MsgstartGrpin), btns)

	default:
		u.logger.Warn(" user start command all condition missmatched") //this will never happen
		err = nil
	}

	//refreal checkings
	args := strings.TrimSpace(upx.Update.Message.CommandArguments())

	if args != "" {
		reowenerid, err := strconv.Atoi(args)

		if err == nil {
			if reowenerid == int(upx.User.TgID) {
				Messagesession.SendAlert(C.GetMsg(C.MsgSelfRef), nil)
				return nil
			}

			user, err := u.ctrl.CreateRefrral(int64(reowenerid), upx.User.TgID)

			if err != nil {

				if errors.Is(err, C.ErrUserExitDb) {
					Messagesession.SendAlert(fmt.Sprintf(C.GetMsg(C.MsgRefAlredy), strconv.Itoa(int(user.OwnerID))), nil)
				}

			} else {

				btypeuser, ok, err := u.ctrl.GetUser(&tgbotapi.User{
					ID: int64(reowenerid),
				})

				if !ok {
					if err != nil {
						Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
					} else {
						Messagesession.SendAlert(C.GetMsg(C.MsgRefOwenerNFoun), nil)
					}

				} else {
					Messagesession.SendAlert(fmt.Sprintf(C.GetMsg(C.MsgReferd), btypeuser.Name), nil)
					Messagesession.SendAlert(C.GetMsg(C.Msgcanref), nil)
				}

			}

		}

	}

	//upx = nil
	return err
}

func (u *Usersrv) commandGift(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	if upx.User.IsCapped {
		Messagesession.SendAlert(C.GetMsg(C.MsgGifUsercap), nil)
		return nil
	}

	Usersession, err := controller.NewctrlSession(u.ctrl, upx, false)
	if err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.EditText(C.GetMsg(C.MsgSessionExcist), nil)
		} else {
			Messagesession.EditText(C.GetMsg(C.MsgSessionFail), nil)
		}
		upx = nil
		return nil
	}
	defer Usersession.Close()
	//avblquota := 0

	// if len(upx.User.Configs) == 0 {
	// 	Messagesession.SendAlert("you don't have any configs, you should have at least 1 config to send a gift",  nil)
	// 	return nil
	// }

	Messagesession.Edit(struct {
		LeftQuota string
	}{
		LeftQuota: Usersession.LeftQuotaFromOrigin().BToString(),
	}, nil, C.TmpGifSend)
	var (
		replymg  *tgbotapi.Message
		usersend int
		retry    = 0
	)
	for {

		if upx.Ctx.Err() != nil {
			return err
		}
		if retry > 5 {
			Messagesession.EditText(C.GetMsg(C.Msgretryfail), nil)
			return nil
		}

		if replymg, err = u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID); err != nil {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionOver), nil)
			return err
		}
		Messagesession.Addreply(replymg.MessageID)
		retry++
		if usersend, err = strconv.Atoi(replymg.Text); err != nil {
			Messagesession.SendAlert(C.GetMsg(C.MsgValidInt), nil)
			continue
		}

		if usersend <= 0 {
			Messagesession.SendAlert(C.GetMsg(C.MsgQuotawarnzero), nil)
			continue
		}

		if C.Bwidth(usersend).GbtoByte() > Usersession.LeftQuotaFromOrigin() {
			Messagesession.SendAlert(C.GetMsg(C.Msggifterr), nil)
			continue
		}

		break
	}

	//usersend, err := common.ReciveInt(common.Tgcalls{}, max, )


	btns := botapi.NewButtons([]int16{1})
	btns.Addcancle()

	Messagesession.EditText(C.GetMsg(C.MsgGifreciver), nil)

	if replymg, err = u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID); err != nil {
		return err
	}

	if replymg.Command() == "cancel" {
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert(C.GetMsg(C.MsgGiftcancle), nil)
		return nil
	}
	replymg.Text = strings.ReplaceAll(replymg.Text, "@", "")

	var targetuser *db.User
	var reciver any

	reciver, err = strconv.Atoi(replymg.Text)
	if err != nil {
		reciver = replymg.Text
		if replymg.Text == upx.User.Username.String {
			Messagesession.SendAlert("Lol, You can't send Gift You'r self", nil)
			return nil
		}
	} else {
		if reciver.(int) == int(upx.User.TgID) {
			Messagesession.SendAlert("Lol, You can't send Gift You'r self", nil)
			return nil
		} 
	}





	targetuser, err = u.ctrl.Gift(upx, reciver, C.Bwidth(usersend).GbtoByte())

	if err != nil {
		Messagesession.DeleteAllMsg()
		switch {
		case errors.Is(err, C.ErrConfigNotFound):
			Messagesession.SendAlert(C.GetMsg(C.MsgGifRecnOconfig), nil)
		case errors.Is(err, C.ErrDbopration):
			Messagesession.SendAlert(C.GetMsg(C.MsgDberr), nil)
		case errors.Is(err, C.ErrConfigNotFound):
			Messagesession.SendAlert(C.GetMsg(C.MsgUserNotFoun), nil)
		case errors.Is(err, C.ErrUserCanootReciveUserCapped):
			Messagesession.SendAlert(C.GetMsg(C.MsgTargetcapped), nil)
		default:
			Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		}

		return errors.Join(errors.New("errored when gifting " + upx.User.Info()), err)
	}

	//TODO: add template here
	//u.ctrl.Addquemg(upx.Ctx, )

	btns.Reset([]int16{2})
	btns.AddUrlbutton("Thanks Him", fmt.Sprintf("tg://user?id=%v", upx.User.TgID))

	u.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
		Template: struct {
			*botapi.CommonUser
			Gift     string
			FromUser string
		}{
			CommonUser: &botapi.CommonUser{
				Name:     targetuser.Name,
				Username: targetuser.Username.String,
				TgId:     targetuser.TgID,
			},
			FromUser: upx.User.Name,
			Gift:     C.Bwidth(usersend).GbtoByte().BToString(),
		},
		TemplateName: C.TmplRecived,
		Buttons:      btns,
		DestinatioID: targetuser.TgID,
		Lang:         upx.User.Lang,
	})
	Messagesession.SendAlert(fmt.Sprintf(C.GetMsg(C.MsgGiftSent), C.Bwidth(usersend).GbtoByte().BToString(), targetuser.Name), nil)
	
	u.logger.Info(fmt.Sprintf("User [%s] Gifted %d GB to %s", upx.User.Name, usersend, targetuser.Name ))
	
	/*
		old way of sending msg
		u.botapicaller.SendContext(upx.Ctx, &botapi.Msgcommon{
			Infocontext: &botapi.Infocontext{
				ChatId:  targetuser.TgID,
				User_id: targetuser.TgID,
			},
			Text: "Congratulation you have recived " + C.Bwidth(usersend).String() + " gift data from " + upx.User.Name,
		})
	*/

	return nil
}

func (u *Usersrv) commandDistribute(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	if upx.User.IsDistributedUser {
		Messagesession.SendAlert(C.GetMsg(C.MsgDisAlready), nil)
		return nil
	}

	if upx.User.IsCapped {
		Messagesession.SendAlert(C.GetMsg(C.MsgDisCapped), nil)
		return nil
	}

	Usersession, err := controller.NewctrlSession(u.ctrl, upx, false)
	if err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.EditText(C.GetMsg(C.MsgSessionExcist), nil)
		} else {
			Messagesession.EditText(C.GetMsg(C.MsgSessionFail), nil)
		}
		upx = nil
		return nil

	}
	defer Usersession.Close()

	btns := botapi.NewButtons([]int16{1, 1})
	btns.Addbutton(C.BtnConform, C.BtnConform, "")
	btns.AddClose(false)

	Messagesession.EditText(C.GetMsg(C.MsgDisConform), btns)
	replcallback, err := u.callback.GetcallbackContext(upx.Ctx, btns.ID())
	if err != nil {
		return nil
	}

	if ok, err := closeback(replcallback.ID, Messagesession.DeleteAllMsg, func() error {
		return nil
	}); ok {
		return err
	}

	if replcallback.Data == C.BtnConform {
		Usersession.GetUser().IsDistributedUser = true
		Usersession.DeactivateAll()
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert(C.GetMsg(C.MsgDisSucsess), nil)

		u.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
			Template: struct {
				*botapi.CommonUser
				Disquota string
			}{
				Disquota: (upx.User.CalculatedQuota - upx.User.MonthUsage).BToString(),
				CommonUser: &botapi.CommonUser{
					Name:     upx.User.Name,
					Username: upx.Chat.UserName,
					TgId:     upx.User.TgID,
				},
			},
			TemplateName: C.TmpDisGroup,
			Lang:         upx.User.Lang,
			DestinatioID: u.ctrl.GroupID,
		})

	}

	u.logger.Info(upx.User.Name + " Is distributed his quota " + upx.User.Info() )

	return nil
}

func (u *Usersrv) commandCap(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	if upx.User.IsCapped {
		Messagesession.SendAlert(C.GetMsg(C.MsgcapAlready), nil)
		Messagesession.SendExtranal(struct {
			EndDate string
		}{
			EndDate: upx.User.Captime.AddDate(0, 0, 30).Format("2006-01-02 15:04:05"),
		}, nil, C.TmpcapQuota, true)
		return nil
	}

	Usersession, err := controller.NewctrlSession(u.ctrl, upx, false)
	if err != nil {
		if errors.Is(err, C.ErrSessionExcit) {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionExcist), nil)
		} else {
			Messagesession.SendAlert(C.GetMsg(C.MsgSessionFail), nil)
		}
		upx = nil
		return nil

	}
	defer Usersession.Close()

	if Usersession.LeftQuota() <= 0 {
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert(C.GetMsg(C.Msgcapexced), nil)
		return nil
	}

	btns := botapi.NewButtons([]int16{1, 1})
	btns.Addbutton(C.BtnContinue, C.BtnContinue, "")
	btns.AddClose(false)

	capble_quota := (Usersession.GetUser().CalculatedQuota - Usersession.GetFullUsage().Full())

	Messagesession.Edit(struct {
		Leftquota    string
		CapbleQuouta string
	}{
		Leftquota:    Usersession.LeftQuota().BToString(),
		CapbleQuouta: capble_quota.BToString(),
	}, btns, C.TmpcapWarn)

	answer, err := u.callback.GetcallbackContext(upx.Ctx, btns.ID())
	if err != nil {
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return nil
	}

	switch answer.Data {
	case C.BtnClose:
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert("closed", nil)
		return nil
	}

	Messagesession.Edit(struct {
		LeftQuota    string
		CapbleQuouta string
	}{
		LeftQuota:    Usersession.LeftQuota().BToString(),
		CapbleQuouta: capble_quota.BToString(),
	}, nil, C.Tmpcapreply)

	Newcap := C.Bwidth(0)
	var retry = 0
	for {

		if upx.Ctx.Err() != nil {
			return C.ErrContextDead
		}

		if retry > 5 {
			Messagesession.SendAlert(C.GetMsg(C.Msgretryfail), nil)
			return nil
		}
		replymg, err := u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.Id, upx.Chat_ID)
		if err != nil {
			return err
		}
		if replymg == nil {
			Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
			continue
		}
		Messagesession.Addreply(replymg.MessageID)

		newCap, err := strconv.Atoi(replymg.Text)
		retry++
		if err != nil {
			Messagesession.SendAlert(C.GetMsg(C.MsgValidInt), nil)
			continue
		}
		if newCap <= 0 {
			Messagesession.SendAlert(C.GetMsg(C.Msgcapzerod), nil)
			continue
		}

		

		if C.Bwidth(newCap).GbtoByte() > capble_quota {
			Messagesession.SendAlert(C.GetMsg(C.MsgcapAlready), nil)
			continue
		}
		// if C.Bwidth(newCap).GbtoByte() > Usersession.TotalUsage() {
		// 	Messagesession.SendAlert(C.Getmsg(MsgcapUsage), nil)
		// 	continue
		// }
		Newcap = C.Bwidth(newCap).GbtoByte()
		break

	}

	btns.Reset([]int16{1, 1})
	btns.Addbutton(C.BtnConform, C.BtnConform, "")
	btns.Addcancle()

	Messagesession.EditText(C.GetMsg(C.MsgcapAlready), btns)

	answer, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID())
	if err != nil {
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
		return nil
	}

	switch answer.Data {
	case C.BtnCancle:
		Messagesession.DeleteAllMsg()
		Messagesession.SendAlert(C.GetMsg(C.MsgcapAlready), nil)
		return nil

	}

	Usersession.GetUser().IsCapped = true
	Usersession.GetUser().CappedQuota = Newcap
	Usersession.GetUser().Captime = time.Now()

	if err = u.ctrl.RecalculateConfigquotas(upx.User.User); err != nil {
		Messagesession.SendAlert(C.GetMsg(C.MsgcapRecalFail), nil)
	}

	if err = Usersession.Close(); err != nil {
		if errors.Is(err, C.ErrContextDead) {
			Messagesession.SendAlert(C.GetMsg(C.MsgContextDead), nil)
			return C.ErrContextDead
		} else {
			Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
			return C.ErrDbopration
		}
	}

	Messagesession.DeleteAllMsg()
	Messagesession.SendAlert(C.GetMsg(C.MsgcapSuccses), nil)

	return nil
}

func (u *Usersrv) commandReffral(upx *update.Updatectx) error {
	Messagesession := botapi.NewMsgsession(u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	refred, refverified, err := u.ctrl.ReffralCount(upx.User.TgID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Messagesession.SendAlert(C.GetMsg(C.MsgcapAlready), nil)
		}
	}

	btns := botapi.NewButtons([]int16{2})
	btns.AddBtcommon(C.BtnClaim)
	btns.AddBtcommon(C.BtnGetLink)
	btns.AddClose(true)

	// check loop begining
	var firstsend bool = true

	for i := 0; i < 5; i++ { //max press 5

		if upx.Ctx.Err() != nil {
			tmpctx, cancle := context.WithTimeout(u.ctx, 1*time.Minute)
			Messagesession.SetNewcontext(tmpctx)
			Messagesession.DeleteAllMsg()
			Messagesession.SendAlert("context dead, session over", nil)
			cancle()
			return err
		}

		if firstsend {
			Messagesession.Edit(struct {
				*botapi.CommonUser
				Refred   string
				Verified string
			}{
				CommonUser: &botapi.CommonUser{
					Name:     upx.User.Name,
					Username: upx.FromChat().UserName,
					TgId:     upx.User.TgID,
				},
				Refred:   strconv.Itoa(int(refred)),
				Verified: strconv.Itoa(int(refverified)),
			}, btns, C.TmpRefHome)
		}

		callback, err := u.callback.GetcallbackContext(upx.Ctx, btns.ID())
		if err != nil {
			return err
		}

		switch callback.Data {

		case C.BtnClaim:

			if refred == 0 {
				Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgRefNoANyUser), true)
				continue
			}

			btns.Reset([]int16{2})
			btns.Addbutton("Claim Verified", "Claim Verified", "")
			btns.Addbutton("claim All", "claim All", "")
			btns.AddClose(true)

			Messagesession.Edit(C.GetMsg(C.MsgRefClaimNote), btns, "")

			if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
				return err
			}

			switch callback.Data {
			case "Claim Verified":
				if refverified == 0 {
					Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgRefNousers), true)
					continue
				}

				btns.Reset([]int16{2})
				btns.Addbutton(C.BtnConform, C.BtnConform, "")
				btns.Addbutton(C.BtnCancle, C.BtnCancle, "")

				Messagesession.Edit(C.GetMsg(C.MsgRefClaimConform), btns, "")

				if callback, err = u.callback.GetcallbackContext(upx.Ctx, btns.ID()); err != nil {
					return err
				}
				if err = checkconform(callback.Data, Messagesession); err != nil {
					Messagesession.SendAlert(C.GetMsg(C.MsgClaimCancle), nil)
					return nil
				}

				newpoints, err := u.ctrl.ClaimReferVerified(upx.User.TgID)
				if err != nil {
					Messagesession.Edit(C.GetMsg(C.MsgRefClaimError), nil, "")
					return nil
				}
				Messagesession.Edit(fmt.Sprintf(C.GetMsg(C.MsgRefClaimed), newpoints), nil, "")
				return nil

				//u.ctrl.UrlTestOut()

			case "claim All":
				Messagesession.SendAlert(C.GetMsg(C.MSgRefClaimAllunsupport), nil)

			}

		case C.BtnGetLink:
			Messagesession.DeleteAllMsg()

			btns.Reset([]int16{1})
			btns.AddUrlbutton("Connected Bot", u.ctrl.Botlink+"?start="+strconv.Itoa(int(upx.User.TgID)))

			Messagesession.SendAlert(fmt.Sprintf(C.GetMsg(C.MsgRefLink), u.ctrl.Botlink+"?start="+strconv.Itoa(int(upx.User.TgID))), nil)

			_, err := Messagesession.SendExtranal(struct {
				Botlink string
				*botapi.CommonUser
			}{
				CommonUser: &botapi.CommonUser{
					Name:     upx.User.Name,
					Username: upx.Chat.UserName,
					TgId:     upx.User.TgID,
				},
				Botlink: u.ctrl.Botlink + "?start=" + strconv.Itoa(int(upx.User.TgID)),
			}, btns, C.TmpRefshare, true)

			return err

		case C.BtnClose:
			Messagesession.DeleteLast()
			return nil
		}

		firstsend = false

	}

	return nil
}

// TODO: implemet this function later
func (u *Usersrv) commandContact(upx *update.Updatectx) error {
	// Create contact session here
	upx.Ctx, upx.Cancle = context.WithTimeout(u.ctx, 2*time.Minute)

	Messagesession := botapi.NewMsgsession(u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)
	Messagesession.SendAlert(`
	⏳ You have 2 minutes of chat time!
	If an admin is online, they'll reply within this time. If not, don't worry—they'll get back to you as soon as possible.
	💡 If you’d like to cancel this chat, simply send /cancel at any time.
	
	`, nil)

	timeovermg := C.GetMsg(C.GetMsg(C.MsgContactTimeover))
	for {
		if upx.Ctx.Err() != nil {
			break
		}
		msg, err := u.defaultsrv.ExcpectMsg(upx.User.Id, upx.FromChat().ID)
		if err != nil {
			break
		}
		if msg.Text == "/cancel" {
			timeovermg = C.GetMsg(C.MsgContactCancle)
			break
		}
		//Messagesession.ForwardMgTo(u.ctrl.SudoAdmin, int64(msg.MessageID))
		u.ctrl.Addquemg(upx.Ctx, &botapi.Msgcommon{
			Infocontext: &botapi.Infocontext{
				ChatId: u.ctrl.SudoAdmin,
			},
			Text: fmt.Sprintf("%v,\n@%v\n\n message: \n\n%v", upx.User.TgID, upx.Chat.UserName, msg.Text),
		})
		if msg.Text == "" {
			Messagesession.CopyMessageTo(u.ctrl.SudoAdmin, int64(msg.MessageID))
		}
		
	}

	u.ctrl.Addquemg(u.ctx, &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId: upx.User.TgID,
		},
		Text: timeovermg,
	})
	return nil

}

func (u *Usersrv) commandSuggesion(upx *update.Updatectx) error {

	Messagesession := botapi.NewMsgsession(u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)
	_, err := Messagesession.SendAlert(C.GetMsg(C.MsgSugess), nil)
	if err != nil {
		Messagesession.SendAlert(C.GetMsg(C.Msgwrong), nil)
	}
	repmsg, err := u.defaultsrv.ExcpectMsg(upx.User.Id, upx.FromChat().ID)

	if err != nil {
		return err
	}

	u.ctrl.Addquemg(upx.Ctx, &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId: u.ctrl.SudoAdmin,
		},
		Text: "msg from user  " + upx.User.Name + " UserName @" + upx.FromChat().UserName + " userid " + strconv.Itoa(int(upx.User.TgID)) + " sugess  msg := " + repmsg.Text,
	})

	Messagesession.SendAlert(C.GetMsg(C.Msgsugessdone), nil)
	return nil
}

func (u *Usersrv) Canhandle(upx *update.Updatectx) (bool, error) {
	return upx.Service == C.Userservicename, nil
}

func (u *Usersrv) Recheck(upx *update.Updatectx) error {
	var userMessage string
	
	if upx.User.Isverified() {
		userMessage = "already verified"
		return nil
	} else {
		err := u.ctrl.RefreshUser(upx.Ctx, upx.Dbuser())
		userMessage = "rechecking verificity done"
		if err != nil {
			userMessage = "rechecking verificity failed"
		}
	}
	
	u.ctrl.Addquemg(upx.Ctx, &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId: upx.User.TgID,
		},
		Text: userMessage,
	})
	return nil
}

func (u *Usersrv) Name() string {
	return C.Userservicename
}
