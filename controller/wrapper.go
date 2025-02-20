package controller

import (
	"context"
	"database/sql"
	"errors"
	"net/netip"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sadeepa24/connected_bot/sbox/singapi"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"github.com/sadeepa24/connected_bot/tg/update/bottype"
	"github.com/sagernet/sing-box/option"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Controller struct {
	ctx    context.Context
	sbox   sbox.Sboxcontroller
	db     *db.Database
	botapi botapi.BotAPI
	logger *zap.Logger
	// mu sync.Mutex

	Lockval    *atomic.Int32
	Metaconfig *MetadataConf
	//sboxio     *SboxIO
	*Metadata
	Overview *Overview

	Usermgrsession *sync.Map

	//sboxlog  chan any
	critical *atomic.Int32 //ongoing critical opration count such as chatmember updates, this value should be zero, in order to do a db refresh
	critchan chan interface{}
	//lockctx context.Context

	//cond *sync.Cond
	lockchanbuf []chan struct{}

	basectx    context.Context    //parent context for all ongoing upx
	basecancle context.CancelFunc //cancle function for basecontext all upx will down

	lastDbRefresh time.Time // this value only changed by watchman, all other routing read it so no race condition occure,

	signals chan any // share signals and message types to watchman (*botapi.Msgcommon, botapi.Upmessage, controller.UserCount, //TODO: forcedbrefresh signal) 
}

func New(ctx context.Context, db *db.Database, logger *zap.Logger, metaconf *MetadataConf, btapi botapi.BotAPI, sboxpath string) (*Controller, error) {

	//sboxlog := make(chan any, 2000) //buffer for reciving logs from sbox core

	if metaconf.WatchMgbuf <= 0 {
		metaconf.WatchMgbuf = 100
	}

	var err error
	boxapi, boxopts, err := singapi.NewsingAPI(ctx, sboxpath, logger)
	if err != nil {
		return nil, errors.Join(err, errors.New("sing api creation failed"))
	}

	basectx, basecanc := context.WithCancel(ctx)
	cn := &Controller{
		ctx:            ctx,
		db:             db,
		logger:         logger,
		basectx:        basectx,
		Overview: &Overview{
			Mu: &sync.RWMutex{},
		},
		sbox: boxapi,
		basecancle:     basecanc,
		signals:         make(chan any, metaconf.WatchMgbuf),
		Usermgrsession: &sync.Map{},
		Metadata: &Metadata{
			Inbounds:      []sbox.Inboud{},
			Outbounds:     []sbox.Outbound{},
			rawoptions: boxopts,
			inboundasMap:  make(map[int]sbox.Inboud, len(boxopts.Inbounds)),
			outboundasMap: make(map[int]sbox.Outbound, len(boxopts.Outbounds)),
			Botlink:       metaconf.Botlink,
			GroupLink:     metaconf.GroupLink,
			Channelink:    metaconf.Channelink,
		},
		critical: new(atomic.Int32),
		//critchan: make(chan interface{}),
		//cond: &sync.Cond{},
		Metaconfig: metaconf,
		botapi:     btapi,
		Lockval:    new(atomic.Int32),
		// sboxio: &SboxIO{
		// 	Inbounds:  boxopts.Inbounds,
		// 	outbounds: boxopts.Outbounds,
		// },
		//sboxlog: sboxlog,
	}



	return cn, nil
}

type ForceResetUsage uint16 //use to send Newrefresh signal wit force reset all usage database checkcount will reset
type UserCount int //sending usercount updates
type RefreshSignal uint16 //use to send Newrefresh signal 
type BroadcastSig string //use to send Broadcast signal with broadcast msg 

// returning channel can be used many things, user update count, que sending msg to user
// when buffring usercount update type should be UserCount
func (c *Controller) Getmgque() chan any {
	return c.signals
}

// msg should be type controller.UserCount, *botapi.Msgcommon, botapi.UpMessage:
// remove ctx argument later
func (w *Controller) Addquemg(upxctx context.Context, msg any) {
	if upxctx.Err() != nil {
		return
	}
	w.signals <- msg
}

func (c *Controller) Init() error {
	var (
		dbMeta     = &db.Metadata{Id: 1} // dbmeta is the loaded values from database not from configure file
		err        error
		dbnotfound bool
	)
	if err = c.Metadata.Init(*c.Metaconfig); err != nil {
		return err
	}

	if c.Metaconfig.DefaultDomain == "" || c.Metaconfig.DefaultPublicIp == "" {
		return errors.New("default domain or public ip not found")
	} else {
		//TODO: verify ip and domain dns
	}

	c.DefaultDomain = c.Metaconfig.DefaultDomain
	c.DefaultPubip = c.Metaconfig.DefaultPublicIp

	if c.Metaconfig == nil {
		return errors.New("metaconfig not found ")
	}
	if c.sbox == nil {
		return errors.New("sbox creation failed")
	}
	if err = c.startbox(); err != nil {
		return err
	}

	if err = c.db.Model(&db.Metadata{}).First(dbMeta).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbnotfound = true //which mean running very first time 
		} else {
			return err
		}
	}

	// Count users who are in a group (is_in_group = true)
	if err := c.db.Model(&db.User{}).Where("is_in_group = ? AND is_in_channel = ?", true, true).Count(&dbMeta.VerifiedUserCount).Error; err != nil {
		return err
	}

	//intilize All inbounds to map
	for _, in := range c.rawoptions.Inbounds {
		if in.Type != C.Vless {
			return errors.New("this type inbound not supported yet " + in.Type)
		}

		vlessout, ok := in.Options.(*option.VLESSInboundOptions)
		if !ok {
			return errors.New("this type inbound not supported yet " + in.Type)
		}


		if in.Id == nil {
			return errors.New("inbound id not found for " + in.Tag)
		}

		if in.Domain == "" {
			in.Domain = c.DefaultDomain
		}
		if in.Public_Ip == "" {
			in.Public_Ip = c.DefaultPubip
		}

		inbdremake := sbox.Inboud{
			Id:          int64(*in.Id),
			Name:        in.Tag,
			Tag:         in.Tag,
			Type:        in.Type,
			Option:      &in,
			Custom_info: in.Custom_info,
			Domain:      in.Domain,
			PublicIp:    in.Public_Ip,
			Support:     in.SupportInfo,
		}

		switch in.Type {
		case C.Vless:
			inbdremake.ListenAddres = vlessout.ListenOptions.Listen.Build(netip.IPv4Unspecified()).String()
			inbdremake.Listenport = int(vlessout.ListenPort)

			if vlessout.TLS != nil {
				inbdremake.Tlsenabled = vlessout.TLS.Enabled
			}
			if vlessout.Transport != nil {
				inbdremake.Transporttype = vlessout.Transport.Type
				inbdremake.Transportoption = *vlessout.Transport
			}
		default:
			return C.ErrNotsupported

		}
		c.inboundasMap[*in.Id] = inbdremake

		c.Inbounds = append(c.Inbounds, c.inboundasMap[*in.Id])
		if in.Tag == "default" {
			c.defaultinbound = c.inboundasMap[*in.Id]
		}
	}

	if c.defaultinbound.Type == "" {
		return errors.New("default inbound not found")
		//c.defaultinbound = c.Metadata.Inbounds[0]
	}

	//intilize All outbounds to map
	//c.sboxio.outbounds = append(c.sboxio.outbounds, c.rawoptions.Endpoints)
	for _, out := range c.rawoptions.Outbounds {
		if out.Id == nil {
			return errors.New("outbound id not found for " + out.Tag)
		}
		if out.Type == "block" || out.Type == "dns" || out.Type == "selector" {
			continue
		}
		c.outboundasMap[*out.Id] = sbox.Outbound{
			Id:          int64(*out.Id),
			Name:        out.Tag,
			Tag:         out.Tag,
			Type:        out.Type,
			///Option:      &out,
			Custom_info: out.Custom_info,
			Latency:     new(atomic.Int32),
		}
		c.Outbounds = append(c.Outbounds, c.outboundasMap[*out.Id])

		if out.Type == C.Direct {
			c.defaultoutbound = c.outboundasMap[*out.Id]
		}
	}

	for _, endpt := range c.rawoptions.Endpoints {
		if endpt.Id == nil {
			return errors.New("endpoint id not found for " + endpt.Tag)
		}

		_, loaded := c.outboundasMap[*endpt.Id]
		if loaded {
			return errors.New("outbound and endpoint id conflicts outbound and endpoint id canoot be same")
		}
		c.outboundasMap[*endpt.Id] = sbox.Outbound{
			Id:          int64(*endpt.Id),
			Name:        endpt.Tag,
			Tag:         endpt.Tag,
			Type:        endpt.Type,
			///Option:      &out,
			Custom_info: endpt.Custom_info,
			Latency:     new(atomic.Int32),
		}
		c.Outbounds = append(c.Outbounds, c.outboundasMap[*endpt.Id])
	}

	if c.rawoptions.Route == nil {
		return errors.New("route cannopt be empty")
	}

	if c.defaultinbound.Type == "" {
		return errors.New("default outbound not found create direct outbound")
	}

	//if already db intilize verify all new and old  inbounds and make changes as needed
	if !dbnotfound {

		infromdb := []*db.Inbound{}
		outfromdb := []*db.Outbound{}

		// verify all new inbound from config according to exting db inbound
		// reconfigure all inbounds according to new inbounds
		// all inbounds which are'nt avalble nolonger will replace by defaultoutbound
		if err := c.db.Model(&db.Inbound{}).Find(&infromdb).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		for _, dbIn := range infromdb {
			sboxin, ok := c.inboundasMap[int(dbIn.ID)]

			if !ok {
				c.logger.Warn("not found new inbound for inbound from db " + dbIn.Name)
				c.logger.Warn(dbIn.Name + " Will replace by default inbound")
				//c.DefaultInboud()
				c.db.Model(&db.Config{}).Where("inbound_id = ?", dbIn.ID).Update("inbound_id", c.defaultinbound.Id)
				c.db.Model(&db.Inbound{}).Delete(dbIn)

			} else if sboxin.Type != dbIn.Type {
				return errors.New("type conflicst same id has diffrent type inbounds")
			}
		}

		if err := c.db.Model(&db.Outbound{}).Find(&outfromdb).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

		}
		// verify all new outbound from config according to exting db outbound
		// reconfigure all oubounds according to new outbounds from config
		// all oubounds which are'nt avalble nolonger will replace by defaultoubound
		for _, outDb := range outfromdb {
			sboxout, ok := c.outboundasMap[int(outDb.ID)]

			if !ok {
				c.logger.Warn("not found new outbound for oubound from db " + outDb.Name)
				c.logger.Warn(outDb.Name + " Will replace by default outbound")
				//c.DefaultInboud()
				c.db.Model(&db.Config{}).Where("outbound_id = ?", outDb.ID).Update("outbound_id", c.defaultoutbound.Id)
				c.db.Model(&db.Outbound{}).Delete(outDb)

			} else if sboxout.Type != outDb.Type {
				return errors.New("outbound type conflict db and config outbound type missmatch")
			}
		}

		/*if len(infromdb) > len(c.Inbounds) {
			return errors.New("config inbounds not enougf")
		}
		if len(outfromdb) > len(c.Outbounds) {
			return errors.New("config outbound not enougf")
		}


		for _, in := range infromdb {
			if in.Type != c.inboundasMap[int(in.ID)].Type {
				return errors.New("inbound type conflict db and config inbound type missmatch")
			}
		}
		*/

	}

	//replacing all inbounds according to new data
	for _, in := range c.Metadata.Inbounds {

		if in.Type != "vless" {
			return errors.New("this type inbound not supported yet " + in.Type)
		}
		if err := c.db.Model(&db.Inbound{}).Where("id = ?", in.Id).Save(&db.Inbound{
			ID:   int16(in.Id),
			Tag:  in.Tag,
			Name: in.Name,
			Type: in.Type,
		}).Error; err != nil {
			return err
		}

	}

	//replacing all outbound according to new outbounds
	for _, out := range c.Metadata.Outbounds {
		if err := c.db.Model(&db.Outbound{}).Where("id = ?", out.Id).Save(&db.Outbound{
			ID:   int16(out.Id),
			Tag:  out.Tag,
			Name: out.Name,
			Type: out.Type,
		}).Error; err != nil {
			return err
		}

	}

	if err := c.db.Model(&db.User{}).Count(&dbMeta.Dbusercount).Error; err != nil {
		return err
	}

	if c.Metaconfig.RefreshRate <= 0 || c.Metaconfig.RefreshRate > 24 {
		return errors.New("refresh rate should between 0 and 24")
	}
 	
	//initilizing db first time
	dbMeta.LoginLimit = int32(c.Metaconfig.LoginLimit)
	
	if dbnotfound { 

		if dbMeta.BandwidthAvelable, err = C.BwidthString(c.Metaconfig.BandwidthAvelable); err != nil {
			return err
		}

		dbMeta.ChannelId = c.Metaconfig.ChannelID
		dbMeta.GroupID = c.Metaconfig.GroupID
		dbMeta.CommonQuota = dbMeta.BandwidthAvelable
		dbMeta.ResetCount = (30 * 24) / c.Metaconfig.RefreshRate
		dbMeta.RefreshRate = c.Metaconfig.RefreshRate
		dbMeta.PublicDomain = c.Metaconfig.DefaultDomain
		dbMeta.PublicIp = c.Metaconfig.DefaultPublicIp
		dbMeta.CommonWarnRatio = c.Metaconfig.GetWarnRate()
		
		var userct int64
		if err = c.db.Model(&db.User{}).Count(&userct).Error; err != nil {
			dbMeta.Dbusercount = 0
		}
		dbMeta.Dbusercount = userct
		//Load to Database
	}

	if dbMeta.Maxconfigcount > c.Metaconfig.Maxconfigcount {
		c.logger.Warn("Decrement of Maxconfigcount detected. This will not happen as users may have already created configs equal to Maxconfigcount.")
	} else {
		dbMeta.Maxconfigcount = c.Metaconfig.Maxconfigcount
	}

	if c.Metaconfig.RefreshRate != dbMeta.RefreshRate {
		c.logger.Info("Refresh Rate Change Detected. Recalculating Refresh Rates.")
		oldRefreshRate := dbMeta.RefreshRate
		dbMeta.CheckCount = (dbMeta.CheckCount * oldRefreshRate) / c.Metaconfig.RefreshRate //Recalculating ResetCount according to new refresh rate
		dbMeta.ResetCount = (30 * 24) / c.Metaconfig.RefreshRate
		dbMeta.RefreshRate = c.Metaconfig.RefreshRate
	}

	if c.Metaconfig.GetWarnRate() != dbMeta.CommonWarnRatio {
		c.logger.Info("Warn rate change detected, resetting all warn rates of users")
		if err := c.db.Model(&db.User{}).Update("warn_ratio", c.Metaconfig.GetWarnRate()).Error; err != nil {
			return errors.New("errored when changing warn rate")
		}
	}

	if c.Metaconfig.DefaultDomain != dbMeta.PublicDomain {
		c.logger.Info("Defaul Domain Changed")
		c.signals <- BroadcastSig("Default Domain Changed Use New Public Domain " + c.Metaconfig.DefaultDomain)
		dbMeta.PublicDomain = c.Metaconfig.DefaultDomain
	}

	if c.Metaconfig.DefaultPublicIp != dbMeta.PublicIp {
		c.logger.Info("Defaul Public Ip Changed")
		c.signals <- BroadcastSig("Default Public Ip Changed Use New Public Ip (if you are using public domain and the public domain did not change, simply ignore this message )" + c.Metaconfig.DefaultPublicIp)
		dbMeta.PublicIp = c.Metaconfig.DefaultPublicIp
	}


	if c.Metaconfig.GroupID == 0 || c.Metaconfig.ChannelID == 0 {
		return errors.New("channel or group id not found ")
	}
	var Bandwidth C.Bwidth
	if Bandwidth, err = C.BwidthString(c.Metaconfig.BandwidthAvelable); err != nil {
		return err
	}
	c.Metadata.GroupID = dbMeta.GroupID
	c.Metadata.ChannelId = dbMeta.ChannelId

	if Bandwidth != dbMeta.BandwidthAvelable {
		c.Metadata.BandwidthAvelable = Bandwidth
	} else {
		c.Metadata.BandwidthAvelable = dbMeta.BandwidthAvelable
	}

	c.Metadata.LoginLimit = dbMeta.LoginLimit
	c.Metadata.RefreshRate = dbMeta.RefreshRate
	c.Metadata.ResetCount = dbMeta.ResetCount //static
	c.Dbusercount.Swap(int32(dbMeta.Dbusercount))
	c.VerifiedUserCount.Swap(int32(dbMeta.VerifiedUserCount))
	c.Metadata.CheckCount.Swap(dbMeta.CheckCount)
	c.CommonQuota.Swap(dbMeta.CommonQuota.Int64())
	
	if err = c.db.Save(dbMeta).Error; err != nil {
		return err
	}


	return nil

	//c.db.Model(&db.Metadata{}).Find()
}


