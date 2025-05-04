package service

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/sbox/conf"

	// option "github.com/sadeepa24/connected_bot/builder/sbox_option/v2"
	//"github.com/sagernet/sing-box/option"
	"github.com/sadeepa24/connected_bot/builder/v2"
	"github.com/sadeepa24/connected_bot/common"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"github.com/sagernet/sing-box/option"
)

type confBuilder struct {
	common.Tgcalls
	
	ctx            context.Context
	cancel context.CancelFunc

	State          int
	Messagesession *botapi.Msgsession
	btns           *botapi.Buttons
	wiz            *Xraywiz
	Builder        *builder.Builder

	dbconfs 	[]db.Config
	allin 		map[int16]conf.Inboud
	sboxconfs 	[]db.SboxConfigs
	userId 		int64 //config owner

	store *builder.Store
	counter *atomic.Int32

	disableTimeout bool
	disableAutoSave bool
	disableFieldSkip bool
	singlemode bool
}

var _ builder.Connector = (*confBuilder)(nil)

type ExitErr struct {
	error
}

func (c *confBuilder) Select(options []string, msg any) (string, error) {	
	c.btns.Reset([]int16{2})
	lp:
	for _, option := range options {
		//TODO: remove switch after seprating options from singbox maybe in next version
		
		switch option {
		case builder.Pass:
			c.btns.Passline()
			continue lp
		case "Id", "Custom_info", "Domain", "SupportInfo", "Public_Ip":
			if !c.disableFieldSkip {
				continue lp
			}
		}
		c.btns.AddBtcommon(option)
	}
	callback, err := c.Callbackreciver(msg, c.btns) //TODO: change msg: add exportedjson object
	if err != nil {
		return "", ExitErr{err}
	}
	return callback.Data, nil
}
func (c *confBuilder) ReciveVal(msg string) (string, error) {
	mg, err := c.Sendreciver(msg)
	if err != nil {
		return "", err
	}
	if mg.Text == "." {
		return "", nil
	}
	return mg.Text, nil
}
func (c *confBuilder) AlertSend(msg string) (error) {
	c.Alertsender(msg)
	return nil
}

var (

	btnins = "Inbounds"
	btnouts = "Outbounds"
	btnends = "Endpoint"
	btnrouter = "Router"
	btnexperimental = "Experimental"
	btndns = "DNS"
	btnlog = "Log"
	btnNTP = "NTP"
	btnboilerplate = "Boiler Plates"
	btnparse = "Parse Config"
	btnasfile = "As File"

	btnallin = "All Inbounds"
	btnaddin = "Add Inbound"
	btnremin = "Remove Inbound"

	btnallout = "All Outbounds"
	btnaddout = "Add Outbound"
	btnremout = "Remove Outbound"

	btnallend = "All Endpoints"
	btnaddend = "Add Endpoint"
	btnremend = "Remove Endpoint"

	btncreate = "Create"
	btnFrom = "From"
	btnLoad = "LoadConfig"

	btnwizIn = "Inbounds"
	btnwizOut = "Outbounds"
	btnwizEnd = "Endpoint"
	btnwizRrules = "Route rules"
	btnwizdrules = "Dns rules"
	btnwizFull = "Full Config"

	btnaddnew = "⚙ ADD NEW ⚙"
	btndelete = "⚙ Delete ⚙"
)


const (
	builderzero int = iota
	builderhome
	buildclosed 
	builderinbound
	builderAddin 
	builderPlates 
	builderoutbound 
	builderAddout 
	builderLoadout 
	builderendpoint
	builderAddendpoint 
	builderrouter 
	builderdns 
	builderexperimental

)

