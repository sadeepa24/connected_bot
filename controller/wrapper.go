package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	sbConf "github.com/sadeepa24/connected_bot/sbox/conf"
	"github.com/sadeepa24/connected_bot/sbox/singapi"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"github.com/sadeepa24/connected_bot/tg/update"
	"github.com/sadeepa24/connected_bot/tg/update/bottype"
	"github.com/sagernet/sing-box/connectedbot/opts"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Controller struct {
	ctx    context.Context
	//sbox  
	Boxapi sbox.Controller
	db     *db.Database
	botapi botapi.BotAPI
	logger *zap.Logger

	Lockval    *atomic.Int32
	wLockCounter *atomic.Int32
	UpdateCounter *atomic.Int64
	metaconfig *C.MetadataConf
	*Metadata
	Overview *db.Overview

	Usermgrsession *sync.Map

	
	critical *atomic.Int32 //ongoing critical opration count such as chatmember updates, this value should be zero, in order to do a db refresh
	critchan chan interface{}

	lockchan chan struct{}

	basectx    context.Context    //parent context for all ongoing upx
	basecancle context.CancelFunc //cancle function for basecontext all upx will down

	lastDbRefresh *atomic.Value // this value only changed by watchman, all other routing read it so no race condition occure,
	signals chan any // share signals and message types to watchman (*botapi.Msgcommon, botapi.Upmessage, controller.UserCount, //TODO: forcedbrefresh signal) 
	boxcallbacks chan CallBackInfo 
	oprations *atomic.Int32
	waitCritical *atomic.Bool
	mu sync.RWMutex //mutext only use to read context I tried to create completly with out sync.bottlenext but i had to add this for small opration and it's allright
}

func New(ctx context.Context, cdb *db.Database, logger *zap.Logger, metaconf *C.MetadataConf, btapi botapi.BotAPI, sboxpath string) (*Controller, error) {
	if metaconf.WatchMgbuf <= 0 {
		metaconf.WatchMgbuf = 100
	}
	var err error
	boxapi, boxopts, err := singapi.NewsingAPI(ctx, sboxpath, logger, )
	if err != nil {
		return nil, errors.Join(err, errors.New("sing api creation failed"))
	}

	basectx, basecanc := context.WithCancel(ctx)
	cn := &Controller{
		ctx:            ctx,
		db:             cdb,
		logger:         logger,
		basectx:        basectx,
		Overview: &db.Overview{
			Mu: &sync.RWMutex{},
		},
		Boxapi: boxapi,
		basecancle:     basecanc,
		signals:         make(chan any, metaconf.WatchMgbuf),
		boxcallbacks: make(chan CallBackInfo, 500),
		Usermgrsession: &sync.Map{},
		lockchan: make(chan struct{}),
		Metadata: &Metadata{
			Inbounds:      []sbConf.Inboud{},
			Outbounds:     []sbConf.Outbound{},
			rawoptions: boxopts,
			inboundasMap:  make(map[int16]sbConf.Inboud, len(boxopts.Inbounds)),
			outboundasMap: make(map[int16]sbConf.Outbound, len(boxopts.Outbounds)),
			Botlink:       metaconf.Botlink,
			GroupLink:     metaconf.GroupLink,
			Channelink:    metaconf.Channelink,
			boxpath: sboxpath,
		},
		critical: new(atomic.Int32),
		metaconfig: metaconf,
		botapi:     btapi,
		Lockval:    new(atomic.Int32),
		wLockCounter: new(atomic.Int32),
		oprations: new(atomic.Int32),
		UpdateCounter: new(atomic.Int64),
		waitCritical: new(atomic.Bool),
		mu: sync.RWMutex{},
		lastDbRefresh: &atomic.Value{},
	}
    
	return cn, nil
}

type ForceResetUsage uint16 //use to send Newrefresh signal wit force reset all usage database checkcount will reset
type UserCount uint16 //sending Active usercount updates
type RefreshSignal uint16 //use to send Newrefresh signal 
type BroadcastSig string //use to send Broadcast signal with broadcast msg 

// returning channel can be used many things, user update count, que sending msg to user
// when buffring usercount update type should be UserCount
func (c *Controller) Getmgque() chan any {
	return c.signals
}
func (c *Controller) GetBoxCallback() chan CallBackInfo {
	return c.boxcallbacks
}

