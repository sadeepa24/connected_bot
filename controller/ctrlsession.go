package controller

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sadeepa24/connected_bot/update"
	"github.com/sadeepa24/connected_bot/update/bottype"
	"go.uber.org/zap"
)

//TODO: db saving configurecurrently always save after closing usersession

// Ctrlsession is not theadsafe use in single thread
type CtrlSession struct {
	ctx    context.Context
	cancle context.CancelFunc

	ctrl   *Controller
	user   *db.User
	//config []*db.Config

	configmap map[int64]*db.Config

	oprations *atomic.Int32
	olduser   db.User

	closed bool
}

type ForceCloser interface {
	ForceClose() error
}

func NewctrlSession(ctrl *Controller, upx *update.Updatectx, ForceCloseOldSession bool) (*CtrlSession, error) {
	if ctrl == nil || upx == nil {
		return nil, errors.New("ctrl or user objects is nil")
	}
	user := upx.Dbuser()

	if forcecloser, loaded := ctrl.Checksession(upx.User.TgID); loaded {
		if ForceCloseOldSession {
			var closer ForceCloser
			var ok bool
			if closer, ok = forcecloser.(ForceCloser); ok {
				if err := closer.ForceClose(); err != nil {
					return nil, C.ErrSessionExcit
				}
				ctrl.logger.Warn("Force closed old session")
			}
		} else {
			return nil, C.ErrSessionExcit
		}
	}

	session := &CtrlSession{
		ctx:       upx.Ctx,
		cancle:    upx.Cancle,
		ctrl:      ctrl,
		user:      user,
		//config:    []*db.Config{},
		oprations: new(atomic.Int32),
		closed:    false,

		//tx: ctrl.db.Begin(),
	}

	st := time.Now()
	if user.ConfigCount != 0 {
		err := ctrl.db.Model(&db.Config{}).Preload("Inbound").Preload("Outbound").Where("user_id = ?", user.TgID).Find(&session.user.Configs).Error
		if err != nil {
			return nil, C.ErrOnDb
		}
	}
	ctrl.logger.Debug("Elpsed time for fetching configs 😀", zap.Duration("duration", time.Since(st)))

	session.configmap = make(map[int64]*db.Config, ctrl.Maxconfigcount+1)
	for i, conf := range session.user.Configs {
		session.configmap[conf.Id] = &session.user.Configs[i]
		//session.config = append(session.config, &session.user.Configs[i])
	}

	ctrl.Addsession(session, user.TgID)
	session.olduser = *user
	return session, nil
}

func (c *CtrlSession) AddNewConfig(inboundid int16, outboundid int16, Quota C.Bwidth, login int16, name string) (*sbox.Userconfig, error) {

	if c.ctx.Err() != nil {
		return nil, C.ErrContextDead
	}
	c.add()
	defer c.done()

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, C.Erruuidcreatefailed
	}

	intag, err := c.ctrl.GetdbInbound(int(inboundid))
	if err != nil {
		return nil, C.ErrInboundNotFound
	}
	outtag, err := c.ctrl.GetdbOutbound(int(outboundid))
	if err != nil {
		return nil, C.ErrOutboundNotFound
	}

	newconfig := &sbox.Userconfig{
		Type:        intag.Type,
		Name:        c.user.Name,
		UsercheckId: int(c.user.CheckID),

		InboundId:   inboundid,
		Inboundtag:  intag.Tag,
		OutboundID:  outboundid,
		Outboundtag: outtag.Tag,

		Usage:      0,
		Quota:      Quota,
		LoginLimit: int32(login),
		TgId: c.user.TgID,
	}


	var dbconf db.Config

	switch intag.Type {

	case "vless":
		newconfig.Vlessgroup = &sbox.Vlessgroup{
			UUID: uid,
		}
		dbconf = db.Config{
			InboundID:  newconfig.InboundId,
			OutboundID: newconfig.OutboundID,
			UUID:       newconfig.UUID,
			Name:       name,
			UserID:     c.user.TgID,
			Active:     true,
			Type:       "vless",
			Download:   0,
			Upload:     0,
			Usage:      newconfig.Usage,
			Quota:      newconfig.Quota,
			LoginLimit: int16(newconfig.LoginLimit),
		}
	default:
		return nil, C.ErrTypeMissmatch

	}

	if c.ctrl.db.Create(&dbconf).Error != nil {
		return nil, C.ErrDatabaseCreate
	}

	c.user.ConfigCount = c.user.ConfigCount + 1
	c.user.UsedQuota += Quota
	newconfig.DbID = dbconf.Id

	c.configmap[dbconf.Id] = &dbconf
	//c.configmap[c.config[len(c.config)-1].Id] = c.config[len(c.config)-1]
	c.user.Configs = append(c.user.Configs, dbconf)
	//c.user.Configs = append(c.user.Configs, *c.config[len(c.config)-1])
	//c.tx.Save(c.config[len(c.config)-1])
	_, err = c.ctrl.AdduserSbox(newconfig)
	return newconfig, err
}

