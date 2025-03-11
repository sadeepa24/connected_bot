package service

import (
	"strings"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/builder/v1"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"

	// option "github.com/sagernet/sing-box/option"
	option "github.com/sadeepa24/connected_bot/builder/sbox_option/v1"
)

type changer interface {
	Can(ottype any) bool
	MakeChange(builder *builder.BuildConfig, name string, statch func(int)) error // statach mean status changer new state when back and close
}

var _ changer = (*singless)(nil)
var _ changer = (*rulechanger)(nil)
var _ changer = (*dnsserverchanger)(nil)


var outboundfields = []string{ "uuid", "server", "server_port", "password", "flow", } 
var tlschangerlist = []string{ "enabled", "disable_sni", "server_name", "insecure", "min_version", "max_version", }
var dialerchange = []string{ "detour", "bind_interface", "protect_path", "routing_mark", "reuse_addr", "connect_timeout", "domain_strategy"}
var transportchg = []string {"path", "host", "method", "idle_timeout", "ping_timeout", "service_name"}
var multiplex = []string { "enabled", "protocol", "min_streams", "max_streams", "padding", }
var rulechg = []string{ "client", "protocol", "auth_user",  "domain", "domain_suffix", "domain_keyword", "domain_regex", "source_geoip", "user", "geosite", "source_ip_cidr", "process_name", "ip_cidr", "port", "source_port", "process_path", "wifi_ssid", "wifi_bssid", "rule_set", }
var dnschg = []string { "address", "address_strategy", "address_resolver", "address_fallback_delay", "strategy", "detour", "client_subnet",}
var rulesetchgers = []string{ "format", "path", "url", "download_detour", }
var clash_chg = []string{ "external_controller", "external_ui", "external_ui_download_url", "external_ui_download_detour", "secret", "default_mode", "access_control_allow_origin", "access_control_allow_private_network", }
var cache_chg = []string{ "enabled", "path", "cache_id", "store_fakeip", "store_rdrc", "rdrc_timeout", }
var tunchanger = []string{ "interface_name", "mtu", "gso", "auto_route", "iproute2_table_index", "iproute2_rule_index", "auto_redirect", "strict_route", "stack", "include_interface", "exclude_interface", "route_address", "route_exclude_address", "route_exclude_address_set", "route_address_set", "exclude_package", "include_package", "include_android_user", "include_uid_range", "include_uid", "exclude_uid_range", "exclude_uid", "endpoint_independent_nat", }
var inopts = []string{ "sniff_enabled", "sniff_override_destination", "sniff_timeout", "domain_strategy", "udp_disable_domain_unmapping", }


func commonmsg(callbackname string) string {
	return `
		Instructions for Sending Values
		🛠️ When updating fields, please follow these guidelines:
			
		Send values according to the button you pressed. ` + callbackname  +`
		
		To remove a field, simply send a . (dot)
		For boolean fields, send false to disable or clear the field.
		To cancel the operation, simply send /cancel.
		Be precise: Send exact values for each field, as the builder only validates types (e.g., string, int), not the content.
		
		Example:
		for source_ip_cidr
		✅ Correct: 192.168.1.0/24, 127.0.0.1/32,  
		🚫 invalid Inputs:  hello.com, "127.0.0.1/32", "", {}, 1.1.1, 1.1.1.600

		for Port
		✅ Correct Inputs: 8080, 80, 443, 1080
		🚫 Invalid Inputs: "8080", hey, $%, "", {}, _-
		
		🔴 Incorrect values will make the configuration invalid!
			
		Special Notes:
			
		For fields like detour, provide the outbound name that exists in the configuration.
		⚠️ If the name doesn’t match, the configuration won’t work.
		
		Proceed with caution to ensure your changes are valid! 🚀

		කොටින්ම සින්හලෙන්, "දන්නවනන් විතරක් කරහන්"

		`
}





