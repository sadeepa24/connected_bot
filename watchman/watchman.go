package watchman

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sagernet/sing-vmess/vless"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Watchman has accsess to everything

type Watchmanconfig struct {
	//Simlogpath string `json:"box_log"`
	Delbuffer  int   `json:"del_buffer"` //msg count to buffer before delete
}

type Watchman struct {
	ctx    context.Context
	db     *db.Database
	ctrl   *controller.Controller
	botapi botapi.BotAPI

	config *Watchmanconfig
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
	config *Watchmanconfig,
	logger *zap.Logger,
	mgstore *botapi.MessageStore,

) (*Watchman, error) {

	if config == nil {
		config = &Watchmanconfig{}
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
		//msgque: make(chan *botapi.Msgcommon, config.Msgbuf),
	}, nil
}

func (w *Watchman) Start() error {
	var err error
	if w.ctrl.Metaconfig.RefreshRate <= 0 {
		return errors.New("refresh rate cannot be lower than 0")
	}

	// if w.simplelog, err = simplelog.Newsimpllogger(w.ctx, w.config.Simlogpath); err != nil {
	// 	return err
	// }

	w.ticker = time.NewTicker(time.Duration(w.ctrl.Metaconfig.RefreshRate) * time.Hour)


	go func() {
		for _, out := range w.ctrl.Outbounds {
			t, err := w.ctrl.UrlTestOut(out.Tag)
			if err != nil {
				w.logger.Error("urltest error " + out.Tag + " err - " + err.Error())
				out.Latency.Swap(-1)
				continue
			}
			out.Latency.Swap(int32(t))
		}
	}()



	go w.startAutoupdater()
	//go w.startSboxLogger()


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
	w.logger.Info("Starting Watchman Adn AutoUpdater")
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
					t, err := w.ctrl.UrlTestOut(out.Tag)
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
					if w.ctrl.GetUserList(&userlist) != nil {
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

// func (w *Watchman) startSboxLogger() {
// 	flush := 0
// 	for val := range w.sboxlog {
// 		w.simplelog.Info(time.Now().String(),  " ", val.(string))
// 		if flush > 100 {
// 			w.simplelog.Sync()
// 			w.logger.Sync()
// 			flush = 0
// 		}
// 		flush++
// 	}
// }

type preprosessd struct {
	cappeduser        int64    //total user count who capped their bandwidth
	verifiedusercount int64
	monthlimiteduser  int64
	distributeduser   int64
	restricted 		  int64
	templimiteduser   int64

	captotal          C.Bwidth //total bandwidth capped
	totaladdtional    C.Bwidth
	savings           C.Bwidth
	userdbyTemplimitedUser C.Bwidth
	
	UsedByLimitedUsers C.Bwidth
	unUsedUser 		  int64 //to calculate mainquota

	configCount int64


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
	savings %v
	
	`,
		p.cappeduser,
		p.captotal,
		p.verifiedusercount,
		p.totaladdtional,
		p.monthlimiteduser,
		p.distributeduser,
		p.savings,
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

// send any uint16 value to stop this chan
func (w *Watchman) messageBufSend(recivechan chan any)  {
	tmpctx, cancel := context.WithTimeout(w.ctx, 15 * time.Minute) //maximum time to send all buffred message if message get 1s to send 900 messages can be send (worst case)
	defer cancel()
	for val := range recivechan {
		if tmpctx.Err() != nil {
			close(recivechan)
			return
		}
		if _, ok := val.(uint16); ok {
			close(recivechan)
			return
		}
		_,err := w.ctrl.SendMsgContext(tmpctx, val)
		if err != nil {
			if errors.Is(err, C.ErrClientRequestFail) {
				w.ctrl.Getmgque() <- val // buffer again to send later
			}
		}
	}
}

func (w *Watchman) sendUsingBufChan(send chan any, msg string, id int64) {
	send <- &botapi.Msgcommon{
		Infocontext: &botapi.Infocontext{
			ChatId: id,
			User_id: id,
		},
		Text: msg,
	}
}
// refresh member verificity
// refresh usage to database
// if docount true CheckkCount will increase by one
// if forceReset true All Usage Will Resets And Checkcount will be zero
func (w *Watchman) RefreshDb(refreshcontext context.Context, docount bool, forceReset bool) error {
	w.ctrl.WatchmanLock() //locking for dbrefresh, all new upx will be paused
	defer w.ctrl.WatchmanUnlock()

	w.ctrl.WaitCriticalop()      // waiting for all critical opration done
	w.ctrl.CancleUpdateContexs() // cancling all non critical ongoing upx

	

	var (
		checkcount = w.ctrl.CheckCount.Load()
		condcheck  = func() bool {
			return ((checkcount == w.ctrl.ResetCount) && docount) || forceReset
		}
		err error
	)

	w.logger.Info("Batch Updating Database Count ", zap.Int32("checkcount", checkcount))

	var chanmax = w.ctrl.VerifiedUserCount.Load()

	if chanmax < 20 {
		chanmax = 40
	}


	msgchan := make(chan any, chanmax)
	go w.messageBufSend(msgchan) // this will recive all messages to user using the above chan, so this function does not wait for request response, everything releted to req, res will be handled by this, chan will close automatically after 15 minitues(max time)


	predata, err := w.PreprosessDb(refreshcontext, msgchan)
	if err != nil {
		w.ctrl.DirectMg("Predata prosseing error Please Make Manual Refresh := " + err.Error(), w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
		return errors.Join(errors.New("predata prosseing failed"), err)
	}


	w.ctrl.VerifiedUserCount.Swap(int32(predata.verifiedusercount))
	MainCommonUserQuota := w.ctrl.BandwidthAvelable // Newcalculated main quota for each user

	if predata.verifiedusercount-predata.unUsedUser > 0 && (predata.cappeduser != predata.verifiedusercount) {
		// Calculate the quota for each user based on various parameters
		// Parameters include: verified user count, capped user, month-limited user, gifted user, usage overridden user
		// Additional quota from users
		// Overused users can't use their whole quota due to usage rollback from last month
		// MainCommonUserQuota = ((w.ctrl.BandwidthAvelable + predata.savings) - (predata.captotal + predata.usedbyrestricted + predata.totaladdtional + predata.usedbydisuser)) / C.Bwidth(predata.verifiedusercount-(predata.cappeduser+predata.distributeduser+predata.monthlimiteduser+predata.restricted))
		// overused user can't just use their whole quota (due adding usage rollback from lastmonth,  this month initial usage = lastmonth excess usage - last month his quota  ),  so it's like increase of bandwidth but finnaly it's same
		//MainCommonUserQuota = ((w.ctrl.BandwidthAvelable + predata.savings) - (predata.captotal + predata.usedbyrestricted + predata.totaladdtional + predata.usedbydisuser)) / C.Bwidth(predata.verifiedusercount-(predata.cappeduser+predata.distributeduser+predata.monthlimiteduser+predata.restricted))
		
		MainCommonUserQuota = ((w.ctrl.BandwidthAvelable + predata.savings) - (predata.captotal + predata.totaladdtional + predata.userdbyTemplimitedUser)) / C.Bwidth(predata.verifiedusercount-predata.unUsedUser)

	}

	// this value used to calculate the old ratio between config quota and old maincommonquota
	// new config quota will calculate based on this ratio
	oldCommonQuota := w.ctrl.CommonQuota.Swap(MainCommonUserQuota.Int64())
	w.ctrl.Overview.Mu.Lock()
	w.ctrl.Overview.QuotaForEach = MainCommonUserQuota
	w.ctrl.Overview.Mu.Unlock()

	//w.ctrl.Metadata.Lock()
	// oldQuota := C.Bwidth(w.ctrl.UserQuota.Swap(newQuota.Int64())) // Old quota which is used to calculate userquota lasttime
	//w.ctrl.Metadata.Unlock()

	var listUser []db.User
	w.db.Model(&db.User{}).FindInBatches(&listUser, C.Dbbatchsize, func(tx *gorm.DB, batch int) error {
		// Retrieve the current batch of records
		w.logger.Debug("fetching users batch", zap.Int("batch", batch), zap.Int("Count", len(listUser)))
		for _, user := range listUser {
			if refreshcontext.Err() != nil {
				w.ctrl.WatchmanUnlock()
				w.logger.Warn("🔴🔴🔴 Force stopping DB updating, Db update stops middle of db update. Db may malformed 🔴🔴🔴" + user.Name)
				w.sendUsingBufChan(msgchan, "🔴🔴🔴 force stopped when db refresh, you may need to start bot with last backup. see logs for more info", w.ctrl.SudoAdmin )
				return fmt.Errorf("context cancled db refresh stops from record id %v, err %v ", user.TgID, refreshcontext.Err())
			}
			tx.Model(&db.Config{}).Where("user_id = ?", user.TgID).Find(&user.Configs)

			//recalcuted the gift quota according to new ratio
			if oldCommonQuota > 0 && user.GiftQuota != 0 {

				k := C.Bwidth(oldCommonQuota) / C.Bwidth(user.GiftQuota)
				user.GiftQuota = MainCommonUserQuota / k

			}

			//calculating gift quota accrording to newst ratio
			if user.GiftQuota != 0 {
				allgifts := []db.Gift{}
				tx.Model(&db.Gift{}).Where("recive_valid = ? OR send_valid = ?", true, true).Where("sender = ? OR reciver = ?", user.TgID, user.TgID).Find(&allgifts)

				for _, gift := range allgifts {
					if gift.Isgifttimeover() {


						presentGift := ((MainCommonUserQuota / gift.ComQuota) * gift.Bandwidth)

						switch user.TgID {
						case gift.Sender:

							user.GiftQuota = user.GiftQuota + C.Bwidth(presentGift)
							gift.SendValid = false
						case gift.Reciver:
							user.GiftQuota = user.GiftQuota - C.Bwidth(presentGift)
							gift.ReciveValid = false
						}
						// tx.Save(&user)
						tx.Save(&gift)

					}
				}

			}

			// storing old quota for calculating
			oldQuota := user.CalculatedQuota
			user.CalculatedQuota = MainCommonUserQuota + user.GiftQuota
			userVerifycity := user.IsInChannel && user.IsInGroup
			if user.IsCapped && user.CappedQuota > user.CalculatedQuota {
				w.sendUsingBufChan(msgchan, "you'r are no longer capped user, due our main quota is lower than you'r capped quota", user.TgID)
				user.IsCapped = false
				user.CappedQuota = 0
			} else if user.IsCapped {
				user.CalculatedQuota = user.CappedQuota
			}
		
			user.ConfigCount = int16(len(user.Configs))

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
					w.sendUsingBufChan(msgchan, "you have config that don't have any quota please remove it or increase quota", user.TgID)
					user.Configs[i].Active = false
					continue
				}

				usedquota += newConfigQuota
				dbin, err := w.ctrl.GetdbInbound(int(user.Configs[i].InboundID))
				if err != nil {
					_, dbin = w.ctrl.DefaultInboud()
				}
				dbout, err := w.ctrl.GetdbOutbound(int(user.Configs[i].OutboundID))
				if err != nil {
					_, dbout = w.ctrl.Defaultoutboud()
				}
				user.Configs[i].Quota = newConfigQuota

				var (
					forceremove bool
					//justActivated bool
				)

				if (newConfigQuota - user.Configs[i].Usage > 0) && userVerifycity && !user.IsDistributedUser && !user.IsMonthLimited && !user.Restricted && !user.Templimited {
					status, err := w.ctrl.AddResetUserSbox(&sbox.Userconfig{
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
						switch {
						case errors.Is(err, vless.ErrInboundNotFound):
							w.sendUsingBufChan(msgchan, C.GetMsg(C.MsgNoInbound), user.TgID)
						case errors.Is(err, C.ErrTypeMissmatch), errors.Is(err, vless.ErrInvalidInbound):
							w.sendUsingBufChan(msgchan, C.GetMsg(C.MsgwtchErrtypemiss), user.TgID)
						case errors.Is(err, vless.ErrVlessService):
							w.sendUsingBufChan(msgchan, C.GetMsg(C.MsgwtchErruseradd), user.TgID )
						}
						status = sbox.Sboxstatus{
							Download: 0,
							Upload: 0,
						}
						err = nil

					}
					user.Configs[i].UpdateUsages(status)
					user.MonthUsage += status.FullUsage()
					if status.FullUsage() > 0 {
						err = tx.Create(&db.UsageHistory{
							Usage:    status.Download + status.Upload,
							Download: status.Download,
							Upload:   status.Upload,
							Date:     time.Now(),
							UserID:   user.TgID,
							ConfigID: user.Configs[i].Id,
							Name: user.Name,
						}).Error
						if err != nil {
							w.logger.Error("error when creating usage history", zap.Error(err))
						}
					}
					if user.Configs[i].Usage >= user.Configs[i].Quota  {
						forceremove = true
					} else {
						if !user.Configs[i].Active {
							w.sendUsingBufChan(msgchan, "Good News Configuration "+ user.Configs[i].Name+" Online Again Due to Bandiwdth Change 🔄", user.TgID)
						}
						user.Configs[i].Active = true
					}

				}
				
				
				if user.Configs[i].Active  && (forceremove || newConfigQuota - user.Configs[i].Usage <= 0) {
					if (user.Configs[i].Quota - user.Configs[i].Usage) <= 0 {
						w.sendUsingBufChan(msgchan, "⚠️ Your configuration "+user.Configs[i].Name+" has exceeded its usage limit. The config will not function until it is renewed. 🔄", user.TgID)
					}
					status, err := w.ctrl.RemoveUserSbox(&sbox.Userconfig{
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
						Type:        user.Configs[i].Type,
						InboundId:   dbin.ID,
						OutboundID:  dbout.ID,
						LoginLimit:  int32(user.Configs[i].LoginLimit),
						TgId: user.TgID,
					})
					if err == nil && status.Download + status.Upload > 0 && !forceremove {
						
						if status.FullUsage() > 0 {
							user.Configs[i].UpdateUsages(status)
							user.MonthUsage += status.FullUsage()
						}

						tx.Create(&db.UsageHistory{
							Usage:    status.Download + status.Upload,
							Download: status.Download,
							Upload:   status.Upload,
							Date:     time.Now(),
							UserID:   user.TgID,
							ConfigID: user.Configs[i].Id,
						})
					}
					user.Configs[i].Active = false

				}
				//end new
				// if err = tx.Save(&user.Configs[i]).Error; err != nil {
				// 	time.Sleep(800 * time.Millisecond)
				// 	tx.Save(&user.Configs[i])
				// }
			}
			user.UsedQuota = usedquota
			
			

			if oldUsage == user.MonthUsage && user.Verified() && docount && user.MonthUsage <= user.CalculatedQuota { //which means user did n't use the config for last refresh cycle
				user.EmptyCycle++
				if user.EmptyCycle == user.WarnRatio && user.WarnRatio != 0 {
					user.Templimited = true	// hecan't use the service until he remove this war manually
					user.WarnRatio = user.WarnRatio/2
					w.sendUsingBufChan(msgchan, C.GetMsg(C.MsgTemplimit), user.TgID)
					for i := range user.Configs {
						dbin, err := w.ctrl.GetdbInbound(int(user.Configs[i].InboundID))
						if err != nil {
							_, dbin = w.ctrl.DefaultInboud()
						}
						dbout, err := w.ctrl.GetdbOutbound(int(user.Configs[i].OutboundID))
						if err != nil {
							_, dbout = w.ctrl.Defaultoutboud()
						}
						w.ctrl.RemoveUserSbox(&sbox.Userconfig{
							Vlessgroup: &sbox.Vlessgroup{
								UUID: user.Configs[i].GetUUID(),
							},
							UsercheckId: int(user.CheckID),
							Name:        user.Name,
							Inboundtag:  dbin.Tag, //TODO: fetch this correctly
							Outboundtag: dbout.Tag,
							Usage:       user.Configs[i].Usage,
							Quota:       0,
							DbID:        user.Configs[i].Id,
							Type:        user.Configs[i].Type,
							InboundId:   dbin.ID,
							OutboundID:  dbout.ID,
							LoginLimit:  int32(user.Configs[i].LoginLimit),
							TgId: user.TgID,
						})
					}

					if user.WarnRatio == 0 {
						w.sendUsingBufChan(msgchan, C.GetMsg(C.MsgTempOver), user.TgID)
					}
				}

			} else if docount && user.MonthUsage <= user.CalculatedQuota  {
				user.EmptyCycle = 0
			}
			if user.UsedQuota > user.CalculatedQuota {
				w.logger.Warn("violation, usedquota > calculatedquota detected from " + user.String())
				w.sendUsingBufChan(msgchan, "We have detetcted you have bigger quota than we allocated to fix this we overide you'r config's quota", user.TgID)
				user.UsedQuota = user.CalculatedQuota
				quotaforeach := user.CalculatedQuota / C.Bwidth(user.ConfigCount)
				for i := range user.Configs {
					user.Configs[i].Quota = quotaforeach
				}
				
			}
			
			if condcheck() {			
				if user.IsDistributedUser && !user.Restricted {
					w.sendUsingBufChan(msgchan, C.GetMsg(C.MsgDistributeOver), user.TgID)
				}
				if (user.IsMonthLimited || user.WarnRatio == 0 ) && !user.Restricted {
					w.sendUsingBufChan(msgchan, "You'r Limitation is over", user.TgID)
				}
		
				user.AddPoint(10)
				user.SavedQuota = 0

				var (
					configusageReset bool
				)
				if user.MonthUsage > user.CalculatedQuota+user.AdditionalQuota {

					//TODO: add template here
					msgchan <- &botapi.Msgcommon{
						Infocontext: &botapi.Infocontext{
							ChatId: user.TgID,
						},
						Text: C.GetMsg(C.MsgwtchUsagereset),
					}
					useraccutalused := user.MonthUsage

					user.AlltimeUsage += user.CalculatedQuota
					user.MonthUsage = user.MonthUsage - user.CalculatedQuota
					user.SavedQuota = user.MonthUsage // because of user can't use this, it's a saving for this month

					for i, conf := range user.Configs {
						// recalculate excess usage for each configs
						// ratio between useractualused and conf usage should be equal to ratio between new user.Monthusage(excess usage from last month) and new conf usage 
						// using this we can calculate conf excess usage
						// newconfusage = user.Monthusage(new) * (oldconfusage/useraccutalused)

						if conf.Quota == 0 {
							continue
						}
						user.Configs[i].Usage = user.MonthUsage * (conf.Usage / useraccutalused)
					}
					configusageReset = true
					user.IsMonthLimited = false

				} else if user.MonthUsage < ((user.CalculatedQuota*3)/4) && !user.IsMonthLimited && !user.IsDistributedUser && !user.Restricted && !(user.WarnRatio != 0) { 
					//check whether user used 75% from his quota if not user will limited next 30 days
					msgchan <- &botapi.Msgcommon{
						Infocontext: &botapi.Infocontext{
							ChatId: user.TgID,
						},
						Text: C.GetMsg(C.MsgQuotanotUsed),
					}
					user.IsMonthLimited = true
					user.AlltimeUsage += user.MonthUsage
					user.MonthUsage = 0
				} else  {

					//TODO: add template here
					msgchan <- &botapi.Msgcommon{
						Infocontext: &botapi.Infocontext{
							ChatId: user.TgID,
						},
						Text: C.GetMsg(C.MsgresetUsage),
					}
					user.IsMonthLimited = false
					user.AlltimeUsage += user.MonthUsage
					user.MonthUsage = 0
				}
				if !configusageReset {
					for i := range user.Configs {
						user.Configs[i].Usage = 0
						user.Configs[i].Upload = 0
						user.Configs[i].Download = 0
					}
				}
				user.WarnRatio = w.ctrl.GetWarnRate()
				user.Templimited = false
				user.IsDistributedUser = false
			}
			
			if len(user.Configs) > 0 {
				if err = tx.Save(&user.Configs).Error; err != nil {
					time.Sleep(800 * time.Millisecond)
					tx.Save(&user.Configs)
				}
			}

			if err = tx.Save(user).Error; err != nil {
				time.Sleep(800 * time.Millisecond)
				tx.Save(user)
			}

		}
		w.logger.Info("batch prosess done", zap.Int("batch", batch))
		tx.Commit()
		return nil // Return nil to continue to the next batch
	},
	)

	//updating metadata
	var dbmeta = &db.Metadata{ //only one order in db for metadata
		Id: 1,
	}

	if err = w.db.Model(&db.Metadata{}).First(dbmeta).Error; err != nil {
		time.Sleep(800 * time.Millisecond)
		w.db.Model(&db.Metadata{}).First(dbmeta)
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
	w.ctrl.Overview.TotalUpdates = dbmeta.TotalUpdates
	w.ctrl.Overview.Mu.Unlock()

	if docount {
		w.ctrl.CheckCount.Add(1)
		dbmeta.CheckCount = dbmeta.CheckCount + 1
	}

	if err = w.db.Save(dbmeta).Error; err != nil {
		time.Sleep(800 * time.Millisecond)
		w.db.Save(dbmeta)

	}

	w.activeUser = predata.verifiedusercount-predata.unUsedUser
	w.lastRefreshActiveUser = predata.verifiedusercount-predata.unUsedUser
	// it's safe to send backup here
	// because any other goroutine can't access this db while this function is running
	w.sendDbBackup()
	msgchan <- uint16(1) // to tell buffring is over
	runtime.GC()
	// if w.CheckClose() != nil {
	// 	w.close <- struct{}{}
	// }

	return nil
}



func (w *Watchman) PreprosessDb(refreshcontext context.Context, msgchan chan any) (preprosessd, error) {

	/*
		var (
			err error
			isinchannel bool
			is_ingroup bool
		)
		if _, isinchannel, err = w.botapi.GetchatmemberCtx(context.Background(), user.TgID, w.ctrl.ChannelId); err != nil {
			isinchannel = user.IsInChannel
		}

		if _, user.IsInGroup, err = w.botapi.GetchatmemberCtx(context.Background(), user.TgID, w.ctrl.GroupID); err != nil {
			is_ingroup = user.IsInGroup
		}
		user.IsInChannel = isinchannel
		user.IsInGroup = is_ingroup
	*/

	var (
		preData = preprosessd{}
	)
	
	// var checkcount = w.ctrl.CheckCount.Load()
	// var condcheck = func() bool {
	// 	return checkcount == w.ctrl.ResetCount
	// }
	var activeConfCount int64
	var users []db.User
	w.db.Model(&db.User{}).FindInBatches(&users, C.Dbbatchsize, func(tx *gorm.DB, batch int) error {
		// Retrieve the current batch of records
		for _, user := range users {
			if refreshcontext.Err() != nil {
				w.ctrl.WatchmanUnlock()
				w.logger.Warn("Force stopping DB updating, Db update stops from record " + user.Name)
				return fmt.Errorf("context cancled db refresh stops from record id %v, err %v ", user.TgID, refreshcontext.Err())
			}
			// if err := w.db.Model(&db.Config{}).Where("user_id = ?", user.TgID).Find(&user.Configs).Error; err != nil {
			// 	continue
			// }
			if user.IsCapped {
				if user.Iscaptimeover(int(user.CapDays)) {
					user.IsCapped = false
					user.CappedQuota = 0
					w.sendUsingBufChan(msgchan, "you're captime is over, you're no longer capped if you want to set a cap again use /setcap", user.TgID)
					//tx.Model(&db.User{}).First(&user).Update("is_capped", false)
					tx.Save(&user)
				} else {
					preData.cappeduser++
					preData.unUsedUser++
					preData.captotal += user.CappedQuota
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
			

			if user.Verified() && (user.Restricted || user.IsDistributedUser || user.IsMonthLimited || user.Templimited || user.ConfigCount == 0) {
				if user.IsCapped {
					preData.unUsedUser--
					if user.GiftQuota > 0 && user.CappedQuota < C.Bwidth(w.ctrl.CommonQuota.Load()) {
						preData.UsedByLimitedUsers -= user.GiftQuota  //when user does not use gift quota 
					}
				} else if user.GiftQuota > 0 {
					preData.UsedByLimitedUsers -= user.GiftQuota // also adding because they can't use what the recive as gift
				}
				preData.unUsedUser++
				preData.UsedByLimitedUsers += user.MonthUsage
			} else {
				activeConfCount += int64(user.ConfigCount)
			}
			preData.configCount += int64(user.ConfigCount)
			preData.totaladdtional += user.AdditionalQuota
			preData.savings += user.SavedQuota

			

		}

		return nil // Return nil to continue to the next batch
	},
	)
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
	overview.AllTime = alltime+month_usage
	overview.BandwidthAvailable = w.ctrl.BandwidthAvelable
	overview.Restricte = preData.restricted
	overview.CUser = preData.verifiedusercount - preData.unUsedUser
	overview.QuotaForEach = C.Bwidth(w.ctrl.CommonQuota.Load())
	overview.LastRefresh = time.Now()
	overview.Mu.Unlock()

	return preData, nil
}
// DO not call outside refresh db
func (w *Watchman) sendDbBackup() {
	dbraw, err := os.Open(w.db.DatabasePath())
	if err != nil {
		w.logger.Error("Db Backup Send Failed: errored when reading database for backup create", zap.Error(err))
		return
	}
	defer dbraw.Close()

	req, err :=  botapi.CreateMultiPartReq(w.ctx, "POST", w.botapi.CreateFullUrl("sendDocument"), map[string]string{
		"chat_id": strconv.Itoa(int(w.ctrl.SudoAdmin)),
		"caption": `latest database after last refresh
		Time: `+ time.Now().String() + `
		
		`,
	}, map[string]botapi.Filepart{
		"document": {
			Name: "database.db",
			Reader: dbraw,
		},
	})
	if err != nil {
		w.logger.Error("Db Backup Send Failed: request making failed" +  err.Error())
		return
	}
	
	apires, err := w.botapi.SendRawReq(req)
	
	if err != nil {
		w.logger.Error("Db Backup Send Failed: request send failed when uploading backup database ", zap.Error(err))
		return
	}
	if !apires.Ok {
		w.logger.Error("Db Backup Send Failed: Bad Response From Telegram: " + apires.Description)
	}


}