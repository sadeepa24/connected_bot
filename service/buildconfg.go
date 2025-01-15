package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/builder"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	option "github.com/sadeepa24/connected_bot/sbox_option/v1"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	singconst "github.com/sagernet/sing-box/constant"
	"gorm.io/gorm"
)

type BuildState struct {
	ctx            context.Context
	State          int
	Messagesession *botapi.Msgsession
	//upx            *update.Updatectx
	dbuser 			*db.User
	userID 			int64 //user TG id
	btns           *botapi.Buttons
	wiz            *Xraywiz

	lastcallback *tgbotapi.CallbackQuery
	Builder      *builder.BuildConfig

	lastSelectedOut  string
	lastSelectIn     string
	lastSelectDnssrv string
	lastSelectRuleSet string
	lastconfig       string

	otadders []outwizard
	changers []changer

	sendreciver     common.Sendreciver
	callbackreciver common.Callbackreciver
	alertsender     common.Alertsender
}



func (b *BuildState) run() error {
	
	var err error
	main:
	for {

		if b.ctx.Err() != nil {
			return nil
		}

		if b.lastcallback != nil {
			if b.lastcallback.Data == C.BtnClose {
				b.Messagesession.EditText("builder closed", nil)
				break
			}
		}

		switch b.State {
		case initiate:
			err = b.home()
		case createconfig:
			err = b.createConfig()
		case gotoconfig:
			err = b.confighandle()
		case inbound:
			err = b.inbound()
		case outbound:
			err = b.outbound()
		case dns:
			err = b.dns()
		case route:
			err = b.route()
		case outhandler:
			err = b.outhandler()
		case allOutbound:
			err = b.allout()
		case addOutbound:
			err = b.addout()
		case handlein:
			err = b.handlein()
		case addinbound:
			err = b.addin()
		case dnsRules:
			err = b.dnsrules()
		case dnsrvhandle:
			err = b.dnserverhandle()
		case dnssrvs:
			err = b.dnsserver()
		case dnssrvaddr:
			err = b.dnsadder()
		case routeRules:
			err = b.routeRules()
		case rtruleadd:
			err = b.ruleAdd()
		case rtsets:
			err = b.routeruleset()
		case experimental:
			err = b.experimental()
		case clash_api:
			err = b.clashapi()
		case rulesetadd:
			err = b.ruleSetAdd()
		case rulesetchg:
			err = b.ruleSetChg()
		case clash_apichg:
			err = b.change_clashapi()
		case cache_file:
			err = b.cache_file()
		case cache_file_chg:
			err = b.cacheFileChange()

		//case confdelete:
		case buildclose:
			b.Messagesession.EditText("builder closed", nil)
			break main

		default:
			err = b.home()
		}

		if err != nil {
			break
		}

	}
	return nil
}

const (
	initiate     = 0
	createconfig = 1
	gotoconfig   = 2

	//confdelete = 7

	dns      = 5
	dnsRules = 17

	inbound       = 3
	handlein         = 8
	addinbound    = 9
	removeInbound = 10
	changeInbound = 11

	outbound   = 4
	buildclose = 19

	allOutbound = 12
	outhandler  = 18
	addOutbound = 13
	//removeOutbound = 14
	//changeOutbound = 15

	route      = 6
	routeRules = 16

	dnssrvs      = 22
	dnsrvhandle  = 21
	dnssrvaddr   = 23
	rtruleadd    = 24
	rtsets       = 25
	experimental = 26
	clash_api    = 27
	rulesetadd = 28
	rulesetchg = 29
	clash_apichg = 30
	cache_file = 31
	cache_file_chg = 32
)

func NewBuildState(userid int64) {

}



func (b *BuildState) inbound() error {
	var err error
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("all inbounds")
	b.btns.AddBtcommon("add inbound")
	b.btns.AddBtcommon("remove inbound")
	b.btns.AddBtcommon("change inbound")
	b.btns.AddCloseBack()

	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportAllInbounds(), b.btns); err != nil {
		return err
	}

	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = gotoconfig
	case "all inbounds":
		

		b.btns.Reset([]int16{2})
		for _, in := range b.Builder.GetAllInNames() {
			b.btns.AddBtcommon(in)
		}

		b.btns.AddCloseBack()

		if b.lastcallback, err =  b.callbackreciver("Select inbound to continue", b.btns); err != nil {
			return err
		}
		switch b.lastcallback.Data {
		case C.BtnClose, C.BtnBack:
			return nil;
		}
		b.State = handlein
		b.lastSelectIn = b.lastcallback.Data

	default:
		b.Messagesession.Callbackanswere(b.lastcallback.ID, "option does not support yet", true)
		b.State = inbound

		// case "all inbound":
		// 	b.State = allIn
		// case "add inbound":
		// 	b.State = addinbound
		// case "remove inbound":
		// 	b.State = removeInbound
		// case "change inbound":
		// 	b.State = changeInbound

	}

	return nil
}

func (b *BuildState) addin() error {

	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("common configs")
	b.btns.AddBtcommon("custom inbound")
	b.btns.AddBtcommon("sbox inbound")
	b.btns.AddCloseBack()

	b.Messagesession.Edit(b.Builder.ExportAllInbounds(), b.btns, "")

	var err error
	if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
		return err
	}

	switch b.lastcallback.Data {
	case "common configs":

	}

	return nil
}

func (b *BuildState) handlein() error {
	for _, handler := range b.changers {
		if handler.Can(b.Builder.GetInType(b.lastSelectIn)) {
			return handler.MakeChange(b.Builder, b.lastSelectIn, func(st int) { b.State = st })
		}
	}
	b.Messagesession.Callbackanswere(b.lastcallback.ID, "there is no change option found for this type inbounds", true)
	b.State = inbound
	return nil
}