func (c *Controller) GetBaseContext() context.Context {
	return c.basectx
}

// canceling all ongoing upx
func (c *Controller) CancleUpdateContexs() {
	c.basecancle()
	c.basectx, c.basecancle = context.WithCancel(c.ctx)
}

func (c *Controller) GetUser(user *tgbotapi.User) (*bottype.User, bool, error) {
	if user == nil {
		return nil, false, errors.New("cannot fetch user from nil user object")
	}
	dbUser, err := c.db.GetUser(user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, C.ErrDatabasefuncer
	}
	gotuser := bottype.Newuser(user, dbUser)
	return gotuser, true, nil
}
func (c *Controller) GetUserList(in *[]int64) error {
	if c.db.Model(&db.User{}).Pluck("tg_id", in).Error != nil {
		return C.ErrDbopration
	}
	return nil
}
func (c *Controller) GetVerifiedUserList(in *[]int64) error { 
	if err := c.db.Model(&db.User{}).
		Where("is_in_group = ? AND is_in_channel = ?", true, true).
		Pluck("tg_id", in).Error; err != nil {
		return C.ErrDbopration
	}
	return nil
}
func (c *Controller) GetUnVerifiedUserList(in *[]int64) error {
	if err := c.db.Model(&db.User{}).
		Where("is_in_group = ? AND is_in_channel = ?", false, false).
		Pluck("tg_id", in).Error; err != nil {
		return C.ErrDbopration
	}
	return nil
}
func (c *Controller) GetUserById(userId int64) (*db.User, error) {
	var user = &db.User{
		TgID: userId,
	}	
	return user, c.db.Model(&db.User{}).First(user).Error
}
func (c *Controller) GetUserByUserName(userName string) (*db.User, error) {
	var user = &db.User{}
	err := c.db.Model(&db.User{}).Where("username = ?", userName).First(user).Error
	if user.Username.String != userName {
		return user, errors.New("user not found")
	}
	return user, err
}


