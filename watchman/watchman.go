package watchman

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Watchman has accsess to everything
type Watchman struct {
	ctx    context.Context
	db     *db.Database
	ctrl   *controller.Controller
	botapi botapi.BotAPI

	config *C.Watchmanconfig
	logger *zap.Logger

	ticker *time.Ticker
	close  chan struct{}

	DeleteQue chan int64

	
	//simplelog *simplelog.SimpleLog

	mgstore *botapi.MessageStore
	//lastUserCount int32 //User count on db when running function RefreshDb last time
	
	
	// When a new UserCount signal is received, the updater checks the following condition:
	// If lastRefreshActiveUser + (lastRefreshActiveUser / 4) < activeUser, a new refresh cycle is triggered.
	activeUser int64             // Real-time count of users currently using the service.
	lastRefreshActiveUser int64  // User count at the time of the last database refresh.


	//msgque chan *botapi.Msgcommon

}

func New(ctx context.Context,
	ctrl *controller.Controller,
	btapi botapi.BotAPI,
	db *db.Database,
	config *C.Watchmanconfig,
	logger *zap.Logger,
	mgstore *botapi.MessageStore,

) (*Watchman, error) {

	if config == nil {
		config = &C.Watchmanconfig{}
	}

	if config.Delbuffer <= 0 {
		config.Delbuffer = 10
	}

	return &Watchman{
		ctx:       ctx,
		db:        db,
		ctrl:      ctrl,
		botapi:    btapi,
		close:     make(chan struct{}),
		logger:    logger,
		config:    config,
		DeleteQue: make(chan int64, config.Delbuffer),
		mgstore:   mgstore,
	}, nil
}

func (w *Watchman) Start() error {
	var err error
	if w.ctrl.Metadata.RefreshRate <= 0 {
		return errors.New("refresh rate cannot be lower than 0")
	}
	w.ticker = time.NewTicker(time.Duration(w.ctrl.Metadata.RefreshRate) * time.Hour)
	go func() {
		for _, out := range w.ctrl.Outbounds {
			t, err := w.ctrl.Boxapi.UrlTest(out.Tag)
			if err != nil {
				w.logger.Error("urltest error " + out.Tag + " err - " + err.Error())
				out.Latency.Swap(-1)
				continue
			}
			out.Latency.Swap(int32(t))
		}
	}()

	go w.startAutoupdater()


	startrefresh, cancle := context.WithTimeout(w.ctx, 5*time.Minute)
	refreshdone := make(chan struct{})
	go func(ctx context.Context) {
		//w.ctrl.RefreshUrlTest()
		err = w.RefreshDb(startrefresh, false, false)
		if err != nil {
			w.logger.Error("fatal error db start refresh " + err.Error())
			cancle()
			return
		}
		refreshdone <- struct{}{}

	}(startrefresh)

	select {
	case <-refreshdone:
		cancle()
		break
	case <-startrefresh.Done():
		cancle()
		return errors.New("watchman: intlize db refresh failed context timeout or canceled")
	}
	return nil
}

func (w *Watchman) Close() error {

	w.close <- struct{}{} //close chan is not a buffred chan so this opration wait for w.close recive
	w.RefreshDb(w.ctx, false, false)
	<-w.close
	w.logger.Debug("watchman closing done")
	return nil
}