func (b *BuildState) outbound() error {

	b.btns.Reset([]int16{2})

	b.btns.AddBtcommon("all outbound")
	b.btns.AddBtcommon("add outbound")
	b.btns.AddBtcommon("remove outbound")
	// b.btns.AddBtcommon("change outbound")
	b.btns.AddCloseBack()

	var err error
	b.Messagesession.Edit(b.Builder.ExportAllOutbounds(), b.btns, "")
	if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
		return err
	}
	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = gotoconfig
	case "all outbound":
		b.State = allOutbound
	case "add outbound":
		b.State = addOutbound

	case "remove outbound":

		b.btns.Reset([]int16{2})

		for _, out := range b.Builder.GetOutNames() {
			b.btns.AddBtcommon(out)
		}
		b.btns.AddCloseBack()

		b.Messagesession.Edit("select outbound to remove", b.btns, "")

		if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnClose:
			return nil
		case C.BtnBack:
			b.State = outhandler
		default:
			if err = b.Builder.RemoveOutbound(b.lastcallback.Data); err != nil {
				b.Messagesession.SendAlert("outbound remove failed", nil)
			} else {
				b.Messagesession.SendAlert("outbound remove succsesfull", nil)
			}

			b.State = outbound
		}

	}

	return nil
}

func (b *BuildState) allout() error {
	b.btns.Reset([]int16{2})

	for _, out := range b.Builder.GetOutNames() {
		if out == "block" || out == "default" || out == "dns-out" {
			continue
		}
		b.btns.AddBtcommon(out)
	}
	b.btns.AddCloseBack()

	b.Messagesession.Edit("all outbound here select one to make chanmges", b.btns, "")

	var err error
	if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
		return err
	}

	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = outbound
	case C.BtnClose:
		return nil
	default:
		b.State = outhandler
		b.lastSelectedOut = b.lastcallback.Data
	}

	return nil
}

func (b *BuildState) outhandler() error {

	for _, handler := range b.changers {
		if handler.Can(b.Builder.OutType(b.lastSelectedOut)) {
			return handler.MakeChange(b.Builder, b.lastSelectedOut, func(st int) { b.State = st })
		}
	}
	b.Messagesession.Callbackanswere(b.lastcallback.ID, "there is no change option found for this type outbounds", true)
	b.State = allOutbound
	return nil

	/*				b.btns.Reset([]int16{2})
					b.btns.AddBtcommon("change sni")
					b.btns.AddBtcommon("change server")
					b.btns.AddBtcommon("remove")
					b.btns.AddCloseBack()

					b.Messagesession.Edit(b.Builder.ExportOutbound(b.lastSelectedOut), b.btns, "")
					var err error
					if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
						return err
					}

					switch b.lastcallback.Data {
					case C.BtnBack:
						b.State  = allOutbound
					case "change sni":

						b.Messagesession.EditText("send new sni ", nil)

						reply, err := b.wiz.defaultsrv.ExcpectMsgContext(b.ctx, b.upx.User.TgID, b.upx.User.TgID)
						if err != nil {
							return err
						}
						if reply.IsCommand() {
							b.Messagesession.SendAlert("send a sni not command", nil)
							return nil
						}
						if reply.Text == "" {
							b.Messagesession.SendAlert("no sni foundplease retry", nil)
						}

						if err = b.Builder.ChangeSni(b.lastSelectedOut, reply.Text); err != nil {
							b.Messagesession.SendAlert("sni change failed", nil)
						}

					}
					return nil
	*/
}

func (b *BuildState) addout() error {
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("common configs")
	b.btns.AddBtcommon("custom outbound")
	b.btns.AddBtcommon("via xray link")
	b.btns.AddBtcommon("load from your config")
	b.btns.AddBtcommon("json outbound")
	b.btns.AddCloseBack()
	b.Messagesession.Edit(b.Builder.ExportAllOutbounds(), b.btns, "")

	var err error
	if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
		return err
	}

	switch b.lastcallback.Data {

	case "custom outbound":
		b.btns.Reset([]int16{})
		for _, otadder := range b.otadders {
			b.btns.AddBtcommon(otadder.Name())
		}

		b.btns.AddCloseBack()

		b.Messagesession.Edit("select you desired outboiund type to add", b.btns, "")
		if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnClose:
			return nil
		case C.BtnBack:
			b.State = addOutbound
		default:
			for _, adder := range b.otadders {
				if adder.Name() == b.lastcallback.Data {
					if aoutbound, err := adder.Excute(); err != nil {
						b.Messagesession.SendAlert("outbound adding failed", nil)
						b.State = addOutbound

					} else {
						if err = b.Builder.AddRawOut(aoutbound); err != nil {
							b.Messagesession.SendAlert("outbound adding failed", nil)
							b.State = addOutbound
						}

					}
					break
				}
			}
		}

	case "load from your config":
		allconfs, err := b.wiz.ctrl.GetUserConfigs(b.userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				b.Messagesession.SendAlert("you don't have any configs to load to create config use /create command", nil)
				return nil

			}
			b.Messagesession.SendAlert("somythin went wrong", nil)
			return nil
		}
		b.btns.Reset([]int16{2})
		C.ExcuteSlice(allconfs, func(t *db.Config) {
			b.btns.AddBtcommon(t.Name)
		})
		b.btns.AddCloseBack()
		b.Messagesession.Edit("select config to load", b.btns, "")

		if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnClose:
			return nil
		case C.BtnBack:
			b.State = addOutbound

		default:
			if b.Builder.CheckOutbound(b.lastcallback.Data) {
				b.Messagesession.Callbackanswere(b.lastcallback.ID, "selected config already loded", true)
				return nil
			}
			conf := C.GetFromSlice(allconfs, func(conf db.Config) bool {
				return conf.Name == b.lastcallback.Data
			})
			if conf != nil {
				in, ok := b.wiz.ctrl.Getinbound(int(conf.InboundID))
				if !ok {
					b.Messagesession.SendAlert("config loading failed", nil)
					return nil

				}
				err = b.Builder.AddOutbound(*conf, in, "")
				if err == nil {
					b.Messagesession.SendAlert("config loading sucsess", nil)
				}
			}

		}

	case "common configs":
		b.btns.Reset([]int16{2})

		for _, name := range b.wiz.confstore.Alloutbounds() {
			if b.Builder.CheckOutbound(name) {
				b.btns.Addbutton(name+C.GetMsg(C.ButtonSelectEmjoi), name, "")
				continue
			}
			b.btns.AddBtcommon(name)

		}

		b.btns.AddCloseBack()

		if b.lastcallback, err = b.callbackreciver("select or deselect common outbound", b.btns); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnClose, C.BtnBack:
			return nil
		default:
			if b.Builder.CheckOutbound(b.lastcallback.Data) {
				if err = b.Builder.RemoveOutbound(b.lastcallback.Data); err != nil {
					b.Messagesession.Callbackanswere(b.lastcallback.ID, "outbound remove errored - "+err.Error(), true)
				}
				return nil
			}
			builderOut, err := b.wiz.confstore.GetOutbound(b.lastcallback.Data)
			if err != nil {
				b.Messagesession.Callbackanswere(b.lastcallback.ID, "fetching outbound failoed - "+err.Error(), true)
				return nil
			}
			b.alertsender("selected outbound info, if you want to cancle the fill prosess send /cancel anytime  " + builderOut.Info)
			if err = builderOut.FillReqirments(func(msg any) (string, error) {
				mg, err := b.sendreciver(msg)
				if err != nil {
					return "", err
				}
				if mg.Command() == "/cancel" {
					return "", errors.New("user cancled outbound filling")
				}
				return mg.Text, nil
			}); err != nil {
				b.Messagesession.Callbackanswere(b.lastcallback.ID, "outbound addin failed -"+" "+err.Error(), true)
				return nil
			}
			if err = b.Builder.AddRawOut(&builderOut.Out); err != nil {
				b.Messagesession.Callbackanswere(b.lastcallback.ID, "outbound addin failed -"+" "+err.Error(), true)
			}

		}

	case "via xray link":
		
		b.Messagesession.Callbackanswere(b.lastcallback.ID, "not avlable yet", true)
		//TODO: Build later Outbound should have own method to parse quary parameters
		/*
		protocol://
			$(uuid)
			@
			remote-host
			:
			remote-port
		?
			<protocol-specific fields>
			<transport-specific fields>
			<tls-specific fields>
		#$(descriptive-text)
		
	

		
		link, err := b.sendreciver("send you'r xray link config")
		if err != nil {
			return err
		}
		linkurl, err := url.Parse(link.Text)
		if err != nil {
			return err
		}

		switch linkurl.Scheme {
		case singconst.TypeVLESS:
			//values := linkurl.Query()

		case singconst.TypeVMess:
		default:
			b.alertsender("adding this protocole not supported yet")
		}

		
		*/
	

	case "json outbound":
		b.Messagesession.Edit("send your json formated outbound config structure should be sing box", nil, "")
		mg, err := b.wiz.defaultsrv.ExcpectMsgContext(b.ctx, b.userID, b.userID)
		if err != nil {
			return err
		}

		if err = b.Builder.AddoutJson([]byte(mg.Text)); err != nil {
			b.wiz.logger.Error(err.Error())
			b.Messagesession.SendAlert("outbound struct adding failed", nil)
			b.State = addOutbound
			return nil
		}
		b.Messagesession.SendAlert("outbound adding succses", nil)

	case C.BtnBack:
		b.State = outbound
	}

	return nil
}