func (c *Controller) SearchUserByUsername(username string) (*db.User, bool, error) {
	var dbuser *db.User
	err := c.db.Model(&db.User{}).Where("username = ?", username).First(dbuser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, C.ErrDatabasefuncer
	}
	return dbuser, true, nil
}

// Only checks reciver can recive the gift if not return err,
// Caller should check sender is valid
// input quota should be BYte format
func (c *Controller) Gift(upx *update.Updatectx, to any, quota C.Bwidth) (*db.User, error) {

	var touser = &db.User{}
	var fromuser = upx.User.User
	var err error

	if usertxt, ok := to.(string); ok {
		if err = c.db.Model(&db.User{}).Where("username = ?", usertxt).Preload("Configs").First(touser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, C.ErrDbnotfound
			}
			return nil, C.ErrDbopration
		}

	} else if userid, ok := to.(int); ok {
		if err = c.db.Model(&db.User{}).Where("tg_id = ?", userid).Preload("Configs").First(touser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, C.ErrDbnotfound
			}
			return nil, C.ErrDbopration
		}
	} else {
		return nil, errors.New("invalid reciver")
	}

	if err = c.db.Model(fromuser).Preload("Configs").First(fromuser).Error; err != nil {
		return nil, err
	}

	if touser.IsCapped {
		return touser, C.ErrUserCanootReciveUserCapped
	}

	// if touser.GiftQuota != 0 {
	// 	return touser, C.ErrUserGiftAlready
	// }

	if len(touser.Configs) <= 0 {
		return touser, C.ErrConfigNotFound
	}

	touser.GiftQuota = touser.GiftQuota + C.Bwidth(quota)
	fromuser.GiftQuota = fromuser.GiftQuota - (C.Bwidth(quota))

	c.RecalculateConfigquotas(fromuser)
	c.RecalculateConfigquotas(touser)

	tx := c.db.Begin()

	if tx.Error != nil {
		tx.Rollback()
		return nil, C.ErrDbopration
	}

	if err = tx.Save(fromuser).Error; err != nil {
		tx.Rollback()
		return nil, C.ErrDbopration
	}
	if err = tx.Save(touser).Error; err != nil {
		tx.Rollback()
		return nil, C.ErrDbopration
	}
	if fromuser.ConfigCount > 0 {
		if err = tx.Save(&fromuser.Configs).Error; err != nil {
			tx.Rollback()
			return nil, C.ErrDbopration
		}
	}

	if touser.ConfigCount > 0 {
		if err = tx.Save(&touser.Configs).Error; err != nil {
			tx.Rollback()
			return nil, C.ErrDbopration
		}
	}

	//record
	tx.Model(&db.Gift{}).Create(&db.Gift{
		Date:        time.Now(),
		Valid:       true,
		ComQuota:    C.Bwidth(c.CommonQuota.Load()),
		SendValid:   true,
		ReciveValid: true,
		Sender:      fromuser.TgID,
		Reciver:     touser.TgID,
		Bandwidth:   quota,
	})

	//TODO: remove this
	tx.Model(&db.GiftLog{}).Create(&db.GiftLog{
		SendID:    fromuser.TgID,
		RecivedID: touser.TgID,
		Bandwidth: quota,
		Date:      time.Now(),
	})

	return touser, tx.Commit().Error

}

