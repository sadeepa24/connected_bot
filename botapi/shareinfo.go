package botapi

import (
	"encoding/json"
	"io"

	C "github.com/sadeepa24/connected_bot/constbot"
)

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

	content []byte `json:"-"` //for Read method
	called  bool   `json:"-"` // for Read method

}

func (m *Callbackanswere) Read(p []byte) (int, error) {
	var err error
	if !m.called {
		m.content, err = json.Marshal(m)
		if err != nil {
			return 0, err
		}
		m.called = true
	}
	n := copy(p, m.content)
	m.content = m.content[n:]

	if len(m.content) == 0 {
		return n, io.EOF
	}
	return n, nil
}

func (m *Callbackanswere) Close() error {
	return nil
}

type Meadiacommon struct {
	//sending newly media
	Photo       any    `json:"photo,omitempty"`
	Video       any    `json:"video,omitempty"`
	Caption     string `json:"caption,omitempty"`
	Has_spoiler bool   `json:"has_spoiler,omitempty"`

	//this field for editing meadia
	Media *InputMedia `json:"media,omitempty"`
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

func (i *InputMedia) Marshal() json.RawMessage {
	if content, err := json.Marshal(i); err != nil {
		return json.RawMessage{}
	} else {
		return json.RawMessage(content)
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

	content  []byte `json:"-"` //for Read method
	called   bool   `json:"-"` // for Read method
	Endpoint string `json:"-"`
}

func (m *Msgcommon) Read(p []byte) (int, error) {
	var err error
	if !m.called {

		m.content, err = json.Marshal(m)

		if err != nil {
			return 0, err
		}
		m.called = true
	}
	n := copy(p, m.content)
	m.content = m.content[n:]

	if len(m.content) == 0 {
		return n, io.EOF
	}
	return n, nil
}

func (m *Msgcommon) Close() error {
	return nil
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