func (c *confBuilder) zero() error {
	if c.singlemode {
		c.State = builderhome
		return nil
	}
	if c.sboxconfs == nil {
		var err error
		c.sboxconfs, err = c.wiz.ctrl.GetSboxConfig(c.userId)
		if err != nil {return err}
	}

	for _, sconf := range c.sboxconfs {
		c.btns.AddBtcommon(sconf.Name)
	}
	c.btns.Passline()
	c.btns.AddBtcommon(btnaddnew)
	c.btns.AddBtcommon(btndelete)
	c.btns.AddClose(true)
	callback, err := c.Callbackreciver("select config or create new one", c.btns)
	if err != nil {
		return err
	}
	var path string
	switch callback.Data {
	case C.BtnClose:
		c.State = buildclosed
		if c.Builder != nil {
			c.Builder.Close()
		}
	case btnaddnew:
		if c.wiz.ctrl.MaxBuildConf <= len(c.sboxconfs) {
			c.Alertsender("you can't create config more than " + strconv.Itoa(c.wiz.ctrl.MaxBuildConf))
			return nil
		}
		name, err := common.ReciveName(c.Tgcalls)
		if err != nil {
			return err
		}
		new, err := c.wiz.ctrl.CreateSboxConf(c.userId, name)
		if err != nil {
			c.Messagesession.SendError(err, "new config create failed")
			return nil
		}
		c.sboxconfs = append(c.sboxconfs, new)
	case btndelete:
		c.btns.Reset([]int16{2})
		for _, sconf := range c.sboxconfs {
			c.btns.Addbutton(sconf.Name, strconv.Itoa(int(sconf.ID)), "")
		}
		c.btns.Passline()
		c.btns.Addcancle()
		callback, err = c.Callbackreciver("select config to delete", c.btns)
		if err != nil {return err}
		if callback.Data == C.BtnCancle {
			return nil
		}
		id, err := strconv.Atoi(callback.Data)
		if err != nil {return nil}
		c.wiz.ctrl.DeleteSboxConf(int64(id))
		
		for i, sconf := range c.sboxconfs {
			if sconf.ID == int64(id) {
				c.sboxconfs[i] = c.sboxconfs[len(c.sboxconfs)-1]
				c.sboxconfs = c.sboxconfs[:len(c.sboxconfs)-1]
				break
			}
		}

	default:
		C.ExcuteSliceTill(c.sboxconfs, func(t *db.SboxConfigs) bool {
			if t != nil && t.Name == callback.Data {
				path = c.wiz.ctrl.ConfigFolder+t.ConfPath
				return false
			}
			return true
		})
		if c.Builder != nil {
			err := c.Builder.ResetConf(path)
			if err != nil {return err}
		} else {
			c.Builder, err = builder.NewBuilderFromFile(c, path)
			if err != nil {return err}
			c.Builder.CallBack = func() builder.Buf {
				return &botapi.PreBuuf{
					Buffer: &bytes.Buffer{},
				}
			}
		}
		c.State = builderhome
	}
	return nil
}
func (c *confBuilder) home() error {
	
	c.btns.AddButtonSlice([]string{btnins, btnouts, btnrouter, btndns, btnexperimental, btnends, btnNTP, btnlog, btnboilerplate})
	c.btns.Passline()
	c.btns.AddBtcommon(btnparse)
	c.btns.Passline()
	c.btns.AddBtcommon(btnasfile)
	c.btns.AddCloseBack()

	callback, err := c.Callbackreciver(c.Builder.Export(), c.btns)
	if err != nil {
		return err
	}
	switch callback.Data {
	case btnins:
		c.State = builderinbound
	case btnouts:
		c.State = builderoutbound
	case btnrouter:
		c.State = builderrouter
	case btndns:
		c.State = builderdns
	case btnexperimental:
		c.State = builderexperimental
	case btnends:
		c.State = builderendpoint
	case btnboilerplate:
		c.State = builderPlates
	case btnasfile:
		f := c.Builder.ExportBuffer(&bytes.Buffer{})
		c.wiz.ctrl.SendAsFile(f, "config.json", "", c.Messagesession.UserID())
	case btnlog:
		return c.Builder.LogFieldChange()
	case btnNTP:
		return c.Builder.NTPFieldChange()
	case btnparse:
		new, err := c.Sendreciver("send new config")
		if err != nil {
			return err
		}
		err = c.Builder.ResetAny(new.Text)
		if err != nil {
			c.Messagesession.SendAlert("Parsing Config Failed" + err.Error(), nil)
		}
		return nil
	case C.BtnClose:
		c.State = buildclosed
		if c.Builder != nil {
			return c.Builder.Close()
		}
	case C.BtnBack:
		c.State = builderzero
		err := c.Builder.Save()
		if err != nil {
			c.Messagesession.SendError(err, "current config save failed")
		}
	}
	return nil
}
func (c *confBuilder) inbound() error {

	c.btns.AddButtonSlice([]string{btnallin, btnaddin, btnremin})
	c.btns.AddCloseBack()
	callback, err := c.Callbackreciver(c.Builder.ExportAllIn(), c.btns)
	if err != nil {return err}
	switch callback.Data {
	case btnallin, btnremin:
		allin := []string{}
		for _, in := range c.Builder.AllIn() {
			allin = append(allin, in.Tag)
		}
		if len(allin) == 0 {
			c.Alertsender("no inbound available for editing please add inbound	")
			return nil
		}
		sin, err := c.Select(allin, c.Builder.ExportAllIn())
		if err != nil {
			return err
		}

		switch callback.Data {
		case btnallin:
			err = c.Builder.InboundFieldEditor(sin)
			if err != nil && c.ctx.Err() != nil {
				return err
			}
		case btnremin:
			err = c.Builder.RemoveInbound(sin)
			if err != nil {
				c.Alertsender("inbound remove failed " + err.Error())
			}
		}
		return nil
	case btnaddin:
		c.State = builderAddin
	case C.BtnBack:
		c.State = builderhome
	case C.BtnClose:	
		c.State = buildclosed
	}
	return nil
}
func (c *confBuilder) addin() error {
	c.btns.Reset([]int16{2})
	c.btns.AddButtonSlice([]string{btnFrom, btncreate})
	c.btns.AddCloseBack()


	callback, err := c.Callbackreciver("select add inbound method", c.btns)
	if err != nil {
		return err
	}
	switch callback.Data {
	case btnFrom:
		var userval *tgbotapi.Message
		userval, err = c.Sendreciver("send you'r inbound config block must be json")
		if err != nil {
			return err
		}
		if userval.Text == "." {
			return nil
		}
		err = c.Builder.AddInbound(userval.Text)
	case btncreate:
		err = c.Builder.AddInbound(builder.Create(1))
	case C.BtnBack:
		c.State = builderinbound
	case C.BtnClose:
		c.State = buildclosed
	}
	if err != nil {
		c.AlertSend("inbound adding failed "+ err.Error())
	}
	return nil
}


