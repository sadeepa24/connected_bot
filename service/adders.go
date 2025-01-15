package service

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofrs/uuid"
	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/builder"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	option "github.com/sadeepa24/connected_bot/sbox_option/v1"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
)

func tlsconstruct(sendreciver common.Sendreciver, callback common.Callbackreciver, btns *botapi.Buttons) (*option.OutboundTLSOptionsContainer, error) {
	TLS := &option.OutboundTLSOptions{
		Enabled:  true,
		Insecure: true,
	}
	mg, err := sendreciver("send your sni")
	if err != nil {
		return nil, err
	}
	TLS.ServerName = mg.Text

	if mg, err = sendreciver("send tls max version"); err != nil {
		return nil, err
	}
	TLS.MaxVersion = mg.Text

	return &option.OutboundTLSOptionsContainer{
		TLS: TLS,
	}, nil
}

type name interface {
	Name() string
}

type outwizard interface {
	name
	Excute() (*option.Outbound, error)
}

type transportwizard interface {
	name
	Excute() (*option.V2RayTransportOptions, error)
}

type vlesswiz struct {
	btns            *botapi.Buttons
	sendreciver     common.Sendreciver
	callback        common.Callbackreciver
	alertsender     common.Alertsender
	transportadders []transportwizard
}

func (v *vlesswiz) Name() string {
	return C.Vless
}

func (v *vlesswiz) Excute() (*option.Outbound, error) {

	var (
		mg  *tgbotapi.Message
		err error
	)

	mg, err = v.sendreciver("please send a new tagName for vless outbound (should be uniq)")
	if err != nil {
		return nil, err
	}

	outbound := &option.Outbound{
		Type: C.Vless,
		VLESSOptions: option.VLESSOutboundOptions{
			ServerOptions: option.ServerOptions{},
		},
	}
	outbound.Tag = mg.Text

	if mg, err = v.sendreciver("send yopu ip addr"); err != nil {
		return nil, err
	}

	if _, err = netip.ParseAddr(mg.Text); err != nil {
		v.alertsender("ip addr error")
		return nil, errors.New("ip addr recive error")
	}
	outbound.VLESSOptions.Server = mg.Text
	if mg, err = v.sendreciver("send yopu port"); err != nil {
		return nil, err
	}
	var port int
	if port, err = strconv.Atoi(mg.Text); err != nil {
		v.alertsender("send valid port cancling wizard")
		return nil, errors.New("port error")
	}
	outbound.VLESSOptions.ServerPort = uint16(port)

	v.btns.Reset([]int16{2})
	v.btns.AddBtcommon("generate")
	v.btns.AddBtcommon("send your one")
	var callback *tgbotapi.CallbackQuery
	if callback, err = v.callback("uuid option", v.btns); err != nil {
		return nil, err
	}
	var uid = uuid.UUID{}

	switch callback.Data {
	case "generate":
		if uid, err = uuid.NewV4(); err != nil {
			v.alertsender("uuid error")
			return nil, err
		}

	case "send your one":
		if mg, err = v.sendreciver("send yopu uuid"); err != nil {
			return nil, err
		}
		if uid, err = uuid.FromString(mg.Text); err != nil {
			v.alertsender("send valid uuid cancling wizard")
			return nil, errors.New("uuid error")
		}

	}
	outbound.VLESSOptions.UUID = uid.String()

	v.btns.Reset([]int16{2})
	v.btns.AddBtcommon("true")
	v.btns.AddBtcommon("false")
	callback, err = v.callback("Do you want to enable tls", v.btns)
	if err != nil {
		return nil, err
	}

	switch callback.Data {
	case "true":
		tlsob, err := tlsconstruct(v.sendreciver, v.callback, v.btns)
		if err != nil {
			return nil, err
		}
		outbound.VLESSOptions.TLS = tlsob.TLS
	default:
		break
	}

	v.btns.Reset([]int16{2})
	for _, tra := range v.transportadders {
		v.btns.AddBtcommon(tra.Name())
	}

	if callback, err = v.callback("select transport", v.btns); err != nil {
		return nil, err
	}

	for _, tra := range v.transportadders {
		if tra.Name() == callback.Data {
			if transopt, err := tra.Excute(); err != nil {
				// v.send.Edit("transport create failed", nil, "")
				return nil, err
			} else {
				outbound.VLESSOptions.Transport = transopt
			}

			break
		}
	}

	return outbound, nil
}