func (b *BuildState) dns() error {
	var err error
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("rules")
	b.btns.AddBtcommon("servers")
	b.btns.AddCloseBack()

	b.Messagesession.Edit("choos what you want to do with dns settings", b.btns, "")
	if b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID()); err != nil {
		return err
	}
	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = gotoconfig
	case "rules":
		b.State = dnsRules
	case "servers":
		b.State = dnssrvs
	}
	return nil
}

func (b *BuildState) dnsserver() error {
	b.btns.Reset([]int16{})
	b.btns.AddBtcommon("all server")
	b.btns.AddBtcommon("add server")
	b.btns.AddBtcommon("remove server")
	b.btns.AddBtcommon("set default")
	b.btns.AddCloseBack()

	var err error
	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportDns(), b.btns); err != nil {
		return err
	}

	switch b.lastcallback.Data {

	case C.BtnBack:
		b.State = dns

	case "add server":
		b.State = dnssrvaddr

	case "all server":

		b.btns.Reset([]int16{2})

		avbl := 0
		for _, srv := range b.Builder.GetDnsServers() {
			if srv.Tag == "block" || srv.Tag == "default" {
				continue
			}
			avbl++
			b.btns.AddBtcommon(srv.Tag)
		}

		if avbl == 0 {
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "no any editable dns servers", true)
			return nil
		}

		b.btns.AddCloseBack()

		if b.lastcallback, err = b.callbackreciver("select a dns server", b.btns); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnBack, C.BtnClose:
			return nil
		default:
			b.lastSelectDnssrv = b.lastcallback.Data
			b.State = dnsrvhandle
		}

	case "remove server", "set default":
		b.btns.Reset([]int16{2})
		for _, srv := range b.Builder.GetDnsServers() {
			b.btns.AddBtcommon(srv.Tag)
		}
		b.btns.AddCloseBack()

		action := b.lastcallback.Data

		if b.lastcallback, err = b.callbackreciver("select a dns server", b.btns); err != nil {
			return err
		}

		if b.lastcallback.Data == C.BtnBack || b.lastcallback.Data == C.BtnClose {
			return nil
		}

		switch action {

		case "remove server":
			if err = b.Builder.RemoveDnsServer(b.lastcallback.Data); err != nil {
				b.alertsender("dns server remove failed - " + err.Error())
				return nil
			}
			b.alertsender("dns server setting sucses")
		case "set default":
			if err = b.Builder.SetDefaultDns(b.lastcallback.Data); err != nil {
				b.alertsender("default server setting failed - " + err.Error())
				return nil
			}
			b.alertsender("default server setting sucsess")

		}

	}

	return nil
}

