package constbot

import "time"

const (
	Groupservicename    string = "groupsrv"
	Adminservicename    string = "adminsrv"
	Callbackservicename string = "callback"
	Userservicename     string = "usersrv"
	Xraywizservicename  string = "xraywiz"
	Defaultservicename  string = "defaultsrv"
	InlineServiceName string = "inline"

	DefaultPoint int64 = 10

	Channel string = "channel"
	Group   string = "group"

	Inbound  string = "inbound"
	Outbound string = "outbound"

	Statusmember string = "member"
	Statusleft   string = "left"
	Statuskicked string = "kicked"

	AsKB Bwidth = 1024
	AsMB Bwidth = AsKB * AsKB
	AsGB Bwidth = AsKB * AsMB

	GBtoByte Bwidth = 1024 * 1024 * 1024
	MBtoByte Bwidth = 1024 * 1024
	KBtoByte Bwidth = 1024

	KB string = "KB"
	MB string = "MB"
	GB string = "GB"

	Dbbatchsize int = 100

	Vless  string = "vless"
	Trojan string = "trojan"
	Common string = "common"
	Direct string = "direct"

	UpdateTimeout = 4 * time.Minute

	MaxLoginLimit int = 5

	CmdConfigBlocks string = "confblocks"

	//ParseModes
	ParseHtml       string = "html"
	ParseMarkdown   string = "mdw1"
	ParseMarkdownv2 string = "mdw2"

	HelpPags    int16 = 3
	InfoPage    int16 = 4
	BuilderPage int16 = 5

	//Mediatypes
	MedVideo string = "video"
	MedPhoto string = "photo"
	MedAudio string = "audio"

	//apimethods
	ApiMethodCaptionEdit string = "editMessageCaption"
	ApiMethodEdimgmed    string = "editMessageMedia"
	ApiMethodEditText    string = "editMessageText"
	ApiMethodSendMG      string = "sendMessage"
	ApiMethodSendPhoto   string = "sendPhoto"
	ApiMethodSendVid     string = "sendVideo"
	ApiMethodAnswereInline     string = "answerInlineQuery"

	SkipDelim string = "."

	MaxCharacterMg = 4096
)