type websocketwiz struct {
	sendrec common.Sendreciver
}

func (w *websocketwiz) Name() string {
	return "ws"
}

func (w *websocketwiz) Excute() (*option.V2RayTransportOptions, error) {
	websocet := &option.V2RayTransportOptions{
		Type: "ws",
		WebsocketOptions: option.V2RayWebsocketOptions{
			Headers: option.HTTPHeader{},
		},
	}
	mg, err := w.sendrec("send path for ws ex:=/")
	if err != nil {
		return nil, err
	}
	websocet.WebsocketOptions.Path = mg.Text

	if mg, err = w.sendrec(" send websocket host"); err != nil {
		return nil, err
	}
	websocet.WebsocketOptions.Headers["host"] = option.Listable[string]{mg.Text}

	return websocet, err
}

func alltransportadder(sendrec common.Sendreciver, callback common.Callbackreciver, alertsender common.Alertsender, btns *botapi.Buttons) []transportwizard {
	transport := []transportwizard{} //TODO: change later
	transport = append(transport, &websocketwiz{
		sendrec: sendrec,
	})
	return transport
}

func alloutadders(sendrec common.Sendreciver, callback common.Callbackreciver, alertsender common.Alertsender, btns *botapi.Buttons, tra []transportwizard) []outwizard {
	outwizard := []outwizard{} //TODO: change later
	//transpoprts := alltransportadder(sendrec, callback, btns)
	outwizard = append(outwizard, &vlesswiz{
		btns:            btns,
		sendreciver:     sendrec,
		callback:        callback,
		alertsender:     alertsender,
		transportadders: tra,
	})

	return outwizard
}

var listableRulesFields = map[string]string{
	"client":         "Client",
	"auth_user":      "AuthUser",
	"protocol":       "Protocol",
	"source_ip_cidr": "Source_ip_cidr",
	"domain":         "Domain",
	"domain_suffix":  "DomainSuffix",
	"domain_keyword": "DomainKeyword",
	"domain_regex":   "DomainRegex",
	"user":           "User",
}

const dnsrule string = "dnsRule"
const routerule string = "routerule"

type RuleSetter interface {
	SetOut(out string)
	SetList(name string, list []string) error
}

