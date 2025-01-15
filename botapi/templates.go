package botapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	C "github.com/sadeepa24/connected_bot/constbot"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

//io.Writer for excution
type Message struct {
	Msg                string
	ParseMode          string
	Includemed         bool
	MedType            string
	MediaId            string
	ContinueMed        bool
	Disabled           bool
	SkipText           bool
	Continue_Skip_Text bool
	SuperContinue      bool

	MeadiaSkip bool
}

func (m *Message) String() string {
	return m.Msg
}

func (m *Message) Write(p []byte) (n int, err error) {
	m.Msg = m.Msg + string(p)
	return len(p), nil
}

type MgItem struct {
	Msgtmpl            string             `json:"msg_template" yaml:"msg_template"`
	ParseMode          string             `json:"parse_mode" yaml:"parse_mode" `
	Includemed         bool               `json:"include_media" yaml:"include_media"`
	Mediatype          string             `json:"media_type" yaml:"media_type"`
	MediaId            string             `json:"media_id" yaml:"media_id" `
	tmpl               *template.Template `json:"-" yaml:"-"`
	ContinueMed        bool               `json:"continue_media" yaml:"continue_media"`
	Disabled           bool               `json:"disabled" yaml:"disabled"`
	SkipText           bool               `json:"skip_text" yaml:"skip_text"`
	Continue_Skip_Text bool               `json:"contin_skip_text" yaml:"contin_skip_text"`
	SuperContinue      bool               `json:"supercontinue" yaml:"supercontinue"`
	AltMediaUrl        string 			  `json:"alt_med_url" yaml:"alt_med_url"`
	AltMediaPath       string 			  `json:"alt_med_path" yaml:"alt_med_path"`

	MeadiaSkip		   bool  			  `json:"media_skip" yaml:"media_skip"`

}

type MessageStore struct {
	Templates map[string]*MgItem
}

func NewMessageStore(path string) (*MessageStore, error) {
	
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var messages map[string]map[string]MgItem

	switch {
	case strings.Contains(path, ".yaml"):
		err = yaml.Unmarshal(file, &messages)
	case strings.Contains(path, ".json"):
		err = json.Unmarshal(file, &messages)
	}

	if err != nil {
		return nil, err
	}
	
	mgmap := map[string]*MgItem{}

	for name, allmg := range messages {
		for langcode, item := range allmg {
			t, err := template.New(name + langcode).Parse(item.Msgtmpl)
			if err != nil {
				return nil, err
			}
			item.tmpl = t

			switch item.ParseMode {
			case "HTML":
				item.ParseMode = C.ParseHtml
			case "MarkdownV2":
				item.ParseMode = C.ParseMarkdownv2
			case "Markdown":
				item.ParseMode = C.ParseMarkdown
			default:
				item.ParseMode = ""
			}

			mgmap[name+langcode] = &item
		}
	}

	return &MessageStore{
		Templates: mgmap,
	}, nil
}

