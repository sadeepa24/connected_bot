package builder

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	option "github.com/sadeepa24/connected_bot/builder/sbox_option/v1"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	// option "github.com/sagernet/sing-box/option"
)

type DnsRule struct {
	Rule       option.DNSRule `json:"rule,omitempty"`
	Info       string         `json:"info,omitempty"`
	Reqirments ReqirmentsRule `json:"reqirments,omitempty"`
}
type RouteRule struct {
	Rule       option.Rule    `json:"rule,omitempty"`
	Info       string         `json:"info,omitempty"` // info about rule object
	Reqirments ReqirmentsRule `json:"reqirments,omitempty"`
}

type Outbound struct {
	Out        option.Outbound    `json:"out,omitempty"`
	Info       string             `json:"info,omitempty"`
	Reqirments ReqirmentsOutbound `json:"reqirments,omitempty"`
	Price      int64              `json:"price,omitempty"` //TODO: may be later
}

type DnsServer struct {
	Server option.DNSServerOptions   `json:"server,omitempty"`
	Info string 					 `json:"info,omitempty"`
}

type RuleSet struct {
	RuleSet option.RuleSet `json:"rule_set,omitempty"`
	Info string 		   `json:"info,omitempty"`
}


// common requrments for all outbounds common
type ReqirmentsOutbound struct {
	NeedUUID          bool `json:"uuid,omitempty"`
	NeedServer        bool `json:"server,omitempty"`
	NeedPort          bool `json:"port,omitempty"`
	NeedSni           bool `json:"server_name,omitempty"`
	NeedTag           bool `json:"tag,omitempty"`
	NeedTransportHost bool `json:"transporthost,omitempty"`
}

type ReqirmentsRule struct {
	IsStatic        bool `json:"static,omitempty"`
	NeedPort        bool `json:"needport,omitempty"`
	NeedPortRange   bool `json:"needportRange,omitempty"`
	NeedDomain      bool `json:"needdomain,omitempty"`
	NeedProtocole   bool `json:"needprotocole,omitempty"`
	NeedIp_cidr     bool `json:"needip_cidr,omitempty"`
	NeedRuleSet     bool `json:"needruleSet,omitempty"`
	NeedToChangeOut bool `json:"changeout,omitempty"`
}

type RuleSetter interface {
	SetOut(out string)
	SetList(name string, list []string) error
}

func (o *Outbound) FillReqirments(callback func(msg any) (string, error)) error {
	//reqcopy := Reqirments{}
	r := o.Reqirments

	if r.NeedUUID {
		uid, err := callback("send uuid")
		if err != nil {
			return errors.New("callback reciving error")
		}
		_, err = uuid.FromString(uid)
		if err != nil {
			return errors.New("uuid parsing error ")
		}
		o.Out.Set("", uid)
	}
	if o.Reqirments.NeedServer {
		addr, err := callback("send server ip or domain")
		if err != nil {
			return errors.New("callback reciving error")
		}
		o.Out.SetServer(addr)
	}
	if o.Reqirments.NeedTransportHost {
		addr, err := callback("send transport host (one of ws, http, httpupgrade according to outbound)")
		if err != nil {
			return errors.New("callback reciving error")
		}
		o.Out.SetTransPortHost(addr)
	}
	if o.Reqirments.NeedPort {
		strport, err := callback("send server port")
		if err != nil {
			return errors.New("callback reciving error")
		}
		port, err := strconv.Atoi(strport)

		if err != nil {
			return errors.New("invalid port")
		}
		o.Out.SetPort(uint16(port))
	}
	if o.Reqirments.NeedSni {
		sni, err := callback("send sni")
		if err != nil {
			return errors.New("callback reciving error")
		}
		o.Out.Set(sni, "")
	}
	if o.Reqirments.NeedTag {
		tag, err := callback("send tagname")
		if err != nil {
			return errors.New("callback reciving error")
		}
		o.Out.Tag = tag
	}

	return nil
}