func (b *BuildState) dnsadder() error {
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("from json")
	b.btns.AddBtcommon("add custom server")
	b.btns.AddBtcommon("from common servers")
	b.btns.AddCloseBack()
	var err error
	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportDns(), b.btns); err != nil {
		return err
	}

	switch b.lastcallback.Data {
	case C.BtnClose:
		return nil
	case C.BtnBack:
		b.State = dnssrvs
	case "from common servers":
		dnssrv:
		for {
			b.btns.Reset([]int16{2})
			for _, servers := range b.wiz.confstore.AllDnsServer() {
				if b.Builder.CheckDnsServer(servers.Server.Tag) {
					b.btns.Addbutton(servers.Server.Tag + C.GetMsg(C.ButtonSelectEmjoi), servers.Server.Tag, "")
					continue
				}

				b.btns.AddBtcommon(servers.Server.Tag)
			}
			b.btns.AddCloseBack()
			var err error
			if b.lastcallback, err = b.callbackreciver(b.Builder.ExportDns(), b.btns); err != nil {
				return err
			}
			switch b.lastcallback.Data {
			case C.BtnClose, C.BtnBack:
				return nil
			default:

				if b.Builder.CheckDnsServer(b.lastcallback.Data) {
					b.Messagesession.Callbackanswere(b.lastcallback.Data, "already added", true)
					continue dnssrv
				}

				b.btns.Reset([]int16{2})
				b.btns.AddBtcommon(C.BtnConform)
				b.btns.AddBtcommon(C.BtnCancle)

				server, err := b.wiz.confstore.DnsServerbyTag(b.lastcallback.Data)
				if err != nil {
					b.alertsender("dns common server fetching failed")
					continue dnssrv
				}

				if b.lastcallback, err = b.callbackreciver(server.Info, b.btns); err != nil {
					return err
				}

				switch b.lastcallback.Data {
				case C.BtnCancle:
					continue dnssrv
				}

				if err = b.Builder.AddDnsServer(server.Server); err != nil {
					b.alertsender("dns server adding failed - " + err.Error())
				}	
			}

		}

	case "from json":
		rep, err := b.sendreciver("send your json dns server object")
		if err != nil {
			return err
		}
		err = b.Builder.AddDnsServerRaw([]byte(rep.Text))
		if err != nil {
			b.alertsender("json obj add failed err - " + err.Error())
		}
		return nil
	case "add custom server":
		dnsServer := option.DNSServerOptions{}
		var mg *tgbotapi.Message
		if mg, err = b.sendreciver("send a tag name for dnsserver this will be the reffrance for this server"); err != nil {
			return err
		}
		dnsServer.Tag = mg.Text
		if mg, err = b.sendreciver("send address of the server example tcp://1.1.1.1"); err != nil {
			return err
		}
		dnsServer.Address = mg.Text
		if mg, err = b.sendreciver("send address reseolver this used to resolve dns server addres if it is a domain if you want skip this simply send ."); err != nil {
			return err
		}
		if mg.Text == "." {

		} else {
			dnsServer.AddressResolver = mg.Text
		}
		if mg, err = b.sendreciver("send outbound tag for dns server if the outbound tag is not valid default tag will use"); err != nil {
			return err
		}

		if b.Builder.CheckOutbound(mg.Text) {
			dnsServer.Detour = mg.Text
		}
		if err = b.Builder.AddDnsServer(dnsServer); err != nil {
			b.alertsender("dns server adding failed")
			return nil
		}
		b.alertsender("dns server add sucsess")
	}
	return nil
}

func (b *BuildState) dnserverhandle() error {
	for _, changer := range b.changers {
		server, err := b.Builder.GetDnsServer(b.lastSelectDnssrv)
		if err != nil {
			b.alertsender("selected server not found")
			b.State = dnssrvs
			return nil
		}
		if changer.Can(server) {
			return changer.MakeChange(b.Builder, server.Tag, func(i int) {
				b.State = i
			})
		}
	}
	b.alertsender("no any supported changer found")
	b.State = dnssrvs
	return nil
}