func allchangers(sendrec common.Sendreciver, callback common.Callbackreciver, alertsender common.Alertsender, btns *botapi.Buttons, tra []transportwizard) []changer {
	//TODO: Remove All Interface Changer Create 1 Universal Changer
	
	
	changers := []changer{}
	changers = append(changers, &singless{
		btns:            btns,
		sendreciver:     sendrec,
		callbackreciver: callback,
		alertsender:     alertsender,
		transwiz:        tra,
	})

	changers = append(changers, &rulechanger{
		btns:            btns,
		sendreciver:     sendrec,
		callbackreciver: callback,
		alertsender:     alertsender,
		ruletype:        dnsrule,
	})
	changers = append(changers, &rulechanger{
		btns:            btns,
		sendreciver:     sendrec,
		callbackreciver: callback,
		alertsender:     alertsender,
		ruletype:        routerule,
	})

	changers = append(changers, &dnsserverchanger{
		btns:            btns,
		sendreciver:     sendrec,
		callbackreciver: callback,
		alertsender:     alertsender,
	})

	changers = append(changers, &rulesetchanger{
		btns:            btns,
		sendreciver:     sendrec,
		callbackreciver: callback,
		alertsender:     alertsender,
	})

	changers = append(changers, &clash_api_chg{
		btns:            btns,
		sendreciver:     sendrec,
		callbackreciver: callback,
		alertsender:     alertsender,
	})
	changers = append(changers, &clash_file_chg{
		btns:            btns,
		sendreciver:     sendrec,
		callbackreciver: callback,
		alertsender:     alertsender,
	})

	changers = append(changers, &inboundchanger{
		btns:            btns,
		sendreciver:     sendrec,
		callbackreciver: callback,
		alertsender:     alertsender,
		transwiz:        tra,
	})


	return changers
}


// common for vless vmess trojan
type singless struct {
	btns            *botapi.Buttons
	sendreciver     common.Sendreciver
	callbackreciver common.Callbackreciver
	alertsender     common.Alertsender
	transwiz        []transportwizard
}

func (c *singless) Can(ottype any) bool {
	switch st := ottype.(type) {
	case string:
		switch st {
		case "vless", "vmess", "trojan":
			return true
		}
	}
	return false
}

func (c *singless) MakeChange(builder *builder.BuildConfig, outname string, statch func(int)) error {
	c.btns.Reset([]int16{2})
	//c.btns.AddBtcommon("change tag name")
	// c.btns.AddBtcommon("change sni")
	// c.btns.AddBtcommon("change server")
	// c.btns.AddBtcommon("change uuid")
	c.btns.AddBtcommon("change outbound field")
	c.btns.AddBtcommon("change tls settings")
	c.btns.AddBtcommon("change dialer settings")
	c.btns.AddBtcommon("change transport settings")
	c.btns.AddBtcommon("replace tls settings")
	c.btns.AddBtcommon("replace transport settings")
	c.btns.AddBtcommon("change multiplex settings")
	c.btns.AddBtcommon("remove tls")
	c.btns.AddCloseBack()

	callback, err := c.callbackreciver(builder.ExportOutbound(outname), c.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case "change sni":
		mg, err := c.sendreciver("send new sni for this outbound")
		if err != nil {
			return err
		}
		if err = builder.ChangeSni(outname, mg.Text); err != nil {
			c.alertsender("sni change failed")
		}
		return nil
	case "change outbound field":
		return c.changeloop(builder, outname, builder.ChangeSelfOutbound, outboundfields)
	case "change transport settings":
		return c.changeloop(builder, outname, builder.ChangeTransport, transportchg)
	case "change dialer settings":
		return c.changeloop(builder, outname, builder.ChangeDialer, dialerchange)
	case "change tls settings":
		return c.changeloop(builder, outname, builder.ChangeTLS, tlschangerlist)
	case "change uuid":
		mg, err := c.sendreciver("send new uuid for this outbound")
		if err != nil {
			return err
		}
		if err = builder.ChangeUUID(outname, mg.Text); err != nil {
			c.alertsender("uuid change failed err - " + err.Error())
		}
		return nil
	case "replace tls settings":
		tlsob, err := tlsconstruct(c.sendreciver, c.callbackreciver, c.btns)
		if err != nil {
			return err
		}
		builder.ChangeTlsSettings(outname, tlsob.TLS)
	case "replace transport settings":
		c.btns.Reset([]int16{2})
		for _, tra := range c.transwiz {
			c.btns.AddBtcommon(tra.Name())
		}
		c.btns.AddCloseBack()

		if callback, err = c.callbackreciver("select a transport", c.btns); err != nil {
			return err
		}
		switch callback.Data {
		case C.BtnClose:
			statch(buildclose)
			return nil
		case C.BtnBack:
			statch(allOutbound)
		default:
			for _, tra := range c.transwiz {
				if tra.Name() == callback.Data {
					newtra, err := tra.Excute()
					if err != nil {
						return err
					}
					if err = builder.ChageTransport(outname, newtra); err != nil {
						c.alertsender("error occured when changing transport")
					}
					return nil

				}
			}
		}
	case "change server":
		mg, err := c.sendreciver("send new server ip addr or domain")
		if err != nil {
			return err
		}
		if err = builder.ChangeServer(outname, mg.Text); err != nil {
			c.alertsender("server change failed")
		}
		return nil
	case "remove tls":
		if err = builder.RemoveTls(outname); err != nil {
			c.alertsender("error occured when removing tls ")
		}
	case "change multiplex settings":
		return c.changeloop(builder, outname, builder.ChangeMultiplex, multiplex)
	case C.BtnBack:
		statch(allOutbound)
		return nil
	case C.BtnClose:
		statch(buildclose)
	default:
		c.alertsender("this does not supported yet")

	}

	return nil
}

