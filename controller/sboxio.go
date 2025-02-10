package controller

import option "github.com/sagernet/sing-box/option"

type SboxIO struct {
	Inbounds  []option.Inbound
	outbounds []option.Outbound
}