func (c *CtrlSession) ActivateConfig(confid int64) (sbox.Sboxstatus, error) {
	conf, dbconf, err := c.getsboxconf(confid)
	c.add()
	defer c.done()
	var stsatus sbox.Sboxstatus
	if err != nil {
		return stsatus, err
	}

	if (conf.Quota - conf.Usage) <= 0 {
		return stsatus, C.ErrQuotaExceed
	}
	dbconf.Active = true

	return c.ctrl.sbox.AddUser(conf)

}

func (c *CtrlSession) ConfigCloseConn(confid int64) error {
	conf, _, err := c.getsboxconf(confid)
	if err != nil {
		return err
	}
	c.add()
	defer c.done()
	return c.ctrl.sbox.CloseConns(conf)
}

func (c *CtrlSession) ReActivateConfig(confid int64) (sbox.Sboxstatus, error) {
	conf, dbconf, err := c.getsboxconf(confid)
	c.add()
	defer c.done()

	var stsatus sbox.Sboxstatus

	if err != nil {
		return stsatus, err
	}

	if (conf.Quota - conf.Usage) <= 0 {
		return stsatus, C.ErrQuotaExceed
	}
	dbconf.Active = true

	status, err := c.ctrl.sbox.AddUserReset(conf)

	if err != nil {
		return stsatus, err
	}
	c.ctrl.db.Model(&db.UsageHistory{}).Create(&db.UsageHistory{
		Upload:   status.Upload,
		Download: status.Download,
		UserID:   c.user.TgID,
		Usage:    (status.Download + status.Upload),
		Date:     time.Now(),
		ConfigID: confid,
	})

	c.user.MonthUsage = (status.Download + c.user.MonthUsage + status.Upload)

	return status, err

}

func (c *CtrlSession) ChangeLoginLimit(confid int64, newlimit int32) (sbox.Sboxstatus, error) {
	conf, dbconf, err := c.getsboxconf(confid)
	if err != nil {
		return sbox.Sboxstatus{}, err
	}
	c.add()
	defer c.done()

	conf.LoginLimit = newlimit
	dbconf.LoginLimit = int16(newlimit)
	return c.ReActivateConfig(confid)
}


func (c *CtrlSession) ActivateAll() error {

	if c.user.IsRemoved || !(c.user.IsInChannel && c.user.IsInGroup) || c.user.IsMonthLimited || c.user.Restricted || c.user.IsDistributedUser {
		return errors.New("cannot activate configs user is not verified")
	}
	c.add()
	defer c.done()
	var err error
	for _, conf := range c.user.Configs {
		if _, errr := c.ActivateConfig(conf.Id); errr != nil {
			err = errors.Join(err, errr)
		}

	}
	return err
}

func (c *CtrlSession) DeactivateConfig(configID int64) (sbox.Sboxstatus, error) {

	if c.ctx.Err() != nil {
		return sbox.Sboxstatus{}, C.ErrContextDead
	}
	c.add()
	defer c.done()
	conf, dbconf, err := c.getsboxconf(configID)

	if err != nil {
		return sbox.Sboxstatus{}, err
	}
	dbconf.Active = false
	status, err := c.ctrl.sbox.RemoveUser(conf)
	if err != nil {
		return sbox.Sboxstatus{}, err
	}

	err = c.ctrl.db.Model(&db.UsageHistory{}).Create(&db.UsageHistory{
		Upload:   status.Upload,
		Download: status.Download,
		UserID:   c.user.TgID,
		Usage:    (status.Download + status.Upload),
		Date:     time.Now(),
		ConfigID: configID,
	}).Error

	dbconf.Usage += status.Download + status.Upload
	dbconf.Download += status.Download
	dbconf.Upload += status.Upload

	c.user.MonthUsage = (status.Download + c.user.MonthUsage + status.Upload)
	return status, err
}

func (c *CtrlSession) DeactivateAll() error {
	var err error
	c.add()
	defer c.done()
	for _, conf := range c.user.Configs {
		if _, errr := c.DeactivateConfig(conf.Id); errr != nil {
			err = errors.Join(err, errr)
		}

	}
	return err
}

