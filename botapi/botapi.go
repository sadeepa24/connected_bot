package botapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	C "github.com/sadeepa24/connected_bot/constbot"
	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
)

type BotAPI interface {
	Makerequest(ctx context.Context, method, endpoint string, body *BotReader) (*tgbotapi.APIResponse, error)
	SendRawReq(req *http.Request) (*tgbotapi.APIResponse, error)
	SendContext(ctx context.Context, msg *Msgcommon) (*tgbotapi.Message, error)
	AnswereCallbackCtx(ctx context.Context, Callbackanswere *Callbackanswere) error
	GetchatmemberCtx(ctx context.Context, Userid int64, Chatid int64) (*tgbotapi.ChatMember, bool, error)
	//Createkeyboard(keyboard *InlineKeyboardMarkup) ([]byte,error)
	Send(msg *Msgcommon) (*tgbotapi.Message, error)
	SendError(error, int64)
	DeleteMsg(ctx context.Context, msgid int64, chatid int64) error
	GetMgStore() *MessageStore
	SetWebhook(webhookurl, secret, ip_addr string, allowd_ob []string) error
	CreateFullUrl(endpoint string) string
	GetFile(file_Id string) (io.ReadCloser, error)
}

type Botapi struct {
	ctx    context.Context
	token  string
	Client *http.Client
	//tgapi *tgbotapi.BotAPI
	mainurl string
	mgstore *MessageStore
}

var _ BotAPI = (*Botapi)(nil)

func NewBot(ctx context.Context, token, mainurl string, mgstore *MessageStore) *Botapi {
	return &Botapi{
		Client: &http.Client{
			Timeout:  5 * time.Minute, //TODO: change later
			//Transport: &http.Transport{},
		},
		ctx:     ctx,
		token:   token,
		mainurl: mainurl,
		mgstore: mgstore,
	}
}
func (b *Botapi) GetFile(filed_id string) (io.ReadCloser, error) {
	res, err := b.Makerequest(b.ctx, "POST", "getFile", CreateReder(struct{
		File_id string `json:"file_id,omitempty"`
	}{
		File_id: filed_id,
	}))
	if err != nil {
		return nil, err
	}

	if res.Ok {
		var file = &tgbotapi.File{}
		err = json.Unmarshal(res.Result, file)
		if err != nil {
			return nil, err
		}
		if file.FilePath == "" {
			return nil, errors.New("file path does not avbl")
		}
		res, err := b.Client.Get("https://api.telegram.org/file/bot" + b.token + "/" + file.FilePath)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusOK {
			return nil, errors.New("request diffrent status code - " + res.Status)
		}
		return res.Body, nil
	}

	return nil, errors.New("tg response err " + res.Description)
}


func (b *Botapi) SetWebhook(webhookurl, secret, ip_addr string, allo_updates []string) error {
	setwebhook := tgbotapi.WebhookInfo{
		URL:                  webhookurl,
		HasCustomCertificate: false,
		IPAddress:            ip_addr,
		AllowedUpdates:       allo_updates,
		Secret_token:         secret,
	}

	reply, err := b.Makerequest(b.ctx, http.MethodPost, "setWebhook", &BotReader{RealOb: setwebhook})
	if err != nil {
		return errors.Join(errors.New("request for Webhook set, sent failed "), err)
	}
	if !reply.Ok {
		return C.ErrWebhookSetFailed
	}
	return nil
}

func (b *Botapi) CreateFullUrl(endpoint string) string {
	return  b.mainurl+b.token+"/"+endpoint
}

