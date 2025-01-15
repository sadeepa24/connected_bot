package events

import (
	"os"
	"strconv"
	"strings"
	"time"

	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
)

const (
	aplusconfpath = "./configs/aplusconf.json"
)

type Aplusconf struct {
	duration int    //duration in days
	makeDate string // init date
	price    int64  //how much point is  nedded to this event

	ctrl *controller.Controller
}

var _ Event = &Aplusconf{}

func (a *Aplusconf) Name() string {
	return "Aplus Working Config"
}
func (a *Aplusconf) Info() string {
	return `
	You can watch Aplus Ewings videos effortlessly with this config! 🎥✨
	
	All Cloudflare sites work seamlessly without a VPN 🌐. You can confirm this by checking your IP on ip.sb or speedtest.net 📊.
	
	Enjoy unlimited streaming, downloading, and uploading on websites or services using Cloudflare CDN 🚀📥📤!
	
	`
}
func (a *Aplusconf) Expired() bool {
	//cannot be an error
	makedate, _ := time.Parse("2006-02-02 15:04:05", a.makeDate)
	bo := time.Now().After(makedate.AddDate(0, 0, a.duration))
	return bo
}
func (a *Aplusconf) Price() int64 {
	return a.price
}

func (a *Aplusconf) strPrice() string {
	return strconv.Itoa(int(a.price))
}

func (a *Aplusconf) Excute(eventctx Eventctx) error {

	btns := eventctx.Btns

	btns.Reset([]int16{2})
	btns.AddBtcommon(C.BtnClaim)
	//btns.AddBtcommon("set cf bug host")
	btns.AddClose(false)

	callback, err := eventctx.Callbackreciver(`
	
	you have available `+strconv.Itoa(int(eventctx.Upx.User.Points))+`
	to claim this event, you need `+strconv.Itoa(int(a.Price()))+` points
	
	`+"\nEvent Info - "+a.Info(), btns)

	if err != nil {
		return err
	}

	switch callback.Data {

	case C.BtnClaim:
		if eventctx.Upx.User.Points < a.price {
			eventctx.Callbackreciver("you can't claim this because you don't have enougf points, use reffral system to earn points ", nil)
			return nil
		}
		btns.Reset([]int16{})
		btns.AddBtcommon(C.BtnCancle)
		btns.AddBtcommon(C.BtnConform)

		if callback, err = eventctx.Callbackreciver("Are you sure about this?, to claim this you have to spent "+a.strPrice(), btns); err != nil {
			return nil
		}

		if callback.Data == C.BtnCancle {
			eventctx.Callbackreciver("cancled", nil)
			return nil
		}
		if a.ctrl.AddEvent(eventctx.Upx.User.TgID, a.Name()) != nil {
			eventctx.Alertsender("somthing went wrong please try again")
		}
		a.ctrl.UpdatePoint((eventctx.Upx.User.Points - a.price), eventctx.Upx.User.TgID)
		return a.ExcuteComplete(eventctx)
		//builder.NewBuilder()
	case C.BtnClose:
		eventctx.Callbackreciver("closed", nil)
		return nil
	}

	return nil
}

func (a *Aplusconf) ExcuteComplete(eventctx Eventctx) error {
	btns := eventctx.Btns

	btns.Reset([]int16{1})
	btns.AddBtcommon("Add to Builder")
	btns.AddBtcommon("cancel")
	var premg string
	olConf, _ := a.ctrl.GetSpecificConf(eventctx.Upx.User.TgID, a.Name())
	if olConf.Name == a.Name() {
		btns.Reset([]int16{1})
		btns.AddBtcommon("replace current")
		btns.AddBtcommon("cancel")
		// callback, err = eventctx.Callbackreciver("you already added this config\n event info = " + a.Info(), btns)
		// if err != nil {
		// 	return err
		// }
		premg = "you already added this config\n\n event info = "

	}
	callback, err := eventctx.Callbackreciver(premg+a.Info(), btns)
	if err != nil {
		return err
	}
	if callback.Data == "cancel" {
		return nil
	}
	srcFile, err := os.ReadFile(aplusconfpath)
	if err != nil {
		eventctx.Callbackreciver("src config file error", nil)
		return err
	}
	confname := strconv.Itoa(int(eventctx.Upx.User.TgID)) + "-" + a.Name() + ".json"

	dstFile, err := os.OpenFile("./configs/"+confname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		eventctx.Callbackreciver("src config file error", nil)
		return err
	}
	defer dstFile.Close()

	bughost, err := eventctx.Sendreciver(`
	⚠️ Important Notice
	Please send your Cloudflare Bug Host IP
	If you provide an incorrect IP address, it could mess up the entire configuration.
	
	💡 Don't worry—you can manually add or correct it later if needed.
	`)

	if err != nil {
		return err
	}
	srcFile = []byte(strings.ReplaceAll(string(srcFile), "<bughost>", bughost.Text))

	if _, err = dstFile.Write(srcFile); err != nil {
		eventctx.Alertsender("config creation failed")
		return err
	}

	if olConf.Name != a.Name() {
		a.ctrl.CreateSboxConf(eventctx.Upx.User.TgID, a.Name())
	}

	eventctx.Callbackreciver(`
	✅ Success! AplusWorking Config Added
	
	To start using this config, you need to add at least one outbound. Follow these steps carefully:
	
	1️⃣ Send the /buildconf command.
	2️⃣ Select the  `+a.Name()+` config.
	3️⃣ Press the Outbound button.
	4️⃣ Choose an option to add an outbound (we recommend loading it from your existing configs).
	
	⚠️ Important:
	
	Do NOT change our DNS server—it’s crucial for the config to work as expected.
	Avoid modifying other settings like DNS, Routing, or Inbounds unless you're confident about what you're doing.
	
	`, nil)

	//a.ctrl.

	return nil
}
