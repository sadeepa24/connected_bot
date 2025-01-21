package box

import (
	"errors"
	"fmt"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/connectedbot"
	"github.com/sagernet/sing-vmess/vless"
)

func (b *Box) AddUser(BotUser connectedbot.BotUser, inboundtag string) (connectedbot.StatusOutput, error) {
	for _, in := range b.inbounds {
		if in.Tag() == inboundtag {
			inb, ok := in.(adapter.BotFace)
			if !ok {
				continue
			}
			return inb.AddUser(BotUser)
		}
	}
	return connectedbot.StatusOutput{}, fmt.Errorf("inbound not found")
}

func (b *Box) RemoveUser(BotUser connectedbot.BotUser, inboundtag string) (connectedbot.StatusOutput, error) {
	for _, in := range b.inbounds {
		if in.Tag() == inboundtag {
			inb, ok := in.(adapter.BotFace)
			if !ok {
				continue
			}
			return inb.RemoveUser(BotUser)
		}
	}
	return connectedbot.StatusOutput{}, fmt.Errorf("inbound not found")

}
func (b *Box) GetastatusUser(BotUser connectedbot.BotUser, inboundtag string) (connectedbot.StatusOutput, error) {
	for _, in := range b.inbounds {
		if in.Tag() == inboundtag {
			inb, ok := in.(adapter.BotFace)
			if !ok {
				continue
			}
			return inb.GetastatusUser(BotUser)
		}
	}
	return connectedbot.StatusOutput{}, fmt.Errorf("inbound not found")

}

func (b *Box) AddUserReset(BotUser connectedbot.BotUser, inboundtag string) (connectedbot.StatusOutput, error) {
	for _, in := range b.inbounds {
		if in.Tag() == inboundtag {
			inb, ok := in.(adapter.BotFace)
			if !ok {
				continue
			}
			return inb.AddUserReset(BotUser)
		}
	}
	return connectedbot.StatusOutput{}, vless.ErrInboundNotFound

}

func (b *Box) CloseAll(BotUser connectedbot.BotUser, inboundtag string)  error {
	for _, in := range b.inbounds {
		if in.Tag() == inboundtag {
			inb, ok := in.(adapter.BotFace)
			if !ok {
				continue
			}
			return inb.CloseAll(BotUser)
		}
	}
	return vless.ErrInboundNotFound

}



func (b *Box) Addoutbounduser(user, outbound string) {
	adder ,ok := b.router.(adapter.OutboundAdder)
	
	
	if ok {
		adder.AddOutboundUser(user, outbound)
	}
}

func (b *Box) RemoveAllRule(user string) {
	adder ,ok := b.router.(adapter.OutboundAdder)
	if ok {
		adder.RemoveAllRule(user)
	}
}

func (b *Box) GetOutBound(tag string) (adapter.Outbound, error) {
	for _, out := range b.outbounds {
		if out.Tag() == tag {
			return out, nil
		}
	}
	return nil, errors.New("outboun d not found")
}