// this verify that all media can sendble
func (m *MessageStore) Init(sender BotAPI, sudoadminID int64, logger *zap.Logger) error {
	cacheMap :=  map[string]tgbotapi.Message{}
	tmplLoop:
	for key, template := range m.Templates {
		if template.Includemed {
			//var filecontent []byte
			var ( 
			 	//err error
			 	toreq io.Reader
			 	endpoint string	
			)

			ContentType := "application/json"

			var forcheck string

			if template.AltMediaPath != "" {

				rawfile, err := os.ReadFile(template.AltMediaPath) 
				if err != nil {
					logger.Error(err.Error())
					continue tmplLoop
				}
				var (
					body bytes.Buffer
				)
				forcheck = template.AltMediaPath
				
				multiparwriter := multipart.NewWriter(&body)

				var (
					filepart io.Writer
				)

				switch template.Mediatype {
				case C.MedPhoto:
					filepart, err = multiparwriter.CreateFormFile("photo",  key+".jpg")
					endpoint = C.ApiMethodSendPhoto
				case C.MedVideo:
					filepart, err = multiparwriter.CreateFormFile("video",  key+".mp4")
					endpoint = C.ApiMethodSendVid
				default:
					continue tmplLoop
				}

				if err != nil {
					logger.Error(err.Error())
					continue tmplLoop
				}
				if _, err = filepart.Write(rawfile); err != nil {
					logger.Error(err.Error())
					continue tmplLoop
				}

				err = multiparwriter.WriteField("chat_id", strconv.Itoa(int(sudoadminID)))
				if err != nil {
					logger.Error(err.Error())
					continue tmplLoop
				}

				multiparwriter.Close()
				toreq = &body
				ContentType = multiparwriter.FormDataContentType()

			} else if template.AltMediaUrl != "" {
				mg := &Msgcommon{
					Infocontext: &Infocontext{
						ChatId: sudoadminID,
					},
					Meadiacommon: &Meadiacommon{},
				}
				
				switch template.Mediatype {
				case C.MedPhoto:
					mg.Photo = template.AltMediaUrl
					endpoint = C.ApiMethodSendPhoto
				case C.MedVideo:
					mg.Video = template.AltMediaUrl
					endpoint = C.ApiMethodSendVid
				}
				forcheck = template.AltMediaUrl
				toreq = mg

			} else {
				continue tmplLoop
			}

			var message tgbotapi.Message

			if mg, ok := cacheMap[forcheck]; ok {
				message = mg
			} else {

				ctx, canc := context.WithTimeout(context.Background(), 10  * time.Second)
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.CreateFullUrl(endpoint), toreq)
				if err != nil {
					canc()
					logger.Error(err.Error())
					continue tmplLoop
				}
				
				req.Header.Set("Content-Type", ContentType)
				
				res, err := sender.SendRawReq(req)
				canc()
				if err != nil {
					logger.Error(err.Error())
					continue tmplLoop
				}
				if err = json.Unmarshal(res.Result, &message); err != nil {
					
					logger.Error("marshell error - "+ err.Error())
					continue tmplLoop
				}
				cacheMap[forcheck] = message

			}
			switch template.Mediatype {
			case C.MedPhoto:
				if len(message.Photo) > 0 {
					template.MediaId = message.Photo[len(message.Photo)-1].FileID
				}
			case C.MedVideo:
				if message.Video != nil {
					template.MediaId = message.Video.FileID
				}
			}


		}
	}

	return nil
}

func (m *MessageStore) GetMessage(name, lang string, obj any) (*Message, error) {
	if obj == nil {
		return nil, errors.New("nil object ")
	}
	if lang == "" {
		lang = "en"
	}
	msg, ok := m.Templates[name+lang]

	mg := &Message{
		Msg: "",
	}

	if !ok {
		mg.Msg = "sorry no template found may be change the lang it will help message can't be displayed"
		return mg, nil
	}
	if msg.Disabled {
		return nil, C.ErrApierror
	}

	mg.ParseMode = msg.ParseMode
	mg.SkipText = msg.SkipText
	mg.MeadiaSkip = msg.MeadiaSkip

	if msg.Includemed {
		mg.MedType = msg.Mediatype
		mg.MediaId = msg.MediaId
		mg.ContinueMed = msg.ContinueMed
		mg.Includemed = msg.Includemed
		mg.Continue_Skip_Text = msg.Continue_Skip_Text
		mg.SuperContinue = msg.SuperContinue
	}

	return mg, msg.tmpl.Execute(mg, obj)

}

func (m *MessageStore) MsgWithouerro(name, lang string, obj any) string {
	mg, err := m.GetMessage(name, lang, obj)
	if err != nil {
		//TODO: should return  default mgs
		return " errored from template will replace"
	}
	return mg.String()
}

// common input types
// All of type below are for template render
type CommonUser struct {
	Name     string
	Username string
	TgId     int64
}

type CommonUsage struct {
	AddtionalQuota  string
	CalculatedQuota string
	MUsage          string
	Alltime         string
}

type UsageAll struct {
}