// msg should be type controller.UserCount, *botapi.Msgcommon, botapi.UpMessage:
// remove ctx argument later
func (w *Controller) Addquemg(msg any) {
	w.signals <- msg
}

func (c *Controller) Init() error {
	var (
		dbMeta     = &db.Metadata{Id: 1} // dbmeta is the loaded values from database not from configure file
		err        error
		dbnotfound bool
	)
	if c.metaconfig == nil {
		return errors.New("metaconfig not found ")
	}
	if c.Boxapi == nil {
		return errors.New("sbox creation failed")
	}
	if err = c.startbox(); err != nil {
		return err
	}

	if err = c.Metadata.Init(*c.metaconfig, c.logger); err != nil {
		return err
	}

	if c.metaconfig.DefaultDomain == "" || c.metaconfig.DefaultPublicIp == "" {
		return errors.New("default domain or public ip not found")
	} else {
		//TODO: verify ip and domain dns
	}

	c.DefaultDomain = c.metaconfig.DefaultDomain
	c.DefaultPubip = c.metaconfig.DefaultPublicIp

	if err = c.db.Model(&db.Metadata{}).First(dbMeta).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbnotfound = true //which mean running very first time 
		} else {
			return err
		}
	}

	// Count users who are in a group (is_in_group = true)
	if !dbnotfound {
		if err := c.db.Model(&db.User{}).Where("is_in_group = ? AND is_in_channel = ?", true, true).Count(&dbMeta.VerifiedUserCount).Error; err != nil {
			return err
		}
		if err := c.db.Model(&db.Config{}).Where("1 = 1").Count(&dbMeta.TotalConfigCount).Error; err != nil {
			return err
		}
	}

	
	//intilize All inbounds to map
	for _, in := range c.rawoptions.Inbounds {
		if in.Id == nil {
			continue
		}
		if in.Domain == "" {
			in.Domain = c.DefaultDomain
		}
		if in.Public_Ip == "" {
			in.Public_Ip = c.DefaultPubip
		}
		var inbdremake sbConf.Inboud
		err = inbdremake.AddOption(in)
		if err != nil {
			return err
		}
		c.inboundasMap[*in.Id] = inbdremake
		c.Inbounds = append(c.Inbounds, inbdremake)
		if in.Tag == "default" {
			c.defaultinbound = inbdremake
		}
	}

	if c.defaultinbound.Type == "" {
		c.defaultinbound = c.Inbounds[0]
		c.logger.Warn("Default Inbound Not Found First Inbound will bes used as default")
	}

	//intilize All outbounds to map
	for _, out := range c.rawoptions.Outbounds {
		if out.Type == "block" || out.Type == "dns" || out.Type == "selector" {
			continue
		}
		
		if out.Id == nil {
			return errors.New("outbound id not found for " + out.Tag)
		}

		var oubd sbConf.Outbound
		oubd.AddOption(out)
		c.outboundasMap[*out.Id] = oubd
		c.Outbounds = append(c.Outbounds, c.outboundasMap[*out.Id])
		
		if out.Tag == "default" {
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
		var outbd sbConf.Outbound
		outbd.AddOptionEndpoint(endpt)
		c.outboundasMap[*endpt.Id] = outbd
		c.Outbounds = append(c.Outbounds, c.outboundasMap[*endpt.Id])
	}
	if c.rawoptions.Route == nil {
		return errors.New("route cannopt be empty")
	}

	if c.defaultoutbound.Type == "" {
		for _, ou := range c.outboundasMap {
			if ou.Type == "direct" {
				c.defaultoutbound = ou
				break
			}
		}
		if c.defaultoutbound.Type == "" {
			return errors.New("default outbound not found create direct outbound")
		}
	}

	//if already db intilize verify all new and old  inbounds and make changes as needed
	if !dbnotfound {
		err = c.initallconfigs(dbMeta.TotalConfigCount, false)
		if err != nil {
			return errors.New("failed to init all configs " + err.Error())
		}
	}

	//replacing all inbounds according to new data
	for _, in := range c.Metadata.Inbounds {
		if err := c.db.Save(&db.Inbound{
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

	if c.metaconfig.RefreshRate <= 0 || c.metaconfig.RefreshRate > 24 {
		return errors.New("refresh rate should between 0 and 24")
	}
	dbMeta.LoginLimit = int32(c.metaconfig.LoginLimit)
	
	//initilizing db first time
	dbMeta.ChannelId = c.metaconfig.ChannelID
	dbMeta.GroupID = c.metaconfig.GroupID
	if dbnotfound { 
		if dbMeta.BandwidthAvelable, err = C.BwidthString(c.metaconfig.BandwidthAvelable); err != nil {
			return err
		}
		dbMeta.CommonQuota = dbMeta.BandwidthAvelable
		dbMeta.ResetCount = (30 * 24) / c.metaconfig.RefreshRate
		dbMeta.RefreshRate = c.metaconfig.RefreshRate
		dbMeta.PublicDomain = c.metaconfig.DefaultDomain
		dbMeta.PublicIp = c.metaconfig.DefaultPublicIp
		dbMeta.CommonWarnRatio = c.GetWarnRate()
		var userct int64
		if err = c.db.Model(&db.User{}).Count(&userct).Error; err != nil {
			dbMeta.Dbusercount = 0
		}
		dbMeta.Dbusercount = userct
		//Load to Database
	}

	if dbMeta.Maxconfigcount > c.metaconfig.Maxconfigcount {
		c.logger.Warn("Decrement of Maxconfigcount detected. This will not happen as users may have already created configs equal to Maxconfigcount.")
	} else {
		dbMeta.Maxconfigcount = c.metaconfig.Maxconfigcount
	}

	if c.metaconfig.RefreshRate != dbMeta.RefreshRate {
		c.logger.Info("Refresh Rate Change Detected. Recalculating Refresh Rates.")
		oldRefreshRate := dbMeta.RefreshRate
		dbMeta.CheckCount = (dbMeta.CheckCount * oldRefreshRate) / c.metaconfig.RefreshRate //Recalculating ResetCount according to new refresh rate
		dbMeta.ResetCount = (30 * 24) / c.metaconfig.RefreshRate
		dbMeta.RefreshRate = c.metaconfig.RefreshRate
	}

	if c.GetWarnRate() < int16(c.RefreshRate) {
		return errors.New("warn rate cannot be zero or lower than RefreshRate warnRate " +  strconv.Itoa(int(c.GetWarnRate())) + " refreshRate " + strconv.Itoa(int(c.RefreshRate)))
	}

	if c.GetWarnRate() != dbMeta.CommonWarnRatio {
		c.logger.Info("Warn rate change detected, resetting all warn rates of users")
		if err := c.db.Model(&db.User{}).Where("1 = 1").Update("warn_ratio", c.GetWarnRate()).Error; err != nil {
			return errors.New("errored when changing warn rate")
		}
		dbMeta.CommonWarnRatio = c.GetWarnRate()
	}

	if c.metaconfig.DefaultDomain != dbMeta.PublicDomain {
		c.logger.Info("Default Domain Changed")
		c.signals <- BroadcastSig("Default Domain Changed Use New Public Domain " + c.metaconfig.DefaultDomain)
		dbMeta.PublicDomain = c.metaconfig.DefaultDomain
	}

	if c.metaconfig.DefaultPublicIp != dbMeta.PublicIp {
		c.logger.Info("Default Public Ip Changed")
		c.signals <- BroadcastSig("Default Public Ip Changed Use New Public Ip (if you are using public domain and the public domain did not change, simply ignore this message )" + c.metaconfig.DefaultPublicIp)
		dbMeta.PublicIp = c.metaconfig.DefaultPublicIp
	}


	if c.metaconfig.GroupID == 0 || c.metaconfig.ChannelID == 0 {
		return errors.New("channel or group id not found ")
	}
	var Bandwidth C.Bwidth
	if Bandwidth, err = C.BwidthString(c.metaconfig.BandwidthAvelable); err != nil {
		return err
	}
	if Bandwidth == 0 {
		return errors.New("bandwidth cannot be zero")
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
	c.CommonQuota.Swap(dbMeta.CommonQuota)
	
	if err = c.db.Save(dbMeta).Error; err != nil {
		return err
	}


	return nil
}


//this function verify all config's inbound with new sbox config
func (c *Controller) initallconfigs(totalConfig int64, force bool) error {


	outfromdb := []*db.Outbound{}
	if err := c.db.Model(&db.Outbound{}).Find(&outfromdb).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	// verify all new outbound from config according to exting db outbound
	// reconfigure all oubounds according to new outbounds from config
	// all oubounds which are'nt avalble nolonger will replace by defaultoubound
	for _, outDb := range outfromdb {
		_, ok := c.outboundasMap[outDb.ID]

		if !ok {
			c.logger.Warn("not found new outbound for oubound from db " + outDb.Name)
			c.logger.Warn(outDb.Name + " Will replace by default outbound")
			//c.DefaultInboud()
			c.db.Model(&db.Config{}).Where("outbound_id = ?", outDb.ID).Update("outbound_id", c.defaultoutbound.Id)
			c.db.Delete(outDb)

		}
	}
	
	infromdb := []*db.Inbound{}
	if err := c.db.Model(&db.Inbound{}).Find(&infromdb).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	changedID := map[int16]bool{} //ids which is not available with new sbox config 
	for _, in := range infromdb {
		_, ok := c.inboundasMap[(in.ID)]
		if !ok {
			c.logger.Warn("Inbound from database (" + in.Tag + ") not found in the new configuration. All inbound lists of configurations using this inbound will be replaced or removed.")
			changedID[in.ID] = true
		}
		
	}

	if len(changedID) == 0 && !force {
		return nil
	}

	bufct := totalConfig / 10 
	if bufct < 20 {
		bufct = 30
	}
	if bufct > 100 {
		bufct = 100
	}
	bufsender := NewBufSender(c.ctx, c,  int(bufct), time.Duration(totalConfig * 3) * time.Second)
	go bufsender.Start()
	defer bufsender.Close()
	
	// verify all config's inbounds according to exting db inbound
	// reconfigure all inbounds according to new inbounds
	// all inbounds which are nolonger availble  will be removed

	var listConfig []db.Config
	save := make([]*db.Config, 0, C.Dbbatchsize)
	c.db.Model(&db.Config{}).FindInBatches(&listConfig, C.Dbbatchsize, func(tx *gorm.DB, batch int) error {
		var newinbounds []int16
		for i := range listConfig {
			if listConfig[i].Password == "" {
				listConfig[i].Password = strconv.Itoa(int(listConfig[i].UserID)) + strconv.Itoa(int(rand.Int63()))
			}
			if listConfig[i].CreatedAt.IsZero() {
				listConfig[i].CreatedAt = time.Now()
			}
			if len(listConfig[i].InboundIds) == 0 {
				listConfig[i].InboundIds = append(listConfig[i].InboundIds, c.defaultinbound.Id)
				bufsender.Send("you'r config " + listConfig[i].Name +"'s inbound has been changed due to zero inbound ", listConfig[i].UserID )
				save = append(save, &listConfig[i])
				continue
			}			
			for _, id := range listConfig[i].InboundIds {
				if changedID[id] {
					continue
				}
				newinbounds = append(newinbounds, id)
			}
			mustchange := len(listConfig[i].InboundIds) != len(newinbounds)
			if len(newinbounds) == 0 {
				newinbounds = append(newinbounds, c.defaultinbound.Id)
				mustchange = true
			}
			if mustchange {
				bufsender.Send("you'r config " + listConfig[i].Name +"'s inbound has been changed due to configuration changes please check new inbound or reconfigure you'r config's inbound as you need ", listConfig[i].UserID )
				listConfig[i].InboundIds = newinbounds
				save = append(save, &listConfig[i])
			}
			newinbounds = newinbounds[:0]
		}
		if len(save) > 0 {
			tx.Save(&save)
		}
		save = save[:0]
		return nil
	})
	c.db.Unscoped().Where("1 = 1").Delete(&db.Inbound{})
	bufsender.Over()
	return nil
}

func (c *Controller) RefreshAllConfig() error {
	c.IncCriticalOp()
	var totalconfs int64
	c.Overview.Mu.RLock()
	totalconfs = c.Overview.TotalConfCount
	c.Overview.Mu.RUnlock()
	err := c.initallconfigs(totalconfs, true)
	c.DecCriticalOp()
	if err != nil {
		return err
	}
	c.signals <- RefreshSignal(1)
	return nil
}

func (c *Controller) GetUser(user *tgbotapi.User) (*bottype.User, bool, error) {
	if user == nil {
		return nil, false, errors.New("cannot fetch user from nil user object")
	}
	dbUser, err := c.db.GetUser(user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, C.ErrDatabasefuncer
	}
	gotuser := bottype.Newuser(user, dbUser)
	return gotuser, true, nil
}
func (c *Controller) GetAllUserList(in *[]int64) error {
	if c.db.Model(&db.User{}).Pluck("tg_id", in).Error != nil {
		return C.ErrDbopration
	}
	return nil
}



var availableuserList = []string{
	C.UserLstAll,    
	C.UserLstTempLimited,  
	C.UserLstMonthLimited, 
	C.UserLstActive,       
	C.UserLstGroup,       
	C.UserLstDistributed, 
	C.UserLstVerified,    
	C.UserLstUnVerified,   
	C.UserLstRestricted,
	C.UserLstAddtional,
	//C.UserLstOnline,
}
func (c *Controller) AvailableUserList() []string {
	return availableuserList
}
func (c *Controller) GetUserList(listType string, in *[]int64) error  {
	switch listType {
	case C.UserLstAll:
		return c.GetAllUserList(in)
	case C.UserLstVerified:
		if err := c.db.Model(&db.User{}).
			Where("is_in_group = ? AND is_in_channel = ?", true, true).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}
	case C.UserLstUnVerified:
		if err := c.db.Model(&db.User{}).
			Where("is_in_group = ? OR is_in_channel = ?", false, false).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}
	case C.UserLstTempLimited:
		if err := c.db.Model(&db.User{}).
			Where("temp_limited = ?", true).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}
	case C.UserLstMonthLimited:
		if err := c.db.Model(&db.User{}).
			Where("is_month_limited = ?", true).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}
	case C.UserLstDistributed:
		if err := c.db.Model(&db.User{}).
			Where("is_dis_user = ?", true).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}		
	case C.UserLstGroup:
		if err := c.db.Model(&db.User{}).
			Where("is_in_group = ?", true).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}		
	case C.UserLstActive:
		if err := c.db.Model(&db.User{}).
			Where("is_removed = ? AND is_month_limited = ? AND temp_limited = ?", false, false, false).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}
	case C.UserLstRestricted:
		if err := c.db.Model(&db.User{}).
			Where("restricted = ?", true).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}
	case C.UserLstAddtional:
		if err := c.db.Model(&db.User{}).
			Where("additional_quota > ?", 0).
			Pluck("tg_id", in).Error; err != nil {
			return C.ErrDbopration
		}
	default:
		return C.ErrUnknownUserListType
	}
	return nil
}

