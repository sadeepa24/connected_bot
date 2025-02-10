package sboxoption

type VLESSInboundOptions struct {
	ListenOptions
	Users []VLESSUser `json:"users,omitempty"`
	InboundTLSOptionsContainer
	Multiplex *InboundMultiplexOptions `json:"multiplex,omitempty"`
	Transport *V2RayTransportOptions   `json:"transport,omitempty"`
}

func (v VLESSInboundOptions) GetPath() string {
	if v.Transport == nil {
		return ""
	}
	switch v.Transport.Type {
	case "ws":
		return v.Transport.WebsocketOptions.Path
	}
	return ""
}

func (v VLESSInboundOptions) TransportType() string {
	if v.Transport == nil {
		return ""
	}
	return v.Transport.Type
}

type VLESSUser struct {
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
	Flow     string `json:"flow,omitempty"`
	Maxlogin int    `json:"maxlogin,omitempty"`
}

type VLESSOutboundOptions struct {
	DialerOptions
	ServerOptions
	UUID    string      `json:"uuid"`
	Flow    string      `json:"flow,omitempty"`
	Network NetworkList `json:"network,omitempty"`
	OutboundTLSOptionsContainer
	Multiplex      *OutboundMultiplexOptions `json:"multiplex,omitempty"`
	Transport      *V2RayTransportOptions    `json:"transport,omitempty"`
	PacketEncoding *string                   `json:"packet_encoding,omitempty"`
}

func (v *VLESSOutboundOptions) ParseUrlQuary(quary map[string][]string) {
	v.Flow = gefirstval(quary["flow"])
	
	v.ServerOptions = ServerOptions{}
	
	if sni, ok := quary["sni"]; ok {
		v.TLS = &OutboundTLSOptions{
			Enabled: true,
			ServerName: gefirstval(sni),
			Insecure: true,
		
			UTLS: &OutboundUTLSOptions{
				Fingerprint: gefirstval(quary["fp"]),
			},
			ALPN: quary["alpn"],


		}
	}


}

func gefirstval(mp []string) string {
	if len(mp) > 0 {
		return mp[0]
	}
	return ""
}