// user struct should have been preloaded configs
// this method does not save to db, caller should 
func (c *Controller) RecalculateConfigquotas(user *db.User) error {
	oldQuota := user.CalculatedQuota
	user.CalculatedQuota = C.Bwidth(c.CommonQuota.Load()) + user.GiftQuota

	if user.IsCapped {
		user.CalculatedQuota = user.CappedQuota
	}

	for i := range user.Configs {

		k := oldQuota.Float64() / user.Configs[i].Quota.Float64()      // findig ratio between oldquota and old configs quota
		newConfigQuota := C.Bwidth(user.CalculatedQuota.Float64() / k) // subpressing quota according to ratio, k is the constant

		dbin, err := c.GetdbInbound(int(user.Configs[i].InboundID))
		if err != nil {
			_, dbin = c.DefaultInboud()
		}
		dbout, err := c.GetdbOutbound(int(user.Configs[i].OutboundID))
		if err != nil {
			_, dbout = c.Defaultoutboud()
		}

		user.Configs[i].Quota = newConfigQuota
		status, err := c.AddResetUserSbox(&sbox.Userconfig{
			Vlessgroup: &sbox.Vlessgroup{
				UUID: user.Configs[i].GetUUID(),
			},
			Type: user.Configs[i].Type,

			UsercheckId: int(user.CheckID),
			Name:        user.Name,
			Inboundtag:  dbin.Tag,
			Outboundtag: dbout.Tag,
			InboundId:   dbin.ID,
			DbID:        user.Configs[i].Id,
			OutboundID:  dbout.ID,
			Usage:       user.Configs[i].Usage,
			Quota:       newConfigQuota,
			LoginLimit:  int32(user.Configs[i].LoginLimit),
			TgId: user.TgID,
		})
		if err != nil {
			c.DirectMg("config adding failed you may need to contact admin with error err - " + err.Error(), user.TgID, user.TgID)
		}

		user.Configs[i].Usage += (status.Download + status.Upload)
		user.MonthUsage += (status.Download + status.Upload)
		user.Configs[i].Download += status.Download
		user.Configs[i].Upload += status.Upload

		if (user.Configs[i].Quota-user.Configs[i].Usage) <= 0 || user.IsDistributedUser || (user.IsCapped && user.CappedQuota > C.Bwidth(c.CommonQuota.Load())) || (user.MonthUsage >= user.CalculatedQuota) {

			// if (user.Configs[i].Quota - user.Configs[i].Usage) <= 0 {
			// 	c.DirectMg("your config "+user.Configs[i].Name+" Usage is over, config wo'nt work until renew", user.TgID, user.TgID)
			// }

			c.RemoveUserSbox(&sbox.Userconfig{
				Vlessgroup: &sbox.Vlessgroup{
					UUID: user.Configs[i].GetUUID(),
				},
				UsercheckId: int(user.CheckID),
				Name:        user.Name,
				Inboundtag:  dbin.Tag, //TODO: fetch this correctly
				Outboundtag: dbout.Tag,
				Usage:       user.Configs[i].Usage,
				Quota:       newConfigQuota,
				DbID:        user.Configs[i].Id,
				LoginLimit:  int32(user.Configs[i].LoginLimit),
				TgId: user.TgID,
			})
		}
		if err == nil && !user.IsDistributedUser {
			c.db.Create(&db.UsageHistory{
				Usage:    status.Download + status.Upload,
				Download: status.Download,
				Upload:   status.Upload,
				Date:     time.Now(),
				UserID:   user.TgID,
				ConfigID: user.Configs[i].Id,
			})
		}
		c.db.Save(&user.Configs[i])
	}

	return nil
}