func (c *Controller) GetOnlineConfList() ([]int64, error) {
	activeusrConfigs := c.Boxapi.GetAllUserStatus()
	added := map[int]bool{} 
	ids := []int64{}
	for userId := range activeusrConfigs {
		if len(activeusrConfigs[userId].Ip) > 0 && !added[userId] {
			ids = append(ids, int64(userId))
			added[userId] = true
		}
	}
	return ids, nil
}

func (c *Controller) GetUserById(userId int64) (*db.User, error) {
	var user = &db.User{
		TgID: userId,
	}	
	return user, c.db.Model(&db.User{}).First(user).Error
}

func (c *Controller) GetUserByConfID(confId int64) (*db.User, error) {
	var conf = &db.Config{
		Id: confId,
	}
	err := c.db.Model(&db.Config{}).First(conf).Error
	if err != nil {
		return nil, err
	}
	var user = &db.User{
		TgID: conf.UserID,
	}
	return user, c.db.Model(&db.User{}).First(user).Error
}

func (c *Controller) GetUserByUserName(userName string) (*db.User, error) {
	if userName == "" {
		return nil, errors.New("user name Cannot be empty")
	}
	var user = &db.User{}
	err := c.db.Model(&db.User{}).Where("username = ?", userName).First(user).Error
	if user.Username != userName {
		return user, errors.New("user not found")
	}
	return user, err
}