func (c *CtrlSession) DeleteConfig(configID int64) error {
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	c.add()
	defer c.done()

	var err error
	if _, err = c.DeactivateConfig(configID); err != nil {
		return C.ErrOnDeactivation
	}

	conf, dbconf, err := c.getsboxconf(configID)
	if err != nil {
		return err
	}
	configQuota := conf.Quota

	delete(c.configmap, configID)
	
	if c.ctrl.db.Model(&db.Config{}).Delete(dbconf).Error != nil {
		return C.ErrDbopration
	}

	/*for i, conf := range c.config {
		if conf.Id == configID {
			if c.ctrl.db.Model(&db.Config{}).Delete(conf).Error != nil {
				return C.ErrDbopration
			}
			c.config = append(c.config[:i], c.config[i+1:]...)
			break
		}
	}*/


	for i, config := range c.user.Configs {
		if config.Id == configID {
			c.user.Configs = append(c.user.Configs[:i], c.user.Configs[i+1:]...)
			break
		}
	}

	c.user.ConfigCount--
	c.user.DeletedConfCount = c.user.DeletedConfCount + 1
	c.user.UsedQuota = c.user.UsedQuota - configQuota


	return nil
}

func (c *CtrlSession) GetConfig(confid int64) (*db.Config, error) {
	conf, ok := c.configmap[confid]
	c.add()
	defer c.done()
	if !ok {
		return nil, C.ErrConfigNotFound
	}
	return conf, nil
}

// return current vpn usage, old usage from db
// total usage for now = vpn usage +  old
// Every thing in byte format
// retur today, monthusage
func (c *CtrlSession) GetUsage() (C.Bwidth, C.Bwidth) {
	var status = sbox.Sboxstatus{
		Download: 0,
		Upload:   0,
	}
	for _, config := range c.user.Configs {
		conf, _, err := c.getsboxconf(config.Id)
		if err != nil {
			continue
		}
		cstatus, err := c.ctrl.sbox.GetstatusUser(conf)
		if err != nil {
			continue
		}
		status.Download += cstatus.Download
		status.Upload += cstatus.Upload

	}
	return (status.Download + status.Upload), c.user.MonthUsage
}

func (c *CtrlSession) GetFullUsage() bottype.FullUsage {

	bf := bottype.FullUsage{}

	for _, config := range c.user.Configs {
		conf, dbconf, err := c.getsboxconf(config.Id)
		if err != nil {
			continue
		}
		cstatus, err := c.ctrl.sbox.GetstatusUser(conf)
		if err != nil {
			continue
		}
		bf.Download = bf.Download + dbconf.Download + cstatus.Download
		bf.Upload = bf.Upload + dbconf.Upload + cstatus.Upload
		bf.Uploadtd += cstatus.Upload
		bf.Downloadtd += cstatus.Download

	}

	return bf

}

// return total usage for this month
func (c *CtrlSession) TotalUsage() C.Bwidth {
	var status = sbox.Sboxstatus{
		Download: 0,
		Upload:   0,
	}
	for _, config := range c.user.Configs {
		conf, _, err := c.getsboxconf(config.Id)
		if err != nil {
			continue
		}
		cstatus, err := c.ctrl.sbox.GetstatusUser(conf)
		if err != nil {
			continue
		}
		status.Download += cstatus.Download
		status.Upload += cstatus.Upload

	}
	return status.Download + status.Upload + c.user.MonthUsage
}

// returns today, month, usage as byte
func (c *CtrlSession) GetconfigUsage(confid int64) (C.Bwidth, C.Bwidth, error) {
	var (
		conf *sbox.Userconfig
		err  error
	)
	if conf, _, err = c.getsboxconf(confid); err != nil {
		return 0, 0, err
	}
	cstatus, err := c.ctrl.sbox.GetstatusUser(conf)

	return cstatus.Download + cstatus.Upload, conf.Usage + cstatus.Download + cstatus.Upload, err
}

func (c *CtrlSession) GetconfigUsageTotal(confid int64) C.Bwidth {
	td, m, err := c.GetconfigUsage(confid)
	if err != nil {
		return 0
	}
	return td + m
}