func (b *BuildState) dnsrules() error {
	b.btns.Reset([]int16{2})

	b.btns.AddBtcommon("⚙️ json rule ⚙️")
	b.btns.AddBtcommon("⚙️ common rule ⚙️")
	b.btns.AddBtcommon("⚙️ create rule ⚙️")
	b.btns.AddBtcommon("⚙️ remove rule ⚙️")
	b.btns.AddBtcommon("⚙️ all rules ⚙️")
	for _, rule := range b.wiz.confstore.DnsRuleTags {
		if b.Builder.CheckDnsRule(rule) {
			b.btns.Addbutton(rule+" "+C.GetMsg(C.ButtonSelectEmjoi), rule, "")
			continue
		}
		b.btns.AddBtcommon(rule)

	}
	b.btns.AddCloseBack()

	var err error
	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportDns(), b.btns); err != nil {
		return err
	}

	switch b.lastcallback.Data {

	case "⚙️ common rule ⚙️":
		allrules := b.wiz.confstore.AllDnsRule()
		if len(allrules) == 0 {
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "no any changeble rule found", true)
			return nil
		}
		b.btns.Reset([]int16{2})
		C.ExcuteSlice(allrules, func(t *builder.DnsRule) {
			b.btns.AddBtcommon(t.Rule.Tag)
		})
		b.btns.AddCloseBack()
		if b.lastcallback, err = b.callbackreciver("select a rule", b.btns); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnClose, C.BtnBack:
			return nil
		}

		rule, ok := b.wiz.confstore.FullDnsRuleByname(b.lastcallback.Data)

		if !ok {
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "rule not found ", true)
			return nil

		}
		b.alertsender("selected rule info\nif you want to cancle send /cancle\n" + rule.Info)
		if err = rule.Reqirments.FillReqirments(b.callbackreciver, func(msg any) (*tgbotapi.Message, error) {
			mg, err := b.sendreciver(msg)
			if err != nil {
				return nil, err
			}
			if mg.Command() == "/cancle" {
				return nil, errors.New("user cancled filling")
			}
			return mg, err
		}, &rule.Rule); err != nil {
			b.alertsender("rule adding failed - " + err.Error())
			return nil
		}

		if err = b.Builder.AddDnsRuleObj(rule.Rule); err != nil {
			b.alertsender("cannot add rule err = " + err.Error())
			return nil
		}
		b.alertsender("rule adding succses")
	case "⚙️ json rule ⚙️":
		mg, err := b.sendreciver("send json formated rule object  if rule include inbound or outbound tag which is not in config will replace by default inbounds, same for the dns server")
		if err != nil {
			return err
		}
		b.Messagesession.Addreply(mg.MessageID)

		rule := mg.Text
		if mg, err = b.sendreciver("send a uniq tag name for this rule reffrance"); err != nil {
			return err
		}
		err = b.Builder.AddRawDns([]byte(rule), mg.Text)
		if err != nil {
			b.wiz.logger.Error(err.Error())
			b.alertsender("dns rule addin error")
		}
	case "⚙️ remove rule ⚙️":
		b.btns.Reset([]int16{2})
		//b.btns.AddBtcommon()
		for _, rule := range b.Builder.GetDnsRules() {
			if rule.Tag == "" {
				continue
			}

			b.btns.AddBtcommon(rule.Tag)
		}
		b.btns.AddBack(true)

		if b.lastcallback, err = b.callbackreciver("select the desired rule to delete", b.btns); err != nil {
			return err
		}
		switch b.lastcallback.Data {
		case C.BtnBack:
			return nil
		default:
			b.Builder.RemoveDnsRule(b.lastcallback.Data)
			return nil
		}
	case "⚙️ create rule ⚙️":
		tagmg, err := b.sendreciver("send reffrance name for this rule, do not give this name for other rule should be uniq")
		if err != nil {
			return err
		}
		rule, err := CreateRule(common.OptionExcutors{
			Callbackreciver: b.callbackreciver,
			Sendreciver:     b.sendreciver,
			Alertsender:     b.alertsender,
			// Upx:             b.upx, //does not reqire upx
			MessageSession:  b.Messagesession,
			Ctrl:            b.wiz.ctrl,
			Btns:            b.btns,
		}, b.Builder, routerule, &option.DNSRule{
			Type: singconst.RuleTypeDefault,
			Tag:  tagmg.Text,
		})

		if err != nil {
			b.alertsender("rule creation failed")
			return nil
		}
		if ruler, ok := rule.(*option.DNSRule); ok {
			err = b.Builder.AddDnsRuleObj(*ruler)
			if err != nil {
				b.alertsender("rule creation failed err - " + err.Error())
			}
		}
	case "⚙️ all rules ⚙️":
		if len(b.Builder.GetDnsRules()) == 0 {
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "no any dns rules avlable", true)
			return nil
		}
		b.btns.Reset([]int16{2})
		for _, rule := range b.Builder.GetDnsRules() {
			b.btns.AddBtcommon(rule.Tag)
		}

		b.btns.AddCloseBack()

		if b.lastcallback, err = b.callbackreciver("select dns rule", b.btns); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnClose, C.BtnBack:
			return nil
		}

		if !b.Builder.CheckDnsRule(b.lastcallback.Data) {
			b.Messagesession.Callbackanswere(b.lastcallback.Data, "somthing went wrong selected dns rule not found", true)
			return nil
		}

		for _, handler := range b.changers {
			if handler.Can(b.Builder.GetDnsRule(b.lastcallback.Data)) {
				return handler.MakeChange(b.Builder, b.lastcallback.Data, func(st int) { b.State = st })
			}
		}
		b.Messagesession.Callbackanswere(b.lastcallback.ID, "there is no change option found for this type outbounds", true)
		b.State = allOutbound
		return nil

	case C.BtnBack:
		b.State = dns
	case C.BtnClose:
		return nil
	default:
		if b.Builder.CheckDnsRule(b.lastcallback.Data) {
			b.Builder.RemoveDnsRule(b.lastcallback.Data)
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "Dns rule removed", true)
			return nil

		}

		dnsRule, loaded := b.wiz.confstore.FullDnsRuleByname(b.lastcallback.Data)
		if !loaded {
			return nil
		}
		if !dnsRule.Reqirments.IsStatic {
			if err = dnsRule.Reqirments.FillReqirments(b.callbackreciver, b.sendreciver, &dnsRule.Rule); err != nil {
				b.alertsender("dns rule adding failed err - " + err.Error())
				return nil
			}
			if err = b.Builder.AddDnsRuleObj(dnsRule.Rule); err != nil {
				b.alertsender("dns rule adding failed err - " + err.Error())
				return nil
			}
		} else {
			if err = b.Builder.AddDnsRule(b.lastcallback.Data, ""); err != nil {
				b.Messagesession.Callbackanswere(b.lastcallback.ID, "Dns rule addin error eoccured", true)
				b.wiz.logger.Error(err.Error())
				return nil
			}
		}

		b.Messagesession.Callbackanswere(b.lastcallback.ID, "Dns rule added added rule info - "+b.wiz.confstore.DnsRuleMust(b.lastcallback.Data).Info, true)
	}
	return nil
}

func (b *BuildState) route() error {
	var err error
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("rules")
	b.btns.AddBtcommon("ruleset")
	b.btns.AddBtcommon("set final")

	b.btns.AddCloseBack()

	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportRoute(), b.btns); err != nil {
		return err
	}
	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = gotoconfig
	case "rules":
		b.State = routeRules
	case "set final":
		b.btns.Reset([]int16{2})
		for _, out := range b.Builder.GetAllOutNames() {
			switch out {
			case "block", "dns":
				continue
			}
			b.btns.AddBtcommon(out)
		}
		if b.lastcallback, err = b.callbackreciver("select outbound", b.btns); err != nil {
			return err
		}

		if err = b.Builder.SetRouteFinal(b.lastcallback.Data); err != nil {
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "Setting Final Failed err := "+err.Error(), true)
		}

	case "ruleset":
		b.State = rtsets
	}

	return nil
}

func (b *BuildState) routeruleset() error {
	var err error

	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("all rule set")
	b.btns.AddBtcommon("add rule set")
	b.btns.AddBtcommon("remove rule set")
	b.btns.AddCloseBack()

	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportRoute(), b.btns); err != nil {
		return err
	}

	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = route
	case "all rule set":
		b.State = rulesetchg
		b.btns.Reset([]int16{2})
		allset := b.Builder.GetRouteRuleSetTags()
		if len(allset) == 0 {
			b.Messagesession.Callbackanswere(b.lastcallback.Data, "there are nopt any rule set", true)
			return nil
		}
		for _, set :=  range  allset {
			b.btns.AddBtcommon(set)
		}

		if b.lastcallback, err = b.callbackreciver(b.Builder.ExportRoute(), b.btns); err != nil {
			return err
		}
		b.lastSelectRuleSet = b.lastcallback.Data
	case C.BtnClose:
		return nil
	case "add rule set":
		b.State = rulesetadd

	case "remove rule set":
		b.btns.Reset([]int16{2})
		for _, set :=  range b.Builder.GetRouteRuleSetTags() {
			b.btns.AddBtcommon(set)
		}
		if b.lastcallback, err = b.callbackreciver("select rule set to delete", b.btns); err != nil {
			return err
		}
		err = b.Builder.RemoveRouteRuleSet(b.lastcallback.Data)
		if err != nil {
			b.alertsender("rule set remove failed err - " + err.Error())
		}
	default:
		b.Messagesession.Callbackanswere(b.lastcallback.ID, "this option does not support yet", true)
	}

	return nil
}