func (c *Controller) SearchUserByUsername(userName string) (*db.User, bool, error) {
	if userName == "" {
		return nil, false, errors.New("user name Cannot be empty")
	}
	var dbuser *db.User
	err := c.db.Model(&db.User{}).Where("username = ?", userName).First(dbuser).Error
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
			return nil, C.CErrDbopration
		}

	} else if userid, ok := to.(int); ok {
		if err = c.db.Model(&db.User{}).Where("tg_id = ?", userid).Preload("Configs").First(touser).Error; err != nil {
			return nil, C.CErrDbopration
		}
	} else {
		return nil, errors.New("invalid reciver")
	}
	c.CloseSession(touser.TgID)
	var giftcount int64 
	c.db.Model(&db.Gift{}).Where("reciver = ?", touser.TgID).Count(&giftcount)

	if giftcount >= c.MaxGiftCount {
		return nil, C.WrapError(errors.New("user gift limit exceed"), "gift limit exceed")
	}

	if err = c.db.Model(fromuser).Preload("Configs").First(fromuser).Error; err != nil {
		return nil, err
	}

	if touser.IsCapped {
		return touser, C.WrapError(C.ErrUserCanootReciveUserCapped, C.ErrUserCanootReciveUserCapped.Error())
	}
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
		return nil, C.CErrDbopration
	}

	if err = tx.Save(fromuser).Error; err != nil {
		tx.Rollback()
		return nil, C.CErrDbopration
	}
	if err = tx.Save(touser).Error; err != nil {
		tx.Rollback()
		return nil, C.CErrDbopration
	}
	if fromuser.ConfigCount > 0 {
		if err = tx.Save(&fromuser.Configs).Error; err != nil {
			tx.Rollback()
			return nil, C.CErrDbopration
		}
	}

	if touser.ConfigCount > 0 {
		if err = tx.Save(&touser.Configs).Error; err != nil {
			tx.Rollback()
			return nil, C.CErrDbopration
		}
	}

	//record
	tx.Model(&db.Gift{}).Create(&db.Gift{
		Date:        time.Now(),
		Sender:      fromuser.TgID,
		Reciver:     touser.TgID,
		Bandwidth:   quota,
	})
	c.db.CreateGiftLog(&db.GiftLog{
		SendID:    fromuser.TgID,
		RecivedID: touser.TgID,
		Bandwidth: quota,
		Date:      time.Now(),
	})

	return touser, tx.Commit().Error

}
func (c *Controller) CancelGift(gift db.Gift, sender *db.User) error {
	c.IncCriticalOp()
	defer c.DecCriticalOp()
	var touser = &db.User{
		TgID: gift.Reciver,
	}
	err := c.db.Model(&db.User{}).Preload("Configs").First(touser).Error
	if err != nil {
		return err
	}
	c.CloseSession(touser.TgID)
	err = c.db.Model(&db.User{}).Preload("Configs").First(sender).Error
	if err != nil {
		return err
	}
	touser.GiftQuota -=  gift.Bandwidth
	sender.GiftQuota +=  gift.Bandwidth
	c.RecalculateConfigquotas(touser)
	c.RecalculateConfigquotas(sender)
	tx := c.db.Begin()
	if err := tx.Save(touser).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(sender).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&gift).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Save(sender.Configs)
	tx.Save(touser.Configs)
	err = tx.Commit().Error
	if err == nil {
		c.DirectMg(fmt.Sprintf("gift bandwidth %s recived from %d was canceled", gift.Bandwidth.BToString(), gift.Sender), gift.Reciver, gift.Reciver)
		c.DirectMg(fmt.Sprintf("gift bandwidth %s that you have sent to %d was canceled", gift.Bandwidth.BToString(), gift.Sender), gift.Sender, gift.Sender)
	}
	return err
}

