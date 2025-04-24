package sboxoption

type RedirectInboundOptions struct {
	ListenOptions
}

type TProxyInboundOptions struct {
	ListenOptions
	Network NetworkList `json:"network,omitempty"`
}