func (b *BuildState) ruleSetAdd() error {
	var err error
	
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("create rule set")
	b.btns.AddBtcommon("json obj")
	b.btns.AddBtcommon("common rule set")
	b.btns.AddCloseBack()
	
	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportRoute(), b.btns); err != nil {
		return err
	}
	
	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = rtsets
	case C.BtnClose:
		return nil
	case "create rule set":
		ruleset, err := CreateRuleSet(common.OptionExcutors{
			Callbackreciver: b.callbackreciver,
			MessageSession: b.Messagesession,
			Sendreciver: b.sendreciver,
			Alertsender: b.alertsender,
			Btns: b.btns,
			Ctrl: b.wiz.ctrl,

			//does not require upx
			
		}, b.Builder)

		if err != nil {
			b.alertsender("rule set creation failed")
			return nil
		}
		err = b.Builder.AddRouteRuleSet(ruleset)

		if err!= nil {
			b.alertsender("rule adding failed err - " + err.Error())
		}
	case "json obj":
		rulesetrw, err := b.sendreciver("send your json rule set obj")
		if err != nil {
			return err
		}
		err = b.Builder.AddRouteRuleSetRaw([]byte(rulesetrw.Text))
		if err!= nil {
			b.alertsender("json rule adding failed err - " + err.Error())
		}
	
	case "common rule set":

		rset:

		for {
			b.btns.Reset([]int16{2})
			for _, rset := range b.wiz.confstore.AllRUleSet() {
				if b.Builder.CheckRuleSet(rset.RuleSet.Tag) {
					b.btns.Addbutton(rset.RuleSet.Tag + C.GetMsg(C.ButtonSelectEmjoi), rset.RuleSet.Tag, "" )
					continue
				}
				b.btns.AddBtcommon(rset.RuleSet.Tag)
			}
			b.btns.AddCloseBack()

			var err error
			if b.lastcallback, err = b.callbackreciver(b.Builder.ExportRoute(), b.btns); err != nil {
				return err
			}

			switch b.lastcallback.Data {
			case C.BtnClose, C.BtnBack:
				return nil
			default:
				if b.Builder.CheckRuleSet(b.lastcallback.Data) {
					b.Messagesession.Callbackanswere(b.lastcallback.ID, "already added", true)
					continue rset
				}

				b.btns.Reset([]int16{2})
				b.btns.AddBtcommon(C.BtnConform)
				b.btns.AddBtcommon(C.BtnCancle)

				ruleset, err := b.wiz.confstore.RuleSetBytag(b.lastcallback.Data)
				if err != nil {
					b.alertsender("common ruleset fetching failed")
					continue rset
				}

				if b.lastcallback, err = b.callbackreciver(ruleset.Info, b.btns); err != nil {
					return err
				}

				switch b.lastcallback.Data {
				case C.BtnCancle:
					continue rset
				}

				if err = b.Builder.AddRouteRuleSet(ruleset.RuleSet); err != nil {
					b.alertsender("ruleset adding failed - " + err.Error())
				}
				
			}
		}
	}
	return nil
}

func (b *BuildState) ruleSetChg() error {
	//var err error
	ruleset, err := b.Builder.GetRouteRuleSet(b.lastSelectRuleSet)
	if err != nil {
		b.alertsender("rule set not found ")
		b.State = rtsets
		return nil
	}
	for _, chg := range b.changers {
		if chg.Can(ruleset) {
			return chg.MakeChange(b.Builder, b.lastSelectRuleSet, func(i int) {
				b.State = i
			})
		} 
	}
	b.alertsender("no any changer found for rule set")
	b.State = rtsets
	return nil
}

func (b *BuildState) routeRules() error {
	var err error

	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("add rules")
	b.btns.AddBtcommon("all rules")

	b.btns.AddCloseBack()
	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportRoute(), b.btns); err != nil {
		return err
	}

	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = route
	case C.BtnClose:
		return nil

	case "all rules":
		if len(b.Builder.GetRouteRules()) == 0 {
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "no any routing rules avlable", true)
			return nil
		}
		b.btns.Reset([]int16{2})
		for _, rule := range b.Builder.GetRouteRules() {
			b.btns.AddBtcommon(rule.Tag)
		}

		b.btns.AddCloseBack()

		if b.lastcallback, err = b.callbackreciver("select route rule", b.btns); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnClose, C.BtnBack:
			return nil
		}

		if !b.Builder.CheckRule(b.lastcallback.Data) {
			b.Messagesession.Callbackanswere(b.lastcallback.Data, "somthing went wrong selected rule not found", true)
			return nil
		}

		for _, handler := range b.changers {
			if handler.Can(b.Builder.GetRouteRule(b.lastcallback.Data)) {
				return handler.MakeChange(b.Builder, b.lastcallback.Data, func(st int) { b.State = st })
			}
		}
		b.Messagesession.Callbackanswere(b.lastcallback.ID, "there is no change option found for this type outbounds", true)
		b.State = allOutbound
		return nil

	case "add rules":
		b.State = rtruleadd
	}

	return err
}

