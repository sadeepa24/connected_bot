package botapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"

	C "github.com/sadeepa24/connected_bot/constbot"
)


type Error struct {
	error
	usermsg string
	exit bool
	code uint16
}

const (
	ErrJsonOp = iota
	ErrReqFail
	ErrReqCreate
	ErrBadReq
	ErrMsgDisable
	ErrTmplRender
)


var _ C.Error = (*Error)(nil)

func (e Error) Exit() bool {return e.exit}
func (e Error) UserMsg() string {return e.usermsg }
     



type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton represents a button in the inline keyboard
type InlineKeyboardButton struct {
	Text         string `json:"text,omitempty"`
	CallbackData string `json:"callback_data,omitempty"` // For button actions
	URL          string `json:"url,omitempty"`           // For opening a URL
}

type Infocontext struct {
	ChatId  int64 `json:"chat_id,omitempty"`
	User_id int64 `json:"user_id,omitempty"`
}

type Callbackanswere struct {
	Callback_query_id string `json:"callback_query_id"`
	Show_alert        bool   `json:"show_alert,omitempty"`
	Cache_time        int16  `json:"cache_time,omitempty"`
	Text              string `json:"text,omitempty"`
}

// MarshalJSON custom marshal method for Callbackanswere
func (c *Callbackanswere) MarshalJSON() ([]byte, error) {
	bt := make([]byte, 0, 512)
	bt = append(bt, '{')
	bt = append(bt, `"callback_query_id":"`+c.Callback_query_id+`",`...)
	if c.Show_alert {
		bt = append(bt, `"show_alert":true,`...)
	}
	if c.Cache_time != 0 {
		bt = append(bt, `"cache_time":`+strconv.Itoa(int(c.Cache_time))+`,`...)
	}
	if c.Text != "" {
		bt = append(bt, `"text":"`+c.Text+`"`...)
	} else {
		bt = bt[:len(bt)-1] 
	}
	bt = append(bt, '}')
	return bt, nil
}

type Meadiacommon struct {
	//sending newly media
	Photo       string    `json:"photo,omitempty"`
	Video       string    `json:"video,omitempty"`
	Caption     string `json:"caption,omitempty"`
	Has_spoiler bool   `json:"has_spoiler,omitempty"`

	Media *InputMedia `json:"media,omitempty"`
}
//custom method only for this package not standard way
func (in *Meadiacommon) cMarshal(add bool, m *Msgcommon) { 
	if !add {
		in.Caption = "r"
	}
	l := len(m.content)
	if in.Photo != "" {
		m.content = append(m.content, `"photo":"`+in.Photo+`",`...)
	}
	if in.Video != "" {
		m.content = append(m.content, `"video":"`+in.Video+`",`...)
	}
	if in.Has_spoiler {
		m.content = append(m.content, `"has_spoiler":true,`...)
	}

	if in.Caption != "" && in.Media == nil {
		m.content = append(m.content, `"caption":"`...)
		if add {
			m.content = append(m.content, in.Caption...)
			m.content = append(m.content, `"}`...)
		}
	} else if len(m.content) > l  && in.Media == nil {
		m.content = m.content[:len(m.content)-1]
	}

	if in.Media != nil {
		in.Media.cMarshal(add, m)
	}
}



func (m *Msgcommon) SetMedType(medtype string, medid string) {
	if m.Meadiacommon == nil {
		return
	}

	switch medtype {
	case C.MedPhoto:
		m.Photo = medid
		m.Endpoint = C.ApiMethodSendPhoto
	case C.MedVideo:
		m.Video = medid
		m.Endpoint = C.ApiMethodSendVid
	}
}

// use only for editmeadia
type InputMedia struct {
	Type      string `json:"type"`
	Media     string `json:"media"`
	Caption   string `json:"caption,omitempty"`
	ParseMode string `json:"parse_mode,omitempty"`
}


//custom method only for this package not standard way
func (in *InputMedia) cMarshal(add bool, m *Msgcommon) { 
	//var jsot []byte
	m.content = append(m.content, `"media":{"type":"`+in.Type+`","media":"`+in.Media+`",` ...)
	if in.ParseMode != "" {
		m.content = append(m.content, `"parse_mode":"`+in.ParseMode+`",`...)
	}
	if !add {
		in.Caption = "r"
	}
	if in.Caption != "" {
		m.content = append(m.content, `"caption":"`...)
		if add {
			m.content = append(m.content, in.Caption...)
			m.content = append(m.content, `"}}`...)
		}
	} else {
		m.content[len(m.content)-1] = '}'
		m.content = append(m.content, '}')
	}
}