func (c *CtrlSession) GetConfigFullUsage(confid int64) (bottype.FullUsage, sbox.Sboxstatus) {
	var (
		conf    *sbox.Userconfig
		dbconf  *db.Config
		err     error
		btusage bottype.FullUsage
	)
	if conf, dbconf, err = c.getsboxconf(confid); err != nil {
		return bottype.FullUsage{}, sbox.Sboxstatus{}
	}
	btusage = bottype.FullUsage{
		Uploadtd:   0,
		Downloadtd: 0,
		Download:   dbconf.Download,
		Upload:     dbconf.Upload,
	}

	cstatus, err := c.ctrl.sbox.GetstatusUser(conf)

	if err == nil {
		btusage.Download += cstatus.Download
		btusage.Upload += cstatus.Upload
		btusage.Downloadtd = cstatus.Download
		btusage.Uploadtd = cstatus.Upload
	}
	return btusage, cstatus
}

func (c *CtrlSession) GetconfigQuota(confid int64) C.Bwidth {
	var (
		conf *sbox.Userconfig
		err  error
	)
	if conf, _, err = c.getsboxconf(confid); err != nil {
		return 0
	}

	return conf.Quota
}

func (c *CtrlSession) Getstatus(confid int64) (sbox.Sboxstatus, error) {
	if c.ctx.Err() != nil {
		return sbox.Sboxstatus{}, C.ErrContextDead
	}

	var (
		conf *sbox.Userconfig
		err  error
	)
	if conf, _, err = c.getsboxconf(confid); err != nil {
		return sbox.Sboxstatus{}, err
	}
	return c.ctrl.sbox.GetstatusUser(conf)

}

func (c *CtrlSession) FullUsageHistory() {
}

func (c *CtrlSession) ChangeInbound(confid, inboundid int64) error {
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	c.add()
	defer c.done()
	conf, dbconf, err := c.getsboxconf(confid)
	if err != nil {
		return err
	}
	in, ok := c.ctrl.Getinbound(int(inboundid))
	if !ok {
		return C.ErrInboundNotFound
	}

	dbconf.InboundID = int16(in.Id)
	status, err := c.ctrl.sbox.RemoveUser(conf)
	if err != nil {
		if !errors.Is(err, C.ErrResultMalformed) {
			return err
		}

	}
	conf.InboundId = int16(in.Id)
	conf.Inboundtag = in.Tag
	conf.Usage += (status.Download + status.Upload)

	c.CreateUsagehistory(status, confid)
	_, err = c.ctrl.sbox.AddUser(conf)
	return err

}

func (c *CtrlSession) CreateUsagehistory(status sbox.Sboxstatus, confid int64) error {
	return c.ctrl.db.Model(&db.UsageHistory{}).Create(&db.UsageHistory{
		Upload:   status.Upload,
		Download: status.Download,
		UserID:   c.user.TgID,
		Usage:    (status.Download + status.Upload),
		Date:     time.Now(),
		ConfigID: confid,
	}).Error
}

func (c *CtrlSession) ChangeOutbound(confid, inboundid int64) error {
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	c.add()
	defer c.done()

	conf, dbconf, err := c.getsboxconf(confid)
	if err != nil {
		return err
	}
	out, ok := c.ctrl.Getoutbound(int(inboundid))
	if !ok {
		return C.ErrOutboundNotFound
	}

	dbconf.OutboundID = int16(out.Id)
	c.ctrl.sbox.RemoveAllRuleForuser(conf.GetuniqName())
	// if err != nil {
	// 	if !errors.Is(err, C.ErrResultMalformed) {
	// 		return err
	// 	}
	// }
	conf.OutboundID = int16(out.Id)
	conf.Outboundtag = out.Tag
	//conf.Usage += (status.Download + status.Upload) / C.AsGB
	_, err = c.ctrl.sbox.AddUser(conf)
	return err
}

func (c *CtrlSession) ChangeIO(ioboundname string, confid, ioid int64) error {
	switch ioboundname {
	case C.Inbound:
		return c.ChangeInbound(confid, ioid)
	case C.Outbound:
		return c.ChangeOutbound(confid, ioid)
	}
	return errors.New("no inbound or outbound found")
}

func (c *CtrlSession) Chatupdate(chat string, val bool) {
	switch chat {
	case C.Group:
		c.user.IsInGroup = val
	case C.Channel:
		c.user.IsInChannel = val
	}
}

func (c *CtrlSession) Banuser(chat string) {
	switch chat {
	case C.Group:
		c.user.IsInGroup = false
		c.user.GroupBanned = true
	case C.Channel:
		c.user.IsInChannel = false
		c.user.ChannelBanned = true
	}
	c.user.IsRemoved = true
	c.DeactivateAll()
}
//used by admin
func (c *CtrlSession) Restrict() {
	c.user.Restricted = true
	c.DeactivateAll()
}
func (c *CtrlSession) RemoveRestrict() {
	c.user.Restricted = false
	c.ActivateAll()
}