func (c *Controller) DirectMg(text string, UserId int64, ChatID int64) error {
	mgcontext, cancle := context.WithTimeout(c.ctx, 2*time.Minute)
	c.botapi.SendContext(mgcontext, &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId:  ChatID,
			User_id: UserId,
		},
		Text: text,
	})
	cancle()
	return nil
}

func (c *Controller) Newuser(user *tgbotapi.User, chat *tgbotapi.Chat) (*bottype.User, error) {
	if user == nil || chat == nil {
		return nil, C.ErrChatOrUserNofound
	}

	var (
		inchan  bool
		ingroup bool
		err     error
		recheck bool
	)

	if _, inchan, err = c.botapi.GetchatmemberCtx(context.Background(), user.ID, c.Metadata.ChannelId); err != nil {
		recheck = true
		c.logger.Error("error when checking user is in channel err " + err.Error())
	}
	if _, ingroup, err = c.botapi.GetchatmemberCtx(context.Background(), user.ID, c.Metadata.GroupID); err != nil {
		recheck = true
		c.logger.Error("error when checking user is in group err " + err.Error())
	}

	newuser := &db.User{
		Joined:  time.Now(),
		TgID:    user.ID,
		CheckID: uint(c.Metadata.Dbusercount.Load()),
		Name:    user.FirstName + " " + user.LastName,
		Username: sql.NullString{
			String: user.UserName,
			Valid:  true,
		},
		CalculatedQuota:   C.Bwidth(c.CommonQuota.Load()),
		DeletedConfCount:  0,
		AddtionalConfig:   0,
		WarnRatio: c.Metaconfig.GetWarnRate(),
		RecheckVerificity: recheck,
		Lang:        "en",
		Points:      C.DefaultPoint,
		IsTgPremium: false,

		IsInChannel:   inchan,
		ConfigCount:   0,
		IsInGroup:     ingroup,
		IsBotStarted:  false,
		GroupBanned:   false,
		ChannelBanned: false,
		
		//IsVipUser:     false,
		// WebToken: sql.NullString{
		// 	String: "no token", //TODO: change after making wqb app
		// 	Valid:  true,
		// },
	}

	dbUser, err := c.db.AddUser(newuser)
	if err != nil {
		return nil, C.ErrDatabaseCreate
	}
	c.Metadata.Dbusercount.Add(1)
	gotuser := bottype.Newuser(user, dbUser)
	gotuser.Newuser = true

	return gotuser, nil
}

