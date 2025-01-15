package builder

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/netip"
	"os"
	"time"

	"github.com/gofrs/uuid"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	option "github.com/sadeepa24/connected_bot/sbox_option/v1"
	sboxConst "github.com/sagernet/sing-box/constant"
	singJson "github.com/sagernet/sing/common/json"
	"go.uber.org/zap"
)

type BuildConfig struct {
	ctx context.Context //builder life

	path   string
	Isnew  bool
	store  *ConfigStore
	config option.Options
	//callbacks map[string]any

	DefaultBuilt bool

	outbounds  map[string]*option.Outbound
	inbounds   map[string]*option.Inbound
	routeRules map[string]*option.Rule    //thease two map used to deal with rules it's easy to deal with map rather than slice
	dnsRules   map[string]*option.DNSRule // after all changes convet thease to into original slice
	dnsServers map[string]*option.DNSServerOptions
	ruleSet    map[string]*option.RuleSet

	DefaultDnsServerTag string
	DefaultOutbound     string
	DefaultInbound      string

	//OptEntry *OptionsEntry

	//DnsOpt *option.DNSOptions
	RouteOpt *option.RouteOptions

	logger *zap.Logger

	exportafter func(b []byte) any
}

const (
	outSelector string = "selector"
)

func NewBuilder(ctx context.Context, path string, store *ConfigStore, logger *zap.Logger) (*BuildConfig, error) {

	build := &BuildConfig{
		ctx:        ctx,
		logger:     logger,
		path:       path,
		store:      store,
		inbounds:   map[string]*option.Inbound{},
		outbounds:  map[string]*option.Outbound{},
		routeRules: map[string]*option.Rule{},
		dnsRules:   map[string]*option.DNSRule{},
		config:     option.Options{},
	}

	file, err := os.OpenFile("./configs/"+path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	st, err := file.Stat()
	if err != nil {
		return nil, err
	}
	build.Isnew = st.Size() == 0

	if !build.Isnew {
		entry, err := readConfigAt(file)
		if err != nil {
			return nil, err
		}
		build.config = entry.options
		build.RouteOpt = entry.options.Route
	} else {
		build.preconfig()
	}

	//build.DnsOpt = build.config.DNS
	build.inbounds = C.SliceToMapPtr(build.config.Inbounds, func(op option.Inbound) string { return op.Tag })
	build.outbounds = C.SliceToMapPtr(build.config.Outbounds, func(op option.Outbound) string { return op.Tag })
	build.routeRules = C.SliceToMapPtr(build.config.Route.Rules, func(op option.Rule) string { return op.Tag })
	build.dnsRules = C.SliceToMapPtr(build.config.DNS.Rules, func(op option.DNSRule) string { return op.Tag })
	build.dnsServers = C.SliceToMapPtr(build.config.DNS.Servers, func(op option.DNSServerOptions) string { return op.Tag })
	build.ruleSet = C.SliceToMapPtr(build.config.Route.RuleSet, func(op option.RuleSet) string { return op.Tag })
	if err = build.BuildDefault(); err != nil {
		return nil, err
	}

	return build, nil
}

// optional
func (b *BuildConfig) AddExperminatl() error {

	return nil
}

func (b *BuildConfig) preconfig() {
	b.config = option.Options{
		Route:     &option.RouteOptions{},
		Inbounds:  []option.Inbound{},
		Outbounds: []option.Outbound{},
		DNS:       &option.DNSOptions{},
	}
}

func (b *BuildConfig) AddClashApi() error {
	if b.config.Experimental == nil {
		b.config.Experimental = &option.ExperimentalOptions{}
	}

	// b.config.Experimental.CacheFile = &option.CacheFileOptions{
	// 	Enabled: true,
	// }
	// default
	b.config.Experimental.ClashAPI = &option.ClashAPIOptions{
		ExternalController: "127.0.0.1:9090",
		ExternalUI:         "yacd",
		DefaultMode:        "clash",
	}
	return nil
}

func (b *BuildConfig) AddCacheFile() error {
	if b.config.Experimental == nil {
		b.config.Experimental = &option.ExperimentalOptions{}
	}

	b.config.Experimental.CacheFile = &option.CacheFileOptions{
		Enabled: true,
		Path: "",
		CacheID: "connected_cache",
		StoreFakeIP: true,
		StoreRDRC: true,
		RDRCTimeout: option.Duration(10 * time.Second),
	}
	return nil
}

func (b *BuildConfig) RemoveClashApi() {
	if b.config.Experimental == nil {
		return
	}
	b.config.Experimental.ClashAPI = nil
}
func (b *BuildConfig) RemoveCacehFile() {
	if b.config.Experimental == nil {
		return
	}
	b.config.Experimental.CacheFile = nil
}

func (b *BuildConfig) CacheFile() error {
	if b.config.Experimental == nil {
		b.config.Experimental = &option.ExperimentalOptions{}
	}
	// b.config.Experimental.CacheFile = &option.CacheFileOptions{
	// 	Enabled: true,
	// }

	return nil
}

// BuildDefault
func (b *BuildConfig) BuildDefault() error {
	var err error
	if err = b.BuildDefaultLog(); err != nil {
		return err
	}
	if err = b.BuildDefaultDns(); err != nil {
		return err
	}
	if err = b.BuildDefaultInbounds(); err != nil {
		return err
	}
	if err = b.BuildDefaultOutbound(); err != nil {
		return err
	}
	if err = b.BuildDefaultRoute(); err != nil {
		return err
	}
	b.CacheFile()

	b.DefaultBuilt = true

	return nil
}

func (b *BuildConfig) BuildDefaultLog() error {
	b.config.Log = &option.LogOptions{
		Disabled:  false,
		Level:     "warn",
		Timestamp: false,
	}
	return nil
}

// DNS handlin
func (b *BuildConfig) BuildDefaultDns() error {
	if b.config.DNS == nil {
		b.config.DNS = &option.DNSOptions{
			Servers:          []option.DNSServerOptions{},
			DNSClientOptions: option.DNSClientOptions{},
			Rules:            []option.DNSRule{},
		}
		//b.DnsOpt = b.config.DNS
	}

	var (
		blockavbl   bool
		defaultavbl bool
		lclavbl bool
	)

	for _, server := range b.dnsServers {
		switch server.Tag {
		case "block":
			blockavbl = true
		case "default":
			defaultavbl = true
		case "local":
			lclavbl = true
		}
	}

	if !defaultavbl {
		b.dnsServers[b.store.defaultDns.Tag] = &b.store.defaultDns
		//b.DnsOpt.Servers = append(b.DnsOpt.Servers, b.store.DefaultDns())

	}
	if !blockavbl {
		// b.DnsOpt.Servers = append(b.DnsOpt.Servers, option.DNSServerOptions{
		// 	Tag: "block",
		// 	Address: "rcode://success",
		// })
		b.dnsServers["block"] = &option.DNSServerOptions{
			Tag:     "block",
			Address: "rcode://success",
		}
	}

	if !lclavbl {
		b.dnsServers["local"] = &option.DNSServerOptions{
			Tag:     "local",
			Address: "local",
		}
	}

	if _, ok := b.dnsServers[b.config.DNS.Final]; !ok || b.config.DNS.Final == "" {
		b.config.DNS.Final = "default"
	}
	b.DefaultDnsServerTag = "default"
	return nil
}

func (b *BuildConfig) AddDnsServer(dnsServer option.DNSServerOptions) error {

	if dnsServer.Tag == "" {
		return errors.New("dns server tag not found")
	}

	if _, ok := b.dnsServers[dnsServer.Tag]; ok {
		return errors.New("there is already dnsserver called " + dnsServer.Tag)
	}

	if dnsServer.Detour != "" {
		if _,ok := b.outbounds[dnsServer.Detour]; !ok {
			return errors.New("detour "+ dnsServer.Detour +" not found please set detour in config")
		}
	}
	if dnsServer.AddressResolver != "" {
		if _, ok := b.dnsServers[dnsServer.AddressResolver]; !ok {
			return errors.New("dns resolver "+ dnsServer.AddressResolver +" not found please set avalble resolver in config")
		}
	}
	b.dnsServers[dnsServer.Tag] = &dnsServer
	return nil
}


func (b *BuildConfig) SetDefaultDns(serverTag string) error {
	var ok bool
	for _, dnsser := range b.config.DNS.Servers {
		if dnsser.Tag == serverTag {
			ok = true
			break
		}
	}
	if !ok {
		b.config.DNS.Final = b.DefaultDnsServerTag
		return errors.New("server tag name not found in dnsserver list")
	}
	b.DefaultDnsServerTag = serverTag
	b.config.DNS.Final = serverTag
	return nil
}

func (b *BuildConfig) AddDnsRule(tagName string, server string) error {

	if b.config.DNS == nil {
		if err := b.BuildDefaultDns(); err != nil {
			return err
		}
	}

	if tagName == "" {
		//TODO: get info from client and prosess rule here
		return nil
	}

	if _, ok := b.dnsRules[tagName]; ok {
		return errors.New("rule already excist")
	}

	rule, err := b.store.DnsRuleByname(tagName)
	if err != nil {
		return err
	}

	if len(rule.DefaultOptions.RuleSet) > 0 {
		for _, tagofset := range rule.DefaultOptions.RuleSet {
			if _, ok := b.ruleSet[tagofset]; ok {
				continue
			}
			set, err := b.store.RuleSetBytag(tagofset)
			if err != nil{
				return errors.New("rule set - " + tagofset + " not found ")
			}
			b.ruleSet[set.RuleSet.Tag] = &set.RuleSet
		}

	}

	var (
		srvok   bool
		ruleout bool
	)
	for _, srv := range b.dnsServers {
		if srv.Tag == server {
			srvok = true
		}
		if rule.DefaultOptions.Server == srv.Tag {
			ruleout = true
		}
	}

	if !srvok {
		server = b.DefaultDnsServerTag
	}
	if !ruleout {
		rule.DefaultOptions.Server = server
	}

	if len(rule.DefaultOptions.Outbound) > 0 {
		for i, def := range rule.DefaultOptions.Outbound {
			_, ok := b.outbounds[def]
			if ok {
				continue
			}
			rule.DefaultOptions.Outbound[i] = b.DefaultOutbound
		}

	}

	//b.config.DNS.Rules = append(b.config.DNS.Rules, rule)
	b.dnsRules[rule.Tag] = &rule
	return nil
}

func (b *BuildConfig) RemoveDnsRule(ruleTag string) {
	delete(b.dnsRules, ruleTag)
}

func (b *BuildConfig) RemoveDnsServer(tag string) error {
	switch tag {
	case b.DefaultDnsServerTag, "block", "default":
		return errors.New("cannot delete this dns server it is neccry to config boiler plate")
	}

	C.ExcuteMap(b.dnsRules, func(val *option.DNSRule, key string) {
		if val.DefaultOptions.Server == tag {
			val.DefaultOptions.Server = b.DefaultDnsServerTag
		}
		if val.LogicalOptions.Server == tag {
			val.LogicalOptions.Server = b.DefaultDnsServerTag
		}

	})

	if b.config.DNS.Final == tag {
		b.config.DNS.Final = "default"
	}

	delete(b.dnsServers, tag)

	return nil
}

func (b *BuildConfig) DnsClient() error {
	//TODO: need many callbacks
	// get nedded info via calling callbacks
	clientopt := option.DNSClientOptions{}
	b.config.DNS.DNSClientOptions = clientopt

	return nil
}

func (b *BuildConfig) CheckDnsRule(tag string) bool {
	_, ok := b.dnsRules[tag]
	return ok
}

func (b *BuildConfig) CheckDnsServer(tag string) bool {
	_, ok := b.dnsServers[tag]
	return ok
}

func (b *BuildConfig) CheckRuleSet(tag string) bool {
	_, ok := b.ruleSet[tag]
	return ok
}

func (b *BuildConfig) ChangeDnsServerOpts(tag, changer, val string) error {
	dnsServer, ok := b.dnsServers[tag]
	if !ok {
		return errors.New("dns server not found " + tag)
	}
	return dnsServer.Changer(changer, val)
}

func (b *BuildConfig) AddDnsServerRaw(in []byte) error {
	var dnsServer option.DNSServerOptions
	if err := json.Unmarshal(in, &dnsServer); err != nil {
		return err
	}

	return b.AddDnsServer(dnsServer)

} 

func (b *BuildConfig) AddRawDns(in []byte, tagName string) error {
	var dnsrule option.DNSRule

	if err := dnsrule.UnmarshalJSON(in); err != nil {
		return err
	}
	dnsrule.Tag = tagName
	//dnsrule.IsValid()

	switch dnsrule.Type {
	case sboxConst.RuleTypeLogical:
		return errors.New("logical rules cannot add")
	}

	if len(dnsrule.DefaultOptions.Outbound) > 0 {
		for i, out := range dnsrule.DefaultOptions.Outbound {
			if _, ok := b.outbounds[out]; ok {
				dnsrule.DefaultOptions.Outbound[i] = b.DefaultOutbound
			}

		}
	}
	if len(dnsrule.DefaultOptions.Inbound) > 0 {
		for i, in := range dnsrule.DefaultOptions.Inbound {
			if _, ok := b.inbounds[in]; ok {
				dnsrule.DefaultOptions.Inbound[i] = b.DefaultInbound
			}

		}
	}
	var srvok bool
	for _, v := range b.config.DNS.Servers {
		if v.Tag == dnsrule.DefaultOptions.Server {
			srvok = true
			break
		}
	}
	if !srvok {
		dnsrule.DefaultOptions.Server = b.DefaultDnsServerTag
	}

	b.dnsRules[tagName] = &dnsrule

	return nil
}

func (b *BuildConfig) AddDnsRuleObj(rule option.DNSRule) error {

	if b.config.DNS == nil {
		if err := b.BuildDefaultDns(); err != nil {
			return err
		}
	}

	tagName := rule.Tag
	server := rule.DefaultOptions.Server //TODO: only support default option

	if tagName == "" {
		//TODO: get info from client and prosess rule here
		return nil
	}

	if _, ok := b.dnsRules[tagName]; ok {
		return errors.New("rule exit")
	}

	if len(rule.DefaultOptions.RuleSet) > 0 {
		for _, tagofset := range rule.DefaultOptions.RuleSet {
			if _, ok := b.ruleSet[tagofset]; ok {
				continue
			}
			set, err := b.store.RuleSetBytag(tagofset)
			if err != nil{
				return errors.New("rule set - " + tagofset + " not found ")
			}
			b.ruleSet[set.RuleSet.Tag] = &set.RuleSet
		}

	}

	var (
		srvok   bool
		ruleout bool
	)
	for _, srv := range b.dnsServers {
		if srv.Tag == server {
			srvok = true
		}
		if rule.DefaultOptions.Server == srv.Tag {
			ruleout = true
		}
	}

	if !srvok {
		server = b.DefaultDnsServerTag
	}
	if !ruleout {
		rule.DefaultOptions.Server = server
	}

	if len(rule.DefaultOptions.Outbound) > 0 {
		for i, def := range rule.DefaultOptions.Outbound {
			_, ok := b.outbounds[def]
			if ok {
				continue
			}
			rule.DefaultOptions.Outbound[i] = b.DefaultOutbound
		}

	}

	//b.config.DNS.Rules = append(b.config.DNS.Rules, rule)
	b.dnsRules[rule.Tag] = &rule
	return nil
}

// Routing Handling
func (b *BuildConfig) BuildDefaultRoute() error {
	if b.config.Route != nil {
		b.config.Route.Final = b.DefaultOutbound
		b.config.Route.AutoDetectInterface = true
		b.routeRules["dfdns"] = &option.Rule{
			Tag: "dfdns",
			Type: sboxConst.RuleTypeDefault,
			DefaultOptions: option.DefaultRule{
				Protocol: option.Listable[string]{"dns"},
				Outbound: "dns-out",
			},
		}
		
		return nil
	}

	b.config.Route = &option.RouteOptions{
		Final:               b.DefaultOutbound, //TODO: select default final later
		AutoDetectInterface: true,
		Rules:               []option.Rule{},
	}
	b.RouteOpt = b.config.Route

	return nil
}

// should update fetched rules again after
func (b *BuildConfig) RemoveRouteRule(tag string) {
	delete(b.routeRules, tag)
}
func (b *BuildConfig) RemoveRouteRuleSet(tag string) error {
	_, ok := b.ruleSet[tag]

	if !ok {
		return errors.New("cannot find rule set called "+ tag )
	}

	for _, rr := range b.dnsRules {
		if C.IsInSlice(rr.DefaultOptions.RuleSet, func(e string) bool { return tag == e }) {
			return errors.New("can't remove rule set, ruleset used in dns rules")
		}
		if rr.Type == "logical" {
			for _, rin := range rr.LogicalOptions.Rules {
				if C.IsInSlice(rin.DefaultOptions.RuleSet, func(e string) bool { return tag == e }) {
					return errors.New("can't remove rule set, ruleset used in loghical dns rules")
				}
			} 
		}
	}

	for _, rr := range b.routeRules {
		if C.IsInSlice(rr.DefaultOptions.RuleSet, func(e string) bool { return tag == e }) {
			return errors.New("can't remove rule set, ruleset used in route rules")
		}
		if rr.Type == "logical" {
			for _, rin := range rr.LogicalOptions.Rules {
				if C.IsInSlice(rin.DefaultOptions.RuleSet, func(e string) bool { return tag == e }) {
					return errors.New("can't remove rule set, ruleset used in logical routing rules")
				}
			} 
		}
	}

	delete(b.ruleSet, tag)

	return nil
	
}

func (b *BuildConfig) AddRouteRuleSet(ruleset option.RuleSet) error {
	_, ok := b.ruleSet[ruleset.Tag]
	if ok {
		return errors.New("there is already ruleset called "+ ruleset.Tag)
	}

	if ruleset.RemoteOptions.DownloadDetour != "" {
		_, ok := b.outbounds[ruleset.RemoteOptions.DownloadDetour]
		if !ok {
			return errors.New("outbound detour not found "+ ruleset.RemoteOptions.DownloadDetour)
		}
	}
	b.ruleSet[ruleset.Tag] = &ruleset
	return nil
}

func (b *BuildConfig) AddRouteRuleSetRaw(rulesetraw []byte) error {
	var ruleset option.RuleSet
	if err := json.Unmarshal(rulesetraw, &ruleset); err != nil {
		return err
	}
	return b.AddRouteRuleSet(ruleset)
}

func (b *BuildConfig) AddRouteRule(tagName string, outbound string) error {

	if b.config.Route == nil {
		if err := b.BuildDefaultRoute(); err != nil {
			return err
		}
	}

	if tagName == "" {
		//TODO: get info from client and prosess rule here
		//TODO: callback function use here
		return nil
	}

	if _, ok := b.routeRules[tagName]; ok {
		return errors.New("rule exit")
	}

	//getting predefined rule
	rule, err := b.store.RouteRuleByname(tagName)
	if err != nil {
		return err
	}

	if len(rule.DefaultOptions.RuleSet) > 0 {
		for _, tagofset := range rule.DefaultOptions.RuleSet {
			if _, ok := b.ruleSet[tagofset]; ok {
				continue
			}
			set, err := b.store.RuleSetBytag(tagofset)
			if err != nil {
				return errors.New("rule set - " + tagofset + " not found ")
			}
			b.ruleSet[set.RuleSet.Tag] = &set.RuleSet
		}
	}

	_, ok := b.outbounds[outbound]
	if !ok {
		outbound = b.DefaultOutbound
	}

	if rule.DefaultOptions.Outbound == "" {
		rule.DefaultOptions.Outbound = outbound
	} else {
		_, ok := b.outbounds[rule.DefaultOptions.Outbound]
		if !ok {
			rule.DefaultOptions.Outbound = outbound
		}
	}

	//b.config.Route.Rules = append(b.config.Route.Rules, rule)
	b.routeRules[rule.Tag] = &rule
	return nil
}

func (b *BuildConfig) SetListTorule(tagname, listname string, listable []string) error {
	if rule, ok := b.routeRules[tagname]; ok {
		if err := rule.SetList(listname, listable); err != nil {
			return err
		}
		return nil
	}
	return errors.New("rule not found")
}
func (b *BuildConfig) SetRuleOutbound(tagname, outname string) error {
	if rule, ok := b.routeRules[tagname]; ok {
		rule.SetOut(outname)
		return nil
	}
	return errors.New("rule not found")
}

func (b *BuildConfig) AddRawRoute(in []byte, ref string) error {
	if _, ok := b.routeRules[ref]; ok {
		return errors.New(" route rule called  " + ref + " already extig in config please send diffrecnt reffrance name")
	}

	rt := option.Rule{}

	err := rt.UnmarshalJSON(in)
	if err != nil {
		return errors.New(" your rule object invalid please check the json object again")
	}

	switch rt.Type {
	case sboxConst.RuleTypeLogical:

		if _, ok := b.outbounds[rt.LogicalOptions.Outbound]; !ok {
			return errors.New("outbound not found in config please change outbound tag in rule object")
		}

		if len(rt.LogicalOptions.Rules) > 0 {
			for _, inb := range rt.DefaultOptions.Inbound {
				if _, ok := b.inbounds[inb]; !ok {
					return errors.New(inb + " not found in the configs inbound object")
				}
			}
		}

	case sboxConst.RuleTypeDefault:
		if _, ok := b.outbounds[rt.DefaultOptions.Outbound]; !ok {
			return errors.New("outbound not found in config please change outbound tag in rule object")
		}

		if len(rt.DefaultOptions.Inbound) > 0 {
			for _, inb := range rt.DefaultOptions.Inbound {
				if _, ok := b.inbounds[inb]; !ok {
					return errors.New(inb + " not found in the configs inbound object")
				}
			}
		}

	}

	b.routeRules[ref] = &rt

	return nil
}

func (b *BuildConfig) ChangeRuleSet(rulesetatg, changer, val string) error {
	ruleset, ok := b.ruleSet[rulesetatg]
	if !ok {
		return errors.New("cannot find ruleset")
	}

	return ruleset.Change(changer, val)
}
func (b *BuildConfig) ChangeClash(changer, val string) error {
	if b.config.Experimental != nil {
		if b.config.Experimental.ClashAPI != nil  {
			return b.config.Experimental.ClashAPI.Changer(changer, val)
		}
		return errors.New("there is no clash api to change")
	}
	return errors.New("there is no experimental to change clash api")
}
func (b *BuildConfig) ChangeCache(changer, val string) error {
	if b.config.Experimental != nil {
		if b.config.Experimental.CacheFile != nil  {
			return b.config.Experimental.CacheFile.Changer(changer, val)
		}
		return errors.New("there is no cache file fields to change")
	}
	return errors.New("there is no experimental to change clash file fields")
}



func (b *BuildConfig) AddRouteRuleOb(rule option.Rule) error {
	if _, ok := b.routeRules[rule.Tag]; ok {
		return errors.New("rule already exit, please remove it and try again")
	}

	//TODO: change rule set option later

	if len(rule.DefaultOptions.RuleSet) > 0 {
		for _, tagofset := range rule.DefaultOptions.RuleSet {
			if _, ok := b.ruleSet[tagofset]; ok {
				continue
			}
			set, err := b.store.RuleSetBytag(tagofset)
			if err != nil {
				return errors.New("rule set - " + tagofset + " not found ")
			}
			b.ruleSet[set.RuleSet.Tag] = &set.RuleSet
		}

	}

	_, ok := b.outbounds[rule.DefaultOptions.Outbound]
	if !ok {
		rule.DefaultOptions.Outbound = b.DefaultOutbound
	}

	b.routeRules[rule.Tag] = &rule
	return nil
}

func (b *BuildConfig) CheckRule(tag string) bool {
	_, ok := b.routeRules[tag]
	return ok
}

func (b *BuildConfig) GetDnsRule(tag string) option.DNSRule {
	rule, ok := b.dnsRules[tag]
	if !ok {
		return option.DNSRule{}
	}
	return *rule
}
func (b *BuildConfig) SetListToDnsRule(tagname, listname string, listable []string) error {
	if rule, ok := b.dnsRules[tagname]; ok {
		if err := rule.SetList(listname, listable); err != nil {
			return err
		}
		return nil
	}
	return errors.New("rule not found")
}

func (b *BuildConfig) SetRouteFinal(final string) error {
	out, loaded := b.outbounds[final]
	if !loaded {
		return errors.New("outbound not found")
	}
	switch out.Type {
	case "dns":
		return errors.New("outbound does not support as final")
	}
	b.config.Route.Final = out.Tag
	return nil
}

// Outbounds handling
func (b *BuildConfig) BuildDefaultOutbound() error {

	var (
		selectorcheck bool
		blockcheck    bool
		directcheck   bool
		dnsoutcheck   bool
	)

	_, selectorcheck = b.outbounds["default"]
	_, directcheck = b.outbounds["direct"]
	_, dnsoutcheck = b.outbounds["dns-out"]
	_, blockcheck = b.outbounds["block"]

	if !selectorcheck {
		ot := []string{}

		for _, out := range b.outbounds {
			if out.Type == outSelector || out.Type == "block" || out.Type == "dns" {
				continue
			}
			ot = append(ot, out.Tag)
			//b.config.Outbounds[selectorindex].SelectorOptions.Outbounds = append(b.config.Outbounds[selectorindex].SelectorOptions.Outbounds, out.Tag)
		}

		b.outbounds["default"] = &option.Outbound{
			Type: outSelector,
			Tag:  "default",
			SelectorOptions: option.SelectorOutboundOptions{
				Outbounds: ot,
			},
		}

	}

	if !blockcheck {

		b.outbounds["block"] = &option.Outbound{
			Type: "block",
			Tag:  "block",
		}
	}

	if !directcheck {

		b.outbounds["direct"] = &option.Outbound{
			Type: "direct",
			Tag:  "direct",
		}
	}

	if !dnsoutcheck {
		b.outbounds["dns-out"] = &option.Outbound{
			Type: "dns",
			Tag:  "dns-out",
		}
	}

	b.DefaultOutbound = "default"

	return nil
}

func (b *BuildConfig) GetSelector() *option.Outbound {
	//TODO:
	// this function mus return a selector if not nil pointer panic occured
	return b.outbounds["default"]
}

func (b *BuildConfig) RemoveOutbound(tagName string) error {

	switch tagName {
	case "selector", "direct", "block":
		return errors.New("cannot remove main outbounds")
	}

	delete(b.outbounds, tagName)

	b.RemoveSelector(tagName)

	C.ExcuteMap(b.routeRules, func(t *option.Rule, key string) {
		if t.DefaultOptions.Outbound == tagName {
			delete(b.routeRules, key)
		}
	})

	C.ExcuteMap(b.dnsRules, func(val *option.DNSRule, key string) {
		for j, ot := range val.DefaultOptions.Outbound {
			if ot == tagName {
				val.DefaultOptions.Outbound = append(val.DefaultOptions.Outbound[j:], val.DefaultOptions.Outbound[:j+1]...)
			}
		}
	})

	for _, server := range b.dnsServers {
		if server.Detour == tagName {
			b.dnsServers[server.Tag].Detour = b.DefaultOutbound
		}
	}

	if b.config.DNS.Final == tagName {
		b.config.DNS.Final = b.DefaultOutbound
	}

	return nil
}

// this can add outbound via userconfig only support vless ws
func (b *BuildConfig) AddOutbound(sboxot db.Config, serverIN sbox.Inboud, sni string) error {

	if _, ok := b.outbounds[sboxot.Name]; ok {
		return errors.New("outbound already exit")
	}
	if sni == "" {
		//TODO: change later
		sni = "connectebot"
	}

	vlessOutbound := option.Outbound{
		Type: sboxot.Type,
		Tag:  sboxot.Name,
	}

	switch sboxot.Type {
	case C.Vless:
		vlessOutbound.VLESSOptions = option.VLESSOutboundOptions{
			UUID: sboxot.UUID.String(),
			ServerOptions: option.ServerOptions{

				Server:     serverIN.Domain,
				ServerPort: uint16(serverIN.Port()),
			},
		}
	default:
		return errors.New("other type not supported yet")
	}

	//obfshost := b.ReqSni()
	obfshost := sni

	switch serverIN.Transporttype {
	case "ws":
		vlessOutbound.VLESSOptions.Transport = &option.V2RayTransportOptions{
			Type: serverIN.Transporttype,
			WebsocketOptions: option.V2RayWebsocketOptions{

				Headers: option.HTTPHeader{
					"host": option.Listable[string]{
						sni,
					},
				},
				Path: "/",
			},
		}
	default:
		return errors.New("ther transport does not support yet")
	}

	if serverIN.Tlsenabled {
		vlessOutbound.VLESSOptions.TLS = &option.OutboundTLSOptions{
			Enabled:    true,
			Insecure:   true,
			MinVersion: "1.1",
			MaxVersion: "1.3",
			ServerName: obfshost,
		}
	}

	//b.config.Outbounds = append(b.config.Outbounds, vlessOutbound)
	b.outbounds[vlessOutbound.Tag] = &vlessOutbound

	if out, ok := b.outbounds["default"]; ok {
		if out.Type == outSelector {
			out.SelectorOptions.Outbounds = append(out.SelectorOptions.Outbounds, vlessOutbound.Tag)
		}
	}

	return nil
}

func (b *BuildConfig) GetInType(inname string) string {
	in, ok := b.inbounds[inname]
	if !ok {
		return ""
	}
	return in.Type
}

func (b *BuildConfig) ChangeSni(outname string, newsni string) error {
	out, ok := b.outbounds[outname]
	if !ok {
		return errors.New("outbound not found")
	}
	return out.Set(newsni, "")
}
func (b *BuildConfig) ChangeTLS(outname, changer, value string) error {
	out, ok := b.outbounds[outname]
	if !ok {
		return errors.New("outbound not found")
	}
	return out.ChangeTLS(changer, value)
}
func (b *BuildConfig) ChangeSelfOutbound(outname, changer, value string) error {
	out, ok := b.outbounds[outname]
	if !ok {
		return errors.New("outbound not found")
	}
	return out.Change(changer, value)
}
func (b *BuildConfig) ChangeSelfInbound(inname, changer, value string) error {
	in, ok := b.inbounds[inname]
	if !ok {
		return errors.New("outbound not found")
	}
	return in.Change(changer, value)
}
func (b *BuildConfig) ChangeInboundOption(inname, changer, value string) error {
	in, ok := b.inbounds[inname]
	if !ok {
		return errors.New("outbound not found")
	}
	return in.ChangeInboundOpts(changer, value)
}

func (b *BuildConfig) ChangeTransport(outname, changer, value string) error {
	out, ok := b.outbounds[outname]
	if !ok {
		return errors.New("outbound not found")
	}
	return out.ChangeTransPort(changer, value)
}
func (b *BuildConfig) ChangeDialer(outname, changer, value string) error {
	out, ok := b.outbounds[outname]
	if !ok {
		return errors.New("outbound not found")
	}
	return out.ChangeDialer(changer, value)
}
func (b *BuildConfig) ChangeMultiplex(outname, changer, value string) error {
	out, ok := b.outbounds[outname]
	if !ok {
		return errors.New("outbound not found")
	}
	return out.ChangeMultiplex(changer, value)
}





func (b *BuildConfig) ChangeServer(outname string, newserver string) error {
	out, ok := b.outbounds[outname]
	if !ok {
		return errors.New("outbound not found")
	}
	return out.SetServer(newserver)
}
func (b *BuildConfig) ChangeUUID(outname string, newuuid string) error {
	out, ok := b.outbounds[outname]
	if !ok {
		return errors.New("outbound not found")
	}
	if _, err := uuid.FromString(newuuid); err != nil {
		return errors.New("uuid is not valid")
	}
	return out.Set("", newuuid)
}
// get the outbound from user
func (b *BuildConfig) AddExtranalOutbound() error {
	return nil
}

func (b *BuildConfig) GetOutNames() []string {
	names := []string{}

	C.ExcuteMap(b.outbounds, func(ot *option.Outbound, key string) {
		if ot.Type == "block" || ot.Type == outSelector || ot.Tag == "direct" || ot.Type == "dns" {
			return
		}
		names = append(names, ot.Tag)
	})
	return names
}

func (b *BuildConfig) GetAllOutNames() []string {
	return C.MapToSliceKey(b.outbounds)
}
func (b *BuildConfig) GetAllInNames() []string {
	return C.MapToSliceKey(b.inbounds)
}

func (b *BuildConfig) OutType(tag string) string {
	out, ok := b.outbounds[tag]
	if ok {
		return out.Type
	}
	return ""
}

func (b *BuildConfig) AddRawOut(out *option.Outbound) error {
	_, ok := b.outbounds[out.Tag]
	if ok {
		return errors.New("already exit outboun with this tag")
	}
	b.outbounds[out.Tag] = out
	b.AddToSelector(out.Tag)
	return nil
}

func (b *BuildConfig) AddoutJson(bound []byte) error {
	var outbound option.Outbound
	if err := json.Unmarshal(bound, &outbound); err != nil {
		return err
	}
	_, ok := b.outbounds[outbound.Tag]
	if ok {
		return errors.New("already exit outboun with this tag")
	}
	b.outbounds[outbound.Tag] = &outbound
	b.AddToSelector(outbound.Tag)
	return nil
}

func (b *BuildConfig) AddToSelector(tag string) {
	b.GetSelector().SelectorOptions.Outbounds = append(b.GetSelector().SelectorOptions.Outbounds, tag)
}

func (b *BuildConfig) RemoveSelector(tag string) {
	b.GetSelector().SelectorOptions.Outbounds = C.RemoveItem(b.GetSelector().SelectorOptions.Outbounds, func(t string) bool { return t == tag })
}

func (b *BuildConfig) CheckOutbound(tag string) bool {
	_, ok := b.outbounds[tag]
	return ok
}

func (b *BuildConfig) ChangeTlsSettings(tag string, Tls *option.OutboundTLSOptions) error {
	out, ok := b.outbounds[tag]
	if ok {
		out.SetTLS(Tls)
	}
	return errors.New("outbound not found")
}

func (b *BuildConfig) ChageTransport(tag string, transport *option.V2RayTransportOptions) error {
	out, ok := b.outbounds[tag]
	if ok {
		return out.SetTransPort(transport)
	}
	return errors.New("outbound not found")
}

func (b *BuildConfig) RemoveTls(tag string) error {
	out, ok := b.outbounds[tag]
	if ok {
		return out.SetTLS(nil)
	}
	return errors.New("outbound not found")
}

// Inbounds handling
func (b *BuildConfig) AddInbound(preName string) error {

	//TODO: should add default inbound tun
	// can add proxy inbounds stuff
	return nil
}

func (b *BuildConfig) RemoveInbound(tagName string) error {

	if tagName == b.DefaultInbound {
		return errors.New("you cannot delete default inmbound")
	}

	delete(b.inbounds, tagName)

	C.ExcuteMap(b.dnsRules, func(val *option.DNSRule, key string) {
		for j, ot := range val.DefaultOptions.Inbound {
			if ot == tagName {
				val.DefaultOptions.Inbound = append(val.DefaultOptions.Inbound[j:], val.DefaultOptions.Inbound[:j+1]...)
			}
		}
	})

	C.ExcuteMap(b.routeRules, func(val *option.Rule, key string) {
		if len(val.DefaultOptions.Inbound) > 0 {
			for i, in := range val.DefaultOptions.Inbound {
				if in == tagName {
					val.DefaultOptions.Inbound = append(val.DefaultOptions.Inbound[i:], val.DefaultOptions.Inbound[:i+1]...)
				}
			}
		}
	})

	return nil
}

func (b *BuildConfig) BuildDefaultInbounds() error {
	var (
		tunavbl bool
	)

	_, tunavbl = b.inbounds["tun-in"]

	if !tunavbl {
		tunin := option.Inbound{
			Type: "tun",
			Tag:  "tun-in",
			TunOptions: option.TunInboundOptions{
				InboundOptions: option.InboundOptions{
					SniffEnabled:             true,
					SniffOverrideDestination: false,
					SniffTimeout:             option.Duration(100 * time.Millisecond),
				},
				AutoRoute:   true,
				StrictRoute: true,

				Address: option.Listable[netip.Prefix]{netip.MustParsePrefix("172.19.0.1/30")}, //TODO: add later

			},
		}
		b.inbounds["tun-in"] = &tunin

	}
	b.DefaultInbound = "tun-in"

	return nil
}

//interacting

func (b *BuildConfig) GetRouteRules() map[string]*option.Rule {
	return b.routeRules
}
func (b *BuildConfig) GetRouteRule(tag string) option.Rule {
	rule, ok := b.routeRules[tag]
	if !ok {
		return option.Rule{}
	}
	return *rule
}
func (b *BuildConfig) GetRouteRuleSetTags() []string {
	return C.MapToSliceKey(b.ruleSet)
}
func (b *BuildConfig) GetRouteRuleSets() map[string]*option.RuleSet {
	return b.ruleSet
}

func (b *BuildConfig) GetRouteRuleSet(tag string) (*option.RuleSet, error) {

	ruleset, ok := b.ruleSet[tag]
	if !ok {
		return nil, errors.New("rule set not found")
	}
	return ruleset, nil
}


func (b *BuildConfig) GetDnsRules() map[string]*option.DNSRule {
	return b.dnsRules
}

func (b *BuildConfig) GetDnsServers() map[string]*option.DNSServerOptions {
	return b.dnsServers
}
func (b *BuildConfig) GetDnsServer(tag string) (*option.DNSServerOptions, error) {
	srv, ok := b.dnsServers[tag]
	var err error
	if !ok {
		err = errors.New("dns server not found")
	}
	return srv, err
}

func (b *BuildConfig) Close() error {
	//	b.config.Route.Rules = C.MapToSlice(b.routeRules)
	//b.config.DNS.Rules = C.MapToSlice(b.dnsRules)
	//TODO:

	b.config.Outbounds = C.MapToSlicePtr(b.outbounds)
	b.config.Inbounds = C.MapToSlicePtr(b.inbounds)
	b.config.DNS.Rules = C.MapToSlicePtr(b.dnsRules)
	b.config.Route.Rules = C.MapToSlicePtr(b.routeRules)
	b.config.DNS.Servers = C.MapToSlicePtr(b.dnsServers)

	//TODO: change path
	file, err := os.OpenFile("./configs/"+b.path, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return err
	}
	defer file.Close()

	content, err := json.Marshal(b.config)
	if err != nil {
		return err
	}
	file.Truncate(0)
	_, err = file.Write(content)


	return err

}

func (b *BuildConfig) ExportOutbound(tag string) any {
	out := b.outbounds[tag]
	if out == nil {
		return "no outbound for the tag"
	}
	return b.commonExport(out)
}
func (b *BuildConfig) ExportInbound(tag string) any {
	in := b.inbounds[tag]
	if in == nil {
		return nil
	}
	return b.commonExport(in)
}

func (b *BuildConfig) ExportAllInbounds() any {
	b.config.Inbounds = C.MapToSlicePtr(b.inbounds)
	return b.commonExport(b.config.Inbounds)
}

func (b *BuildConfig) ExportAllOutbounds() any {
	b.config.Outbounds = C.MapToSlicePtr(b.outbounds)

	return b.commonExport(b.config.Outbounds)
}

func (b *BuildConfig) ExportDns() any {
	b.config.DNS.Rules = C.MapToSlicePtr(b.dnsRules)
	b.config.DNS.Servers = C.MapToSlicePtr(b.dnsServers)

	C.ExcuteSlice(b.config.DNS.Rules, func(r *option.DNSRule) {
		r.Tag = ""
	})

	return b.commonExport(b.config.DNS)
}

func (b *BuildConfig) ExportDnsServer(tag string) any {
	srv, ok := b.dnsServers[tag] 
	if !ok {
		return "cannot find dns server"
	}
	return b.commonExport(srv)
}

func (b *BuildConfig) ExportRuleSet(tag string) any {
	set, ok := b.ruleSet[tag] 
	if !ok {
		return "cannot find dns server"
	}
	return b.commonExport(set)
}


func (b *BuildConfig) ExportExpermental() any {
	if b.config.Experimental == nil {
		return "experimental didn't add yet"
	}
	return b.commonExport(b.config.Experimental)
}

func (b *BuildConfig) ExportRoute() any {
	b.config.Route.Rules = C.MapToSlicePtr(b.routeRules)
	b.config.Route.RuleSet = C.MapToSlicePtr(b.ruleSet)
	C.ExcuteSlice(b.config.Route.Rules, func(r *option.Rule) {
		r.Tag = ""
	})

	return b.commonExport(b.config.Route)
}
func (b *BuildConfig) ExportRouteRule(tag string) any {
	rule, ok := b.routeRules[tag] 
	if !ok {
		return "cannot find route rule"
	}
	return b.commonExport(rule)
}
func (b *BuildConfig) ExportDnsRule(tag string) any {
	rule, ok := b.dnsRules[tag] 
	if !ok {
		return "cannot find dns rule"
	}
	return b.commonExport(rule)
}

func (b *BuildConfig) Export() any {

	b.config.Outbounds = C.MapToSlicePtr(b.outbounds)
	b.config.Inbounds = C.MapToSlicePtr(b.inbounds)
	b.config.DNS.Rules = C.MapToSlicePtr(b.dnsRules)
	b.config.Route.Rules = C.MapToSlicePtr(b.routeRules)
	b.config.DNS.Servers = C.MapToSlicePtr(b.dnsServers)
	b.config.Route.RuleSet = C.MapToSlicePtr(b.ruleSet)

	C.ExcuteSlice(b.config.Route.Rules, func(r *option.Rule) {
		r.Tag = ""
	})
	C.ExcuteSlice(b.config.DNS.Rules, func(r *option.DNSRule) {
		r.Tag = ""
	})

	return b.commonExport(b.config)
}

func (b *BuildConfig) commonExport(v any) any {
	content, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		b.logger.Error(err.Error())
		return []byte{}
	}
	if b.exportafter != nil {
		return b.exportafter(content)
	}
	return content
}

func (b *BuildConfig) SetExportCallback(callback func(b []byte) any) {
	b.exportafter = callback
}

// copied struct from sing box
type OptionsEntry struct {
	content []byte
	path    string
	options option.Options
}

// copied function from sing box (bit changed)
func readConfigAt(file io.Reader) (*OptionsEntry, error) {
	var (
		configContent []byte
		err           error
	)

	configContent, err = io.ReadAll(file)

	if err != nil {
		return nil, errors.New("config read error")
	}
	options, err := singJson.UnmarshalExtended[option.Options](configContent)
	if err != nil {
		return nil, err
	}
	return &OptionsEntry{
		content: configContent,
		options: options,
	}, nil
}
