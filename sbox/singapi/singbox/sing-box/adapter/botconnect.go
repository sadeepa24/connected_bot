package adapter

import "github.com/sagernet/sing-box/connectedbot"

type BotFace interface {
	AddUser(connectedbot.BotUser) (connectedbot.StatusOutput, error)
	RemoveUser(connectedbot.BotUser) (connectedbot.StatusOutput, error)
	GetastatusUser(connectedbot.BotUser) (connectedbot.StatusOutput, error)
	AddUserReset(connectedbot.BotUser) (connectedbot.StatusOutput, error)
	CloseAll(connectedbot.BotUser) error
}


type OutboundAdder interface {
	AddOutboundUser(string, string)
	RemoveAllRule(string)
}