func (c *singless) changeloop(builder *builder.BuildConfig, outname string, chg func(q, sq, qs string) error, fileds []string ) error {
	self:
	for {
		c.btns.Reset([]int16{3})
		for _, otchg := range fileds {
			c.btns.Addbutton(otchg + " ⚙", otchg, "")
		}
		c.btns.AddCloseBack()
		callback, err := c.callbackreciver(builder.ExportOutbound(outname), c.btns)
		if err != nil {
			return err
		}
		switch callback.Data {
		case C.BtnBack:
			break self 
		case C.BtnClose:
			return nil
		}

		replymg, err := c.sendreciver(commonmsg(callback.Data))

		if err != nil {
			return err
		}

		if replymg.Command() == "cancel" {
			continue
		}
		
	
		if replymg.Text == "." {
			replymg.Text = ""	
		}
		err = chg(outname, callback.Data, replymg.Text)
		//err = builder.ChangeSelf(outname, callback.Data, replymg.Text)
		if err != nil {
			c.alertsender("change failed - " + err.Error())
		}

	}
	return nil
} 





type rulechanger struct {
	btns            *botapi.Buttons
	sendreciver     common.Sendreciver
	callbackreciver common.Callbackreciver
	alertsender     common.Alertsender
	ruletype        string
}

func (r *rulechanger) Can(ottype any) bool {
	switch ottype.(type) {
	case option.Rule:
		if r.ruletype == routerule {
			return true
		}
	case option.DNSRule:
		if r.ruletype == dnsrule {
			return true
		}
	}
	return false
}

func (r *rulechanger) MakeChange(builder *builder.BuildConfig, tag string, statch func(int)) error {

	for {
		r.btns.Reset([]int16{3})

		switch r.ruletype {
		case routerule:
			r.btns.Addbutton("change outbound ⭕", "change outbound", "")
		case dnsrule:
			r.btns.Addbutton("change server ⭕", "change server", "")
		}
		r.btns.Addbutton("change action ⭕", "change action", "")
		r.btns.Addbutton("remove rule ⭕", "remove rule", "")

		for _, chanopt := range rulechg {
			r.btns.AddBtcommon(chanopt)
		}

		r.btns.AddCloseBack()
		var exported any
		switch r.ruletype {
		case routerule:
			exported = builder.ExportRouteRule(tag)
		case dnsrule:
			exported = builder.ExportDnsRule(tag)
		}

		callback, err := r.callbackreciver(exported, r.btns)
		if err != nil {
			return err
		}

		switch callback.Data {
		case C.BtnClose:
			statch(buildclose)
			return nil
		case C.BtnBack:
			return nil
		case "change outbound":
			r.btns.Reset([]int16{2})
			for _, outtag := range builder.GetAllOutNames() {
				r.btns.AddBtcommon(outtag)
			}
			callback, err := r.callbackreciver("select outbound", r.btns)
			if err != nil {
				return err
			}
			err = builder.SetRuleOutbound(tag, callback.Data)
			if err != nil {
				r.alertsender("setting outbound failed err - " + err.Error())
			}
		case "change server":
			continue
		case "remove rule":
			switch r.ruletype {
			case routerule:
				builder.RemoveRouteRule(tag)
			case dnsrule:
				builder.RemoveDnsRule(tag)
			}
			return nil
		case "change action":
			continue
		default:
			replist, err := r.sendreciver(`
				🚀 Send a comma-separated list of ` + callback.Data + `.

				🛑 If you want to cancel, just type /cancel.
				✨ Want to clear the list? Simply send .
				⚠️ Important: Please double-check your inputs!
				
				For example: Avoid sending invalid strings like hello.com for source_ip_cidr. Incorrect inputs will make the configuration invalid! 🚫`)
			if err != nil {
				return err
			}
			if replist.Command() == "/cancel" {
				continue
			}

			list := strings.Split(replist.Text, ",")
			if replist.Text == "." {
				list = []string{}
			}

			switch r.ruletype {
			case routerule:
				if builder.SetListTorule(tag, callback.Data, list) != nil {
					r.alertsender("seting " + callback.Data + " failed")
				}
			case dnsrule:
				if builder.SetListToDnsRule(tag, callback.Data, list) != nil {
					r.alertsender("seting " + callback.Data + " failed")
				}
			}
		}
	}
}




