package constbot

import "errors"


type Error interface {
	error
	Exit() bool
	UserMsg() string
}



type customErr struct {
	error
	exit bool
	msg string

}
func (c customErr) UserMsg() string { return c.msg }
func (c customErr) Exit() bool { return c.exit }
func WrapError(err error, usermsg string) error {
	return customErr{
		error: err,
		msg: usermsg,
	}
}


var (
	CErrConfigNotFound =  WrapError(ErrConfigNotFound, "Can't Find Config Please Try Again")
	CErrQuotaExceed  = WrapError(ErrQuotaExceed, "Cannot Activate Config: Config's Quota Already Exceed")
	CErrContextDead = WrapError(ErrContextDead, "Context DeadLine: Maybe Session Timeout")
	CErrRetryFailed = WrapError(ErrRetryFailed, "Don't Be idiot :(")
	CErrNoPerm = WrapError(ErrNoPerm, "Operation not permitted. Please check your account status.")
	CErrNoPoints = WrapError(errors.New("no points"), "Cannot Pause Service User Does not have enogf points")
	CErrDbopration = WrapError(ErrDbopration, "Operation on db failed, may be try again later")
)



var ErrNoPerm = errors.New("operation not permitted due to user's current state")
var ErrContextDead = errors.New("context cancled")
var ErrConfigNotFound = errors.New("config Notfound")
var ErrInboundNotFound = errors.New("inbound Notfound")
var ErrOutboundNotFound = errors.New("outbound Notfound")
var ErrResultMalformed = errors.New("status result malformed")
var ErrNotMsgType = errors.New("type is not valid messagetype")
var ErrRetryFailed = errors.New("retry attemps failed")


var ErrTypeMissmatch = errors.New("iobound type not supported or invalid type")

var ErrChatOrUserNofound = errors.New("chat or user not found")
var ErrBtnClosed = errors.New("user closed btn")

var ErrClosedSession = errors.New("session already closed")

// Databbase
var ErrDatabaseCreate = errors.New("cannot create record")
var ErrDatabasefuncer = errors.New("database calling error")
var ErrDbnotfound = errors.New("user cannot find in db")
var ErrDbopration = errors.New("database opration or transaction failed")

var ErrRequest = errors.New("request errored")
var ErrClientRequestFail = errors.New("requests send failed on client side")
var ErrResponseMissmatch = errors.New("telegram responsed diffrent status code")

var ErrRead = errors.New("body reading error")
var ErrApierror = errors.New("request not resolverd by server ")

var ErrJsonopra = errors.New("json marshling error")
var ErrTgParsing = errors.New("parsing error from telegram")
// parser
var ErrServiceNotFound = errors.New("service not found")

// Usersession
var ErrOnDeactivation = errors.New("deactivation failed")
var ErrQuotaExceed = errors.New("config quota exceed")
var ErrSessionExcit = errors.New("already session excist")
var ErrSessioForceClose = errors.New("old session force closing errored ")
var Erruuidcreatefailed = errors.New("uuid create failed")


var ErrConfigNoQota = errors.New("config doesn't have any quota")

// Metadata
var ErrNotsupported = errors.New("type not supporte yet")

// Gift
var ErrUserCanootReciveUserCapped = errors.New("user cannot recive gifts")
var ErrUserCappedUser = errors.New("user is capped")
var ErrUserGiftAlready = errors.New("user already recived or send a gift")

// Parser
var (
	ErrCommandNotfound    = errors.New("command not found")
	ErrUserMonthLimited   = errors.New("this month limited for the user")
	ErrUserNotVerified    = errors.New("user is not verified user")
	ErrUserIsNotinPrivate = errors.New("user is not private")
	ErrUserIsRestricted = errors.New("restricted user")
	ErrUserTempLimited = errors.New("templimited user")

	ErrUpdateFaile = errors.New("recived updated has nothing to process")
)



// Messagesession
var (
	ErrTmplRender = errors.New("render template failed")
)

// refral
var (
	ErrUserExitDb = errors.New("user already in database")
)

// Configure command
var (
	ErrRecurtionExceed = errors.New("recurtion limit hit")
)

//template

var (
	ErrMsgDisabled = errors.New("msg disabled")
)


var ErrWebhookSetFailed = errors.New("setting web hook failed")
var ErrNilRequest = errors.New("request is nil pointer")
var ErrUnknownUserListType = errors.New("user list type error")
var ErrUserObNil = errors.New("user struct cannot be nil")
var ErrNoService = errors.New("no service")
var ErrNoAnyInbound = errors.New("no inbound")