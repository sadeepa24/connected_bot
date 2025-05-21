package builder

import (
	"context"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
)

var globalCtx context.Context = box.Context(context.Background(), include.InboundRegistry(), include.OutboundRegistry(), include.EndpointRegistry())