func (w *Watchman) startAutoupdater() {
	w.logger.Info("Starting Watchman And AutoUpdater")
update:
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Warn("context cancled autoupdater closing")
			w.logger.Warn("Force Closing DB")
			break update
		case <-w.close:
			w.logger.Sync()
			w.logger.Info("Closing Auto Updater close call recived")
			w.close <- struct{}{}
			break update
		case tick := <-w.ticker.C:
			w.logger.Info("db refresh starting", zap.String("tick", tick.String()), zap.Int32("count", w.ctrl.CheckCount.Load()))
			go func () {
				for _, out := range w.ctrl.Outbounds {
					t, err := w.ctrl.Boxapi.UrlTest(out.Tag)
					if err != nil {
						out.Latency.Swap(-1)
						continue
					}
					out.Latency.Swap(int32(t))
				}
			}()
			refreshctx, cancle := context.WithCancel(w.ctx)
			err := w.RefreshDb(refreshctx, true, false)
			cancle()
			if err != nil {
				w.logger.Error("Db Refresh Failed Due to: ", zap.Error(err))
				w.ctrl.DirectMg("Db refresh Failed; You may need to check what happend", w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
				continue
			}
			w.logger.Info("db refresh done", zap.String("tick", tick.String()), zap.Int32("count", w.ctrl.CheckCount.Load()))
			w.logger.Sync()
		case confid := <- w.ctrl.GetBoxCallback():
			switch confid.Code {
			case C.BoxCallBackTorrent:
				if confid.Status == nil {
					w.logger.Error("torrent callback status is nil", zap.Int64("config id", confid.ConfigId))
					continue
				}
				ips := ""
				for ip := range confid.Status.Status.Ip {
					ips += ip + ", "
				}
				err := w.ctrl.RestrictUserByConfId(confid.ConfigId, "download bittorrent time:" + time.Now().Format("2006-01-02 15:04:05") + " ip: " + ips +" more info " + confid.Status.String())
				if err != nil {
					w.logger.Error("user restriction failed", zap.Int64("config id", confid.ConfigId))
					continue
				}
				w.logger.Info("user restricted due to download torrent", zap.Int64("config id", confid.ConfigId))
			}
		case mg := <-w.ctrl.Getmgque():
			switch unwrapedmg := mg.(type) {
			case controller.RefreshSignal:
				w.ctrl.DirectMg("force refresh added", w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
				if w.ctrl.CheckLock() {
					continue	
				}
				refreshctx, cancle := context.WithCancel(w.ctx)
				err := w.RefreshDb(refreshctx, false, false)
				cancle()
				if err != nil {
					w.ctrl.DirectMg("refresh failed", w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
					w.logger.Error("Force Db Refresh Failed Due to: ", zap.Error(err))
				} else {
					w.ctrl.DirectMg("refresh done", w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
				}
			case controller.BroadcastSig:
				go func ()  {
					userlist := []int64{}
					if w.ctrl.GetAllUserList(&userlist) != nil {
						w.logger.Error("error while feteching userlist to broadcast msg " + string(unwrapedmg) )
						return
					}
					for _, user := range userlist {
						w.ctrl.DirectMg(string(unwrapedmg), user, user)
					}
				}()
			case controller.ForceResetUsage:
				refreshctx, cancle := context.WithCancel(w.ctx)			
				err := w.RefreshDb(refreshctx, false, true)
				cancle()
				if err != nil {
					w.logger.Error("Usercount Db Refresh Failed Due to: ", zap.Error(err))
					continue
				}
				w.logger.Info("db refresh done")
			case controller.UserCount:
				w.activeUser += int64(unwrapedmg)
				if w.ctrl.CheckLock() {
					continue	
				}
				if float32(w.lastRefreshActiveUser)+((float32(w.lastRefreshActiveUser)/4)*3) < float32(w.activeUser) {
					refreshctx, cancle := context.WithCancel(w.ctx)
					err := w.RefreshDb(refreshctx, false, false)
					cancle()
					if err != nil {
						w.logger.Error("Usercount Db Refresh Failed Due to: ", zap.Error(err))
						continue
					}
					w.logger.Info("db refresh done")
				}
				continue
			default:
				repmg, err := w.ctrl.SendMsgContext(w.ctx, mg)
				if err != nil {
					if errors.Is(err, C.ErrClientRequestFail) {
						w.ctrl.Getmgque() <- mg // buffer again
					}
					continue update
				}
				if repmg.Chat != nil && repmg.Chat.ID == w.ctrl.GroupID {
					w.Delmg(repmg.MessageID)
				}
			}
		}
	}
}

type preprosessd struct {
	cappeduser        int64    //total user count who capped their bandwidth
	verifiedusercount int64
	monthlimiteduser  int64
	distributeduser   int64
	restricted 		  int64
	templimiteduser   int64

	captotal          C.Bwidth //total bandwidth capped
	totaladdtional    C.Bwidth // total additional bandwidth from users who can really use it
	UsedByLimitedUsers C.Bwidth
	savings C.Bwidth
	
	unUsedUser 		  int64 //to calculate mainquota

	configCount int64




}

// Calculate the quota for each user based on various parameters
// Parameters include: verified user count, capped user, month-limited user, gifted user, usage overridden user
// Additional quota from users
// Overused users can't use their whole quota due to usage rollback from last month (this is removed)
// MainCommonUserQuota = ((w.ctrl.BandwidthAvelable) - (predata.captotal + predata.usedbyrestricted + predata.totaladdtional + predata.usedbydisuser)) / C.Bwidth(predata.verifiedusercount-(predata.cappeduser+predata.distributeduser+predata.monthlimiteduser+predata.restricted))

//removed
// overused user can't just use their whole quota (due adding usage rollback from lastmonth,  this month initial usage = lastmonth excess usage - last month his quota  ),  so it's like increase of bandwidth but finnaly it's same
//MainCommonUserQuota = ((w.ctrl.BandwidthAvelable) - (predata.captotal + predata.usedbyrestricted + predata.totaladdtional + predata.usedbydisuser)) / C.Bwidth(predata.verifiedusercount-(predata.cappeduser+predata.distributeduser+predata.monthlimiteduser+predata.restricted))
func (p *preprosessd) MainQuota(total C.Bwidth, AddtionalQuotaType uint8) C.Bwidth {
	var aadtional = p.totaladdtional
	if AddtionalQuotaType != 0 {
		aadtional = 0
	}
	if p.verifiedusercount-p.unUsedUser > 0 {
		return ((total + p.savings) - (p.captotal + aadtional + p.UsedByLimitedUsers)) / C.Bwidth(p.verifiedusercount-p.unUsedUser)
	}
	return total
}

// TODO: remove after testings
func (p preprosessd) String() (s string) {
	s = fmt.Sprintf(`
	cappeduser %v
	captotal %v 
	verifiedusercount %v
	totaladdtional %v
	monthlimiteduser %v
	distributeduser %v
	
	`,
		p.cappeduser,
		p.captotal,
		p.verifiedusercount,
		p.totaladdtional,
		p.monthlimiteduser,
		p.distributeduser,
	)

	return
}

func (w *Watchman) CheckClose() error {
	select {
	case <-w.close:
		return errors.New("close signal recived")
	case <-w.ctx.Done():
		return C.ErrContextDead
	default:
		return nil
	}
}

func (w *Watchman) Delmg(delmg int) {

	if len(w.DeleteQue) >= cap(w.DeleteQue)-1 {
		delmg := <-w.DeleteQue
		timeoutctx, cancle := context.WithTimeout(w.ctx, 2*time.Minute)
		defer cancle()
		w.botapi.DeleteMsg(timeoutctx, delmg, w.ctrl.GroupID)
	}
	w.DeleteQue <- int64(delmg)

}


// refresh member verificity
// refresh usage to database
// if docount true CheckkCount will increase by one
// if forceReset true All Usage Will Resets And Checkcount will be zero
const MaxRetry = 10
func (w *Watchman) RefreshDb(refreshcontext context.Context, docount bool, forceReset bool) error {
	w.ctrl.WatchmanLock() //locking for dbrefresh, all new upx will be paused
	w.ctrl.WaitCriticalop()      // waiting for all critical opration done
	w.ctrl.CloseAllUserSession() // closing all usersession safely
	w.ctrl.CancleUpdateContexs() // cancling all non critical ongoing upx

	st := time.Now()
	var (
		checkcount = w.ctrl.CheckCount.Load()
		condcheck  = func() bool {
			return ((checkcount == w.ctrl.ResetCount) && docount) || forceReset
		}
		err error
	)

	w.logger.Info("Batch Updating Database Count ", zap.Int32("checkcount", checkcount))

	var chanmax = w.ctrl.VerifiedUserCount.Load()

	if chanmax < 30 {
		chanmax = 35
	}
	if chanmax > 100 {
		chanmax = 100 
	}

	bufsender := controller.NewBufSender(w.ctx, w.ctrl, int(chanmax), time.Duration(w.ctrl.Overview.TotalUser * 3) * time.Second)
	go bufsender.Start()
	defer func ()  {
		bufsender.Over()
		bufsender.Close()
		w.ctrl.WatchmanUnlock()
	}()
	predata, err := w.PreprosessDb(refreshcontext, bufsender)
	if err != nil {
		bufsender.Send("Predata prosseing error Please Make Manual Refresh := " + err.Error(), w.ctrl.SudoAdmin)
		return errors.Join(errors.New("predata prosseing failed"), err)
	}
	w.ctrl.VerifiedUserCount.Swap(int32(predata.verifiedusercount))


	MainCommonUserQuota := predata.MainQuota(w.ctrl.BandwidthAvelable, w.config.AddtionalQuotaType) // Newcalculated main quota for each user
	// this value used to calculate the old ratio between config quota and old maincommonquota
	// new config quota will calculate based on this ratio
	oldCommonQuota := w.ctrl.CommonQuota.Swap(MainCommonUserQuota)
	
	
	w.ctrl.Overview.Mu.Lock()
	w.ctrl.Overview.QuotaForEach = MainCommonUserQuota
	w.ctrl.Overview.Mu.Unlock()

	if oldCommonQuota > 0 {
		var glistUser []db.User
		todel := make([]db.Gift, 0, C.Dbbatchsize)
		tosaveuser :=  make([]db.User, 0, C.Dbbatchsize)
		tosavegift := make([]db.Gift, 0, C.Dbbatchsize)
		giftsaved := make(map[int64]struct{}, C.Dbbatchsize)
		loadedgift := make(map[int64]struct{}, C.Dbbatchsize)
	
		var (
			chnged bool
			send[]db.Gift
			recive []db.Gift
		)
		err := w.db.Model(&db.User{}).
		Preload("SentGifts").Preload("ReceivedGifts").
		Where("gift_quota != ? OR is_capped = ?", 0, true).
		FindInBatches(&glistUser, C.Dbbatchsize, func(tx *gorm.DB, batch int) error {
			for i := range glistUser {
				chnged = false
				if glistUser[i].GiftQuota != 0 {
					glistUser[i].GiftQuota = 0
					chnged = true
					send = glistUser[i].SentGifts
					recive = glistUser[i].ReceivedGifts

					for s := range glistUser[i].SentGifts {
						send[s].Bandwidth = (MainCommonUserQuota/oldCommonQuota) *  send[s].Bandwidth
						if send[s].Isgifttimeover() {							
							if  _, loaded := loadedgift[send[s].ID]; loaded {
								delete(loadedgift, send[s].ID)
								todel = append(todel, send[s])
							}
							loadedgift[send[s].ID] = struct{}{}
							continue
						}
						if _, alsaved := giftsaved[send[s].ID]; alsaved {
							delete(giftsaved, send[s].ID)
							tosavegift  = append(tosavegift, send[s])
						} else {
							giftsaved[send[s].ID] = struct{}{}
						}
						glistUser[i].GiftQuota -= send[s].Bandwidth
					}
					for r := range recive {
						recive[r].Bandwidth = (MainCommonUserQuota/oldCommonQuota) *  recive[r].Bandwidth
						if recive[r].Isgifttimeover() {
							if glistUser[i].Verified() && (!glistUser[i].CanUse() || glistUser[i].ConfigCount == 0) {
								predata.UsedByLimitedUsers += recive[r].Bandwidth
							}
							if _, loaded := loadedgift[recive[r].ID]; loaded {
								delete(loadedgift, recive[r].ID)
								todel = append(todel, recive[r])
							}
							loadedgift[recive[r].ID] = struct{}{}
							continue
						}
						if _, alsaved := giftsaved[recive[r].ID]; alsaved {
							delete(giftsaved, recive[r].ID)
							tosavegift  = append(tosavegift, recive[r])
						} else {
							giftsaved[recive[r].ID] = struct{}{}
						}
						glistUser[i].GiftQuota += recive[r].Bandwidth
					}
				}
				if glistUser[i].IsCapped && glistUser[i].CappedQuota > MainCommonUserQuota + glistUser[i].GiftQuota {
					if glistUser[i].Verified() {
						predata.cappeduser--
						predata.captotal -= glistUser[i].CappedQuota
						if glistUser[i].CanUse() {
							predata.unUsedUser--
						}
					}
					bufsender.Send("you'r are no longer capped user, due our main quota is lower than you'r capped quota", glistUser[i].TgID)
					glistUser[i].IsCapped = false
					glistUser[i].CappedQuota = 0
					chnged = true
				}
				if chnged {
					tosaveuser  = append(tosaveuser, glistUser[i])
				}
			}
			if len(todel) > 0 {
				tx.Delete(&todel)
			}
			if len(tosaveuser) > 0 {
				tx.Save(&tosaveuser)
			}
			if len(tosavegift) > 0 {
				tx.Save(&tosavegift)
			}
			todel = todel[:0:C.Dbbatchsize]
			tosaveuser = tosaveuser[:0:C.Dbbatchsize]
			tosavegift = tosavegift[:0:C.Dbbatchsize]
			return nil

		},).Error
		if err != nil {
			return err
		}
		glistUser = nil
		todel = nil
		tosaveuser = nil
	}
	MainCommonUserQuota = predata.MainQuota(w.ctrl.BandwidthAvelable, w.config.AddtionalQuotaType)
	
	var (
		listUser []db.User
	)
	allconfigs := make([]*db.Config, 0, C.Dbbatchsize)
	usagehistr := make([]db.UsageHistory, 0, C.Dbbatchsize)
 
	err = w.db.Model(&db.User{}).
	Preload("Configs").
	FindInBatches(&listUser, C.Dbbatchsize, func(tx *gorm.DB, batch int) error {
		if tx.Error != nil {
			return tx.Error
		}
		for i := range listUser {
			user := &listUser[i]
			if refreshcontext.Err() != nil {
				w.logger.Warn("🔴🔴🔴 Force stopping DB updating, Db update stops middle of db update. Db may malformed 🔴🔴🔴" + user.Name)
				bufsender.Send("🔴🔴🔴 force stopped when db refresh, you may need to start bot with last backup. see logs for more info", w.ctrl.SudoAdmin )
				return fmt.Errorf("context cancled db refresh stops from record id %v, err %v ", user.TgID, refreshcontext.Err())
			}
			//tx.Model(&db.Config{}).Where("user_id = ?", user.TgID).Find(&user.Configs)
			// storing old quota for calculating
			oldQuota := user.CalculatedQuota
		
			user.CalculatedQuota = MainCommonUserQuota + user.GiftQuota + user.AdditionalQuota
			userVerifycity := user.IsInChannel && user.IsInGroup
			user.ConfigCount = int16(len(user.Configs))
			if user.IsCapped {
				user.CalculatedQuota = user.CappedQuota
			}
		
			var (
				usedquota C.Bwidth
				oldUsage = user.MonthUsage
			)
			//configs:
			for i := range user.Configs {
				newConfigQuota := C.Bwidth(0)
				if user.Configs[i].Quota != 0 {
					k := oldQuota / user.Configs[i].Quota      // findig ratio between oldquota and old configs quota
					newConfigQuota = user.CalculatedQuota / k  // subpressing quota according to ratio, k is the constant
				} else {
					bufsender.Send("you have config that don't have any quota please remove it or increase quota", user.TgID)
					user.Configs[i].Active = false
					allconfigs = append(allconfigs, &user.Configs[i])
					continue
				}
				usedquota += newConfigQuota
				user.Configs[i].Quota = newConfigQuota
				var (
					forceremove bool
				)
				if (newConfigQuota - user.Configs[i].Usage > 0) && userVerifycity && user.CanUse() && !user.Configs[i].UserOff  {
					status, err := w.ctrl.Boxapi.AddConfigReset(&user.Configs[i])
					if err != nil {
						if cerr, ok := err.(C.Error); ok {
							bufsender.Send("config " +user.Configs[i].Name + " got error while adding to singbox, msg = " + cerr.UserMsg() , user.TgID)
						}
						w.logger.Error("config add failed: config " +user.Configs[i].Name , zap.Error(err))
						status, _ = w.ctrl.Boxapi.RemoveConfig(&user.Configs[i])
						user.Configs[i].Active = false
						err = nil
					}
					user.Configs[i].UpdateUsages(status)
					user.MonthUsage += status.FullUsage()
					if status.FullUsage() > 0 {
						usagehistr = append(usagehistr, db.UsageHistory{
							Usage:    status.Download + status.Upload,
							Download: status.Download,
							Upload:   status.Upload,
							Date:     time.Now(),
							UserID:   user.TgID,
							ConfigID: user.Configs[i].Id,
							Name: user.Name,
						})
						
					}
					if user.Configs[i].Usage >= user.Configs[i].Quota  {
						forceremove = true
					} else {
						if !user.Configs[i].Active {
							bufsender.Send("Good News Configuration "+ user.Configs[i].Name+" Online Again Due to Bandiwdth Change 🔄", user.TgID)
						}
						user.Configs[i].Active = true
					}
				}
				if user.Configs[i].Active  && (!user.CanUse() || forceremove || (newConfigQuota - user.Configs[i].Usage <= 0) || user.Configs[i].UserOff) {
					if (user.Configs[i].Quota - user.Configs[i].Usage) <= 0 {
						bufsender.Send("⚠️ Your configuration "+user.Configs[i].Name+" has exceeded its usage limit. The config will not function until it is renewed. 🔄", user.TgID)
					}
					status, err := w.ctrl.Boxapi.RemoveConfig(&user.Configs[i])
					if err == nil && status.FullUsage() > 0 && !forceremove {
						
						if status.FullUsage() > 0 {
							user.Configs[i].UpdateUsages(status)
							user.MonthUsage += status.FullUsage()
						}
						usagehistr = append(usagehistr, db.UsageHistory{
							Usage:    status.Download + status.Upload,
							Download: status.Download,
							Upload:   status.Upload,
							Date:     time.Now(),
							UserID:   user.TgID,
							ConfigID: user.Configs[i].Id,
							Name: user.Name,
						})
					}
					user.Configs[i].Active = false
				}
				allconfigs = append(allconfigs, &user.Configs[i])
			}
			user.UsedQuota = usedquota
			if user.Verified() && user.CanUse()  && docount && user.MonthUsage <= user.CalculatedQuota && user.ConfigCount > 0 {
				if oldUsage == user.MonthUsage   { //which means user did n't use the config for last refresh cycle
					user.EmptyCycle++
					if user.EmptyCycle >= user.WarnRatio && user.WarnRatio != 0 {
						user.Templimited = true	// hecan't use the service until he remove this war manually
						user.EmptyCycle = 0
						user.WarnRatio = user.WarnRatio/2
						bufsender.Send( C.GetMsg(C.MsgTemplimit), user.TgID)
						for i := range user.Configs {
							w.ctrl.Boxapi.RemoveConfig(&user.Configs[i])
							user.Configs[i].Active = false
						}
						if user.WarnRatio == 0 {
							bufsender.Send(C.GetMsg(C.MsgTempOver), user.TgID)
						}

					}
				} else if (oldUsage != user.MonthUsage){
					user.EmptyCycle = 0
				}
			}
			if user.UsedQuota > user.CalculatedQuota {
				w.logger.Warn("violation, usedquota > calculatedquota detected from " + user.String())
				bufsender.Send("We have detetcted you have bigger quota than we allocated to fix this we overide you'r config's quota", user.TgID)
				user.UsedQuota = user.CalculatedQuota
				quotaforeach := user.CalculatedQuota / C.Bwidth(user.ConfigCount)
				for i := range user.Configs {
					user.Configs[i].Quota = quotaforeach
				}
			}
			if condcheck() {			
				if user.IsDistributedUser && !user.Restricted {
					bufsender.Send(C.GetMsg(C.MsgDistributeOver), user.TgID)
				}
				if (user.IsMonthLimited || user.WarnRatio == 0 ) && !user.Restricted {
					bufsender.Send("You'r Limitation is over", user.TgID)
				}
				user.AddPoint(10)
				//user.SavedQuota = 0
				if user.MonthUsage < ((user.CalculatedQuota*3)/4) && !user.IsMonthLimited && !user.IsDistributedUser && !user.Restricted && !(user.WarnRatio != 0) { 
					//check whether user used 75% from his quota if not user will limited next 30 days
					bufsender.Send(C.GetMsg(C.MsgQuotanotUsed), user.TgID)
					user.IsMonthLimited = true
					user.AlltimeUsage += user.MonthUsage
					user.MonthUsage = 0
				} else  {
					bufsender.Send(C.GetMsg(C.MsgresetUsage), user.TgID)
					user.IsMonthLimited = false
					user.AlltimeUsage += user.MonthUsage
					user.MonthUsage = 0
				}
				for i := range user.Configs {
					user.Configs[i].Usage = 0
					user.Configs[i].Upload = 0
					user.Configs[i].Download = 0
					if user.IsMonthLimited {
						user.Configs[i].Active = false
					}
				}
				
				user.WarnRatio = w.ctrl.GetWarnRate()
				user.IsDistributedUser = false
			}
		}
		if len(allconfigs) > 0 {
			err = w.txsave(&allconfigs, tx, bufsender, batch)
			if err != nil {
				return err
			}
			
		}
		err = w.txsave(&listUser, tx, bufsender, batch)
		if len(usagehistr) > 0 {
			w.db.CreateUsageHistories(&usagehistr)
		}
		usagehistr = usagehistr[:0:cap(usagehistr)]
		allconfigs = allconfigs[:0:cap(allconfigs)]
		if err != nil {
			return err
		}
		w.logger.Info("batch prosess done", zap.Int("batchNum", batch), zap.Int("usercount", len(listUser)), zap.Duration("elapsed time", time.Since(st))) 
		return nil
	},
	).Error
	allconfigs = nil
	usagehistr = nil
	listUser = nil

	if err != nil {
		return err
	}
	
	//updating metadata
	var dbmeta = &db.Metadata{ //only one order in db for metadata
		Id: 1,
	}

	if err = w.db.Model(&db.Metadata{}).First(dbmeta).Error; err != nil {
		time.Sleep(100 * time.Millisecond)
		if w.db.Model(&db.Metadata{}).First(dbmeta).Error != nil {
			return errors.New("db update success but metadata updating failed due to metdata fetch fail, retry with /refreshdb")
		}
	}

	if condcheck() {
		w.ctrl.CheckCount.Swap(0)
		dbmeta.CheckCount = 0
	}
	w.ctrl.SetLastRefreshtime() // updating refreshed time

	//Updating Metadata
	dbmeta.LoginLimit = w.ctrl.LoginLimit
	dbmeta.Maxconfigcount = w.ctrl.Maxconfigcount
	dbmeta.VerifiedUserCount = predata.verifiedusercount
	dbmeta.CommonQuota = MainCommonUserQuota
	dbmeta.TotalUpdates += w.ctrl.UpdateCounter.Swap(0)
	
	w.ctrl.Overview.Mu.Lock()
	w.ctrl.Overview.UpdateTime = time.Now()
	dbmeta.TotalConfigCount = w.ctrl.Overview.TotalConfCount
	w.ctrl.Overview.TotalUpdates = dbmeta.TotalUpdates
	w.ctrl.Overview.Mu.Unlock()

	if condcheck() {
		w.db.CreateOverviewLog(w.ctrl.Overview)
	}

	if docount {
		w.ctrl.CheckCount.Add(1)
		dbmeta.CheckCount = dbmeta.CheckCount + 1
	}
	w.activeUser = predata.verifiedusercount-predata.unUsedUser
	w.lastRefreshActiveUser = predata.verifiedusercount-predata.unUsedUser

	if err = w.db.Save(dbmeta).Error; err != nil {
		time.Sleep(100 * time.Millisecond)
		return w.db.Save(dbmeta).Error

	}
	// it's safe to send backup here
	// because any other goroutine can't access this db while this function is running
	w.logger.Info("total time elapsed db refresh " + time.Since(st).String())
	w.sendDbBackup(!docount || forceReset)
	runtime.GC()
	return nil
}

func (w *Watchman) txsave(value interface{}, tx *gorm.DB, bufsender *controller.MsgBufSender, batch int) error {
	retry := 0
	var err error
	save:
	for {
		if retry > MaxRetry {
			tx.Rollback()
			return fmt.Errorf("all retry attempt failed batch [%d] lasterr %s", batch, err.Error())
		}
		retry++
		err = tx.Save(value).Error
		if err != nil {
			if sqliteErr, ok := err.(sqlite3.Error); ok {
				switch sqliteErr.Code {
				case sqlite3.ErrBusy, sqlite3.ErrFull, sqlite3.ErrIoErr, sqlite3.ErrLocked:
					time.Sleep(200 * time.Millisecond)
					bufsender.Send("Database err while saving: " +err.Error(), w.ctrl.SudoAdmin)
				default:
					bufsender.Send("Unknown DB Error " + err.Error(), w.ctrl.SudoAdmin)
					return err
				}
				continue save
			}
		}
		break save
	}
	return nil
}

func (w *Watchman) PreprosessDb(refreshcontext context.Context, bufsender *controller.MsgBufSender) (*preprosessd, error) {
	var (
		preData = &preprosessd{}
		activeConfCount int64
		users []db.User
	)
	tosaveuser := make([]*db.User, 0, C.Dbbatchsize)
	err := w.db.Model(&db.User{}).FindInBatches(&users, C.Dbbatchsize, func(tx *gorm.DB, batch int) error {
		// Retrieve the current batch of records
		for i := range users {
			user := &users[i]
			if refreshcontext.Err() != nil {
				w.logger.Warn("Force stopping DB updating, Db update stops from record " + user.Name)
				return fmt.Errorf("context cancled db refresh stops from record id %v, err %v ", user.TgID, refreshcontext.Err())
			}
			if user.IsCapped {
				if user.Iscaptimeover(int(user.CapDays)) {
					user.IsCapped = false
					user.CappedQuota = 0
					tosaveuser = append(tosaveuser, user)
					bufsender.Send("you're captime is over, you're no longer capped if you want to set a cap again use /setcap", user.TgID)
				} else if user.Verified() {
					preData.cappeduser++
					preData.unUsedUser++
					preData.captotal += user.CappedQuota
					if user.GiftQuota > 0 {
						if user.CappedQuota > (user.CalculatedQuota-user.GiftQuota) {
							preData.savings = (user.CalculatedQuota - user.CappedQuota)
						} else {
							preData.savings = user.GiftQuota
						}
					}
				}
			}
			// for overview
			if user.Verified() {
				preData.verifiedusercount++
			}
			if user.IsMonthLimited {
				preData.monthlimiteduser++
			}
			if user.IsDistributedUser {
				preData.distributeduser++
			}
			if user.Restricted {
				preData.restricted++
			}
			if user.Templimited {
				preData.templimiteduser++
			}
			
			if user.Verified() && (!user.CanUse() || user.ConfigCount == 0) {
				if user.IsCapped {
					preData.unUsedUser--
					preData.captotal -= user.CappedQuota
				}
				if user.GiftQuota > 0 {
					preData.UsedByLimitedUsers -= user.GiftQuota // also adding because they can't use what the recive as gift
				}
				preData.unUsedUser++
				preData.UsedByLimitedUsers += user.MonthUsage
			} else if !user.Verified() {
				preData.UsedByLimitedUsers += user.MonthUsage
				if user.GiftQuota > 0 {
					preData.UsedByLimitedUsers -= user.GiftQuota // also adding because they can't use what the recive as gift
				}
			} else {
				activeConfCount += int64(user.ConfigCount)
				if !user.IsCapped {
					preData.totaladdtional += user.AdditionalQuota
				}
			}
			preData.configCount += int64(user.ConfigCount)
			
			//preData.savings += user.SavedQuota
		}

		if len(tosaveuser) > 0 {
			err := tx.Save(&tosaveuser).Error
			if err != nil {
				w.logger.Error("db save failed while preprocess ", zap.Int("batch", batch), zap.Error(err))
				return err
			}
			tosaveuser = tosaveuser[:0]
		}
		return nil // Return nil to continue to the next batch
	},
	).Error
	tosaveuser = nil
	if err != nil {
		return preData, err
	}
	overview := w.ctrl.Overview
	var (
		month_usage = C.Bwidth(0)
		alltime = C.Bwidth(0)
	)

	if err := w.db.Model(&db.User{}).Select("COALESCE(SUM(all_time_usage), 0)").Scan(&alltime).Error; err != nil {
		overview.Mu.RLock()
		alltime = overview.AllTime
		overview.Mu.RUnlock()
		w.logger.Error("all_time_usage sum for over view failed: ", zap.Error(err))
	}
	if err := w.db.Model(&db.User{}).Select("COALESCE(SUM(month_usage), 0)").Scan(&month_usage).Error; err != nil {
		overview.Mu.RLock()
		alltime = overview.AllTime
		overview.Mu.RUnlock()
		w.logger.Error("all_time_usage sum for over view failed: ", zap.Error(err))
	}

	overview.Mu.Lock()
	overview.CappedUser = preData.cappeduser
	overview.TotalConfCount = preData.configCount
	overview.ActiveConfCount = activeConfCount
	overview.TempLimitedUser = preData.templimiteduser
	overview.DistributedUser = preData.distributeduser
	overview.VerifiedUserCount = preData.verifiedusercount
	overview.TotalUser = w.ctrl.Dbusercount.Load()
	overview.MonthTotal = month_usage
	overview.MonthLimitedUser = preData.monthlimiteduser
	overview.AllTime = alltime+month_usage
	overview.BandwidthAvailable = w.ctrl.BandwidthAvelable
	overview.BandwidthAddtional = preData.totaladdtional
	overview.Restricte = preData.restricted
	overview.CUser = preData.verifiedusercount + preData.cappeduser - preData.unUsedUser
	overview.QuotaForEach = C.Bwidth(w.ctrl.CommonQuota.Load())
	overview.LastRefresh = time.Now()
	overview.DaysToReset = ((w.ctrl.ResetCount - w.ctrl.CheckCount.Load()) * w.ctrl.RefreshRate) / 24
	overview.Mu.Unlock()

	return preData, nil
}
// DO not call outside refresh db
func (w *Watchman) sendDbBackup(force bool) {
	if w.ctrl.ForceDisableBackup {
		return
	}
	if w.ctrl.CheckCount.Load() % int32(w.ctrl.BackupCycle) != 0 && !force {
		return
	}
	w.ctrl.SendFile(w.db.DatabasePath(), "database.db", `latest database
	Time: `+ time.Now().String() + `
	`, w.ctrl.SudoAdmin)
	if force {
		w.ctrl.SendFile(w.db.UsageDatabasePath(), "usage.db", `latest usage database
		Time: `+ time.Now().String() + `
	`, w.ctrl.SudoAdmin)
	}
}