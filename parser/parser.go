package parser

import (
	"context"
	"errors"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/service"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)

type Parserwrap interface {
	Parse(tgbotapimsg *tgbotapi.Update) error
	Init() error
	Stop() error
}

type Parser struct {
	ctrl      *controller.Controller
	ctx       context.Context
	services  map[string]service.Service //ordinary services
	logger    *zap.Logger
	Callback  service.Service // special service
	Defaulsrv *service.Defaultsrv
	AdminSrc  *service.Adminsrv
	InlineService *service.InlineService
	srvs      []service.Service
	//baseCtxforUpx context.Context
	//baseCancle    context.CancelCauseFunc
	botapi botapi.BotAPI

	GetBaseCtx func() context.Context

	//usrservice  map[string]bool //for fututre
	//xrayservice map[string]bool
}

func New(
	ctx context.Context,
	ctrl *controller.Controller,
	services []service.Service,
	botapi botapi.BotAPI,
	logger *zap.Logger,

) *Parser {

	// basectx, _ := context.WithCancelCause(ctx)
	parser := &Parser{
		ctx:        ctx,
		ctrl:       ctrl,
		services:   make(map[string]service.Service, len(services)),
		logger:     logger,
		srvs:       services,
		botapi:     botapi,
		GetBaseCtx: ctrl.GetBaseContext, //TODO: change later
		//baseCtxforUpx: basectx,
		//baseCancle:    basecancle,

		//xrayservice: make(map[string]bool, 10), //TODO: create a beter way to select service
		//usrservice:  make(map[string]bool, 10),
	}
	return parser
}

func (p *Parser) Init() error {
	return p.registerservice(p.srvs)
}

func (p *Parser) Stop() error {
	return nil
}

func (p *Parser) registerservice(services []service.Service) error {
	if len(services) <= 0 {
		return errors.New("service count must be greater that zero")
		
	}
	for _, srv := range services {
		switch srv.Name() {
		case C.Callbackservicename:
			p.Callback = srv
			continue
		case C.Defaultservicename:
			p.Defaulsrv = srv.(*service.Defaultsrv)
		case C.Adminservicename:
			p.AdminSrc = srv.(*service.Adminsrv)
		case C.InlineServiceName:
			p.InlineService = srv.(*service.InlineService)
		}
		p.services[srv.Name()] = srv
	}
	return nil
}

func (p *Parser) Parse(tgbotapimsg *tgbotapi.Update) error {

	upx, err := p.Readrequest(tgbotapimsg)
	if err != nil {
		return errors.Join(errors.New("tg request read error from parser"), err)
	}

	if p.ctrl.CheckLock() {
		p.logger.Debug("watchman locked when proc update " + tgbotapimsg.Info())
		//TODO: replace upx context to add addtinal deadline due to Db refresh time,
		// Crucial FOr Updates Like CHatmember
		if upx.Update.ChatMember != nil {
			upx.Ctx, upx.Cancle = context.WithTimeout(p.GetBaseCtx(), 2 * time.Second) //replace old context because chatmember update must be proceed
		}
	}
	
	if upx.Update.CallbackQuery != nil {
		upx.Setcallback()
		return p.Callback.Exec(upx)
	}

	if upx.Update.InlineQuery != nil {
		return p.InlineService.Exec(upx)
	}

	if upx.Update.Message != nil {
		if p.Defaulsrv.Ismsgrequired(upx.FromUser().ID, upx.FromChat().ID) {
			return p.Defaulsrv.Exec(upx)
		}
	}
	if upx.FromChat().ID == p.ctrl.SudoAdmin {
		if upx.Update.Message.Command() == "switch" {
			p.AdminSrc.SwapMode()
			var mode string 
			if p.AdminSrc.AdminMode() {
				mode = "from User to Admin"
			} else {
				mode = "from Admin to User"
			}
			p.ctrl.Addquemg(context.Background(), &botapi.Msgcommon{
				Infocontext: &botapi.Infocontext{
					ChatId: p.ctrl.SudoAdmin,
					User_id: p.ctrl.SudoAdmin,
				},
				
				Text: "Mode Changed " + mode,
			})
			return nil
		}
		if p.AdminSrc.AdminMode() {
			return p.AdminSrc.Exec(upx)
		}
	}

	var cannprocUpdate bool
	if cannprocUpdate, err = p.Setuser(upx); err != nil { //loads info from database
		if upx.User != nil {
			err = errors.New("Error When Preprosess user " +  upx.User.Info() + err.Error())
		}
		return err
	}
	if !cannprocUpdate {
		if upx.User != nil {
			p.logger.Error("Cannot Continue With Update " +  upx.User.Info())
		}
		return nil
	}

	if upx.Serviceset {
		return p.addtoservice(upx)
	}

	if upx.Update.MyChatMember != nil || upx.Update.ChatMember != nil {
		return p.chatmemberparse(upx)
	}

	upx.SetDrop(true)
	return p.addtoservice(upx)

}

func (p *Parser) Readrequest(tgbotapimsg *tgbotapi.Update) (*update.Updatectx, error) {
	upx := update.Newupdate(p.GetBaseCtx(), tgbotapimsg)

	if upx.Update.InlineQuery != nil {
		return upx, nil
	}

	switch {
	case upx.FromChat() == nil:
		return nil, errors.New("recived update is not from a chat")
	case upx.FromUser() == nil:
		return nil, errors.New("recived update is not from a enduser")
	case upx.FromUser().IsBot:
		return nil, errors.New("user is not a human")
	case !upx.FromChat().IsPrivate():
		if upx.FromChat().ID != p.ctrl.ChannelId && upx.FromChat().ID != p.ctrl.GroupID {
			return nil, errors.New("user from elsewhere group")
		}

	}
	p.logger.Info("user updated recived " + tgbotapimsg.Info())
	//replacing context
	upx.Ctx, upx.Cancle = context.WithTimeout(p.GetBaseCtx(), C.UpdateTimeout)

	return upx, nil
}

func (u *Parser) addtoservice(upx *update.Updatectx) error {
	if upx.Drop() {
		u.logger.Warn("Dropping update not a valid update context")
		upx.Cancle()
		upx = nil
		return nil
	}

	if service, ok := u.services[upx.Service]; ok {
		return service.Exec(upx)
	}
	return C.ErrServiceNotFound

}

func (p *Parser) Setuser(upx *update.Updatectx) (bool, error) {
	var (
		ok  bool
		err error
	)
	upx.Chat = upx.FromChat()
	upx.Chat_ID = upx.FromChat().ID

	if upx.Update.Message != nil {
		if upx.Update.Message.IsCommand() {
			var servicenm string
			upx.Command, servicenm, err = p.commandparser(upx.Update.Message)
			if err != nil {
				upx.SetDrop(true)
				return false, err
			}

			upx.Setservice(servicenm)
		} else {
			//Already checked Is message required by Default service as reply to question
			return false, C.ErrUpdateFaile
		}
	}

	if upx.User, ok, err = p.ctrl.GetUser(upx.FromUser()); err != nil {
		return false, err
	}

	if !ok {
		upx.User, err = p.ctrl.Newuser(upx.FromUser(), upx.FromChat())
		if err != nil {
			upx.SetDrop(true)
			return false, err
		}
		p.logger.Info("New user added to DB " + upx.User.Info() )
		upx.Setservice(C.Userservicename)

	}
	if upx.User.IsMonthLimited && (upx.Update.Message != nil) && !upx.IsCommand(C.CmdBuild) {
		p.ctrl.Addquemg(upx.Ctx, &botapi.Msgcommon{
			Infocontext: &botapi.Infocontext{
				ChatId: upx.User.TgID,
			},
			Text: C.GetMsg(C.MsgUserMonthLimited),
		})
		return false, nil
	}

	if upx.Dbuser().RecheckVerificity {
		p.logger.Info("rechecking verificty " + upx.User.Info())

		var (
			err1 error
			err2 error
			is   bool
		)
		if !upx.User.IsInChannel {
			if _, is, err1 = p.botapi.GetchatmemberCtx(upx.Ctx, upx.User.Id, p.ctrl.ChannelId); is {
				upx.User.IsInChannel = true
			}
		}

		if !upx.User.IsInGroup {
			if _, is, err2 = p.botapi.GetchatmemberCtx(upx.Ctx, upx.User.Id, p.ctrl.GroupID); is {
				upx.User.IsInGroup = true
			}
		}
		if err1 == nil && err2 == nil {
			upx.Dbuser().RecheckVerificity = false
		}
	}

	switch upx.Command {
	case C.CmdStart, C.CmdHelp, C.CmdNull, C.CmdContact, C.CmdRecheck, C.CmdSource, C.CmdFree:
		break
	default:
		if !upx.Update.FromChat().IsPrivate() {
			//return C.ErrUserIsNotinPrivate
			return false, nil
		}
		if upx.User.Templimited {
			//return C.ErrUserTempLimited
			p.ctrl.Addquemg(upx.Ctx, &botapi.Msgcommon{
				Infocontext: &botapi.Infocontext{
					ChatId: upx.User.TgID,
				},
				Text: C.GetMsg(C.MsgTempLimitAlert),
			})
			return false, nil
		}
		if upx.User.Restricted {
			p.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
				DestinatioID: upx.User.TgID,
				TemplateName: "restricted",
				Lang:         upx.User.Lang,
				Template: struct {
					*botapi.CommonUser
				}{
					CommonUser: &botapi.CommonUser{
						Name:     upx.User.Name,
						Username: upx.Chat.UserName,
						TgId:     upx.User.TgID,
					},
				},
			})
			return false, nil
		}
		if !upx.User.Isverified() {
			p.ctrl.Addquemg(upx.Ctx, botapi.UpMessage{
				DestinatioID: upx.User.TgID,
				TemplateName: C.TmplCommonUnverified,
				Lang:         upx.User.Lang,
				Template: struct {
					*botapi.CommonUser
				}{
					CommonUser: &botapi.CommonUser{
						Name:     upx.User.Name,
						Username: upx.Chat.UserName,
						TgId:     upx.User.TgID,
					},
				},
			})
			return false, nil
			
		}
		if upx.User != nil {
			p.logger.Error("Error When Preprosess user command" +  upx.User.Info(), zap.Error(err))
			return false, nil
		}
	}

	return true, nil
}

func (p *Parser) chatmemberparse(upx *update.Updatectx) error {
	if upx.Update.ChatMember != nil || upx.Update.MyChatMember != nil {
		upx.Setservice(C.Userservicename)
		return p.addtoservice(upx)
	}
	return errors.New("upx passed to chat member, chatmember objects are empty(nil)")
}

// return command, service, error
func (p *Parser) commandparser(msg *tgbotapi.Message) (string, string, error) {
	//TODO: remove this switch and create good way to select service
	switch msg.Command() {
	case C.CmdStart, C.CmdFree,  C.CmdHelp, C.CmdGift, C.CmdRecheck, C.CmdCap, C.CmdDistribute, C.CmdRefer, C.CmdEvents, C.CmdSugess, C.CmdPoints, C.CmdContact, C.CmdSource:
		return msg.Command(), C.Userservicename, nil
	case C.CmdCreate, C.CmdStatus, C.CmdConfigure, C.CmdInfo, C.CmdBuild:
		return msg.Command(), C.Xraywizservicename, nil
	default:
		return msg.Command(), C.Defaultservicename, C.ErrCommandNotfound
	}
}
