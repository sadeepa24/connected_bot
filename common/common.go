package common

import (
	//
	"errors"
	"fmt"
	"strconv"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"go.uber.org/zap"
)

type Sendreciver func(msg any) (*tgbotapi.Message, error) //sendreciver should support only recive mode (when msg == nil)
type Callbackreciver func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error)
type Alertsender func(msg string)

type Tgcalls struct {
	Callbackreciver Callbackreciver
	Sendreciver     Sendreciver
	Alertsender     Alertsender
}

type OptionExcutors struct {
	//Common
	Tgcalls
	Upx             *update.Updatectx
	Btns            *botapi.Buttons
	Usersession     *controller.CtrlSession
	MessageSession  *botapi.Msgsession
	Ctrl            *controller.Controller
	Logger 			*zap.Logger

	//For Exec Rule addr
}

type Initer interface {
	Init() 
}

func ReciveString(call Tgcalls) (string, error) {
	var(
		replymeassage *tgbotapi.Message
		err error
		confName string
	) 

	for {

		if replymeassage, err = call.Sendreciver(nil); err != nil {
			return "", err
		}
		if replymeassage.IsCommand() {
			call.Alertsender("Send Valid String Not Commands")
			continue
		}
		confName = replymeassage.Text
		if replymeassage.Text == "" {
			confName = "noname"
		}

		break

	}

	return confName, nil
	
}

func ReciveInt(call Tgcalls, max, min int) (int, error) {
	var (
		retry int
		replymeassage *tgbotapi.Message
		err error
		out int
	)
	for {
		
		if retry > 5 {
			call.Alertsender(C.GetMsg(C.Msgretryfail))
			return 0, errors.New("retry attemps failed")
		}
		if replymeassage, err = call.Sendreciver(nil); err != nil {
			return 0, err
		}
		if replymeassage == nil {
			continue
		}
		if replymeassage.IsCommand() {
			if replymeassage.Command() == C.CmdCancel {
				call.Alertsender("canceld")
				return 0, errors.New("user canceld value sending")
			}
			call.Alertsender("send valid value or cancel command")
			continue
		}
		if out, err = strconv.Atoi(replymeassage.Text); err != nil {
			call.Alertsender(C.GetMsg(C.MsgValidInt))
			continue
		}
		 if out > max || out <= min {
			call.Alertsender(fmt.Sprintf("int should be between %d, and %d", min, max))
			continue
		 }

		break

	}
	
	return out, nil
}

//recived as GB,
//max, min should be in byte format,
//return as GB,
func ReciveBandwidth(call Tgcalls, max, min C.Bwidth) (C.Bwidth, error) {
	bwith := C.Bwidth(0)
	var err error
	var replymg *tgbotapi.Message
	
	var retry int
	
	for {
		//bth, err := ReciveInt(call, 100000, 0)
		retry++
		if retry > 5 {
			call.Alertsender("yep you have been succses fully prove that you are real idiot")
			return 0, errors.New("user is an idiot")
		}
		if replymg, err = call.Sendreciver(nil); err != nil {
			return 0, err
		}
		if replymg.IsCommand() {
			if replymg.Command() == C.CmdCancel {
				call.Alertsender("canceld")
				return 0, errors.New("user canceld value sending")
			}
			call.Alertsender("send valid value or cancel command")
			continue
		}
		bwith, err = C.ParserBwidth(replymg.Text)
		if err != nil {
			call.Alertsender("recheck your inputs")
			continue
		}
		if bwith > max || bwith <= min {
			call.Alertsender(fmt.Sprintf("bandwidth should be between %s, and %s", min.BToString(), max.BToString()))
			continue
		}
		break

	}

	return bwith.BytetoGB(), nil

}