type dnsserverchanger struct {
	btns            *botapi.Buttons
	sendreciver     common.Sendreciver
	callbackreciver common.Callbackreciver
	alertsender     common.Alertsender
}

func (r *dnsserverchanger) Can(ottype any) bool {
	
	switch ottype.(type) {
	case option.DNSServerOptions:
		return true
	case *option.DNSServerOptions:
		return true
	}
	
	return false
}

func (r *dnsserverchanger) MakeChange(builder *builder.BuildConfig, tag string, statch func(int)) error {
	r.btns.Reset([]int16{3})

	for _, dnschanger := range dnschg {
		r.btns.AddBtcommon(dnschanger)
	}

	r.btns.AddCloseBack()
	
	callback, err := r.callbackreciver(builder.ExportDnsServer(tag), r.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnBack:
		statch(dnssrvs)
		return nil
	case C.BtnClose:
		statch(buildclose)
		return nil
	case "detour":
		//r.alertsender("tip := Use the 'set detour' ⚙️ in the previous step to validate your inputs ✅, 🔗.")
	}


	replymg, err := r.sendreciver(commonmsg(callback.Data))

	if err != nil {
		return err
	}

	if replymg.Command() == "cancel" {
		return nil
	}
	
	if replymg.Text == "." {
		replymg.Text = ""	
	}

	err = builder.ChangeDnsServerOpts(tag, callback.Data, replymg.Text)
	if err!= nil {
		r.alertsender("dns opt change failed err - " + err.Error())
	}
	
	return nil
}




type rulesetchanger struct {
	btns            *botapi.Buttons
	sendreciver     common.Sendreciver
	callbackreciver common.Callbackreciver
	alertsender     common.Alertsender
}

func (r *rulesetchanger) Can(ottype any) bool {
	_, ok := ottype.(*option.RuleSet)
	return ok
}

func (r *rulesetchanger) MakeChange(builder *builder.BuildConfig, tag string, statch func(int)) error {
	r.btns.Reset([]int16{3})

	for _, dnschanger := range rulesetchgers {
		r.btns.AddBtcommon(dnschanger)
	}

	r.btns.AddCloseBack()

	callback, err := r.callbackreciver(builder.ExportRuleSet(tag), r.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnBack:
		statch(rtsets)
		return nil
	case C.BtnClose:
		statch(buildclose)
		return nil
	}


	replymg, err := r.sendreciver(commonmsg(callback.Data))

	if err != nil {
		return err
	}

	if replymg.Command() == "cancel" {
		return nil
	}
	
	if replymg.Text == "." {
		replymg.Text = ""	
	}

	err = builder.ChangeRuleSet(tag, callback.Data, replymg.Text)
	if err!= nil {
		r.alertsender("ruleset opt change failed err - " + err.Error())
	}
	
	return nil

}



type clash_api_chg struct {
	btns            *botapi.Buttons
	sendreciver     common.Sendreciver
	callbackreciver common.Callbackreciver
	alertsender     common.Alertsender
}

func (r *clash_api_chg) Can(ottype any) bool {
	switch val := ottype.(type) {
	case string:
		return val == "clash"
	}
	return false
}

func (r *clash_api_chg) MakeChange(builder *builder.BuildConfig, tag string, statch func(int)) error {
	r.btns.Reset([]int16{3})

	for _, clash_chger := range clash_chg {
		r.btns.AddBtcommon(clash_chger)
	}

	r.btns.AddCloseBack()

	callback, err := r.callbackreciver(builder.ExportExpermental(), r.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnBack:
		statch(clash_api)
		return nil
	case C.BtnClose:
		statch(buildclose)
		return nil
	}


	replymg, err := r.sendreciver(commonmsg(callback.Data))

	if err != nil {
		return err
	}

	if replymg.Command() == "cancel" {
		return nil
	}
	
	if replymg.Text == "." {
		replymg.Text = ""	
	}

	err = builder.ChangeClash(callback.Data, replymg.Text)
	if err!= nil {
		r.alertsender("clash opt change failed err - " + err.Error())
	}
	
	return nil

}




