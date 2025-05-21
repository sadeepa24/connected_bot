package controller

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	sbConf "github.com/sadeepa24/connected_bot/sbox/conf"
	"github.com/sadeepa24/connected_bot/sbox/singapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"github.com/sadeepa24/connected_bot/tg/update/bottype"
)

// Ctrlsession is not theadsafe use in single thread
type CtrlSession struct {
	ctx    context.Context
	cancle context.CancelFunc

	ctrl   *Controller
	user   *db.User

	configmap map[int64]*db.Config
	olduser   db.User

	gift []db.Gift

	closed bool

	lowperm bool
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
				ctrl.logger.Info("Force closed old session")
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
		closed:    false,
	}
	if user.ConfigCount != 0 {
		err := ctrl.db.Model(&db.Config{}).Where("user_id = ?", user.TgID).Find(&session.user.Configs).Error
		if err != nil {
			return nil, err
		}
	}
	session.configmap = make(map[int64]*db.Config, ctrl.Maxconfigcount+1)
	for i, conf := range session.user.Configs {
		session.configmap[conf.Id] = &session.user.Configs[i]
	}

	ctrl.Addsession(session, user.TgID)
	session.olduser = *user
	session.lowperm = user.IsDistributedUser || user.IsRemoved || user.Restricted || user.Templimited || user.IsPaused
	return session, nil
}





func (c *CtrlSession) AddNewConfig(inboundid []int16, outboundid int16, Quota C.Bwidth, login int16, name string) (*db.Config, error) {
	c.ctrl.IncCriticalOp()
	defer c.ctrl.DecCriticalOp()
	
	var dbconf *db.Config
	if c.ctx.Err() != nil {
		return dbconf, C.ErrContextDead
	}
	dbconf = &db.Config{
		InboundIds:  inboundid,
		OutboundID:	 outboundid,
		Name:       name,
		UserID:     c.user.TgID,
		Active:     true,
		Download:   0,
		Upload:     0,
		Usage:      0,
		Quota:      Quota,
		LoginLimit: int16(login),
		CreatedAt: time.Now(),
	}
	c.typeCheckConfig(dbconf)
	if c.ctrl.db.Create(dbconf).Error != nil {
		return dbconf, C.WrapError(C.ErrDatabaseCreate, "Config Did n't create: DB opration failed try again later")
	}

	c.user.ConfigCount++
	c.user.UsedQuota += Quota
	c.configmap[dbconf.Id] = dbconf
	c.user.Configs = append(c.user.Configs, *dbconf)
	if !c.lowperm {
		_, err := c.ctrl.Boxapi.AddConfig(dbconf)
		if boxerr, ok := err.(singapi.Error); ok {
			if boxerr.IsBoxErr() {
				return dbconf, C.WrapError(err, "Config Created But Error occured When adding to singbox: " + boxerr.UserMsg())
			}
			return dbconf, C.WrapError(err, "Config Created But Error occured When adding to singbox you can check the config using configure and change inbound & outbound it may help: " + boxerr.UserMsg())
		}
		if err == nil && c.user.ConfigCount == 1 {
			c.ctrl.IncreaseUserCount(1)
		}
	}
	return dbconf, nil
}
func (c *CtrlSession) DeleteConfig(confid int64) error {
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	c.ctrl.IncCriticalOp()
	defer c.ctrl.DecCriticalOp()
	c.DeactivateConfig(confid)
	conf, ok := c.configmap[confid]
	if !ok {
		return C.WrapError(C.ErrConfigNotFound, "Config Delete Failed, Cannot Find Config May Be Already Deleted")
	}
	if c.ctrl.db.Delete(conf).Error != nil {
		return C.WrapError(C.ErrDbopration, "Config Delete Failed, Please try Again Later: Db opration Failed")
	}
	configQuota := conf.Quota
	delete(c.configmap, confid)
	for i := range c.user.Configs {
		if c.user.Configs[i].Id == confid {
			c.user.Configs[i] = c.user.Configs[len(c.user.Configs)-1]
			c.user.Configs = c.user.Configs[:len(c.user.Configs)-1]
			break
		}
	}
	c.user.ConfigCount--
	c.user.DeletedConfCount++
	c.user.UsedQuota = c.user.UsedQuota - configQuota
	return nil
}