// Endpoint should be without slash "/"
func (b *Botapi) Makerequest(reqctx context.Context, method, endpoint string, body *BotReader) (*tgbotapi.APIResponse, error) {
	if reqctx == nil {
		reqctx = b.ctx
	}
	req, err := http.NewRequestWithContext(reqctx, method, b.mainurl+b.token+"/"+endpoint, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.ContentLength = int64(body.Len())
	if err != nil {
		return nil, err
	}
	return b.SendRawReq(req)

}

func (b *Botapi) SendRawReq(req *http.Request) (*tgbotapi.APIResponse, error) {
	if req == nil {
		return nil, C.ErrNilRequest
	}
	res, err := b.Client.Do(req)
	if err != nil {
		return nil, C.ErrClientRequestFail
	}
	if res.StatusCode == 400 {
		return nil, C.ErrTgParsing
	}
	apires := &tgbotapi.APIResponse{}
	//var update tgbotapi.Update
    decoder := json.NewDecoder(res.Body)
    if err := decoder.Decode(&apires); err != nil {
		return nil, C.ErrJsonopra
    }
	if !apires.Ok {
		return nil, errors.Join(C.ErrApierror, fmt.Errorf("tgrrcode %d Discription %s", apires.ErrorCode, apires.Description))
	}
	return apires, nil
}

func (b *Botapi) GetMgStore() *MessageStore {
	return b.mgstore
}

func (b *Botapi) AnswereInlineQuary(ctx context.Context) error {
	return nil
}




func (b *Botapi) SendContext(ctx context.Context, msg *Msgcommon) (*tgbotapi.Message, error) {
	endpoint := "sendMessage"
	if msg.Message_id != 0 {
		endpoint = "editMessageText"

	}
	if msg.Meadiacommon != nil {
		switch {
		case msg.Message_id != 0:
			endpoint = "editMessageCaption"
		case msg.Meadiacommon.Photo != nil:
			endpoint = "sendPhoto"
		case msg.Meadiacommon.Video != nil:
			endpoint = "sendVideo"
		}

	}
	if msg.Endpoint != "" {
		endpoint = msg.Endpoint
	}

	apires, err := b.Makerequest(ctx, http.MethodPost, endpoint, &BotReader{RealOb: msg})

	if err != nil {
		return nil, err
	}
	message := &tgbotapi.Message{}
	if err = json.Unmarshal(apires.Result, message); err != nil {
		return nil, C.ErrJsonopra
	}
	return message, nil
}



func (b *Botapi) AnswereCallbackCtx(ctx context.Context, Callbackanswere *Callbackanswere) error {
	_, err := b.Makerequest(ctx, http.MethodPost, "answerCallbackQuery", &BotReader{RealOb: Callbackanswere})
	return err

}

func (b *Botapi) GetchatmemberCtx(ctx context.Context, Userid int64, Chatid int64) (*tgbotapi.ChatMember, bool, error) {

	sendchamem := &Msgcommon{
		Infocontext: &Infocontext{
			User_id: Userid,
			ChatId:  Chatid,
		},
	}
	apires, err := b.Makerequest(ctx, http.MethodPost, "getChatMember",  &BotReader{RealOb: sendchamem})

	if err != nil {
		return nil, false, err

	}

	chamember := &tgbotapi.ChatMember{}
	err = json.Unmarshal(apires.Result, chamember)
	if err != nil {
		err = C.ErrJsonopra
	}

	return chamember, chamember.Status == "member" || chamember.Status == "administrator" || chamember.Status == "creator", err
}

func (b *Botapi) Send(msg *Msgcommon) (*tgbotapi.Message, error) {
	return b.SendContext(context.Background(), msg)

}

func (b *Botapi) SendError(err error, UserID int64) {
	b.Send(&Msgcommon{
		Text: err.Error() + "      Please Send This Error To Admin It's Very Important Thin to Give Bug free Service",
		Infocontext: &Infocontext{
			User_id: UserID,
			ChatId:  UserID,
		},
	})
}

func (b *Botapi) DeleteMsg(ctx context.Context, msgid int64, chatid int64) error {
	_, err := b.Makerequest(ctx, http.MethodPost, "deleteMessage", &BotReader{RealOb: &Msgcommon{
		Infocontext: &Infocontext{
			ChatId: chatid,
		},
		Message_id: msgid,
	}})
	return err
}

// This object Is not concurrent safe use it only in one go routine
type Msgsession struct {
	api            BotAPI
	userID         int64
	ChatID         int64
	MessageID      int
	mainctx        context.Context
	infoctx        Infocontext //user info
	sendfirst      bool
	sentmsg        []int64
	Lasscallbackid int64
	Prime          int64 //Deorecated

	msgstore           *MessageStore
	lang               string
	lastsendmeadia     bool //was lastmessagemedia
	continuemedia      bool // continue session with the media
	supercontinue      bool
	continue_skip_text bool // when true meadia wan't send with text

	alertsent    bool
	replyrecived bool

	cached *Msgcommon

	replyque []int

	lastmediatype string
	lastmedia     string
}

func NewMsgsession(upxctx context.Context, api BotAPI, userid int64, chatid int64, lang string,) *Msgsession {
	return &Msgsession{
		api:    api,
		mainctx: upxctx,
		userID: userid,
		ChatID: chatid,
		infoctx: Infocontext{
			User_id: userid,
			ChatId:  chatid,
		},
		sendfirst:      true,
		sentmsg:        []int64{},
		Prime:          0,
		replyque:       []int{},
		msgstore:       api.GetMgStore(),
		lang:           lang,
		lastsendmeadia: false,
	}

}

func (m *Msgsession) SetPrimeLast() {
	if len(m.sentmsg) > 0 {
		m.Prime = int64(m.MessageID)
		for i, mg := range m.sentmsg {
			if mg == m.Prime {
				m.sentmsg = append(m.sentmsg[:i], m.sentmsg[i+1:]...)
				break
			}
		}
	}
}

type Htmlstring string

// Many types supported by this method msg parameter can be Upmessage, struct with template name, or string
func (m *Msgsession) Edit(msg any, buttons *Buttons, name string) (*tgbotapi.Message, error) {
	if msg == nil {
		return nil, errors.New("no msg")
	}
	var (
		sendmsg = &Msgcommon{
			Infocontext: &Infocontext{
				ChatId: m.ChatID,
			},
		}
	)
	if m.continuemedia {
		sendmsg.Endpoint = C.ApiMethodCaptionEdit
	}

	if (m.alertsent || m.replyrecived) && m.MessageID != 0 {
		m.DeleteLast()
		m.alertsent = false
		m.replyrecived = false
	}

	switch realmg := msg.(type) {
	case string:
		sendmsg.Text = realmg
		if len(sendmsg.Text) > 4096 {
			sendmsg.Text = m.partialsend(sendmsg.Text)
		}
		m.preparesentmg(sendmsg, buttons != nil, false)
	case Htmlstring:
		sendmsg.Text = string(realmg)
		if len(sendmsg.Text) >= 4096 {
			sendmsg.Text = m.partialsend(sendmsg.Text)
		}
		sendmsg.Parse_mode = C.ParseHtml
		m.preparesentmg(sendmsg, buttons != nil, false)
	default:
		if upmg, ok := msg.(UpMessage); ok {
			msg = upmg.Template
			name = upmg.TemplateName

		}
		
		var (
			message *Message
			err error
			ok bool
		)
		if message, ok = msg.(*Message); !ok {
			if message, err = m.msgstore.GetMessage(name, m.lang, msg); err != nil {
				if errors.Is(err, C.ErrMsgDisabled) {
					return nil, err
				}
				return nil, C.ErrTmplRender
			}
		}
		
		if message.Includemed {
			if m.lastsendmeadia && !(m.alertsent || m.replyrecived) {
				sendmsg.Meadiacommon = &Meadiacommon{
					Media: &InputMedia{
						Type:      message.MedType,
						Media:     message.MediaId,
						Caption:   message.String(),
						ParseMode: message.ParseMode,
					},
				}
				sendmsg.Endpoint = C.ApiMethodEdimgmed

			} else {
				sendmsg.Meadiacommon = &Meadiacommon{
					Caption: message.String(),
				}
				sendmsg.Parse_mode = message.ParseMode
				sendmsg.SetMedType(message.MedType, message.MediaId)
			}

			m.continuemedia = message.ContinueMed

			if message.SuperContinue { // cannot equal directly because if m.supercontinue == true then we cannot swap it to false again
				m.supercontinue = true
			}

			

			if m.MessageID != 0 && !m.lastsendmeadia {
				m.DeleteLast()
				m.MessageID = 0
				m.sendfirst = true
				m.alertsent = false
				m.replyrecived = false
			}
			m.lastsendmeadia = true

			if m.continuemedia || m.supercontinue {
				m.lastmediatype = message.MedType
				m.lastmedia = message.MediaId
				m.continue_skip_text = message.Continue_Skip_Text
			}

			if m.supercontinue {
				m.continuemedia = true
			}

		} else {
			sendmsg.Infocontext = &m.infoctx
			sendmsg.Text = message.String()
			sendmsg.Parse_mode = message.ParseMode

			// if m.continuemedia && !message.SkipText && (!m.continue_skip_text || buttons != nil) {
			// 	sendmsg.Meadiacommon = &Meadiacommon{
			// 		Caption: sendmsg.Text,
			// 	}

			// 	fmt.Println("trigger 2")
			// 	sendmsg.Text = ""
			// 	sendmsg.Infocontext = &Infocontext{
			// 		ChatId: m.ChatID,
			// 	}

			// 	if (m.alertsent || m.replyrecived) || m.MessageID == 0  {
			// 		fmt.Println("trigger 3")
			// 		sendmsg.SetMedType(m.lastmediatype, m.lastmedia)
			// 	}
			// } else if m.continue_skip_text {
			// 	m.DeleteLast()
			// 	sendmsg.Endpoint = C.ApiMethodEdimgmed
			// 	if (m.alertsent || m.replyrecived) || m.MessageID == 0 || !m.lastsendmeadia {
			// 		sendmsg.Endpoint = C.ApiMethodSendMG
			// 	}
			// }

			m.preparesentmg(sendmsg, buttons != nil, message.MeadiaSkip)

		}
	}

	// delete last sent msg with media when current message does not have media
	// also deete when already alert sent or reply recived
	// by doing so current msg will be the latest msg in chat
	if ((sendmsg.Meadiacommon == nil) && !m.continuemedia && m.MessageID != 0 && m.lastsendmeadia) ||
		((m.alertsent || m.replyrecived) && m.MessageID != 0) {
		m.DeleteLast()
		m.alertsent = false
		m.replyrecived = false
	}

	if buttons != nil {
		sendmsg.Reply_markup = buttons.Getkeyboard()
	}
	if m.MessageID != 0 {
		sendmsg.Message_id = int64(m.MessageID)
	}
	if m.Prime != 0 {
		sendmsg.Message_id = m.Prime
	}

	replymsg, err := m.api.SendContext(m.mainctx, sendmsg)

	if err != nil {
		switch {
		case errors.Is(err, C.ErrClientRequestFail), errors.Is(err, C.ErrApierror), errors.Is(err, C.ErrRead):
			if replymsg, err = m.api.SendContext(m.mainctx, sendmsg); err != nil { //retry
				return nil, err
			}
		case errors.Is(err, C.ErrJsonopra): //TODO: handle later
			return nil, err
		default:
			return nil, err
		}

	}
	//m.cached = sendmsg

	if m.MessageID == 0 {
		m.sentmsg = append(m.sentmsg, int64(replymsg.MessageID))
	}
	m.alertsent = false
	if m.sendfirst {
		m.MessageID = replymsg.MessageID
	}
	m.sendfirst = false
	return replymsg, err
}

func (m *Msgsession) preparesentmg(sendmsg *Msgcommon, btnavbl, doskip bool, ) {
	//sendmsg.Text = m.partialsend(sendmsg.Text)
	if m.continuemedia && (!m.continue_skip_text || btnavbl) && !doskip {
		sendmsg.Meadiacommon = &Meadiacommon{
			Caption: sendmsg.Text,
		}
		if !m.lastsendmeadia {
			m.DeleteLast()
		}
		
		sendmsg.Text = ""
		if (m.alertsent || m.replyrecived) || m.MessageID == 0 || !m.lastsendmeadia {
			m.sendfirst = true
			sendmsg.SetMedType(m.lastmediatype, m.lastmedia)
		}
		m.lastsendmeadia = true
	} else if m.continue_skip_text || doskip {
		m.DeleteLast()
		m.alertsent = true
		
		sendmsg.Endpoint = C.ApiMethodSendMG
	}
}

// remove all btns from last sent msg
// this functiuon does not gurntee anything
func (m *Msgsession) RemoveBtns() error {
	if m.cached == nil {
		return nil
	}
	if m.MessageID == 0 {
		return nil
	}
	m.cached.Reply_markup = Keyboard{}
	m.api.SendContext(m.mainctx, m.cached)
	return nil
}

func (m *Msgsession) partialsend(text string) string {
	parts := len(text)/C.MaxCharacterMg 
	if parts > 0 {
		btns := NewButtons([]int16{1})
		btns.AddBtcommon(C.BtnClose)
		for i := 1; i < parts+1; i++ {
			m.SendAlert(Htmlstring(text[:C.MaxCharacterMg]), btns)
			text = text[C.MaxCharacterMg:]
		}
	}
	return text
}

func (m *Msgsession) EditText(msg string, buttons *Buttons) (*tgbotapi.Message, error) {
	return m.Edit(msg, buttons, "")
}

func (m *Msgsession) SendExtranal(msg any, buttons *Buttons, name string, nodel bool) (*tgbotapi.Message, error) {

	sendmsg := &Msgcommon{
		Infocontext: &Infocontext{
			ChatId: m.ChatID,
		},
		Parse_mode: "",
		Endpoint:   "",
	}

	switch realmg := msg.(type) {
	case string:
		sendmsg.Text = realmg
		sendmsg.Endpoint = C.ApiMethodSendMG
	default:
		rendermg, err := m.msgstore.GetMessage(name, m.lang, realmg)
		if err != nil {
			sendmsg.Text = "error from template please note admin"
			break
		}
		if rendermg.Includemed {

			sendmsg.Meadiacommon = &Meadiacommon{}
			switch rendermg.MedType {
			case C.MedPhoto:
				sendmsg.Meadiacommon.Photo = rendermg.MediaId
				sendmsg.Endpoint = C.ApiMethodSendPhoto
			case C.MedVideo:
				sendmsg.Meadiacommon.Video = rendermg.MediaId
				sendmsg.Endpoint = C.ApiMethodSendVid
			}
			sendmsg.Parse_mode = rendermg.ParseMode
			sendmsg.Caption = rendermg.String()

		} else {
			sendmsg.Text = rendermg.String()
			sendmsg.Parse_mode = rendermg.ParseMode
			sendmsg.Endpoint = C.ApiMethodSendMG
			sendmsg.User_id = m.userID
		}
	}
	if buttons != nil {
		sendmsg.Reply_markup = buttons.Getkeyboard()
	}

	replymg, err := m.api.SendContext(m.mainctx, sendmsg)
	if err != nil {
		if errors.Is(err, C.ErrTgParsing) {
			sendmsg.Parse_mode = ""
			if sendmsg.Meadiacommon != nil && sendmsg.Meadiacommon.Media != nil {
				sendmsg.Meadiacommon.Media.ParseMode = ""
			}
			replymg, err = m.api.SendContext(m.mainctx, sendmsg)

		}
		if err != nil {
			return nil, err
		}
	}

	if !nodel {
		m.sentmsg = append([]int64{int64(replymg.MessageID)}, m.sentmsg...)
	}
	m.alertsent = true

	return replymg, err
}

// Will send a compleatly new msg not relvent edititn msgs
// will automatically remove with delete all
func (m *Msgsession) SendAlert(msg any, buttons *Buttons) (*tgbotapi.Message, error) {

	sendmsg := &Msgcommon{
		Infocontext: &m.infoctx,
		Text:        "",
	}

	switch mg := msg.(type) {
	case string:
		sendmsg.Text = mg

	case Htmlstring:
		sendmsg.Text = string(mg)
		if len(sendmsg.Text) > 4096 {
			sendmsg.Text =  m.partialsend(sendmsg.Text)
		}
		sendmsg.Parse_mode = C.ParseHtml
	}

	if buttons != nil {
		sendmsg.Reply_markup = buttons.Getkeyboard()
	}
	replymsg, err := m.api.SendContext(m.mainctx, sendmsg)
	if err != nil {
		switch {
		case errors.Is(err, C.ErrClientRequestFail), errors.Is(err, C.ErrApierror), errors.Is(err, C.ErrRead):
			if replymsg, err = m.api.SendContext(m.mainctx, sendmsg); err != nil { //retry
				return nil, err
			}
			
		case errors.Is(err, C.ErrJsonopra):
			return nil, err
		default:
			return nil, err
		}
	

	}

	m.sentmsg = append([]int64{int64(replymsg.MessageID)}, m.sentmsg...)
	m.alertsent = true
	return replymsg, err

}

func (m *Msgsession) ForwardMgTo(to int64, mgid, fromchat int64) error {
	forward := &tgbotapi.ForwardMessage{
		Chat_id:      to,
		Message_id:   mgid,
		From_chat_id: m.ChatID,
	}
	_, err := m.api.Makerequest(m.mainctx, http.MethodPost, "forwardMessage", &BotReader{RealOb: forward})
	return err
}

func (m *Msgsession) CopyMessageTo(to int64, mgid int64) error {
	copymg := &tgbotapi.CopyMessage{
		Chat_id:      to,
		Message_id:   mgid,
		From_chat_id: m.ChatID,
	}
	_, err := m.api.Makerequest(m.mainctx, http.MethodPost, "copyMessage", &BotReader{RealOb: copymg})
	return err
}

func (m *Msgsession) CopyMessageRawTo(to, mgid, fromchat int64) error {
	copymg := &tgbotapi.CopyMessage{
		Chat_id:      to,
		Message_id:   mgid,
		From_chat_id: fromchat,
	}
	_, err := m.api.Makerequest(m.mainctx, http.MethodPost, "copyMessage", &BotReader{RealOb: copymg})
	return err
}

func (m *Msgsession) Addreply(id int) {

	m.replyrecived = true

	m.replyque = append(m.replyque, id)
}

// After adding reply, current mg will not delete and resend
func (m *Msgsession) AddreplyNoDelete(id int) {
	m.replyque = append(m.replyque, id)
}

func (m *Msgsession) DeleteReplys() error {
	var terr error
	for _, rep := range m.replyque {
		if err := m.api.DeleteMsg(m.mainctx, int64(rep), m.ChatID); err != nil {
			terr = errors.Join(terr, err)
		}
	}
	return terr
}

func (m *Msgsession) EditNewcontext(ctx context.Context, msg any, buttons *Buttons, name string) (*tgbotapi.Message, error) {
	m.mainctx = ctx
	return m.Edit(msg, buttons, name)
}

// Send New message and set it as last message
func (m *Msgsession) SendNew(msg any, buttons *Buttons, name string) (*tgbotapi.Message, error) {
	currentId := m.MessageID
	m.MessageID = 0
	m.sendfirst = true
	replymsg, err := m.Edit(msg, buttons, name)
	if err != nil {
		m.MessageID = currentId
		m.sendfirst = false
	}
	return replymsg, err

}

func (m *Msgsession) DeleteLast() {

	if len(m.sentmsg) > 0 {

		err := m.api.DeleteMsg(m.mainctx, int64(m.MessageID), m.ChatID)
		if err != nil {
			//
		}
		m.sentmsg = m.sentmsg[:len(m.sentmsg)-1]
		if len(m.sentmsg) > 0 {
			m.MessageID = int(m.sentmsg[len(m.sentmsg)-1])
		}
		m.sendfirst = true
		m.MessageID = 0
		m.lastsendmeadia = false

	}

}

//Delete last Message adn send new and set it as last message

func (m *Msgsession) DeleteAllMsg() error {
	var terr error
	for _, msgid := range m.sentmsg {
		if err := m.api.DeleteMsg(m.mainctx, msgid, m.ChatID); err != nil {
			terr = errors.Join(terr, err)
		}
	}
	if m.Prime != 0 {
		if err := m.api.DeleteMsg(m.mainctx, m.Prime, m.ChatID); err != nil {
			terr = errors.Join(terr, err)
		}
	}
	for _, rep := range m.replyque {
		if err := m.api.DeleteMsg(m.mainctx, int64(rep), m.ChatID); err != nil {
			terr = errors.Join(terr, err)
		}
	}
	return terr
}
func (m *Msgsession) SetNewcontext(ctx context.Context) {
	m.mainctx = ctx
}

func (m *Msgsession) Callbackanswere(quoaryid, text string, alert bool) error {
	return m.api.AnswereCallbackCtx(m.mainctx, &Callbackanswere{
		Callback_query_id: quoaryid,
		Text:              text,
		Show_alert:        alert,
	})
}

// Deprecated
func (m *Msgsession) SendTmpl(name string, obj any, btns *Buttons) (*tgbotapi.Message, error) {

	if obj == nil {
		return nil, errors.New("no obj")
	}

	message, err := m.msgstore.GetMessage(name, m.lang, obj)
	if err != nil {
		return nil, C.ErrTmplRender
	}

	if message.Includemed {
		var commonmg = &Msgcommon{
			Infocontext: &Infocontext{
				ChatId: m.ChatID,
			},
			Meadiacommon: &Meadiacommon{
				Caption: message.String(),
			},
			Parse_mode: message.ParseMode,
		}

		switch message.MedType {
		case C.MedPhoto:

			inmed := &InputMedia{
				Type:      C.MedPhoto,
				Media:     message.MediaId,
				Caption:   message.String(),
				ParseMode: message.ParseMode,
			}

			commonmg.Meadiacommon.Photo = inmed.Marshal()
		case C.MedVideo:

			inmed := &InputMedia{
				Type:      C.MedVideo,
				Media:     message.MediaId,
				Caption:   message.String(),
				ParseMode: message.ParseMode,
			}

			commonmg.Meadiacommon.Video = inmed.Marshal()
		}

		if m.MessageID != 0 {
			commonmg.Endpoint = "editMessageCaption"
		}

		if btns != nil {
			commonmg.Reply_markup = Keyboard{
				Inline_keyboard: btns.InlineKeyboard,
			}
		}

		replymg, err := m.api.SendContext(m.mainctx, commonmg)

		if err != nil {
			if errors.Is(err, C.ErrContextDead) {
				return nil, err
			} else if errors.Is(err, C.ErrClientRequestFail) {
				return m.api.SendContext(m.mainctx, commonmg)
			}
		}

		m.sentmsg = append([]int64{int64(replymg.MessageID)}, m.sentmsg...)

		return replymg, nil

	} else {

		//return m.Edit(message.String(), btns)
		return nil, nil
	}

}

type Buttons struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
	btmatrix       []int16
	lastval        int16
	uniqid         int64
	nextnew        bool
}

