package sbox

// Xray Sing-box both should implement this
type Sboxcontroller interface {
	Start() error
	Close() error

	//Check If already user is in if not add the user
	//Should Calculate How much quota should added to underlaying sbox
	// Always update Usage to 0 when this called
	AddUser(*Userconfig) (Sboxstatus, error)
	AddUserReset(*Userconfig) (Sboxstatus, error)
	RemoveUser(*Userconfig) (Sboxstatus, error)
	RemoveAllRuleForuser(user string)
	//Do not going to update database using this func usage it will automaticaly doing by watchman
	//Userstatus
	GetstatusUser(*Userconfig) (Sboxstatus, error)

	// AddInboud()

	GetAllInbound() ([]Inboud, error)
	AddInbound() error
	RemoveInboud() error
	InboundStatus(string) error

	GetAllOutbound() ([]Outbound, error)
	AddOutbound() error
	RemoveOutboud() error
	OutboundStatus(string) error

	ShareLinkEncode(*Userconfig, string) (string, error)

	SetLogChan(chan any)
	GetLogChan() chan any

	UrlTest(outtag string) (int16, error)
	RefreshUrlTest()

	CloseConns(suser *Userconfig) error
}