func (b *BuildState) ruleAdd() error {
	var err error

	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("json object")
	b.btns.AddBtcommon("common rules")
	b.btns.AddBtcommon("create rules")
	b.btns.AddCloseBack()
	b.lastcallback, err = b.callbackreciver(b.Builder.ExportRoute(), b.btns)
	if err != nil {
		return err
	}
	switch b.lastcallback.Data {
	case C.BtnBack:
		b.State = routeRules

	case C.BtnClose:
		return nil

	case "json object":
		mg, err := b.sendreciver("send json typed rule object outbound should be matched exting outbouds")
		if err != nil {
			return err
		}
		ref, err := b.sendreciver("send reffrance name for that rule so that you can identyfy it lateron")
		if err != nil {
			return err
		}
		if err = b.Builder.AddRawRoute([]byte(mg.Text), ref.Text); err != nil {
			b.alertsender("rule adding failed - " + err.Error())
		}

	case "common rules":
		allrules := b.wiz.confstore.AllRouteRulesWithReq()
		if len(allrules) == 0 {
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "no any changeble rule found", true)
			return nil
		}
		b.btns.Reset([]int16{2})
		C.ExcuteSlice(allrules, func(t *builder.RouteRule) {
			b.btns.AddBtcommon(t.Rule.Tag)
		})
		b.btns.AddCloseBack()
		if b.lastcallback, err = b.callbackreciver("select a rule", b.btns); err != nil {
			return err
		}

		switch b.lastcallback.Data {
		case C.BtnClose, C.BtnBack:
			return nil
		}

		rule, ok := b.wiz.confstore.FullRuleByname(b.lastcallback.Data)
		if !ok {
			b.Messagesession.Callbackanswere(b.lastcallback.ID, "rule not found ", true)
			return nil

		}
		b.alertsender("selected rule info\nif you want to cancle send /cancle\n" + rule.Info)
		if err = rule.Reqirments.FillReqirments(b.callbackreciver, func(msg any) (*tgbotapi.Message, error) {
			mg, err := b.sendreciver(msg)
			if err != nil {
				return nil, err
			}
			if mg.Command() == "/cancle" {
				return nil, errors.New("user cancled filling")
			}
			return mg, err
		}, &rule.Rule); err != nil {
			b.alertsender("rule adding failed - " + err.Error())
			return nil
		}

		if err = b.Builder.AddRouteRuleOb(rule.Rule); err != nil {
			b.alertsender("cannot add rule err = " + err.Error())
			return nil
		}
		b.alertsender("rule adding succses")

	case "create rules":

		tagmg, err := b.sendreciver("send reffrance name for this rule")
		if err != nil {
			return err
		}

		rule, err := CreateRule(common.OptionExcutors{
			Callbackreciver: b.callbackreciver,
			Sendreciver:     b.sendreciver,
			Alertsender:     b.alertsender,
			// Upx:             b.upx, does not require upx
			MessageSession:  b.Messagesession,
			Ctrl:            b.wiz.ctrl,
			Btns:            b.btns,
		}, b.Builder, routerule, &option.Rule{
			Type: singconst.RuleTypeDefault,
			Tag:  tagmg.Text,
		})
		if err != nil {
			b.alertsender("rule creation failed")
			return nil
		}
		if ruler, ok := rule.(*option.Rule); ok {
			err = b.Builder.AddRouteRuleOb(*ruler)
			if err != nil {
				b.alertsender("rule creation failed")
			}
		}

	}

	return nil
}

func (b *BuildState) home() error {
	var err error
	var configs []db.SboxConfigs

	b.btns.Reset([]int16{2})
	if configs, err = b.wiz.ctrl.GetSboxConfig(b.userID); err != nil {
		return err
	}
	C.ExcuteSlice(configs, func(t *db.SboxConfigs) {
		b.btns.AddBtcommon(t.Name)
	})
	b.btns.Passline()
	b.btns.AddBtcommon(C.BtnBuilderCreateConfig)
	b.btns.AddClose(true)

	if b.lastcallback, err = b.callbackreciver(botapi.UpMessage{
		TemplateName: C.TmplBuilderHome,
		Template: struct {
			*botapi.CommonUser
			ConfCount int
		}{
			CommonUser: &botapi.CommonUser{
				Name:     b.dbuser.Name,
				Username: b.dbuser.Username.String,
				TgId:     b.userID,
			},
			ConfCount: len(configs),
		},
	}, b.btns); err != nil {
		return err
	}
	switch b.lastcallback.Data {

	case C.BtnBuilderCreateConfig:
		b.State = createconfig
	default:
		b.lastconfig = b.lastcallback.Data
		b.State = gotoconfig
	}
	return nil
}

func (b *BuildState) confighandle() error {

	buildconf, err := b.wiz.ctrl.GetSpecificConf(b.userID, b.lastconfig)

	if err != nil {
		b.Messagesession.SendAlert("selected config errored somthin wrong", nil)
		b.State = initiate
		return nil
	}

	if b.Builder != nil {
		b.Builder.Close()
	}

	b.Builder, err = builder.NewBuilder(b.ctx, buildconf.ConfPath, b.wiz.confstore, b.wiz.logger)
	if err != nil {
		b.wiz.logger.Error(err.Error())
		b.Messagesession.SendAlert("builder error", nil)
		b.State = initiate
		return nil
	}
	b.Builder.SetExportCallback(func(b []byte) any {

		var message string

		if len(b) > C.MaxCharacterMg {
			const maxChunkSize = C.MaxCharacterMg - len("<pre><code></code></pre>")
			var buf bytes.Buffer
		
			for len(b) > 0 {
				chunkSize := maxChunkSize
				if len(b) < maxChunkSize {
					chunkSize = len(b)
				}
				buf.WriteString("<pre><code>")
				buf.Write(b[:chunkSize])
				buf.WriteString("</code></pre>")

				b = b[chunkSize:]
			}
			message =  buf.String()
		} else {
			message = fmt.Sprintf("<pre><code>%s</code></pre>", string(b))
		}
		return botapi.Htmlstring(message)
	})

	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("inbound")
	b.btns.AddBtcommon("outbound")
	b.btns.AddBtcommon("dns")
	b.btns.AddBtcommon("route")
	b.btns.AddBtcommon("experimental")
	b.btns.AddBtcommon("snapshot")
	b.btns.AddBtcommon(C.BtnDelete)
	b.btns.AddCloseBack()
	b.Messagesession.Edit(b.Builder.Export(), b.btns, "")

	b.lastcallback, err = b.wiz.callback.GetcallbackContext(b.ctx, b.btns.ID())
	if err != nil {
		return err
	}

	switch b.lastcallback.Data {

	case C.BtnBack:
		b.State = initiate
	case "snapshot":
		b.Messagesession.SendAlert(b.Builder.Export(), nil)
	case C.BtnDelete:

		b.Builder.Close()
		if err = b.wiz.ctrl.DeleteConf(buildconf.ID); err != nil {
			b.alertsender("config deletion errored")
		}
		b.alertsender("config deletion sucsess")
		b.State = initiate
		return nil

	case "inbound":
		b.State = inbound
	case "outbound":
		b.State = outbound
	case "dns":
		b.State = dns
	case "route":
		b.State = route
	case "experimental":
		b.State = experimental

	}

	return nil
}