// Create button Schema
// Examle if you give []int16{2, 1, 3,}, when you call Addbuttons button will add  to this ass the map
// Ex :-
//
//	   btn
//	btn,  btn
//
// btn btn btn
func NewButtons(btmatrix []int16) *Buttons {
	btns := &Buttons{
		InlineKeyboard: make([][]InlineKeyboardButton, len(btmatrix)),
		btmatrix:       btmatrix,
		uniqid:         int64(rand.Int31()),
		nextnew:        true,
	}

	if len(btmatrix) > 0 {
		btns.lastval = btmatrix[len(btmatrix)-1]
	}
	return btns
}

func (b *Buttons) ID() int64 {
	return b.uniqid
}
func (b *Buttons) Reset(btmatrix []int16) {
	b.btmatrix = btmatrix
	b.InlineKeyboard = make([][]InlineKeyboardButton, len(btmatrix))
	b.uniqid = int64(rand.Int31n(4556))
	b.nextnew = true

	if len(btmatrix) > 0 {
		b.lastval = btmatrix[len(btmatrix)-1]
	}

}

func (b *Buttons) Addbutton(btnname, data, url string) {
	for i, matic := range b.btmatrix {
		if matic > 0 {
			b.nextnew = false
			cData := Callbackdata{
				Uniqid: b.uniqid,
				Data:   data,
			}
			if len(b.InlineKeyboard) <= i {
				b.InlineKeyboard = append(b.InlineKeyboard, []InlineKeyboardButton{})
			}
			b.InlineKeyboard[i] = append(b.InlineKeyboard[i], InlineKeyboardButton{
				CallbackData: cData.StringV2(),
				Text:         btnname,
				URL:          url,
			})
			b.btmatrix[i] = matic - 1
			if b.btmatrix[i] == 0 {
				b.nextnew = true
			}
			return
		}
	}

	b.btmatrix = append(b.btmatrix, b.lastval)
	b.Addbutton(btnname, data, url)

}