func (c *confBuilder) outbound() error {
	c.btns.Reset([]int16{2})
	c.btns.AddButtonSlice([]string{btnallout, btnaddout, btnremout})
	c.btns.AddCloseBack()
	callback, err := c.Callbackreciver(c.Builder.ExportAllOut(), c.btns)
	if err != nil {return err}
	switch callback.Data {
	case btnallout, btnremout:
		allout := []string{}
		for _, out := range c.Builder.AllOut() {
			allout = append(allout, out.Tag)
		}
		if len(allout) == 0 {
			c.Alertsender("no outbound available for editing please add endpoint")
			return nil
		}
		sout, err := c.Select(allout, "select outbound to edit")
		if err != nil {
		}
		switch callback.Data {
		case btnallout:
			err = c.Builder.OutboundFieldsChange(sout)
			if err != nil {
			}
		case btnremout:
			err = c.Builder.RemoveOutbound(sout)
			if err != nil {
			}
		}
	case btnaddout:
		c.State = builderAddout
	case C.BtnBack:
		c.State = builderhome
	case C.BtnClose:
		c.State = buildclosed
	}
	return nil
}
func (c *confBuilder) addout() error {
	c.btns.Reset([]int16{2})
	c.btns.AddButtonSlice([]string{btnFrom, btncreate, btnLoad})
	c.btns.AddCloseBack()


	callback, err := c.Callbackreciver(c.Builder.ExportAllOut(), c.btns)
	if err != nil {
		return err
	}
	switch callback.Data {
	case btnFrom:
		var userval *tgbotapi.Message
		userval, err = c.Sendreciver("send you'r outbound config block json or urlencode")
		if err != nil {
			return err
		}
		if userval.Text == "." {
			return nil
		}
		err = c.Builder.AddOutbound(userval.Text)
	case btncreate:
		err = c.Builder.AddOutbound(builder.Create(1))
	case btnLoad:
		c.State = builderLoadout
	case C.BtnBack:
		c.State = builderoutbound
	case C.BtnClose:
		c.State = buildclosed
	}
	if err != nil {
		c.AlertSend("outbound adding failed "+ err.Error())
	}
	return nil
}
//FIXME:
func (c *confBuilder) loadconf() error {
	var err error
	if c.dbconfs == nil {
		c.dbconfs, err = c.wiz.ctrl.GetUserConfigs(c.userId)
		if err != nil {
			return nil
		}
	}
	if c.allin == nil {
		c.allin = c.wiz.ctrl.GetAllinbound()
	}
	if len(c.dbconfs) == 0 {
		c.Alertsender("you don't have any config to load create confs using /create")
		c.State = builderAddout
		return nil
	}
	se, err := c.Select(C.SliceFromSlice(c.dbconfs, func(in db.Config) (string, bool)  { 
		return in.Name, true
	}), "select you'r config")
	if err != nil {
		return err
	}
	
	config := C.GetFromSlice(c.dbconfs, func(in db.Config) bool { return in.Name == se })
	if config == nil {
		c.Alertsender("something wrong")
		return nil
	}
	c.btns.Reset([]int16{2})
	for _, s := range config.InboundIds {
		c.btns.Addbutton(c.allin[s].Name, strconv.Itoa(int(s)), "")
	}
	id, err := c.Callbackreciver("Select Inbound TO Load", c.btns)
	if err != nil {
		
		return err
	}
	idint, err := strconv.Atoi(id.Data)
	if err != nil {
		c.Alertsender("something wrong selected in bound id")
		return nil
	}
	inbound := c.allin[int16(idint)]
	c.Messagesession.DeleteLast()
	expinf, err := common.ReciveExpInfo(c.Tgcalls)
	if err != nil {
		c.Alertsender("info recive failed")
		return nil
	}
	err = c.Builder.LoadOutFromDbconf(*config, inbound, expinf)
	if err  != nil {
		c.Alertsender("config loading failed" + err.Error())
	}
	c.State = builderAddout
	return nil
}


