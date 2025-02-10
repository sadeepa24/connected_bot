package sboxoption

import (
	"errors"
	"strconv"

	"github.com/gofrs/uuid"
	C "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
)

// sni servername uuid or password in trojan
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

func (h *Outbound) SetTransPortHost(transportHost string) error {
	switch h.Type {
	case C.TypeVLESS:
		if h.VLESSOptions.Transport == nil {
			return nil
		}
		switch h.VLESSOptions.Transport.Type {
		case C.V2RayTransportTypeHTTPUpgrade:
			h.VLESSOptions.Transport.HTTPUpgradeOptions.Host = transportHost
		case C.V2RayTransportTypeWebsocket:
			h.VLESSOptions.Transport.WebsocketOptions.Headers = HTTPHeader(map[string]Listable[string]{"host": {transportHost}})
		case C.V2RayTransportTypeGRPC:
			h.VLESSOptions.Transport.GRPCOptions.ServiceName = transportHost
		}
	default:
		return E.New("unknown type ")

	}

	return nil
}



func (h *Outbound) Change(changer string, value string) error {
	// use reflect later
	if changer == "uuid" {
		_, err := uuid.FromString(value)
		if err != nil {
			return errors.New("uuid parsing error, send a valid uuid")
		}
	}
	
	switch h.Type {
	case C.TypeVMess:
		out := &h.VMessOptions
		switch changer {
		case"uuid":
			out.UUID = value
		case "server":
			out.Server = value
		case "server_port":
			pt, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			out.ServerPort = uint16(pt)
		default:
			return errors.New("unsupported change field for protole")
		}
		
	case C.TypeTrojan:
		out := &h.TrojanOptions
		switch changer {
		case "Password", "password":
			out.Password = value
		case "server":
			out.Server = value
		case "server_port", "Port":
			pt, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			out.ServerPort = uint16(pt)
		default:
			return errors.New("unsupported change field for protocole")
		}
		return errors.New("protocole not supported for changers for now")

	case C.TypeVLESS:
		out := &h.VLESSOptions
		switch changer {
		case "uuid":
			out.UUID = value
		case"flow":
			out.Flow = value
		case "server":
			out.Server = value
		case "server_port":
			pt, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			out.ServerPort = uint16(pt)
		default:
			return errors.New("unsupported change field for protole")
		}

	case "":
		return E.New("missing outbound type")
	default:
		return E.New("unknown outbound type: ", h.Type)
	}

	return nil
}

func (h *Outbound) ChangeTLS(changer string, value string) error {
	// use reflect later
	switch h.Type {
	case C.TypeVMess:
		return h.VMessOptions.OutboundTLSOptionsContainer.Changer(changer, value)
	case C.TypeTrojan:
		return h.TrojanOptions.OutboundTLSOptionsContainer.Changer(changer, value)
	case C.TypeVLESS:
		return h.VLESSOptions.OutboundTLSOptionsContainer.Changer(changer, value)
		
	case "":
		return E.New("missing or unsupported outbound type for now")
	default:
		return E.New("unknown outbound type: ", h.Type)
	}

}
func (h *Outbound) ChangeDialer(changer string, value string) error {
	// use reflect later
	switch h.Type {
	case C.TypeVMess:
		return h.VMessOptions.DialerOptions.Changer(changer, value)
	case C.TypeTrojan:
		return h.TrojanOptions.DialerOptions.Changer(changer, value)
	case C.TypeVLESS:
		return h.VLESSOptions.DialerOptions.Changer(changer, value)
		
	case "":
		return E.New("missing or unsupported outbound type for now")
	default:
		return E.New("unknown outbound type: ", h.Type)
	}

}

func (h *Outbound) ChangeMultiplex(changer string, value string) error {
	// use reflect later
	switch h.Type {
	case C.TypeVMess:
		if h.VMessOptions.Multiplex != nil {
			h.VMessOptions.Multiplex = &OutboundMultiplexOptions{}
		}
		return h.VMessOptions.Multiplex.Change(changer, value)
	case C.TypeTrojan:
		if h.TrojanOptions.Multiplex != nil {
			h.TrojanOptions.Multiplex = &OutboundMultiplexOptions{}
		}
		return h.TrojanOptions.Multiplex.Change(changer, value)
	case C.TypeVLESS:
		if h.VLESSOptions.Multiplex != nil {
			h.VLESSOptions.Multiplex = &OutboundMultiplexOptions{}
		}
		return h.VLESSOptions.Multiplex.Change(changer, value)
		
	case "":
		return E.New("missing or unsupported outbound type for now")
	default:
		return E.New("unknown outbound type: ", h.Type)
	}

}


func (h *Outbound) ChangeTransPort(changer string, value string) error {
	// use reflect later
	switch h.Type {
	case C.TypeVMess:
		return h.VMessOptions.Transport.Changer(changer, value)

	case C.TypeTrojan:
		return h.TrojanOptions.Transport.Changer(changer, value)

	case C.TypeVLESS:
		return h.VLESSOptions.Transport.Changer(changer, value)

	case "":
		return E.New("missing outbound type")
	default:
		return E.New("unknown outbound type: ", h.Type)
	}

}