func (b *Buttons) PassButtons(count int16) {
	for i, matic := range b.btmatrix {
		if matic > 0 {
			b.btmatrix[i] = matic - count
			if b.btmatrix[i] == 0 {
				b.nextnew = true
			}
		}
	}
}

func (b *Buttons) Passline() {
	if b.nextnew {
		return
	}
	for i, matic := range b.btmatrix {
		if matic > 0 {
			b.btmatrix[i] = 0
			b.nextnew = true
		}
	}
}

// Should Called After adding allbuttons
func (b *Buttons) AddCloseBack() {

	if !b.nextnew {
		b.Passline()
	}

	b.Addbutton(C.BtnBack, C.BtnBack, "")
	b.Addbutton(C.BtnClose, C.BtnClose, "")
}

// should call after adding all neccery buttons
func (b *Buttons) AddClose(newline bool) {
	if newline && !b.nextnew {
		b.Passline()
	}
	b.Addbutton(C.BtnClose, C.BtnClose, "")
}

func (b *Buttons) AddBack(newline bool) {
	if newline && !b.nextnew {
		b.Passline()
	}
	b.Addbutton(C.BtnBack, C.BtnBack, "")
}

func (b *Buttons) AddUrlbutton(name, url string) {
	b.Addbutton(name, name, url)
}

