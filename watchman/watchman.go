package watchman

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sagernet/sing-vmess/vless"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Watchman has accsess to everything

type Watchmanconfig struct {
	//Simlogpath string `json:"box_log"`
	Delbuffer  int    //msg count to buffer before delete
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

	lastUserCount int32 //User count on db when running function RefreshDb last time

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
		return errors.New("refresh rate cannot be loower than 0")
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
		err = w.RefreshDb(startrefresh, false)
		if err != nil {
			w.logger.Fatal("fatal error db start refresh " + err.Error())
		}
		refreshdone <- struct{}{}

	}(startrefresh)

	select {
	case <-refreshdone:
		cancle()
		break
	case <-startrefresh.Done():
		cancle()
		return errors.New("intlize db refresh failed context timeout")
	}
	return nil
}

func (w *Watchman) Close() error {

	w.close <- struct{}{} //close chan is not a buffred chan so this opration wait for w.close recive
	w.RefreshDb(w.ctx, false)
	<-w.close
	w.logger.Debug("watchman closing done")
	return nil
}

func (w *Watchman) startAutoupdater() {
	w.logger.Info("started watch mand Autoupdater")
	w.lastUserCount = w.ctrl.Dbusercount.Load()
update:
	for {
	sele:
		select {

		case <-w.ctx.Done():

			w.logger.Warn("context Cancled Autoupdater closing")
			w.logger.Warn("Force Closing DB")
			break update

		case <-w.close:
			
			w.logger.Sync()
			w.logger.Info("Closing Auto Updater close call recived")
			w.close <- struct{}{}
			break update

		case tick := <-w.ticker.C:
			w.logger.Info("db refresh started", zap.String("tick", tick.String()), zap.Int32("count", w.ctrl.CheckCount.Load()))

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

			err := w.RefreshDb(refreshctx, true)
			cancle()

			if err != nil {
				w.logger.Error("Db Refresh Failed Due to: ", zap.Error(err))
				continue
			}

			w.logger.Info("db refresh done")

		case mg := <-w.ctrl.Getmgque():

			currentcount := w.ctrl.Dbusercount.Load()

			switch unwrapedmg := mg.(type) {
			case controller.RefreshSignal:
				w.ctrl.DirectMg("force refresh added", w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
				if w.ctrl.CheckLock() {
					continue	
				}
				refreshctx, cancle := context.WithCancel(w.ctx)
				err := w.RefreshDb(refreshctx, false)
				cancle()
				if err != nil {
					w.ctrl.DirectMg("refresh failed", w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
					w.logger.Error("Force Db Refresh Failed Due to: ", zap.Error(err))
				} else {
					w.ctrl.DirectMg("refresh done", w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
				}
			case controller.BroadcastSig:
				//TODO: create broadcaster
			case controller.UserCount:
				if w.ctrl.CheckLock() {
					continue	
				}
				if float32(w.lastUserCount)+(float32(w.lastUserCount)/4)*3 < float32(currentcount) {
					refreshctx, cancle := context.WithCancel(w.ctx)

					
					err := w.RefreshDb(refreshctx, false)
					cancle()

					if err != nil {
						w.logger.Error("Usercount Db Refresh Failed Due to: ", zap.Error(err))
						continue
					}
					w.logger.Info("db refresh done")
				}

				continue
			case *botapi.Msgcommon:
				var (
					repmg *tgbotapi.Message
					err   error
				)
				if unwrapedmg.Endpoint == "" {
					unwrapedmg.Endpoint = C.ApiMethodSendMG
				}

				if repmg, err = w.botapi.SendContext(w.ctx, unwrapedmg); err != nil {
					if errors.Is(err, C.ErrClientRequestFail) {
						w.ctrl.Getmgque() <- mg // buffer again
					}
					break sele
				}

				if unwrapedmg.ChatId == w.ctrl.GroupID {
					w.Delmg(repmg.MessageID)
				}
			case botapi.UpMessage:

				texttmpl, err := w.mgstore.GetMessage(unwrapedmg.TemplateName, unwrapedmg.Lang, unwrapedmg.Template)
				if err != nil {
					w.logger.Error("message from msgque chan failed to send", zap.Error(err))
					continue
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
				var repmg *tgbotapi.Message

				if repmg, err = w.botapi.SendContext(w.ctx, &sendmg); err != nil {
					if errors.Is(err, C.ErrClientRequestFail) {
						w.ctrl.Getmgque() <- mg // buffer again
					}
					break sele
				}

				if sendmg.ChatId == w.ctrl.GroupID {
					w.Delmg(repmg.MessageID)
				}

			}

		/*
		case log := <-w.sboxlog:
			w.simplelog.Info("sboxlog ", log.(string))

			if len(w.sboxlog)+200 > cap(w.sboxlog) {
				for len(w.sboxlog) > 0 {
					lg := <-w.sboxlog
					w.simplelog.Info("sboxlog ", lg.(string))
				}
			}

			w.simplelog.Sync()
			w.logger.Sync()
		// case delmg := <- w.DeleteQue:
		*/

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

// func (w *Watchman) Testloop() {
// 	for {
// 		w.sboxlog <- "text buffer test test test test test"
// 	}
// }

type preprosessd struct {
	cappeduser        int64    //total user count who capped their bandwidth
	captotal          C.Bwidth //total bandwidth capped
	verifiedusercount int64
	totaladdtional    C.Bwidth
	monthlimiteduser  int64
	distributeduser   int64
	usedbydisuser     C.Bwidth
	usedbyrestricted  C.Bwidth
	savings           C.Bwidth
	restricted 		  int64


}

// TODO: remove after testings
func (p *preprosessd) String() (s string) {
	s = fmt.Sprintf(`
	cappeduser %v
	captotal %v 
	verifiedusercount %v
	totaladdtional %v
	monthlimiteduser %v
	distributeduser %v
	usedbydisuser %v
	savings %v
	
	`,
		p.cappeduser,
		p.captotal,
		p.verifiedusercount,
		p.totaladdtional,
		p.monthlimiteduser,
		p.distributeduser,
		p.usedbydisuser,
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

// refresh member verificity
// refresh usage to database
// if docount true CheckkCount will increase by one
func (w *Watchman) RefreshDb(refreshcontext context.Context, docount bool) error {
	w.ctrl.WatchmanLock() //locking for dbrefresh, all new upx will be paused
	defer w.ctrl.WatchmanUnlock()

	w.ctrl.WaitCriticalop()      //waiting for all critical opration done
	w.ctrl.CancleUpdateContexs() // cancling all non critical ongoing upx

	var (
		checkcount = w.ctrl.CheckCount.Load()
		condcheck  = func() bool {
			return (checkcount == w.ctrl.ResetCount) && docount
		}
		err error
	)

	w.logger.Info("Batch Updating Database Count ", zap.Int32("checkcount", checkcount))

	predata, err := w.PreprosessDb(refreshcontext)
	if err != nil {
		w.ctrl.DirectMg("Predata prosseing error := " + err.Error(), w.ctrl.SudoAdmin, w.ctrl.SudoAdmin)
		return errors.Join(errors.New("predata prosseing failed"), err)
	}

	w.ctrl.VerifiedUserCount.Swap(int32(predata.verifiedusercount))
	MainCommonUserQuota := w.ctrl.BandwidthAvelable // Newcalculated main quota for each user

	if predata.verifiedusercount-(predata.cappeduser+predata.distributeduser+predata.monthlimiteduser + predata.restricted) > 0 && (predata.cappeduser != predata.verifiedusercount) {

		// yes i Know below line is stupid but here is how it works
		//it calculate quota for each user accrding to predata values
		// many parameters are responsible for calculating the value
		//here all parameter
		// verified user count, capped user, monthlimited user, gifted user, usage overided user
		// addtional quota from users

		MainCommonUserQuota = ((w.ctrl.BandwidthAvelable + predata.savings) - (predata.captotal + predata.usedbyrestricted + predata.totaladdtional + predata.usedbydisuser)) / C.Bwidth(predata.verifiedusercount-(predata.cappeduser+predata.distributeduser+predata.monthlimiteduser+predata.restricted))

	}

	// this value used to calculate the old ratio between config quota and old maincommonquota
	// new config quota will calculate based on this ratio
	// don think much about english
	oldCommonQuota := w.ctrl.CommonQuota.Swap(MainCommonUserQuota.Int64())
	w.ctrl.Overview.Mu.Lock()
	w.ctrl.Overview.QuotaForEach = MainCommonUserQuota
	w.ctrl.Overview.Mu.Unlock()

	//w.ctrl.Metadata.Lock()
	// oldQuota := C.Bwidth(w.ctrl.UserQuota.Swap(newQuota.Int64())) // Old quota which is used to calculate userquota lasttime
	//w.ctrl.Metadata.Unlock()

	w.db.Model(&db.User{}).FindInBatches(&[]db.User{}, C.Dbbatchsize, func(tx *gorm.DB, batch int) error {
		// Retrieve the current batch of records
		var users []db.User
		tx.Find(&users) // Fetch the current batch of records from tx
		w.logger.Debug(" start to prosess a batch")

		for _, user := range users {
			if refreshcontext.Err() != nil {
				w.ctrl.WatchmanUnlock()
				w.logger.Warn("Force stopping DB updating, Db update stops from record " + user.Name)
				return fmt.Errorf("context cancled db refresh stops from record id %v, err %v ", user.TgID, refreshcontext.Err())
			}

			if user.IsAdmin {
				continue
			}
			tx.Model(&db.Config{}).Where("user_id = ?", user.TgID).Find(&user.Configs)

			//recalcuted the gift quota according to new ratio
			if oldCommonQuota > 0 && user.GiftQuota != 0 {

				k := float64(oldCommonQuota) / float64(user.GiftQuota)
				user.GiftQuota = C.Bwidth(MainCommonUserQuota.Float64() / k)

			}

			//canculating gift quota accrording to newst ratio
			if user.GiftQuota != 0 {
				allgifts := []db.Gift{}
				tx.Model(&db.Gift{}).Where("recive_valid = ? OR send_valid = ?", true, true).Where("sender = ? OR reciver = ?", user.TgID, user.TgID).Find(&allgifts)

				for _, gift := range allgifts {
					if gift.Isgifttimeover() {

						presentGift := (MainCommonUserQuota.Float64() / gift.ComQuota.Float64() * gift.Bandwidth.Float64())

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

			if user.IsCapped && !(user.CappedQuota >= user.CalculatedQuota) {
				user.CalculatedQuota = user.CappedQuota

			}

			user.ConfigCount = int16(len(user.Configs))


			var usedquota C.Bwidth

			//configs:
			for i := range user.Configs {

				k := oldQuota.Float64() / user.Configs[i].Quota.Float64()      // findig ratio between oldquota and old configs quota
				newConfigQuota := C.Bwidth(user.CalculatedQuota.Float64() / k) // subpressing quota according to ratio, k is the constant
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

				//new


				if (newConfigQuota - user.Configs[i].Usage > 0) && userVerifycity && !user.IsDistributedUser && !user.IsMonthLimited && !user.Restricted {
					status, err := w.ctrl.AddResetUserSbox(&sbox.Userconfig{
						Vlessgroup: &sbox.Vlessgroup{
							UUID: user.Configs[i].UUID,
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
	
							w.ctrl.DirectMg(C.GetMsg(C.MsgDistributeOver), user.TgID, user.TgID)
	
						case errors.Is(err, C.ErrTypeMissmatch), errors.Is(err, vless.ErrInvalidInbound):
	
							w.ctrl.DirectMg(C.GetMsg(C.MsgwtchErrtypemiss), user.TgID, user.TgID)
	
						case errors.Is(err, vless.ErrVlessService):
	
							w.ctrl.DirectMg(C.GetMsg(C.MsgwtchErruseradd), user.TgID, user.TgID)
	
						}

					} else {

						if !user.Configs[i].Active {
							w.ctrl.DirectMg("Good News Configuration "+ user.Configs[i].Name+" Online Again Due to Bandiwdth Change 🔄", user.TgID, user.TgID)
						}
						user.Configs[i].Active = true
						user.Configs[i].Usage += (status.Download + status.Upload)
						user.MonthUsage += (status.Download + status.Upload)
						user.Configs[i].Download += status.Download
						user.Configs[i].Upload += status.Upload
						
						if status.Download + status.Upload > 0 {
							tx.Create(&db.UsageHistory{
								Usage:    status.Download + status.Upload,
								Download: status.Download,
								Upload:   status.Upload,
								Date:     time.Now(),
								UserID:   user.TgID,
								ConfigID: user.Configs[i].Id,
							})
						}

					}
				} else if user.Configs[i].Active {
					if (user.Configs[i].Quota - user.Configs[i].Usage) <= 0 {
						w.ctrl.DirectMg("⚠️ Your configuration "+user.Configs[i].Name+" has exceeded its usage limit. The config will not function until it is renewed. 🔄", user.TgID, user.TgID)
					}
					status, err := w.ctrl.RemoveUserSbox(&sbox.Userconfig{
						Vlessgroup: &sbox.Vlessgroup{
							UUID: user.Configs[i].UUID,
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
					if err == nil {
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
				if condcheck() {
					user.Configs[i].Usage = 0
					user.Configs[i].Upload = 0
					user.Configs[i].Download = 0

				}
				
				//end new
				if err = tx.Save(&user.Configs[i]).Error; err != nil {
					time.Sleep(800 * time.Millisecond)
					tx.Save(&user.Configs[i])

				}

			}

			user.UsedQuota = usedquota

			if condcheck() {			
				if user.IsDistributedUser && !user.Restricted {
					w.ctrl.DirectMg(C.GetMsg(C.MsgDistributeOver), user.TgID, user.TgID)
				}
				if user.IsMonthLimited && !user.Restricted {
					w.ctrl.DirectMg("You'r Limitation is over", user.TgID, user.TgID)
				}
				
				user.AddPoint(10)
				

				if user.MonthUsage > user.CalculatedQuota+user.AdditionalQuota {

					//TODO: add template here
					w.ctrl.Addquemg(w.ctx, &botapi.Msgcommon{
						Infocontext: &botapi.Infocontext{
							ChatId: user.TgID,
						},
						Text: C.GetMsg(C.MsgwtchUsagereset),
					})

					user.AlltimeUsage += user.CalculatedQuota
					user.MonthUsage = user.MonthUsage - user.CalculatedQuota
					user.SavedQuota = user.CalculatedQuota - user.MonthUsage

					for i, conf := range user.Configs {

						// recalculate excess usage for each configs
						// ratio betweeen conf quota and user quota should equal to ratio between conf excess usage and total excess usage
						// using this we can calculate conf excess usage
						// conf excess usage =  (conf quota * user excess usage )/ user quota

						k := float64(conf.Quota) / (float64(user.CalculatedQuota) + user.AdditionalQuota.Float64())
						user.Configs[i].Usage = C.Bwidth(user.MonthUsage.Float64() * k)
					}

				} else if user.MonthUsage < ((user.CalculatedQuota*3)/4) && !user.IsMonthLimited && !user.IsDistributedUser && !user.Restricted{ 
					//check whether user used 75% from his quota if not user will limited next 30 days
					w.ctrl.Addquemg(w.ctx, &botapi.Msgcommon{
						Infocontext: &botapi.Infocontext{
							ChatId: user.TgID,
						},
						Text: C.GetMsg(C.MsgQuotanotUsed),
					})

					user.IsMonthLimited = true
					user.MonthUsage = 0

				} else  {

					//TODO: add template here
					w.ctrl.Addquemg(w.ctx, &botapi.Msgcommon{
						Infocontext: &botapi.Infocontext{
							ChatId: user.TgID,
						},
						Text: C.GetMsg(C.MsgresetUsage),
					})
					user.IsMonthLimited = false
					user.AlltimeUsage += user.MonthUsage
					user.MonthUsage = 0
				}
				user.IsMonthLimited = false
				user.IsDistributedUser = false
			}

			if err = tx.Model(user).Save(user).Error; err != nil {
				time.Sleep(800 * time.Millisecond)
				tx.Model(user).Save(user)

			}

		}
		w.logger.Debug(" batch prosess done")

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

	if docount {
		w.ctrl.CheckCount.Add(1)
		dbmeta.CheckCount = dbmeta.CheckCount + 1
	}

	if err = w.db.Save(dbmeta).Error; err != nil {
		time.Sleep(800 * time.Millisecond)
		w.db.Save(dbmeta)

	}

	w.lastUserCount = w.ctrl.Dbusercount.Load()

	w.sendDbBackup()


	// if w.CheckClose() != nil {
	// 	w.close <- struct{}{}
	// }

	return nil
}

func (w *Watchman) PreprosessDb(refreshcontext context.Context) (*preprosessd, error) {

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

	var predata = &preprosessd{}

	// var checkcount = w.ctrl.CheckCount.Load()
	// var condcheck = func() bool {
	// 	return checkcount == w.ctrl.ResetCount
	// }

	w.db.Model(&db.User{}).FindInBatches(&[]db.User{}, C.Dbbatchsize, func(tx *gorm.DB, batch int) error {
		// Retrieve the current batch of records
		var users []db.User
		tx.Find(&users) // Fetch the current batch of records from tx

		for _, user := range users {
			if refreshcontext.Err() != nil {
				w.ctrl.WatchmanUnlock()
				w.logger.Warn("Force stopping DB updating, Db update stops from record " + user.Name)
				return fmt.Errorf("context cancled db refresh stops from record id %v, err %v ", user.TgID, refreshcontext.Err())
			}
			if user.IsAdmin {
				continue
			}
			// if err := w.db.Model(&db.Config{}).Where("user_id = ?", user.TgID).Find(&user.Configs).Error; err != nil {
			// 	continue
			// }

			if user.IsCapped {
				if user.Iscaptimeover() {

					user.IsCapped = false
					w.logger.Debug("cap time over user capped quota resets " + user.Name)
					user.CappedQuota = 0

					//tx.Model(&db.User{}).First(&user).Update("is_capped", false)
					tx.Save(&user)
				} else {
					predata.captotal += user.CappedQuota
					predata.captotal -= (user.CappedQuota)
				}
			}

		}

		return nil // Return nil to continue to the next batch
	},
	)

	var err error

	if err = w.db.Model(&db.User{}).Where("is_capped = ?", true).Count(&predata.cappeduser).Error; err != nil {
		return predata, C.ErrDbopration
	}
	if err = w.db.Model(&db.User{}).Where("restricted = ?", true).Count(&predata.restricted).Error; err != nil {
		return predata, C.ErrDbopration
	}

	if err = w.db.Model(&db.User{}).Where("is_in_channel = ? AND is_in_group = ?", true, true).Count(&predata.verifiedusercount).Error; err != nil {
		return predata, C.ErrDbopration
	}

	if err = w.db.Model(&db.User{}).Where("is_dis_user = ?", true).Count(&predata.distributeduser).Error; err != nil {
		return predata, C.ErrDbopration
	}

	if err := w.db.Model(&db.User{}).Select("COALESCE(SUM(capped_quota), 0)").Scan(&predata.captotal).Error; err != nil {
		return predata, C.ErrDbopration
	}

	if err := w.db.Model(&db.User{}).Select("COALESCE(SUM(additional_quota), 0)").Scan(&predata.totaladdtional).Error; err != nil {
		return predata, C.ErrDbopration
	}
	if err := w.db.Model(&db.User{}).Select("COALESCE(SUM(saved_quota), 0)").Scan(&predata.savings).Error; err != nil {
		return predata, C.ErrDbopration
	}
	
	if err := w.db.Model(&db.User{}).Where("is_month_limited = ? AND  is_in_channel = ? AND is_in_group = ? AND restricted = ?", true, true, true, false).Count(&predata.monthlimiteduser).Error; err != nil {
		return predata, C.ErrDbopration
	}
	if err := w.db.Model(&db.User{}).Where("is_dis_user = ?", true).Select("COALESCE(SUM(month_usage), 0)").Scan(&predata.usedbydisuser).Error; err != nil {
		return predata, C.ErrDbopration
	}
	if err := w.db.Model(&db.User{}).Where("restricted = ? AND is_dis_user = ?", true, false).Select("COALESCE(SUM(month_usage), 0)").Scan(&predata.usedbyrestricted).Error; err != nil {
		return predata, C.ErrDbopration
	}


	overview := w.ctrl.Overview

	var (
		month_usage = C.Bwidth(0)
		alltime = C.Bwidth(0)
		oerr error
	)


	if err := w.db.Model(&db.User{}).Select("COALESCE(SUM(all_time_usage), 0)").Scan(&alltime).Error; err != nil {
		overview.Mu.RLock()
		alltime = overview.AllTime
		overview.Mu.RUnlock()
		oerr = err
	}
	if err := w.db.Model(&db.User{}).Select("COALESCE(SUM(month_usage), 0)").Scan(&month_usage).Error; err != nil {
		overview.Mu.RLock()
		alltime = overview.AllTime
		overview.Mu.RUnlock()
		oerr = err
	}

	overview.Mu.Lock()
	overview.CappedUser = predata.cappeduser
	overview.DistributedUser = predata.distributeduser
	overview.VerifiedUserCount = predata.verifiedusercount
	overview.TotalUser = w.ctrl.Dbusercount.Load()
	overview.MonthTotal = month_usage
	overview.AllTime = alltime
	overview.BandwidthAvailable = w.ctrl.BandwidthAvelable
	overview.Restricted = predata.restricted
	overview.Error = oerr
	overview.QuotaForEach = C.Bwidth(w.ctrl.CommonQuota.Load())
	overview.LastRefresh = time.Now()
	overview.Mu.Unlock()

	return predata, nil
}

// DO not call outside refresh db
func (w *Watchman) sendDbBackup() {
	dbraw, err := os.Open(w.db.DatabasePath())
	if err != nil {
		w.logger.Error("errored when reading database for backup create", zap.Error(err))
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
		w.logger.Error("request making failed" +  err.Error())
		return
	}
	
	apires, err := w.botapi.SendRawReq(req)
	
	if err != nil {
		w.logger.Error("request send failed when uploading backup database ", zap.Error(err))
		return
	}
	if !apires.Ok {
		w.logger.Error("Bad Response From Telegram " + apires.Description)
	}


}