type clash_file_chg struct {
	btns            *botapi.Buttons
	sendreciver     common.Sendreciver
	callbackreciver common.Callbackreciver
	alertsender     common.Alertsender
}

func (r *clash_file_chg) Can(ottype any) bool {
	switch val := ottype.(type) {
	case string:
		return val == "cache"
	}
	return false
}

func (r *clash_file_chg) MakeChange(builder *builder.BuildConfig, tag string, statch func(int)) error {
	r.btns.Reset([]int16{3})

	for _, cache_chgr := range cache_chg {
		r.btns.AddBtcommon(cache_chgr)
	}

	r.btns.AddCloseBack()

	callback, err := r.callbackreciver(builder.ExportExpermental(), r.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnBack:
		statch(cache_file)
		return nil
	case C.BtnClose:
		statch(buildclose)
		return nil
	}


	replymg, err := r.sendreciver(commonmsg(callback.Data))

	if err != nil {
		return err
	}

	if replymg.Command() == "cancel" {
		return nil
	}
	
	if replymg.Text == "." {
		replymg.Text = ""	
	}

	err = builder.ChangeCache(callback.Data, replymg.Text)
	if err!= nil {
		r.alertsender("cache opt change failed err - " + err.Error())
	}
	
	return nil

}






type inboundchanger struct {
	btns            *botapi.Buttons
	sendreciver     common.Sendreciver
	callbackreciver common.Callbackreciver
	alertsender     common.Alertsender
	transwiz        []transportwizard
}

func (c *inboundchanger) Can(ottype any) bool {
	return ottype == "tun" 
}

func (c *inboundchanger) MakeChange(builder *builder.BuildConfig, inname string, statch func(int)) error {
	c.btns.Reset([]int16{2})
	// c.btns.AddBtcommon("change sni")
	// c.btns.AddBtcommon("change server")
	// c.btns.AddBtcommon("change uuid")
	c.btns.AddBtcommon("change inbound self")
	c.btns.AddBtcommon("change tls settings")
	c.btns.AddBtcommon("change listen settings")
	c.btns.AddBtcommon("change inbound options")
	c.btns.AddBtcommon("change transport settings")
	c.btns.AddBtcommon("change multiplex settings")
	c.btns.AddBtcommon("replace tls settings")
	c.btns.AddBtcommon("replace transport settings")

	c.btns.AddBtcommon("remove tls")
	c.btns.AddCloseBack()

	callback, err := c.callbackreciver(builder.ExportInbound(inname), c.btns)
	if err != nil {
		return err
	}

	switch callback.Data {
	case "change inbound self":

		switch builder.GetInType(inname) {
		case "tun":
			return c.changeloop(builder, inname, builder.ChangeSelfInbound, tunchanger)
		default:
			c.alertsender("thease inbound type changing not supported yet")
		}
	case "change inbound options":
		return c.changeloop(builder, inname, builder.ChangeInboundOption, inopts)

	
	
	case C.BtnBack:
		statch(inbound)
		return nil
	case C.BtnClose:
		statch(buildclose)
	default:
		c.alertsender("not avlbl yet")

	}

	return nil
}

func (c *inboundchanger) changeloop(builder *builder.BuildConfig, inname string, chg func(q, sq, qs string) error, fileds []string ) error {
	self:
	for {
		c.btns.Reset([]int16{3})
		for _, otchg := range fileds {
			c.btns.Addbutton(otchg + " ⚙", otchg, "")
		}
		c.btns.AddCloseBack()
		callback, err := c.callbackreciver(builder.ExportInbound(inname), c.btns)
		if err != nil {
			return err
		}
		switch callback.Data {
		case C.BtnBack:
			break self 
		case C.BtnClose:
			return nil
		}

		replymg, err := c.sendreciver(commonmsg(callback.Data))

		if err != nil {
			return nil
		}

		if replymg.Command() == "cancel" {
			continue
		}
		
	
		if replymg.Text == "." {
			replymg.Text = ""	
		}
		err = chg(inname, callback.Data, replymg.Text)
		//err = builder.ChangeSelf(outname, callback.Data, replymg.Text)
		if err != nil {
			c.alertsender("change failed - " + err.Error())
		}

	}
	return nil
} 