func (c *confBuilder) endpoint() error {
	c.btns.Reset([]int16{2})
	c.btns.AddButtonSlice([]string{btnallend, btnaddend, btnremend})
	c.btns.AddCloseBack()
	callback, err := c.Callbackreciver(c.Builder.ExportAllEnd(), c.btns)
	if err != nil {return err}
	switch callback.Data {
	case btnallend, btnremend:
		allout := []string{}
		for _, out := range c.Builder.AllEnd() {
			allout = append(allout, out.Tag)
		}
		if len(allout) == 0 {
			c.Alertsender("no endpoints available for editing please add endpoint")
			return nil
		}
		sout, err := c.Select(allout, "select endpoint to edit")
		if err != nil {
			return err
		}
		switch callback.Data {
		case btnallout:
			err = c.Builder.EndpointFieldEdit(sout)
			if err != nil && c.ctx.Err() != nil {
				return err
			}
		case btnremout:
			err = c.Builder.RemoveEndpoint(sout)
			if err != nil && c.ctx.Err() != nil {
				return err
			}
		}
	case btnaddend:
		c.State = builderAddendpoint
	case C.BtnBack:
		c.State = builderhome
	}
	return nil
}
func (c *confBuilder) addend() error {
	c.btns.Reset([]int16{2})
	c.btns.AddButtonSlice([]string{btnFrom, btncreate})
	c.btns.AddCloseBack()


	callback, err := c.Callbackreciver(c.Builder.ExportAllEnd(), c.btns)
	if err != nil {
		return err
	}
	switch callback.Data {
	case btnFrom:
		var userval *tgbotapi.Message
		userval, err = c.Sendreciver("send you'r endpoint config block json")
		if err != nil {
			return err
		}
		if userval.Text == "." {
			return nil
		}
		err = c.Builder.AddEndpoint(userval.Text)
	case btncreate:
		err = c.Builder.AddEndpoint(builder.Create(1))
	case C.BtnBack:
		c.State = builderendpoint
	case C.BtnClose:
		c.State = buildclosed
	}
	if err != nil {
		c.AlertSend("outbound adding failed "+ err.Error())
	}
	return nil
}
func (c *confBuilder) router() error {
	err :=  c.Builder.RouteFieldChange()
	c.State = builderhome
	return err
}
func (c *confBuilder) dns() error {
	err :=  c.Builder.DNSfieldChange()
	c.State = builderhome
	return err
}
func (c *confBuilder) experimental() error {
	err :=  c.Builder.ExperimentalField()
	c.State = builderhome
	return err
}