func (c *CtrlSession) activateconf(conf *db.Config) (sbConf.Sboxstatus, error) {
	var status sbConf.Sboxstatus
	if (conf.Quota - conf.Usage) <= 0 {
		return status, C.CErrQuotaExceed
	}
	conf.Active = true
	return c.ctrl.Boxapi.AddConfig(conf)
}
func (c *CtrlSession) ActivateConfig(confid int64) (sbConf.Sboxstatus, error) {
	var status sbConf.Sboxstatus
	if c.updateperm() {
		return status, C.CErrNoPerm
	}
	conf, ok := c.configmap[confid]
	if !ok {
		return status, C.WrapError(errors.New(strconv.Itoa(int(confid))+ " config cannot found"), "config cannot dound for activation")
	}
	return c.activateconf(conf)
}
func (c *CtrlSession) ActivateAll() error {
	if c.updateperm() {
		return C.CErrNoPerm
	}
	if c.user.IsRemoved || !c.user.Verified() || c.user.IsMonthLimited || c.user.Restricted || c.user.IsDistributedUser  ||c.user.Templimited{
		return C.WrapError(errors.New("user is not in state that config can be activated "), "Err: User is not in state for the opration")
	}
	var err error
	for _, conf := range c.configmap {
		if _, errr := c.activateconf(conf); errr != nil {
			err = errors.Join(err, errr)
		}
	}
	return err
}
func (c *CtrlSession) DeactivateConfig(confid int64) (sbConf.Sboxstatus, error) {
	var stsatus sbConf.Sboxstatus
	if c.ctx.Err() != nil {
		return stsatus, C.ErrContextDead
	}
	conf, ok := c.configmap[confid]
	
	if !ok {
		return stsatus, C.WrapError(C.ErrConfigNotFound, "Config Cannot Found")
	}
	conf.Active = false
	status, err := c.ctrl.Boxapi.RemoveConfig(conf)
	if err != nil {
		return status, err
	}
	if status.FullUsage() > 0 {
		c.CreateUsagehistory(status, conf.Id)
	}
	conf.UpdateUsages(status)
	c.user.MonthUsage = (status.Download + c.user.MonthUsage + status.Upload)
	return status, nil
}
func (c *CtrlSession) DeactivateAll() error {
	var err error
	for _, conf := range c.configmap {
		if _, errr := c.DeactivateConfig(conf.Id); errr != nil {
			err = errors.Join(err, errr)
		}
	}
	return err
}
func (c *CtrlSession) ReActivateConfig(confid int64) (sbConf.Sboxstatus, error) {
	var status sbConf.Sboxstatus
	if c.updateperm() {
		return status, C.CErrNoPerm
	}
	conf, ok := c.configmap[confid]
	if !ok {
		return status, C.CErrConfigNotFound
	}
	if (conf.Quota - conf.Usage) <= 0 {
		return status, C.CErrQuotaExceed
	}
	conf.Active = true
	status, err := c.ctrl.Boxapi.AddConfigReset(conf)
	if err != nil {
		return status, err
	}
	if status.FullUsage() > 0 {
		c.CreateUsagehistory(status, conf.Id)
		conf.UpdateUsages(status)
		c.user.MonthUsage = (status.Download + c.user.MonthUsage + status.Upload)
	}
	return status, nil
}
func (c *CtrlSession) IsLowPerm() bool {
	return c.lowperm
}