// user struct should have been preloaded configs
// this method does not save to db, caller should 
func (c *Controller) RecalculateConfigquotas(user *db.User) error {
	oldQuota := user.CalculatedQuota
	user.CalculatedQuota = c.CommonQuota.Load() + user.GiftQuota + user.AdditionalQuota

	if user.IsCapped && user.CappedQuota <= (user.CalculatedQuota - user.AdditionalQuota) {
		user.CalculatedQuota = user.CappedQuota
	}

	for i := range user.Configs {
		if user.Configs[i].Quota == 0 {
			continue
		}
		k := oldQuota / user.Configs[i].Quota      // findig ratio between oldquota and old configs quota
		user.Configs[i].Quota = user.CalculatedQuota / k // subpressing quota according to ratio, k is the constant
		
		status, err := c.Boxapi.AddConfigReset(&user.Configs[i])
		
		if err != nil {c.DirectMg("config adding failed you may need to contact admin with error err - " + err.Error(), user.TgID, user.TgID)}
		
		user.Configs[i].UpdateUsages(status)
		user.MonthUsage += (status.Download + status.Upload)

		if (user.Configs[i].Quota-user.Configs[i].Usage) <= 0 || (user.MonthUsage >= user.CalculatedQuota) || !user.CanUse() || !user.Verified() || user.Configs[i].UserOff {
			c.Boxapi.RemoveConfig(&user.Configs[i])
			user.Configs[i].Active = false
		}
		if err == nil && !user.IsDistributedUser && status.FullUsage() > 0 {
			c.db.CreateUsageHistory(&db.UsageHistory{
				Usage:    status.Download + status.Upload,
				Download: status.Download,
				Upload:   status.Upload,
				Date:     time.Now(),
				UserID:   user.TgID,
				ConfigID: user.Configs[i].Id,
			})
		}
	}

	return nil
}