// callback reciver should support even msgs without buttons
func (r *ReqirmentsRule) FillReqirments(callback common.Callbackreciver, sendreciver common.Sendreciver, ruleany any) error {
	if r.IsStatic {
		return nil
	}
	rule, ok := ruleany.(RuleSetter)
	if !ok {
		return errors.New("rule does not implemet the interface")
	}

	if r.NeedDomain {

		msg, err := sendreciver("send commma seprated domain list if you want to skip send " + C.SkipDelim)
		if err != nil {
			return err
		}
		if !(msg.Text == C.SkipDelim) {
			rule.SetList("Domain", strings.Split(msg.Text, ","))
		}
	}
	if r.NeedIp_cidr {
		msg, err := sendreciver("send commma seprated ip_cidr list ex := 1.1.1.0/24,104.20.0.0/16 . if you want to skip send " + C.SkipDelim)
		if err != nil {
			return err
		}
		if !(msg.Text == C.SkipDelim) {
			rule.SetList("ip_cidr", strings.Split(msg.Text, ","))
		}
	}
	if r.NeedProtocole {
		msg, err := sendreciver("send commma seprated protcole list   ( tls, bittorrent ) . if you want to skip send " + C.SkipDelim)
		if err != nil {
			return err
		}
		if !(msg.Text == C.SkipDelim) {
			rule.SetList("Protocole", strings.Split(msg.Text, ","))
		}
	}
	if r.NeedPortRange {
		msg, err := sendreciver("send commma seprated port range list  (3000:, 80-150 ) ex := 1000:2000,:3000,4000: . if you want to skip send " + C.SkipDelim)
		if err != nil {
			return err
		}
		if !(msg.Text == C.SkipDelim) {
			rule.SetList("port_range", strings.Split(msg.Text, ","))
		}
	}

	//TODO: add later
	if r.NeedToChangeOut {

	}

	//TODO: add ruleset later
	// if r.NeedRuleSet {
	// 	rset := option.Listable[string]{}
	// 	for r.
	// }

	return nil
}

type StoreOption struct {
	AllDnsRule    []DnsRule                 `json:"dnsrules,omitempty"`
	AllDnsServer  []DnsServer				`json:"dns_servers,omitempty"`
	AllRouteRules []RouteRule               `json:"routerule,omitempty"`
	AllRuleSet    []RuleSet                 `json:"ruleset,omitempty"`
	AllOutbounds  []Outbound                `json:"outbounds,omitempty"`
}

// Main config store object
type ConfigStore struct {
	dnsRules   map[string]DnsRule
	routeRule  map[string]RouteRule
	dnsServers map[string]DnsServer
	ruleSets   map[string]RuleSet
	outbounds  map[string]Outbound

	DnsRuleTags   []string
	RouteRuleTags []string
	DnsServerTags []string

	defaultDns option.DNSServerOptions

	allinfo map[string]string
}

func NewConfStore(path string) (*ConfigStore, error) {

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var StoreOpt StoreOption
	if err = json.Unmarshal(file, &StoreOpt); err != nil {

		return nil, err
	}
	store := &ConfigStore{allinfo: map[string]string{}}

	store.dnsRules = C.SliceToMap(StoreOpt.AllDnsRule, func(rule DnsRule) string {
		store.DnsRuleTags = append(store.DnsRuleTags, rule.Rule.Tag)
		return rule.Rule.Tag
	})
	store.routeRule = C.SliceToMap(StoreOpt.AllRouteRules, func(rule RouteRule) string {
		store.RouteRuleTags = append(store.RouteRuleTags, rule.Rule.Tag)
		return rule.Rule.Tag
	})
	store.dnsServers = C.SliceToMap(StoreOpt.AllDnsServer, func(server DnsServer) string {
		store.DnsServerTags = append(store.DnsServerTags, server.Server.Tag)
		return server.Server.Tag
	})
	store.ruleSets = C.SliceToMap(StoreOpt.AllRuleSet, func(set RuleSet) string { return set.RuleSet.Tag })
	store.outbounds = C.SliceToMap(StoreOpt.AllOutbounds, func(out Outbound) string { return out.Out.Tag })

	if dfsrv, ok := store.dnsServers["default"]; ok {
		store.defaultDns = dfsrv.Server
	} else {
		store.defaultDns = option.DNSServerOptions{
			Tag:     "default",
			Address: "tcp://1.1.1.1",
			Detour:  "default",
		}
	}

	return store, err
}