func (c *CtrlSession) ConfigCloseConn(confid int64) error {
	conf, ok := c.configmap[confid]
	if !ok {
		return C.CErrConfigNotFound
	}
	return c.ctrl.Boxapi.CloseConns(conf)
}
func (c *CtrlSession) ChangeLoginLimit(confid int64, newlimit int16) (sbConf.Sboxstatus, error) {
	var status sbConf.Sboxstatus
	if c.lowperm {
		return status, C.CErrNoPerm
	}
	conf, ok := c.configmap[confid]
	if !ok {
		return status, C.CErrConfigNotFound
	}
	conf.LoginLimit = newlimit
	return c.ReActivateConfig(confid)
}


func (c *CtrlSession) ChangeLang(langcode string) {
	c.user.Lang = langcode
}

func (c *CtrlSession) GetConfig(confid int64) (*db.Config, error) {
	conf, ok := c.configmap[confid]
	if !ok {
		return nil, C.CErrConfigNotFound
	}
	return conf, nil
}

// return current vpn usage, old usage from db
// total usage for now = vpn usage +  old
// Every thing in byte format
// retur today, monthusage
func (c *CtrlSession) GetUsage() (C.Bwidth, C.Bwidth) {
	var status = sbConf.Sboxstatus{
		Download: 0,
		Upload:   0,
	}
	for _, config := range c.configmap {
		cstatus, err := c.ctrl.Boxapi.GetStatusConfig(config)
		if err != nil {
			continue
		}
		status.Download += cstatus.Download
		status.Upload += cstatus.Upload
	}
	return (status.Download + status.Upload), c.user.MonthUsage
}
// return all config's full usage for now
func (c *CtrlSession) GetFullUsage() bottype.FullUsage {
	bf := bottype.FullUsage{}
	for _, config := range c.configmap {
		cstatus, err := c.ctrl.Boxapi.GetStatusConfig(config)
		if err != nil {
			continue
		}
		bf.Download = bf.Download + config.Download + cstatus.Download
		bf.Upload = bf.Upload + config.Upload + cstatus.Upload
		bf.Uploadtd += cstatus.Upload
		bf.Downloadtd += cstatus.Download
	}
	return bf

}
// return total usage for this month
func (c *CtrlSession) TotalUsage() C.Bwidth {
	var status = sbConf.Sboxstatus{
		Download: 0,
		Upload:   0,
	}
	for _, config := range c.configmap {
		cstatus, err := c.ctrl.Boxapi.GetStatusConfig(config)
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
	conf, ok := c.configmap[confid]
	if !ok {
		return 0,0, C.ErrConfigNotFound
	}
	cstatus, err := c.ctrl.Boxapi.GetStatusConfig(conf)
	return cstatus.Download + cstatus.Upload, conf.Usage + cstatus.Download + cstatus.Upload, err
}
func (c *CtrlSession) GetconfigUsageTotal(confid int64) C.Bwidth {
	td, m, err := c.GetconfigUsage(confid)
	if err != nil {
		return 0
	}
	return td + m
}
func (c *CtrlSession) GetConfigFullUsage(confid int64) (bottype.FullUsage, sbConf.Sboxstatus) {
	var (
		conf  *db.Config
		err     error
		btusage bottype.FullUsage
	)
	conf, ok := c.configmap[confid]
	if !ok {
		return btusage, sbConf.Sboxstatus{}
	}
	btusage = bottype.FullUsage{
		Uploadtd:   0,
		Downloadtd: 0,
		Download:   conf.Download,
		Upload:     conf.Upload,
	}
	cstatus, err := c.ctrl.Boxapi.GetStatusConfig(conf)
	if err == nil {
		btusage.Download += cstatus.Download
		btusage.Upload += cstatus.Upload
		btusage.Downloadtd = cstatus.Download
		btusage.Uploadtd = cstatus.Upload
	}
	return btusage, cstatus
}
func (c *CtrlSession) GetconfigQuota(confid int64) C.Bwidth {
	conf, ok := c.configmap[confid]
	if !ok {
		return 0
	}
	return conf.Quota
}
func (c *CtrlSession) Getstatus(confid int64) (sbConf.Sboxstatus, error) {
	conf, ok := c.configmap[confid]
	if !ok {
		return sbConf.Sboxstatus{}, C.CErrConfigNotFound
	}
	return c.ctrl.Boxapi.GetStatusConfig(conf)

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

	return (c.user.CalculatedQuota + c.user.AdditionalQuota) - dedicated
}
func (c *CtrlSession) FullQuota() C.Bwidth {
	if c.user.IsCapped {
		return c.user.CappedQuota
	}
	return (c.user.CalculatedQuota + c.user.AdditionalQuota)
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

	if c.TotalUsage() != c.GetFullUsage().Full() {
		dedicated += (c.TotalUsage() - c.GetFullUsage().Full()) //hidden usage from deleted config
	}

	return C.Bwidth(c.ctrl.CommonQuota.Load()) - dedicated
}
func (c *CtrlSession) LeftUsage() C.Bwidth {
	return c.user.CalculatedQuota + c.user.AdditionalQuota - c.TotalUsage()
}


func (c *CtrlSession) Reseume() error {
	newperm := c.user.IsDistributedUser || c.user.IsRemoved || c.user.Restricted || c.user.Templimited
	if newperm {
		return C.CErrNoPerm
	}
	c.ctrl.IncCriticalOp()
	defer c.ctrl.DecCriticalOp()
	c.user.IsPaused = false
	c.updateperm()
	return c.ActivateAll()
}
func (c *CtrlSession) Pause() error {
	if c.lowperm {
		return C.CErrNoPerm
	}
	if c.user.Points == 0  {
		return C.CErrNoPoints
	}
	c.ctrl.IncCriticalOp()
	defer c.ctrl.DecCriticalOp()
	c.user.Points--
	c.user.IsPaused = true
	c.updateperm()
	return c.DeactivateAll()
}
func (c *CtrlSession) updateperm() bool  {
	c.lowperm = c.user.IsDistributedUser || c.user.IsRemoved || c.user.Restricted || c.user.Templimited || c.user.IsPaused
	return c.lowperm
}



func (c *CtrlSession) AddInboundConf(confid int64, inboundid int16) error {
	return c.addorremin("add", confid, inboundid)
}
func (c *CtrlSession) RemoveInboudConf(confid int64, inboundid int16) error {
	return c.addorremin("rem", confid, inboundid)
}
func(c *CtrlSession)  addorremin(op string, confid int64, inboundid int16 ) error {
	if c.lowperm {
		return C.CErrNoPerm
	}
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	conf, ok := c.configmap[confid]
	if !ok {
		return C.CErrConfigNotFound
	}
	switch op {
	case "rem":
		for i, id := range conf.InboundIds {
			if id == int16(inboundid) {
				conf.InboundIds = append(conf.InboundIds[:i], conf.InboundIds[i+1:]...)
				break
			}
		}
	case "add":
		conf.InboundIds = append(conf.InboundIds, int16(inboundid))
	}
	return c.ctrl.Boxapi.ResetInbounds(conf)
}
func (c *CtrlSession) typeCheckConfig(conf *db.Config) error {
	
	if conf.UUID == "" {
		var uid uuid.UUID 
		for {
			var err error
			uid, err = uuid.NewV4()
			if err != nil {
				return C.Erruuidcreatefailed
			}
			var exists bool
			if err = c.ctrl.db.Raw("SELECT EXISTS(SELECT 1 FROM configs WHERE uuid = ?)", uid.String()).Scan(&exists).Error; err != nil {
				return C.ErrDbopration
			}
			if exists {
				continue
			}
			break
		}
		conf.UUID = uid.String()
	}

	if conf.Password == "" {
		conf.Password = strconv.Itoa(int(conf.UserID)) + strconv.Itoa(int(rand.Int63()))
	}
	return nil
	
}
func (c *CtrlSession) ChangeOutbound(confid int64, outboundID int16) error {
	if c.lowperm {
		return C.CErrNoPerm
	}
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	conf, ok := c.configmap[confid]
	if !ok {
		return C.CErrConfigNotFound
	}
	out, ok := c.ctrl.Getoutbound(outboundID)
	if !ok {
		return C.ErrOutboundNotFound
	}
	conf.OutboundID = int16(out.Id)
	return c.ctrl.Boxapi.ChangeOutbound(conf)
}






func (c *CtrlSession) CreateUsagehistory(status sbConf.Sboxstatus, confid int64) error {
	return c.ctrl.db.CreateUsageHistory(&db.UsageHistory{
		Upload:   status.Upload,
		Download: status.Download,
		UserID:   c.user.TgID,
		Usage:    (status.Download + status.Upload),
		Date:     time.Now(),
		ConfigID: confid,
	})
}
func (c *CtrlSession) Chatupdate(chat string, val bool) {
	switch chat {
	case C.Group:
		c.user.IsInGroup = val
	case C.Channel:
		c.user.IsInChannel = val
	}
	c.user.IsRemoved = !(c.user.IsInChannel && c.user.IsInGroup)
	c.updateperm()
}


func (c *CtrlSession) AllGifts(force bool) ([]db.Gift, error) {
	if c.gift != nil && !force {
		return c.gift, nil
	}
	gifts := []db.Gift{} 
	err := c.ctrl.db.Model(&db.Gift{}).Where("send_valid = ? OR recive_valid = ?", true, true).Where("sender = ? OR reciver = ?", c.user.TgID, c.user.TgID).Find(&gifts).Error
	c.gift = gifts
	return c.gift, err
}


func (c *CtrlSession) GetUser() *db.User {
	return c.user
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
	c.CancelSentGift()
}
func (c *CtrlSession) CancelSentGift() {
	if c.user.GiftQuota != 0 {
		allgifts, err := c.AllGifts(true)
		if err == nil {
			for i := range allgifts {
				if allgifts[i].Sender == c.user.TgID {
					c.ctrl.CancelGift(allgifts[i], c.user)
				}
			}
		}
	}
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
func (c *CtrlSession) SaveConfig(confid int64) error {
	conf, ok := c.configmap[confid]
	if !ok {
		return C.CErrConfigNotFound
	}
	if err := c.ctrl.db.Save(conf).Error; err != nil {
		return DbError{
			error: err,
			msg:   "Database save operation failed. Some changes might not have been saved. Please verify and try again.",
		}
	}
	return nil
}
func (c *CtrlSession) SaveConfigs() error {
	if c.ctx.Err() != nil {
		return C.ErrContextDead
	}
	return c.saveConfigs()
}


func (c *CtrlSession) save() error {
	var errs error
	if err := c.ctrl.db.Save(c.user).Error; err != nil {
		errs = errors.Join(errs, err)
	}

	if len(c.user.Configs) > 0 {
		if err := c.ctrl.db.Save(c.user.Configs).Error; err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if errs != nil {
		return DbError{
			error: errs,
			msg:   "Database save operation failed. Some changes might not have been saved. Please verify and try again.",
		}
	}
	return nil
}
func (c *CtrlSession) saveConfigs() error {
	if c.user.ConfigCount > 0 {
		err := c.ctrl.db.Save(&c.user.Configs).Error
		if err != nil {
			return DbError{
				error: err,
				msg:   "Failed to save configuration. Some changes might not have been saved. Please verify and try again.",
			}
		}
	}
	return nil
}
func (c *CtrlSession) Close() error {
	if c.closed {
		return nil
	}
	var err error
	c.ctrl.RemoveSesion(c.user.TgID)
	c.closed = true
	if err = c.Save(); err != nil {
		return err
	}
	c.configmap = nil
	return nil
}
func (c *CtrlSession) ForceClose() error {
	err := c.Close()
	if c.cancle != nil {
		c.cancle()
	}
	return err
}