func (c *Controller) DirectMg(text string, UserId int64, ChatID int64) error {
	mgcontext, cancle := context.WithTimeout(c.ctx, 2*time.Minute)
	_, err := c.botapi.SendContext(mgcontext, &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId:  ChatID,
			User_id: UserId,
		},
		Text: text,
	})
	cancle()
	return err
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
		Username: user.UserName,
		CalculatedQuota:   C.Bwidth(c.CommonQuota.Load()),
		WarnRatio: c.GetWarnRate(),
		RecheckVerificity: recheck,
		Lang:        c.metaconfig.DefaultLang,
		Points:      C.DefaultPoint,
		IsTgPremium: user.IsPremium,
		IsInChannel:   inchan,
		IsInGroup:     ingroup,
	}

	dbUser, err := c.db.AddUser(newuser)
	if err != nil {
		return nil, DbError{
			error: errors.Join(errors.New("new user adding failed user " + user.String()), err),
		}
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
func (c *Controller) CloseSession(UserId int64) (bool, error) {
	if forcecloser, loaded := c.Checksession(UserId); loaded {
			if closer, ok := forcecloser.(ForceCloser); ok {
				if err := closer.ForceClose(); err != nil {
					return false, err
				}
			}
			return true, nil
	}
	c.Usermgrsession.Delete(UserId)
	return false, nil
}
func (c *Controller) CloseAllUserSession() {
	c.Usermgrsession.Range(func(key, value any) bool {
		cls, ok := key.(ForceCloser) 
		if ok {
			cls.ForceClose()
		}
		return true	
	})
}



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
}
func (c *Controller) GetHelepCmdInfo() *C.HelpCommandInfo {
	return &c.HelperInfo
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
func (c *Controller) DeleteSboxConf(confId int64) error {
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
			if texttmpl.Keyboard !=nil {
				unwrapedmg.Buttons.OverideKeyboard(texttmpl.Keyboard)
			}
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

func (c *Controller) RemoveAllLimits() error {
	c.IncCriticalOp()
	tx := c.db.Begin()
	err := tx.Model(&db.User{}).Where("1 = 1").Update("is_month_limited", false).Error
	if err != nil {
		tx.Rollback()
		c.DecCriticalOp()
		return err
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		c.DecCriticalOp()
		return err
	}
	c.DecCriticalOp()
	if c.CheckLock() {
		return nil
	}
	c.signals <- RefreshSignal(1)
	return nil
}
func (c *Controller) RemoveAllRestriction() error {
	c.IncCriticalOp()
	lst := []int64{}
	c.GetUserList(C.UserLstRestricted, &lst)
	tx := c.db.Begin()
	err := tx.Model(&db.User{}).Where("restricted = ?", true).Update("restricted", false).Error
	if err != nil {
		tx.Rollback()
		c.DecCriticalOp()
		return err
	}
	err = tx.Exec("DELETE FROM restrict_users").Error
	if err != nil {
		tx.Rollback()
		c.DecCriticalOp()
		return err
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		c.DecCriticalOp()
		return err
	}

	if len(lst) > 0 {
		bufsnd := NewBufSender(c.ctx, c, len(lst), 5 * time.Minute)
		go bufsnd.Start()
		for _, user := range lst {
			bufsnd.Send("✅ admin removed you'r restriction, you can use service again 🎉", user)
		}
		bufsnd.Over()
	}
	c.DecCriticalOp()
	if c.CheckLock() {
		return nil
	}
	c.signals <- RefreshSignal(1)
	return nil
}


func (c *Controller) ResetLangCode() error {
	c.IncCriticalOp()
	tx := c.db.Begin()
	err := tx.Model(&db.User{}).Where("1 = 1").Update("lang", c.metaconfig.DefaultLang).Error
	if err != nil {
		tx.Rollback()
		c.DecCriticalOp()
		return err
	}
	tx.Commit()
	c.DecCriticalOp()
	if c.CheckLock() {
		return nil
	}
	return nil
}

func (c *Controller) SendFile(path string, filename string, msgcaption string, chat int64) error {
	file, err := os.Open(path)
	if err != nil {
		c.logger.Error("File Send Failed: read err", zap.Error(err))
		return err
	}
	defer file.Close()
	return c.SendAsFile(file, filename, msgcaption, chat)
}

func (c *Controller) SendAsFile(buf io.Reader, filename string, msgcaption string, chat int64) error {
	req, err :=  botapi.CreateMultiPartReq(c.ctx, "POST", c.botapi.CreateFullUrl("sendDocument"), map[string]string{
		"chat_id": strconv.Itoa(int(chat)),
		"caption": msgcaption,
	}, map[string]botapi.Filepart{
		"document": {
			Name: filename,
			Reader: buf,
		},
	})
	if err != nil {
		c.logger.Error("File Send Failed: request making failed" +  err.Error())
		return err
	}
	apires, err := c.botapi.SendRawReq(req)	
	if err != nil {
		c.logger.Error("File Send Failed: request send failed when uploading file", zap.Error(err))
		return err
	}
	if !apires.Ok {
		c.logger.Error("File Send Failed: Bad Response From Telegram: " + apires.Description)
	}
	return nil
}

// sbox 
func (c *Controller) startbox() error {
	c.Boxapi.SetCallBack(c.ReciveCallback)
	return c.Boxapi.Start()
}
type CallBackInfo struct {
	Code int16
	ConfigId int64
	Status *opts.CallBackResult
}

func (c *Controller) ReciveCallback(code int16, status *opts.CallBackResult) {
	if status == nil {
		return
	}
	select {
		case c.boxcallbacks <- CallBackInfo{
			Code: code,
			ConfigId: int64(status.Status.UserID),
			Status: status,
		}:
		default:
	}
}
func (c *Controller) RestrictUserByConfId(configId int64, reason string)  error {
	c.CheckLock()
	user, err := c.GetUserByConfID(configId)
	if err != nil {
		return err
	}
	if user.Restricted {
		return nil
	}
	ctlsession, err := NewSessionViaUser(c, user, true)
	if err != nil {
		return err
	}
	ctlsession.ctx = c.basectx
	ctlsession.Restrict(reason)
	ctlsession.Close()
	c.DirectMg("you'r are restricted. reason: " + reason, ctlsession.user.TgID, ctlsession.user.TgID)
	return nil
}
func (c *Controller) Close() error { return c.Boxapi.Close() }






//concurrent area
func (c *Controller) WatchmanLock() {
	c.Lockval.Swap(1)
}
func (c *Controller) WatchmanUnlock() {
	c.Lockval.Swap(0)
	waiters := c.wLockCounter.Swap(0)
	for i := 0; i < int(waiters); i++ {
		c.lockchan <- struct{}{}
	}
}
// check is that controller locked by watchman
// if locked this function wait for it to unlock
func (c *Controller) CheckLock() bool {
	if c.Lockval.Load() == 0 {
		return false
	}
	c.wLockCounter.Add(1)
	<-c.lockchan
	return true
}
func (c *Controller) FCheckLock() bool {
	return c.Lockval.Load() != 0 
}
func (c *Controller) IncCriticalOp() {
	c.critical.Add(1)
}
func (c *Controller) DecCriticalOp() {
	if c.critical.Add(-1) == 0  {
		if c.waitCritical.Load() {
			c.critchan <- struct{}{}
		}
	}
}

//only call ones by watchman do not call elsewhere
func (c *Controller) WaitCriticalop() {
	if c.critical.Load() == 0 {
		return
	}
	c.waitCritical.Swap(true)
	<-c.critchan
	c.waitCritical.Swap(false)
}
func (c *Controller) SetLastRefreshtime() {
	c.lastDbRefresh.Store(time.Now())
}
func (c *Controller) GetLastRefreshtime() time.Time {
	return c.lastDbRefresh.Load().(time.Time)
}
func (c *Controller) GetBaseContext() context.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.basectx
}

// canceling all ongoing upx
func (c *Controller) CancleUpdateContexs() {
	c.mu.Lock()
	c.basecancle()
	c.basectx, c.basecancle = context.WithCancel(c.ctx)
	c.mu.Unlock()
}