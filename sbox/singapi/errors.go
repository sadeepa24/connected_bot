package singapi

import C "github.com/sadeepa24/connected_bot/constbot"
type Error struct {
	usermsg string
	boxerr  bool
	error
}

func (e Error) IsBoxErr() bool {
	return e.boxerr
}
func (e Error) UserMsg() string {
	return e.usermsg
}

var ErrInboundNotFound = Error{
	usermsg: "VPN Error: Internal VPN Fault Inbound Cannot be found try changing you'r inbounds or contact admin",
	error: C.ErrInboundNotFound,
}

var ErrOutboundNotFound = Error{
	usermsg: "VPN Error: Internal VPN Fault Outbound Cannot be found try changing you'r outbound or contact admin",
	error: C.ErrOutboundNotFound,
}

var ErrConfigNoQuota = Error{
	usermsg: "You Don't Have Any Leftquota To add this config",
	error: C.ErrConfigNoQota,
}

func boxerror(err error) error {
	if err == nil {
		return nil
	}
	return Error{
		error: err,
		usermsg: "Internal Box Error",
		boxerr: true,
	}

}