func (c *confBuilder) afterexport(opts any, exec func (i any) error ) error {
	c.btns.Reset([]int16{2})
	c.btns.AddBtcommon(C.BtnConform)
	c.btns.AddBtcommon(C.BtnCancle)
	c.Alertsender("if you press conform you'r current config will replace with this one")
	
	callback, err := c.Callbackreciver(c.Builder.ExportExtranal(opts), c.btns)
	if err != nil {
		return err
	}
	if callback.Data == C.BtnCancle {
		return nil
	}
	return exec(opts)
}
func (c *confBuilder) wizardExport() error {
	c.btns.AddButtonSlice([]string{btnwizFull, btnwizIn, btnwizOut, btnwizEnd, btnwizRrules, btnwizdrules})
	c.btns.AddCloseBack()

	callback, err := c.Callbackreciver("select option", c.btns)
	if err != nil {return err}
	switch callback.Data {
	case C.BtnBack:
		c.State = builderhome
	case C.BtnClose:
		c.State = buildclosed
	case btnwizFull:
		if len(c.store.WizConf) == 0 {
			c.Alertsender("no configs available")
			return nil
		}
		c.btns.Reset([]int16{2})
		c.btns.AddButtonSlice(c.store.WizConf)
		callback, err = c.Callbackreciver("select config", c.btns)
		if err != nil {
			return err
		}
		optss, err := builder.ExportConf[option.Options](c.store, callback.Data, c)
		if err != nil {
			c.Alertsender("wizard confgi export failed")
			return nil
		}
		c.afterexport(&optss, c.Builder.ResetAny)
	case btnwizIn:
		if len(c.store.WizIn) == 0 {
			c.Alertsender("no Inbounds available")
			return nil
		}
		c.btns.Reset([]int16{2})
		c.btns.AddButtonSlice(c.store.WizIn)
		callback, err = c.Callbackreciver("select Inbound", c.btns)
		if err != nil {
			return err
		}
		optss, err := builder.ExportConf[option.Inbound](c.store, callback.Data, c)
		if err != nil {
			c.Alertsender("wizard confgi export failed")
			return nil
		}
		c.afterexport(&optss, c.Builder.AddInbound)
	case btnwizOut:
		if len(c.store.WizOut) == 0 {
			c.Alertsender("no Outbounds available")
			return nil
		}
		c.btns.Reset([]int16{2})
		c.btns.AddButtonSlice(c.store.WizOut)
		callback, err = c.Callbackreciver("select Outbound", c.btns)
		if err != nil {
			return err
		}
		optss, err := builder.ExportConf[option.Outbound](c.store, callback.Data, c)
		if err != nil {
			c.Alertsender("wizard confgi export failed")
			return nil
		}
		c.afterexport(&optss, c.Builder.AddOutbound)
	case btnwizEnd:
		if len(c.store.WizEnd) == 0 {
			c.Alertsender("no Endpoints available")
			return nil
		}
		c.btns.Reset([]int16{2})
		c.btns.AddButtonSlice(c.store.WizEnd)
		callback, err = c.Callbackreciver("select Endpoint", c.btns)
		if err != nil {
			return err
		}
		optss, err := builder.ExportConf[option.Endpoint](c.store, callback.Data, c)
		if err != nil {
			c.Alertsender("wizard confgi export failed")
			return nil
		}
		c.afterexport(&optss, c.Builder.AddEndpoint)
	
	case btnwizRrules:
		if len(c.store.WizRrules) == 0 {
			c.Alertsender("no route rules available")
			return nil
		}
		c.btns.Reset([]int16{2})
		c.btns.AddButtonSlice(c.store.WizRrules)
		callback, err = c.Callbackreciver("select Route RUle", c.btns)
		if err != nil {
			return err
		}
		optss, err := builder.ExportConf[option.Rule](c.store, callback.Data, c)
		if err != nil {
			c.Alertsender("wizard confgi export failed")
			return nil
		}
		c.afterexport(&optss, c.Builder.AddRrule)
	case btnwizdrules:
		if len(c.store.WizDrules) == 0 {
			c.Alertsender("no dns rules available")
			return nil
		}
		c.btns.Reset([]int16{2})
		c.btns.AddButtonSlice(c.store.WizDrules)
		callback, err = c.Callbackreciver("select DNS Rule", c.btns)
		if err != nil {
			return err
		}
		optss, err := builder.ExportConf[option.DNSRule](c.store, callback.Data, c)
		if err != nil {
			c.Alertsender("wizard confgi export failed")
			return nil
		}
		c.afterexport(&optss, c.Builder.AddDrule)
	}

	return nil
}
func (c *confBuilder) save() error {
	return c.Builder.Save()
}
func (c *confBuilder) run() {
	go c.timeout()
	defer func ()  {
		c.wiz.builds.Delete(c.userId)
		c.Messagesession.DeleteAllMsg()
		c.Builder.Close()
		if c.cancel != nil {
			c.cancel()	
		}
	}()
	//TODO: remove this in later version after stabling the builder
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic recover")
			fmt.Println(r)
		}
	}()
	var err error
	bldloop:
	for {
		c.btns.Reset([]int16{2})
		if c.ctx.Err() != nil {
			return
		}
		switch c.State {
		case builderzero:
			err = c.zero()
		case buildclosed:
			break bldloop
		case builderhome:
			err = c.home()
		case builderinbound:
			err = c.inbound()
		case builderAddin:
			err = c.addin()
		case builderoutbound:
			err = c.outbound()
		case builderAddout:
			err = c.addout()
		case builderrouter:
			err = c.router()
		case builderdns:
			err = c.dns()
		case builderexperimental:
			err = c.experimental()
		case builderAddendpoint:
			err= c.addend()
		case builderendpoint:
			err = c.endpoint()
		case builderLoadout:
			err = c.loadconf()
		case builderPlates:
			err = c.wizardExport()
		default:
			err = c.zero()
		}
		if err != nil {
			break bldloop
		}
	}
	if c.disableAutoSave {
		c.btns.Reset([]int16{2})
		c.btns.AddBtcommon(C.BtnSave)
		c.btns.AddBtcommon(C.BtnCancle)
		callback, err := c.Callbackreciver("are you sure save config", c.btns)
		if err != nil {
			return
		}
		if callback.Data == C.BtnSave {
			c.save()
		}
	}
}
func (c *confBuilder) timeout() {
	if c.disableTimeout {
		return
	}
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			if c.counter.Add(-1) <= 0 {
				c.Messagesession.SendAlert("builder timeout", nil)
				c.cancel()
				ticker.Stop()
				return
			}
		case <-c.ctx.Done():
			ticker.Stop()
			return
		}
	}
}



