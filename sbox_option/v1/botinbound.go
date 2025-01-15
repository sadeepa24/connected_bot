package sboxoption

import (
	"errors"

	"github.com/gofrs/uuid"
	C "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
)

func (h *Inbound) Change(changer string, value string) error {
	// use reflect later
	if changer == "uuid" {
		_, err := uuid.FromString(value)
		if err != nil {
			return errors.New("uuid parsing error, send a valid uuid")
		}
	}

	switch h.Type {
	case C.TypeTun:
		return h.TunOptions.Changer(changer, value)
	case "":
		return E.New("missing outbound type")
	default:
		return E.New("unknown outbound type or unchangble for now: ", h.Type)
	}

}

func (h *Inbound) ChangeInboundOpts(changer string, value string) error {
	// use reflect later
	if changer == "uuid" {
		_, err := uuid.FromString(value)
		if err != nil {
			return errors.New("uuid parsing error, send a valid uuid")
		}
	}

	switch h.Type {
	case C.TypeTun:
		return h.TunOptions.InboundOptions.Changer(changer, value)
	case "":
		return E.New("missing outbound type")
	default:
		return E.New("unknown outbound type or not changble yet: ", h.Type)
	}

}