func (b *Buttons) Addcancle() {
	b.Addbutton(C.BtnCancle, C.BtnCancle, "")
}

func (b *Buttons) AddBtcommon(btn string) {
	b.Addbutton(btn, btn, "")
}

func (b *Buttons) Getkeyboard() Keyboard {
	return Keyboard{
		Inline_keyboard: b.InlineKeyboard,
	}
}

// func (b *Buttons) GetKeyBoardTgbotapi() tgbotapi.ReplyKeyboardMarkup {
	



// 	tgmap := make([][]tgbotapi.KeyboardButton, len(b.InlineKeyboard))

// 	for i, keymap := range b.InlineKeyboard {
// 		tgmap[i] = make([]tgbotapi.KeyboardButton, len(keymap))
		
// 		for j, key := range keymap {
// 			tgmap[j][i] = tgbotapi.KeyboardButton{
// 				Text: key.Text,
				
// 			} 
// 		}
// 	}
	
	
// 	return tgbotapi.ReplyKeyboardMarkup{
// 		Keyboard: [][]tgbotapi.KeyboardButton(b.InlineKeyboard),
// 	}
// }

func (b *Buttons) Marshell() (json.RawMessage, error) {
	rkeyboard, err := Createkeyboard(&InlineKeyboardMarkup{
		InlineKeyboard: b.InlineKeyboard,
	})
	if err != nil {
		return nil, err
	}
	return json.RawMessage(rkeyboard), nil
}