func (u *Xraywiz) commandBuildV2(upx *update.Updatectx,  Messagesession *botapi.Msgsession) error {
	//return u.removeafterbuilderdevloped(upx, Messagesession)
	_, loaded := u.builds.LoadOrStore(upx.User.TgID, true)
	if loaded {
		Messagesession.SendAlert("Already In build Session Close It first", nil)
		return nil
	}
	ct := new(atomic.Int32)
	ct.Add(2)
	ctx, cancelfunc := context.WithCancel(u.ctrl.GetBaseContext())
	Messagesession.SetNewcontext(ctx)
	
	Tgcalls :=  common.Tgcalls{
		Sendreciver: func(msg any) (*tgbotapi.Message, error) {
			if msg != nil {
				_, err := Messagesession.Edit(msg, nil, "")
				if err != nil {
					return nil, err
				}
			}
			mg, err := u.defaultsrv.ExcpectMsgContext(ctx, upx.User.TgID, upx.User.TgID)
			if err == nil {
				Messagesession.Addreply(mg.MessageID)
			}
			ct.Add(1)
			return mg, err
		},
		Callbackreciver: func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error) {
			_, err := Messagesession.Edit(msg, btns, "")
			if err != nil {
				return nil, err
			}
			ct.Add(1)
			return u.callback.GetcallbackContext(ctx, btns.ID())
		},
		Alertsender: func(msg string) { Messagesession.SendAlert(msg, nil) },
	}
	build := &confBuilder{
		cancel: cancelfunc,
		store: u.bulderstore,
		counter: ct,
		ctx:            ctx,
		Messagesession: Messagesession,
		userId: upx.User.TgID,
		btns:           botapi.NewButtons([]int16{2}),
		wiz:            u,
		Tgcalls: Tgcalls,
	}
	var err error
	build.Builder, err = builder.NewBuilder(build)
	if err != nil {
		return err
	}
	build.Builder.CallBack = func() builder.Buf {
		return &botapi.PreBuuf{
			Buffer: &bytes.Buffer{},
		}
	}
	go build.run()
	return nil
}