func (c *Controller) IncreaseUserCount(count int) {
	c.signals <- UserCount(count)
}

func (c *Controller) Checksession(UserId int64) (any, bool) {
	return c.Usermgrsession.Load(UserId)

}

func (c *Controller) Addsession(closefunc ForceCloser, UserId int64) {
	c.Usermgrsession.Store(UserId, closefunc)

}

func (c *Controller) RemoveSesion(UserId int64) {
	c.Usermgrsession.Delete(UserId)
}

// Do not use this func its slow
func (c *Controller) SetIsbotarted(userID int64, val bool) error {
	return c.db.Model(&db.User{}).Where(&db.User{TgID: userID}).Update("is_bot_started", val).Error
}

func (c *Controller) Getadminchat() (map[int64]string, error) {
	chat := make(map[int64]string)
	if c.Metadata.GroupID != 0 {
		chat[c.Metadata.GroupID] = C.Group
	}
	if c.Metadata.GroupID != 0 {
		chat[c.Metadata.ChannelId] = C.Channel
	}

	return chat, nil
	//return c.db.Getadminchat()
}

func (c *Controller) GetHelepCmdInfo() bottype.HelpCommandInfo {
	// return bottype.HelpCommandInfo{
	// 	CommandPageCount: 3,
	// 	BuilderHelp: 2,
	// 	TutorialPageCount: 2,
	// 	InfoPageCount: 2,

	// }
	return c.HelperInfo
}

// return reffrld, verified, error
func (c *Controller) ReffralCount(owenerid int64) (int64, int64, error) {

	var count int64 = 0
	alluser := []db.Reffral{}
	err := c.db.Model(&db.Reffral{}).Where("owner_id = ? AND expired = ?", owenerid, false).Find(&alluser).Error
	if err != nil {
		return 0, 0, err
	}

	users := []int64{}
	for _, ref := range alluser {
		users = append(users, ref.UserId)
	}
	if err = c.db.Model(&db.User{}).Where("tg_id IN ? AND is_in_channel = ?", users, true).Where("is_in_group = ?", true).Count(&count).Error; err != nil {
		return int64(len(alluser)), count, err

	}
	return int64(len(alluser)), count, err

}

