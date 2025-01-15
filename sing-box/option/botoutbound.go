package option

import (
	C "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
)

//sni servername uuid or password in trojan
func (h *Outbound) Set(sni string, uuid string) error {
	switch h.Type {
	case C.TypeVMess:
		h.VMessOptions.Transport.SetHost(sni)
		
		if h.VMessOptions.TLS != nil && sni != "" {
			h.VMessOptions.TLS.ServerName = sni
		}
		if uuid != "" {
			h.VMessOptions.UUID = uuid
		}
		
	case C.TypeTrojan:
		h.TrojanOptions.Transport.SetHost(sni)
		if h.TrojanOptions.TLS != nil && sni != "" {
			h.TrojanOptions.TLS.ServerName = sni
		}
		if uuid != "" {
			h.TrojanOptions.Password = uuid
		}
		

	case C.TypeVLESS:
		h.VLESSOptions.Transport.SetHost(sni)
		if h.VLESSOptions.TLS != nil && sni != "" {
			h.VLESSOptions.TLS.ServerName = sni
		}
		if uuid != "" {
			h.VLESSOptions.UUID = uuid
		}
		
		
	case "":
		return E.New("missing outbound type")
	default:
		return E.New("unknown outbound type: ", h.Type)
	}

	return nil
}

func (h *Outbound) SetServer(srv string) error {
	switch h.Type {
	case C.TypeVMess:
		h.VMessOptions.Server = srv

	case C.TypeTrojan:
		h.TrojanOptions.Server = srv
	
	case C.TypeVLESS:
		h.VLESSOptions.Server = srv
	
	case C.TypeShadowTLS:
		h.ShadowTLSOptions.Server = srv
	
	case C.TypeHysteria:
		h.HysteriaOptions.Server = srv
	
	case C.TypeHysteria2:
		h.Hysteria2Options.Server = srv
	
	default:
		return E.New("unknown type ")
	
	} 
	return nil
}

func (h *Outbound) SetPort(port uint16) error {
	switch h.Type {
	case C.TypeVMess:
		h.VMessOptions.ServerPort = port

	case C.TypeTrojan:
		h.TrojanOptions.ServerPort = port
	
	case C.TypeVLESS:
		h.VLESSOptions.ServerPort = port
	
	case C.TypeShadowTLS:
		h.ShadowTLSOptions.ServerPort = port
	
	case C.TypeHysteria:
		h.HysteriaOptions.ServerPort = port
	
	case C.TypeHysteria2:
		h.Hysteria2Options.ServerPort = port
	
	default:
		return E.New("unknown type ")
	
	} 
	return nil
}

func (h *V2RayTransportOptions) SetHost(host string) error {
	if host == "" {
		return nil
	}
	switch h.Type {
	case C.V2RayTransportTypeWebsocket:
		h.WebsocketOptions.Headers["host"] = Listable[string]{host}
	case C.V2RayTransportTypeHTTP:
		h.HTTPOptions.Host = Listable[string]{host}
	case C.V2RayTransportTypeGRPC:
		h.GRPCOptions.ServiceName = host
	}
	return nil
}

func (h *Outbound) SetTLS(tls *OutboundTLSOptions) error {
	switch h.Type {
	case C.TypeVMess:
		h.VMessOptions.TLS = tls
	case C.TypeTrojan:
		h.TrojanOptions.TLS = tls
	case C.TypeVLESS:
		h.VLESSOptions.TLS = tls
	case C.TypeShadowTLS:
		h.ShadowTLSOptions.TLS = tls
	case C.TypeHysteria:
		h.HysteriaOptions.TLS = tls
	case C.TypeHysteria2:
		h.Hysteria2Options.TLS = tls
	default:
		return E.New("unknown type ")
	
	}

	return nil
}

func (h *Outbound) SetTransPort(transport *V2RayTransportOptions) error {
	switch h.Type {
	case C.TypeVMess:
		h.VMessOptions.Transport = transport
	case C.TypeTrojan:
		h.TrojanOptions.Transport = transport
	case C.TypeVLESS:
		h.VLESSOptions.Transport = transport
	default:
		return E.New("unknown type ")
	
	}

	return nil
}