type Keyboard struct {
	Inline_keyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

type Msgcommon struct {
	*Infocontext
	Message_thread_id int64    `json:"message_thread_id,omitempty"`
	Text              string   `json:"text,omitempty"`
	Parse_mode        string   `json:"parse_mode,omitempty"`
	Reply_markup      Keyboard `json:"reply_markup,omitempty"`
	Message_id        int64    `json:"message_id,omitempty"`

	//meadia
	*Meadiacommon
	Endpoint string `json:"-"`
	
	//process read method
	reader io.Reader `json:"-"`
	content []byte `json:"-"`
	etcont []byte `json:"-"`
	inite bool  `json:"-"`
	rover bool `json:"-"`
	//etext *string  `json:"-"`
}

var comcont = []byte(`"}`)

func (m *Msgcommon) Reset() {
	m.inite = false
	m.rover = false
}

func (m *Msgcommon) Read(p []byte) (int, error) {
	if !m.inite {
		err := m.init()
		if err != nil {
			return 0, err
		}
	}
	if m.rover  {
		n := copy(p, m.etcont)
		m.etcont = m.etcont[n:]
		if len(m.etcont) == 0 {
			m.content = nil
			return n, io.EOF
		}
	}
	n := 0
	if len(m.content) != 0 {
		n = copy(p, m.content)
		m.content = m.content[n:]
		if len(m.content) != 0 || n == len(p) {
			return n, nil
		}
	}
	if len(m.content) == 0 && m.reader == nil {
		m.content = nil
		return n, io.EOF
	}
	e, rerr := m.reader.Read(p[n:])
	buf, ok := m.reader.(*bytes.Buffer); 
	if ok && buf.Len() == 0 || rerr !=nil{
		m.rover = true
		if n != len(p) {
			ee, err :=  m.Read(p[n+e:])
			return n+e+ee, err
		}
	}
	return n+e, nil
}
var ErrNilInfoCtx = errors.New("Infocontext cannot be nil")
func (m *Msgcommon) init() error {
	if !m.inite {
		m.inite = true
		m.etcont = comcont
		if m.Infocontext == nil {
			m.inite = false
			return ErrNilInfoCtx
		}
		m.content = make([]byte, 0, 512)
		m.content = append(m.content, '{')
		
		if m.ChatId != 0 {
			m.content = append(m.content, `"chat_id":` + strconv.Itoa(int(m.ChatId)) + `,`... )
		}
		if m.User_id != 0 {
			m.content = append(m.content, `"user_id":` + strconv.Itoa(int(m.User_id)) + `,`... )
		}
		if m.Message_id != 0 {
			m.content = append(m.content, `"message_id":` + strconv.Itoa(int(m.Message_id)) + `,`... )
		}
		if m.Message_thread_id != 0 {
			m.content = append(m.content, `"message_thread_id":` + strconv.Itoa(int(m.Message_thread_id)) + `,`... )
		}
		if m.Parse_mode != "" {
			m.content = append(m.content, `"parse_mode":"`+m.Parse_mode+`",`...)
		}
		if len(m.Reply_markup.Inline_keyboard) > 0 {
			m.content = append(m.content, `"reply_markup":{"inline_keyboard":`...)
			mp, _ := Createkeyboard(m.Reply_markup.Inline_keyboard)
			m.content = append(m.content, mp...)
			m.content = append(m.content, `},`...)
		}
		if m.reader != nil {
			m.Text = "r"
		} 
		if m.Meadiacommon != nil {
			m.cMarshal(m.reader==nil, m)
			if m.Meadiacommon.Media != nil && m.reader != nil {
				m.etcont = []byte(`"}}`)
			}
			if m.reader == nil {return nil}
		} else if m.Text != "" {
			m.content = append(m.content, `"text":"`...)
		} else {
			m.content[len(m.content)-1] = '}'
		}
		if m.reader == nil && m.Text != "" {
			m.content = append(m.content, m.Text+`"}`...)
		}
	}
	return nil
}
type length interface {
	Len() int
}

func (m *Msgcommon) Len() int {
	m.init()
	if m.reader != nil {
		if buf, ok := m.reader.(length); ok {
			return buf.Len() + len(m.content) + len(m.etcont)
		}
		return -1
	}
	return len(m.content)
}


// this struct used to send the msg to watman's mg que
// also support by botapi's Message session, when use with msg session no need to provide destination & langs
type UpMessage struct {
	DestinatioID int64
	Template     any
	TemplateName string
	Buttons      *Buttons
	Lang         string
}

type Filesend io.Reader

type BotReader struct {
	RealOb any
	called bool
	content []byte
	len int
	internalrd bool
	rd io.Reader
}


func (m *BotReader) Read(p []byte) (int, error) {
	if !m.called {
		if err := m.init(); err != nil {
			return 0, err
		}
	}
	if m.internalrd {
		return m.rd.Read(p)
	}
	n := copy(p, m.content)
	m.content = m.content[n:]
	if len(m.content) == 0 {
		return n, io.EOF
	}
	return n, nil
}

func (m *BotReader) init() error {
	var err error
	if m.internalrd {
		return nil
	}
	if !m.called {
		m.called = true
		m.rd, m.internalrd = m.RealOb.(io.Reader)
		if m.internalrd {
			return nil
		}
		m.content, err = json.Marshal(m.RealOb) //FIXME: change json 
		if err != nil {
			return err
		}
		m.len = len(m.content)
	}
	return nil
}
func (m *BotReader) Close() error {
	return nil
}

func (m *BotReader) Len() int {
	if !m.called {
		m.init()
	}
	if a, ok := m.RealOb.(length); ok {
		return a.Len()
	}
	return m.len
}


func CreateReder(botob any) *BotReader {
	return &BotReader{
		RealOb: botob,
	}
}