// pointer of rule object
func CreateRule(opts common.OptionExcutors, builder *builder.BuildConfig, ruleType string, rule any) (any, error) {

	var mg *tgbotapi.CallbackQuery
	var err error
	var setout bool

	for {
		opts.Btns.Reset([]int16{2})
		for fieldname := range listableRulesFields {
			opts.Btns.AddBtcommon(fieldname)
		}
		var configmarshl []byte
		switch ruler := rule.(type) {
		case *option.Rule:
			configmarshl, err = ruler.MarshalJSON()
			opts.Btns.AddBtcommon("⚙️ set outbound ⚙️")

		case *option.DNSRule:
			configmarshl, err = ruler.MarshalJSON()
			opts.Btns.AddBtcommon("⚙️ set dns server ⚙️")
		}
		opts.Btns.AddBtcommon("⚙️ Done ⚙️")
		opts.Btns.AddBtcommon("⚙️ set action ⚙️")
		opts.Btns.AddClose(true)

		if err != nil {
			configmarshl = []byte("select what do you want")
		}
		if mg, err = opts.Callbackreciver(botapi.Htmlstring(fmt.Sprintf("<pre><code>%s</code></pre>", string(configmarshl))), opts.Btns); err != nil {
			return rule, err
		}

		switch mg.Data {
		case C.BtnClose:
			return rule, C.ErrBtnClosed
		case "⚙️ Done ⚙️":
			if !setout {
				opts.MessageSession.Callbackanswere(mg.ID, "you can't done because you didn't set out", true)
				continue
			}
			return rule, nil

		case "⚙️ set outbound ⚙️":
			opts.Btns.Reset([]int16{2})
			for _, outname := range builder.GetAllOutNames() {
				opts.Btns.AddBtcommon(outname)
			}
			if mg, err = opts.Callbackreciver("select outbound fo this route rule", opts.Btns); err != nil {
				return rule, err
			}
			rule.(RuleSetter).SetOut(mg.Data)
			setout = true
		case "⚙️ set dns server ⚙️":
			opts.Btns.Reset([]int16{2})
			for _, dnssserver := range builder.GetDnsServers() {
				opts.Btns.AddBtcommon(dnssserver.Tag)
			}
			if mg, err = opts.Callbackreciver("select outbound fo this route rule", opts.Btns); err != nil {
				return rule, err
			}
			rule.(RuleSetter).SetOut(mg.Data)
			setout = true
		case "⚙️ set action ⚙️":
			opts.MessageSession.Callbackanswere(mg.ID, "action will avalble after relese sing-box latest for v1.11", true)
			//TODO: build later after sing-box 1.11 latest
		default:
			reply, err := opts.Sendreciver(`
			🚀 Send a comma-separated list of ` + mg.Data + `.
			🛑 If you want to cancel, just type /cancel.
			✨ Want to clear the list? Simply send .
			⚠️ Important: Please double-check your inputs!
			For example: Avoid sending invalid strings like hello.com for source_ip_cidr. Incorrect inputs will make the configuration invalid! 🚫 
			⚠️ Warning: If you send an incorrectly formatted list, it will be added to the config, and the configuration won't work.
			✅ Make sure to send the exact, properly formatted comma-separated list!
			`)
			if err != nil {
				return rule, err
			}
			list := strings.Split(reply.Text, ",")
			if err = rule.(RuleSetter).SetList(mg.Data, list); err != nil {
				opts.MessageSession.SendAlert("failed creation err = "+err.Error(), nil)
			}
		}

	}
}

func CreateRuleSet(opts common.OptionExcutors, builder *builder.BuildConfig,) (option.RuleSet, error) {
	var callback *tgbotapi.CallbackQuery
	var reply *tgbotapi.Message
	var err error
	var set option.RuleSet
	
	if reply, err = opts.Sendreciver("send a tag name for this rule set tag name should be uniq"); err != nil {
		return set, err
	}
	set.Tag = reply.Text


	opts.Btns.Reset([]int16{2})
	opts.Btns.AddBtcommon("source")
	opts.Btns.AddBtcommon("binary")

	if callback, err = opts.Callbackreciver("select format", opts.Btns); err != nil {
		return set, err
	}
	set.Format = callback.Data

	opts.Btns.Reset([]int16{2})
	opts.Btns.AddBtcommon("inline")
	opts.Btns.AddBtcommon("local")
	opts.Btns.AddBtcommon("remote")

	if callback, err = opts.Callbackreciver("select ruleset type", opts.Btns); err != nil {
		return set, err
	}

	switch callback.Data {
	case "inline":
		set.Type = "inline"
		set.InlineOptions = option.PlainRuleSet{}
		opts.MessageSession.Callbackanswere(callback.ID, "cannot create inline typed rule set yet", true)
		return set, errors.New("cannot create inline typed rule set yet")
	case "local":
		set.Type = "local"
		if reply, err = opts.Sendreciver("send path for local rule set file"); err != nil {
			return set, err
		}
		set.LocalOptions.Path = reply.Text

	case "remote":
		set.Type = "remote"
		set.RemoteOptions = option.RemoteRuleSet{}
		if reply, err = opts.Sendreciver("send url for remote file"); err != nil {
			return set, err
		}
		set.RemoteOptions.URL = reply.Text
	}
	return set, nil
}