func (c *Controller) CreateRefrral(owenerid, userid int64) (*db.Reffral, error) {
	user := &db.Reffral{
		UserId: userid,
	}
	err := c.db.Model(&db.Reffral{}).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newus := &db.Reffral{
				OwnerID: owenerid,
				UserId:  userid,
				Expired: false,
			}

			return newus, c.db.Create(newus).Error
		}
		return nil, err
	}
	return user, C.ErrUserExitDb

}

func (c *Controller) startbox() error {
	return c.sbox.Start()
}

//func (c *Controller) Getinbounds() ([]sbox.Inboud) { return c.Metadata.Inbounds }

func (c *Controller) WatchmanLock() {
	c.Lockval.Swap(1)

}

func (c *Controller) WatchmanUnlock() {
	c.Lockval.Swap(0)
	time.Sleep(1 * time.Millisecond) //make sure value swaped so that no more chan add to list

	for _, chans := range c.lockchanbuf {
		close(chans)
	}
	c.lockchanbuf = []chan struct{}{}

}

// check is that controller locked by watchman
// if locked this function wait for it to unlock
func (c *Controller) CheckLock() bool {
	if c.Lockval.Load() == 0 {
		return false
	}
	tmpchan := make(chan struct{})
	c.addlockchan(tmpchan)
	<-tmpchan
	return true
}

func (c *Controller) addlockchan(lockchan chan struct{}) {
	c.lockchanbuf = append(c.lockchanbuf, lockchan)
}

func (c *Controller) Close() error { return c.sbox.Close() }

func (c *Controller) AdduserSbox(conf *sbox.Userconfig) (sbox.Sboxstatus, error) {
	return c.sbox.AddUser(conf)
}
func (c *Controller) AddResetUserSbox(conf *sbox.Userconfig) (sbox.Sboxstatus, error) {
	return c.sbox.AddUserReset(conf)
}
func (c *Controller) RemoveUserSbox(conf *sbox.Userconfig) (sbox.Sboxstatus, error) {
	return c.sbox.RemoveUser(conf)
}
func (c *Controller) GetstatusUserSbox(conf *sbox.Userconfig) (sbox.Sboxstatus, error) {
	return c.sbox.GetstatusUser(conf)
}

func (c *Controller) UrlTestOut(tag string) (int16, error) {
	return c.sbox.UrlTest(tag)

}

// heavy load on db
// optimize later
func (c *Controller) ClaimReferVerified(owenerid int64) (int, error) {

	alluser := []db.Reffral{}
	err := c.db.Model(&db.Reffral{}).Where("owner_id = ? AND expired = ?", owenerid, false).Find(&alluser).Error
	if err != nil {
		return 0, err
	}
	//tx := c.db.Begin
	users := []int64{}
	for _, ref := range alluser {
		users = append(users, ref.UserId)
	}
	Ousers := []db.User{}
	if err = c.db.Model(&db.User{}).Where("tg_id IN ? AND is_in_channel = ? AND is_in_group = ?", users, true, true).Find(&Ousers).Error; err != nil {
		return 0, err
	}

	verified := []int64{}
	for _, ref := range Ousers {
		verified = append(verified, ref.TgID)
	}
	tx := c.db.Begin()
	if err = tx.Model(&db.Reffral{}).Where("user_id IN ? AND owner_id = ? AND expired = ?", verified, owenerid, false).UpdateColumn("expired", true).Error; err != nil {
		tx.Rollback()
		return 0, C.ErrDbopration
	}
	user := db.User{
		TgID: owenerid,
	}
	if tx.Model(&db.User{}).First(&user).Error != nil {
		tx.Rollback()
		return 0, C.ErrDbopration
	}
	user.Points += int64(len(verified) * 2)

	if err = tx.Save(user).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	return len(verified) * 2, nil
}

func (c *Controller) RefreshUrlTest() {
	c.sbox.RefreshUrlTest()

}

func (c *Controller) GetSboxConfig(userID int64) ([]db.SboxConfigs, error) {
	sboxconfs := []db.SboxConfigs{}
	if err := c.db.Model(&db.SboxConfigs{}).Where("user_id = ?", userID).Find(&sboxconfs).Error; err != nil {
		return nil, err
	}
	return sboxconfs, nil
	//return
}

func (c *Controller) GetSpecificConf(userId int64, name string) (db.SboxConfigs, error) {
	conf := []db.SboxConfigs{}

	if err := c.db.Model(&db.SboxConfigs{}).Where("user_id = ? AND name = ?", userId, name).Find(&conf).Error; err != nil {
		return db.SboxConfigs{}, err
	}
	if len(conf) > 0 {
		return conf[0], nil
	}
	return db.SboxConfigs{}, nil

}