// func (b *Buttons) AddSpecial(btnname, data,url string) {
// 	b.InlineKeyboard = append(b.InlineKeyboard, []InlineKeyboardButton{

// 	})
// 	b.Addbutton(C.BtnCancle, C.BtnCancle, "")
// }

type Callbackdata struct {
	Uniqid int64  `json:"uid"`
	Data   string `json:"data"`
}

// Deprecated Use V2
func (c *Callbackdata) String() string {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(jsonData)
}

// Deprecated Use V2
func (c *Callbackdata) Fill(data string) error {
	datab, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(datab, c)

}

func (c *Callbackdata) StringV2() string {

	str := strconv.Itoa(int(c.Uniqid)) + "eyjy" + c.Data
	return str

}

func (c *Callbackdata) FillV2(data string) error {
	out := strings.Split(data, "eyjy")
	if len(out) != 2 {
		return errors.New("wrong length")
	}
	c.Data = out[1]
	intt, err := strconv.Atoi(out[0])
	if err != nil {
		return err
	}
	c.Uniqid = int64(intt)
	return nil
}

func Createkeyboard(keyboard *InlineKeyboardMarkup) ([]byte, error) {
	if keyboard == nil {
		return []byte{}, fmt.Errorf("nil keyboard enterd")
	}
	return json.Marshal(keyboard)
}


type Filepart struct {
	Reader io.Reader
	Name string
}


func CreateMultiPartReq(ctx context.Context, method, url string,   fields map[string]string,  fileparts map[string]Filepart, ) (*http.Request, error) {
	var (
		body bytes.Buffer
		err error
	)
	
	multiparwriter := multipart.NewWriter(&body)

	for field, fpart := range fileparts {
		if fpart.Reader == nil {
			continue
		}
		
		iopart, err := multiparwriter.CreateFormFile(field, fpart.Name)
		if err != nil {
			return nil, err
		}
		pt, err := io.ReadAll(fpart.Reader)
		if err != nil {
			return nil, err
		}
		_, err = iopart.Write(pt)
		if err != nil {
			return nil, err
		}
	}

	for fiedl, val := range fields {
		err = multiparwriter.WriteField(fiedl, val)
		if err != nil {
			return  nil, err
		}
	}
	multiparwriter.Close()
	ContentType := multiparwriter.FormDataContentType()
	req, err := http.NewRequestWithContext(ctx, method, url, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", ContentType)

	return req, err

}