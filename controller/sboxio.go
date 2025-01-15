package controller

import option "github.com/sadeepa24/connected_bot/sbox_option/v1"

type SboxIO struct {
	Inbounds  []option.Inbound
	outbounds []option.Outbound
}