func (b *BuildState) createConfig() error {

	reply, err := b.sendreciver("send a name for this config name should be uniq")
	if err != nil {
		return err
	}
	if reply.IsCommand() {
		b.Messagesession.Edit("send a name not command", nil, "")
		b.State = initiate
		return nil
	}

	if reply.Text == "" {
		b.Messagesession.Edit("no name found please send valid name", nil, "")
		b.State = initiate
		return nil
	}

	_, err = b.wiz.ctrl.CreateSboxConf(b.userID, reply.Text)
	if err != nil {
		b.Messagesession.Callbackanswere(b.lastcallback.ID, "config creation failed", true)
		b.State = initiate
		return nil
	}

	b.Messagesession.SendAlert("config creatin sucsess", nil)
	b.State = initiate
	return nil
}



func (b *BuildState) experimental() error {
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("cache_file")
	b.btns.AddBtcommon("clash_api")
	b.btns.AddBtcommon("v2ray_api")
	b.btns.AddCloseBack()

	var err error
	if b.lastcallback, err = b.callbackreciver(b.Builder.ExportExpermental(), b.btns); err != nil {
		return err
	}

	switch b.lastcallback.Data {
	case "clash_api":
		b.State = clash_api
	case "cache_file":
		b.State = cache_file
	case C.BtnBack, C.BtnClose:
		b.State = gotoconfig
		return nil
	default:
		b.Messagesession.Callbackanswere(b.lastcallback.ID, "will avalble soon", true)

	}

	return nil
}

func (b *BuildState) clashapi() error {
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("add default")
	b.btns.AddBtcommon("remove")
	b.btns.AddBtcommon("change")
	b.btns.AddCloseBack()
	var err error
	b.lastcallback, err = b.callbackreciver(b.Builder.ExportExpermental(), b.btns)
	if err != nil {
		return err
	}
	switch b.lastcallback.Data {
	case "remove":
		b.Builder.RemoveClashApi()
	case "change":
		b.State = clash_apichg
	case "add default":
		if b.Builder.AddClashApi() != nil {
			b.alertsender("clash api adding errored")
		}
	case C.BtnBack:
		b.State = experimental
	}
	return nil
}

func (b *BuildState) change_clashapi() error {
	
	for _, chg := range b.changers {
		if chg.Can("clash") {
			return chg.MakeChange(b.Builder, "clash", func(i int) {
				b.State = i
			})
		}
	}
	b.State = clash_api
	b.alertsender("no any clash api changers found")
	
	return nil
}

func (b *BuildState) cache_file() error {
	b.btns.Reset([]int16{2})
	b.btns.AddBtcommon("add default")
	b.btns.AddBtcommon("remove")
	b.btns.AddBtcommon("change")
	b.btns.AddCloseBack()
	var err error
	b.lastcallback, err = b.callbackreciver(b.Builder.ExportExpermental(), b.btns)
	if err != nil {
		return err
	}
	switch b.lastcallback.Data {
	case "remove":
		b.Builder.RemoveCacehFile()
	case "change":
		b.State = cache_file_chg
	case "add default":
		if b.Builder.AddCacheFile() != nil {
			b.alertsender("cache file adding errored")
		}
	case C.BtnBack:
		b.State = experimental
	}
	return nil
}

func (b *BuildState) cacheFileChange() error {
	for _, chg := range b.changers {
		if chg.Can("cache") {
			return chg.MakeChange(b.Builder, "cache", func(i int) {
				b.State = i
			})
		}
	}
	b.State = cache_file
	b.alertsender("no any cache file changers found")
	return nil
}


func (u *Xraywiz) commandBuildV2(upx *update.Updatectx) error {
	//TODO: change later context deadline
	newctx, cancle := context.WithTimeout(u.ctx, 5*time.Minute)
	upx.Ctx = newctx
	defer cancle()

	Messagesession := botapi.NewMsgsession(u.botapi, upx.User.TgID, upx.User.TgID, upx.User.Lang)

	if _, ok := u.builds.Load(upx.User.TgID); ok {
		Messagesession.SendAlert("you have already opend a builder session please close it and open new one", nil)
		return nil
	}

	newstate := BuildState{
		ctx:            upx.Ctx,
		State:          0,
		Messagesession: Messagesession,
		userID: upx.User.TgID,
		dbuser: upx.Dbuser(),
		wiz:            u,
		btns:           botapi.NewButtons([]int16{2}),
	}
	u.builds.Store(upx.User.TgID, struct{}{})

	defer func() {
		if newstate.Builder != nil {
			newstate.Builder.Close()
		}
		u.builds.Delete(upx.User.TgID)
	}()

	var sendrec common.Sendreciver = func(msg any) (*tgbotapi.Message, error) {
		_, err := Messagesession.Edit(msg, nil, "")
		if err != nil {
			return nil, err
		}
		mg, err := u.defaultsrv.ExcpectMsgContext(upx.Ctx, upx.User.TgID, upx.User.TgID)
		if err == nil {
			Messagesession.Addreply(mg.MessageID)
		}
		return mg, err
	}

	var callbackrec common.Callbackreciver = func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error) {
		_, err := Messagesession.Edit(msg, btns, "")
		if err != nil {
			return nil, err
		}
		return u.callback.GetcallbackContext(upx.Ctx, btns.ID())
	}

	var alertsender common.Alertsender = func(msg string) {
		Messagesession.SendAlert(msg, nil)
	}

	newstate.alertsender = alertsender
	newstate.callbackreciver = callbackrec
	newstate.sendreciver = sendrec

	tra := alltransportadder(sendrec, callbackrec, alertsender, newstate.btns)
	newstate.otadders = alloutadders(sendrec, callbackrec, alertsender, newstate.btns, tra)
	newstate.changers = allchangers(sendrec, callbackrec, alertsender, newstate.btns, tra)

	err := newstate.run()
	if err != nil {
		u.logger.Error(err.Error())
		if errors.Is(err, C.ErrContextDead) {
			tmpctx, cancle := context.WithTimeout(u.ctx, 10*time.Second)
			Messagesession.SetNewcontext(tmpctx)
			Messagesession.SendAlert("you'r build time over retry", nil)
			cancle()
		}
	}
	return nil
}