func (c *Controller) CreateSboxConf(userId int64, name string) (db.SboxConfigs, error) {
	conf := &db.SboxConfigs{
		UserID:   userId,
		Name:     name,
		ConfPath: strconv.Itoa(int(userId)) + "-" + name + ".json",
	}

	if err := c.db.Model(&db.SboxConfigs{}).Create(conf).Error; err != nil {
		return *conf, err
	}

	return *conf, nil

}

// this give configs according to server not from builder
func (c *Controller) GetUserConfigs(userID int64) ([]db.Config, error) {
	var confs []db.Config
	return confs, c.db.Model(&db.Config{}).Where("user_id = ?", userID).Find(&confs).Error
}

// Deletes buildconfig not releted to server configs
func (c *Controller) DeleteConf(confId int64) error {
	return c.db.Model(&db.SboxConfigs{}).Delete(&db.SboxConfigs{
		ID: confId,
	}).Error

}

func (c *Controller) LoadEvents(userID int64) (map[string]db.Event, error) {
	var events = []db.Event{}
	if err := c.db.Model(&db.Event{}).Where("user_id = ?", userID).Find(&events).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return map[string]db.Event{}, nil
		}
		return nil, err
	}

	return C.SliceToMap(events, func(eve db.Event) string {
		return eve.Name
	}), nil
}

func (c *Controller) AddEvent(userId int64, name string) error {
	//c.db.Model(&db.Event)
	tx := c.db.Begin()
	err := tx.Model(&db.Event{}).Create(&db.Event{
		UserId: userId,
		Name:   name,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (c *Controller) RefreshUser(ctx context.Context, dbuser *db.User) error {
	if dbuser == nil {
		return errors.New("input user nil object")
	}
	var (
		err1 error
		err2 error
		is bool
	)
	if _, is, err1 = c.botapi.GetchatmemberCtx(ctx, dbuser.TgID, c.ChannelId); is {
		dbuser.IsInChannel = true
		
	}
	if _, is, err2 = c.botapi.GetchatmemberCtx(ctx, dbuser.TgID, c.GroupID); is {
		dbuser.IsInGroup = true
	}

	if err1 != nil || err2 != nil {
		dbuser.RecheckVerificity = true
	}
	err := c.db.Save(dbuser).Error

	return err
}

func (c *Controller) UpdatePoint(newpointCount int64, userId int64) error {
	tx := c.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	err := tx.Model(&db.User{}).
		Where("tg_id = ?", userId). // This can be omitted if First is used with userId directly
		Update("points", newpointCount).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	
	return nil
}

func (c *Controller) IncCriticalOp() {
	c.critical.Add(1)
}

func (c *Controller) DecCriticalOp() {
	if c.critical.Add(-1) == 0 && c.critchan != nil {
		c.critchan <- struct{}{}
	}
}

func (c *Controller) WaitCriticalop() {
	if c.critical.Load() == 0 {
		return
	}
	c.critchan = make(chan interface{})
	<-c.critchan
	c.critchan = nil
}

func (c *Controller) GetLastRefreshtime() time.Time {
	c.CheckLock()
	return c.lastDbRefresh
}

// only use In watchman,
// Do not use elsewhere
func (c *Controller) SetLastRefreshtime() {
	c.lastDbRefresh = time.Now()
}

func (c *Controller) SendMsgContext(ctx context.Context, msg any) (*tgbotapi.Message, error) {
	var (
		repmg *tgbotapi.Message
		err   error
	)
	switch unwrapedmg := msg.(type) {
	case *botapi.Msgcommon:
			if unwrapedmg.Endpoint == "" {
				unwrapedmg.Endpoint = C.ApiMethodSendMG
			}
			repmg, err = c.botapi.SendContext(ctx, unwrapedmg)
	case botapi.UpMessage:
		var texttmpl *botapi.Message
		texttmpl, err = c.botapi.GetMgStore().GetMessage(unwrapedmg.TemplateName, unwrapedmg.Lang, unwrapedmg.Template)
		if err != nil {
			c.logger.Error("failed to get message from msgstore template - " + unwrapedmg.TemplateName , zap.Error(err))
			return nil, err
		}
		sendmg := botapi.Msgcommon{
			Parse_mode: texttmpl.ParseMode,
			Infocontext: &botapi.Infocontext{
				ChatId: unwrapedmg.DestinatioID,
			},
		}
		if unwrapedmg.Buttons != nil {
			sendmg.Reply_markup = unwrapedmg.Buttons.Getkeyboard()
		}
		sendmg.Meadiacommon = &botapi.Meadiacommon{}
		sendmg.Caption = texttmpl.String()
		if texttmpl.MedType == C.MedPhoto {
			sendmg.Photo = texttmpl.MediaId
			sendmg.Endpoint = C.ApiMethodSendPhoto
		} else if texttmpl.MedType == C.MedVideo {
			sendmg.Video = texttmpl.MediaId
			sendmg.Endpoint = C.ApiMethodSendVid
		} else {
			sendmg.Meadiacommon = nil
			sendmg.Text = texttmpl.Msg
			sendmg.Endpoint = C.ApiMethodSendMG
		}
		repmg, err = c.botapi.SendContext(ctx, &sendmg)
	default:
		return nil, C.ErrNotMsgType
	}
	return repmg, err
}
