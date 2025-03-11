package sboxoption

import (
	"errors"
	"strconv"

	C "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
)

type _V2RayTransportOptions struct {
	Type               string                  `json:"type"`
	HTTPOptions        V2RayHTTPOptions        `json:"-"`
	WebsocketOptions   V2RayWebsocketOptions   `json:"-"`
	QUICOptions        V2RayQUICOptions        `json:"-"`
	GRPCOptions        V2RayGRPCOptions        `json:"-"`
	HTTPUpgradeOptions V2RayHTTPUpgradeOptions `json:"-"`
}

type V2RayTransportOptions _V2RayTransportOptions

func (o V2RayTransportOptions) MarshalJSON() ([]byte, error) {
	var v any
	switch o.Type {
	case C.V2RayTransportTypeHTTP:
		v = o.HTTPOptions
	case C.V2RayTransportTypeWebsocket:
		v = o.WebsocketOptions
	case C.V2RayTransportTypeQUIC:
		v = o.QUICOptions
	case C.V2RayTransportTypeGRPC:
		v = o.GRPCOptions
	case C.V2RayTransportTypeHTTPUpgrade:
		v = o.HTTPUpgradeOptions
	case "":
		return nil, E.New("missing transport type")
	default:
		return nil, E.New("unknown transport type: " + o.Type)
	}
	return MarshallObjects((_V2RayTransportOptions)(o), v)
}

func (o *V2RayTransportOptions) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, (*_V2RayTransportOptions)(o))
	if err != nil {
		return err
	}
	var v any
	switch o.Type {
	case C.V2RayTransportTypeHTTP:
		v = &o.HTTPOptions
	case C.V2RayTransportTypeWebsocket:
		v = &o.WebsocketOptions
	case C.V2RayTransportTypeQUIC:
		v = &o.QUICOptions
	case C.V2RayTransportTypeGRPC:
		v = &o.GRPCOptions
	case C.V2RayTransportTypeHTTPUpgrade:
		v = &o.HTTPUpgradeOptions
	default:
		return E.New("unknown transport type: " + o.Type)
	}
	err = UnmarshallExcluded(bytes, (*_V2RayTransportOptions)(o), v)
	if err != nil {
		return err
	}
	return nil
}


func (o *V2RayTransportOptions) Changer(changer, value string) error {
	switch o.Type {
	case C.V2RayTransportTypeHTTP:
		
	case C.V2RayTransportTypeWebsocket:
		return o.WebsocketOptions.Change(changer, value)
	case C.V2RayTransportTypeQUIC:
		
		return E.New("cannot change yet " + o.Type)
	case C.V2RayTransportTypeGRPC:
		return E.New("cannot change yet " + o.Type)
	case C.V2RayTransportTypeHTTPUpgrade:
		return E.New("cannot change yet " + o.Type)
	default:
		return E.New("unknown transport type: " + o.Type)
	}
	return nil
}





type V2RayHTTPOptions struct {
	Host        Listable[string] `json:"host,omitempty"`
	Path        string           `json:"path,omitempty"`
	Method      string           `json:"method,omitempty"`
	Headers     HTTPHeader       `json:"headers,omitempty"`
	IdleTimeout Duration         `json:"idle_timeout,omitempty"`
	PingTimeout Duration         `json:"ping_timeout,omitempty"`
}


func (w *V2RayHTTPOptions) Change(chg string, value string) error {
	

	switch chg {
	case "host":
		w.Host = append(w.Host, value)
	case "path":
	  w.Path = value
	case "method":
	  w.Method = value
	case "idle_timeout":
	  return w.IdleTimeout.Set(value)
	case "ping_timeout":
	  return w.PingTimeout.Set(value)
	default:
	  return errors.New("unsupported change field: " + chg)
	}
	return nil
  }



type V2RayWebsocketOptions struct {
	Path                string     `json:"path,omitempty"`
	Headers             HTTPHeader `json:"headers,omitempty"`
	MaxEarlyData        uint32     `json:"max_early_data,omitempty"`
	EarlyDataHeaderName string     `json:"early_data_header_name,omitempty"`
}

func (w *V2RayWebsocketOptions) Change(chg, val string) error {
	switch chg {
	case "path":
		w.Path = val
	case "host":
		w.Headers["host"] = Listable[string]{
			val,
		}
	default:
		return errors.New("unsupported change field to transport")
	}
	return nil
}

type V2RayQUICOptions struct{}

type V2RayGRPCOptions struct {
	ServiceName         string   `json:"service_name,omitempty"`
	IdleTimeout         Duration `json:"idle_timeout,omitempty"`
	PingTimeout         Duration `json:"ping_timeout,omitempty"`
	PermitWithoutStream bool     `json:"permit_without_stream,omitempty"`
	ForceLite           bool     `json:"-"` // for test
}

func (w *V2RayGRPCOptions) Change(chg string, value string) error {
	switch chg {
	case "service_name":
		w.ServiceName = value
	case "idle_timeout":
		return w.IdleTimeout.Set(value)
	case "ping_timeout":
		return w.PingTimeout.Set(value)
	case "permit_without_stream":
		var err error
		w.PermitWithoutStream, err = strconv.ParseBool(value)
		if err != nil {
			return errors.New("invalid value for permitWithoutStream: expected true or false")
		}
	default:
		return errors.New("unsupported change field: " + chg)
	}
	return nil
}


type V2RayHTTPUpgradeOptions struct {
	Host    string     `json:"host,omitempty"`
	Path    string     `json:"path,omitempty"`
	Headers HTTPHeader `json:"headers,omitempty"`
}

func (w *V2RayHTTPUpgradeOptions) Change(chg, val string) error {
	switch chg {
	case "path":
		w.Path = val
	case "host":
		w.Headers["host"] = Listable[string]{
			val,
		}
	default:
		return errors.New("unsupported change field to transport")
	}
	return nil
}
