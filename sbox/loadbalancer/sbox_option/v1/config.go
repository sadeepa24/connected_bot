package sboxoption

import (
	"bytes"

	"github.com/jinzhu/copier"
	singopt "github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

type _Options struct {
	RawMessage   json.RawMessage      `json:"-"`
	Schema       string               `json:"$schema,omitempty"`
	Log          *LogOptions          `json:"log,omitempty"`
	DNS          *DNSOptions          `json:"dns,omitempty"`
	NTP          *NTPOptions          `json:"ntp,omitempty"`
	Inbounds     []Inbound            `json:"inbounds,omitempty"`
	Outbounds    []Outbound           `json:"outbounds,omitempty"`
	Route        *RouteOptions        `json:"route,omitempty"`
	Experimental *ExperimentalOptions `json:"experimental,omitempty"`
}

type Options _Options

func (o *Options) UnmarshalJSON(content []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	err := decoder.Decode((*_Options)(o))
	if err != nil {
		return err
	}
	o.RawMessage = content
	return nil
}

func (o *Options) SagerNetOpt() singopt.Options {
	//TODO: Remove this copy method, and add manual coping
	var singoptions singopt.Options
	copier.Copy(&singoptions, o)
	return singoptions
}

type LogOptions struct {
	Disabled     bool   `json:"disabled,omitempty"`
	Level        string `json:"level,omitempty"`
	Output       string `json:"output,omitempty"`
	Timestamp    bool   `json:"timestamp,omitempty"`
	DisableColor bool   `json:"-"`
}
