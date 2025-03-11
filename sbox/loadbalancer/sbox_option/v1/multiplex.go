package sboxoption

import (
	"errors"
	"strconv"
)

type InboundMultiplexOptions struct {
	Enabled bool           `json:"enabled,omitempty"`
	Padding bool           `json:"padding,omitempty"`
	Brutal  *BrutalOptions `json:"brutal,omitempty"`
}

type OutboundMultiplexOptions struct {
	Enabled        bool           `json:"enabled,omitempty"`
	Protocol       string         `json:"protocol,omitempty"`
	MaxConnections int            `json:"max_connections,omitempty"`
	MinStreams     int            `json:"min_streams,omitempty"`
	MaxStreams     int            `json:"max_streams,omitempty"`
	Padding        bool           `json:"padding,omitempty"`
	Brutal         *BrutalOptions `json:"brutal,omitempty"`
}


func (o *OutboundMultiplexOptions) Change(chg string, value string) error {
	var err error 
	
	switch chg {
	case "enabled":
		o.Enabled, err = strconv.ParseBool(value)
		if err != nil {
				return errors.New("invalid value for enabled: expected true or false")
		}
	case "protocol":
		o.Protocol = value
	case "max_connections":
		o.MaxConnections, err = strconv.Atoi(value)
		if err != nil {
				return errors.New("invalid value for maxConnections: expected integer")
		}
	case "min_streams":
		o.MinStreams, err = strconv.Atoi(value)
		if err != nil {
				return errors.New("invalid value for minStreams: expected integer")
		}
	case "max_streams":
		o.MaxStreams, err = strconv.Atoi(value)
		if err != nil {
				return errors.New("invalid value for maxStreams: expected integer")
		}
	case "padding":
		o.Padding, err = strconv.ParseBool(value)
		if err != nil {
				return errors.New("invalid value for padding: expected true or false")
		}
	default:
		return errors.New("unsupported change field: " + chg)
	}
	if err != nil {
			return err 
	}
	return nil
}



type BrutalOptions struct {
	Enabled  bool `json:"enabled,omitempty"`
	UpMbps   int  `json:"up_mbps,omitempty"`
	DownMbps int  `json:"down_mbps,omitempty"`
}