func (c *ConfigStore) DnsRuleMust(tag string) DnsRule {
	rule, ok := c.dnsRules[tag]
	if !ok {
		return DnsRule{}
	}
	return rule
}

func (c *ConfigStore) DnsRuleByname(name string) (option.DNSRule, error) {
	rule, ok := c.dnsRules[name]
	if !ok {
		return option.DNSRule{}, errors.New("no rule for the givin name")
	}
	return rule.Rule, nil
}

func (c *ConfigStore) RouteRuleByname(name string) (option.Rule, error) {
	rule, ok := c.routeRule[name]
	if !ok {
		return option.Rule{}, errors.New("no rule for the givin name")
	}
	return rule.Rule, nil
}

func (c *ConfigStore) FullRuleByname(name string) (RouteRule, bool) {
	rule, ok := c.routeRule[name]
	return rule, ok
}
func (c *ConfigStore) FullDnsRuleByname(name string) (DnsRule, bool) {
	rule, ok := c.dnsRules[name]
	return rule, ok
}

func (c *ConfigStore) DnsServerbyTag(tag string) (DnsServer, error) {
	server, ok := c.dnsServers[tag]
	if !ok {
		return server, errors.New("no server for the tag - " + tag)
	}
	return server, nil
}

func (c *ConfigStore) RuleSetBytag(tag string) (RuleSet, error) {
	ruleset, ok := c.ruleSets[tag]
	if !ok {
		return ruleset, errors.New("no rule set for the tag - " + tag)
	}
	return ruleset, nil
}

func (c *ConfigStore) DefaultDns() option.DNSServerOptions {
	return c.defaultDns
}


func (c *ConfigStore) GetOutbound(refName string) (Outbound, error) {
	out, ok := c.outbounds[refName]
	if !ok {
		//out.Out.Tag = refName + parentName
		return out, errors.New("no out found")
	}

	//TODO: build proper copier later
	outs := Outbound{
		Out:        option.Outbound{},
		Info:       out.Info,
		Reqirments: out.Reqirments,
		Price:      out.Price,
	}
	outraw, err := out.Out.MarshalJSON()
	if err != nil {
		return out, err
	}
	if err = json.Unmarshal(outraw, &outs.Out); err != nil {
		return out, err
	}

	return outs, nil

}

func (c *ConfigStore) Alloutbounds() []string {
	obs := []string{}
	for tag := range c.outbounds {
		obs = append(obs, tag)
	}
	return obs
}

// return all rules with requrments
func (c *ConfigStore) AllRouteRulesWithReq() []RouteRule {
	rules := []RouteRule{}
	for _, rule := range c.routeRule {
		if rule.Reqirments.IsStatic {
			continue
		}
		rules = append(rules, rule)

	}
	return rules
}

func (c *ConfigStore) AllDnsRule() []DnsRule {
	rules := []DnsRule{}
	for _, rule := range c.dnsRules {
		if rule.Reqirments.IsStatic {
			continue
		}
		rules = append(rules, rule)

	}
	return rules
}

func (c *ConfigStore) AllRUleSet() []RuleSet {
	return C.MapToSlice(c.ruleSets)
}

func (c *ConfigStore) AllDnsServer() []DnsServer {
	return C.MapToSlice(c.dnsServers)
}


/*

requirments

uuid string
tagname string
host, sni string
ip addres string


*/