func (c *CtrlSession) Save() error {
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	return c.save()
}

func (c *CtrlSession) SaveConfigs() error {
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	return c.saveConfigs()
}

func (c *CtrlSession) save() error {
	//return c.tx.Commit().Error
	var errs error
	if err := c.ctrl.db.Save(c.user).Error; err != nil {
		errs = errors.Join(errs, err)
	}

	if len(c.user.Configs) > 0 {
		if err := c.ctrl.db.Save(c.user.Configs).Error; err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (c *CtrlSession) saveConfigs() error {
	if c.user.ConfigCount > 0 {
		return c.ctrl.db.Save(c.user.Configs).Error
	}
	return nil
}

func (c *CtrlSession) Close() error {
	if c.closed {
		return nil
	}

	var err error
	c.ctrl.RemoveSesion(c.user.TgID)
	c.wait()
	c.closed = true
	if err = c.Save(); err != nil {
		//c.ctrl.logger.Error(err.Error())
		return err
	}
	c.configmap = nil
	//c.tx = nil
	return nil
}

func (c *CtrlSession) GetUser() *db.User {
	return c.user
}

func (c *CtrlSession) ForceClose() error {
	c.cancle()
	c.wait()
	err := c.save()
	c.ctrl.RemoveSesion(c.user.TgID)

	c.ctrl.logger.Info("force closed usersession")
	return err
}

// this returns left quota for user
// userquota - quota elpsed for configs
func (c *CtrlSession) LeftQuota() C.Bwidth {
	var dedicated C.Bwidth = 0
	for _, conf := range c.configmap {
		dedicated += conf.Quota
	}
	if c.user.IsCapped {
		return c.user.CappedQuota - dedicated
	}

	return (c.user.CalculatedQuota + c.user.AdditionalQuota) - dedicated // TODO: should change here
}

// this is special for gift command
func (c *CtrlSession) LeftQuotaFromOrigin() C.Bwidth {
	var dedicated C.Bwidth = 0
	for _, conf := range c.configmap {
		dedicated += conf.Quota
	}
	if c.user.IsCapped {
		return c.user.CappedQuota - dedicated
	}

	//to exclude already gifted bandwidth
	if c.user.GiftQuota < 0 {
		dedicated += -(c.user.GiftQuota)
	}

	return C.Bwidth(c.ctrl.CommonQuota.Load()) - dedicated
}

func (c *CtrlSession) LeftUsage() C.Bwidth {

	return c.user.CalculatedQuota + c.user.AdditionalQuota - c.TotalUsage()
}

func (c *CtrlSession) getsboxconf(confid int64) (*sbox.Userconfig, *db.Config, error) {
	config, ok := c.configmap[confid]
	if !ok {
		return nil, nil, C.ErrConfigNotFound
	}

	inbound, err := c.ctrl.GetdbInbound(int(config.InboundID))
	if err != nil {
		return nil, config, C.ErrInboundNotFound
	}
	outbound, err := c.ctrl.GetdbOutbound(int(config.OutboundID))
	if err != nil {
		return nil, config, C.ErrOutboundNotFound
	}
	var (
		vlessgrp  *sbox.Vlessgroup
		trojangrp *sbox.Trojangroup
		commongrp *sbox.Commongroup
	)

	switch config.Type {
	case C.Vless:
		vlessgrp = &sbox.Vlessgroup{
			UUID: config.UUID,
		}
	//TODO: add other config types here when adding them
	default:
		c.ctrl.logger.Warn("Unsupported Config Type Detected "+  config.Type)
	}

	return &sbox.Userconfig{
		Vlessgroup:  vlessgrp,
		Trojangroup: trojangrp,
		Commongroup: commongrp,
		Type:        inbound.Type,
		Name:        c.user.Name,
		Usage:       config.Usage,
		Quota:       config.Quota,
		DbID:        config.Id,
		Outboundtag: outbound.Tag,
		Inboundtag:  inbound.Tag,
		LoginLimit:  int32(config.LoginLimit),
		UsercheckId: int(c.user.CheckID),
		InboundId:   inbound.ID,
		OutboundID:  outbound.ID,
		TgId: c.user.TgID,
	}, config, nil
}

func (c *CtrlSession) add() {
	c.oprations.Add(1)
}

func (c *CtrlSession) done() {
	c.oprations.Add(-1)
}

func (c *CtrlSession) wait() {
	for {
		if c.oprations.Load() == 0 {
			break
		}
	}
}
