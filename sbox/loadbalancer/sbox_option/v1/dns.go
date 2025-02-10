package sboxoption

import (
	"errors"
	"net/netip"
)

type DNSOptions struct {
	Servers        []DNSServerOptions `json:"servers,omitempty"`
	Rules          []DNSRule          `json:"rules,omitempty"`
	Final          string             `json:"final,omitempty"`
	ReverseMapping bool               `json:"reverse_mapping,omitempty"`
	FakeIP         *DNSFakeIPOptions  `json:"fakeip,omitempty"`
	DNSClientOptions
}

type DNSServerOptions struct {
	Tag                  string         `json:"tag,omitempty"`
	Address              string         `json:"address"`
	AddressResolver      string         `json:"address_resolver,omitempty"`
	AddressStrategy      DomainStrategy `json:"address_strategy,omitempty"`
	AddressFallbackDelay Duration       `json:"address_fallback_delay,omitempty"`
	Strategy             DomainStrategy `json:"strategy,omitempty"`
	Detour               string         `json:"detour,omitempty"`
	ClientSubnet         *AddrPrefix    `json:"client_subnet,omitempty"`
}


func (d *DNSServerOptions) Changer(change, value string) error {
	switch change {
	case "address":
		d.Address = value
	case "address_resolver":
		d.AddressResolver = value
	case "address_strategy":
		return d.AddressStrategy.Set(value)
	case "address_fallback_delay":
		return d.AddressFallbackDelay.Set(value)
	case "strategy":
		return d.Strategy.Set(value)
	case "detour":
		d.Detour = value
	case "client_subnet":
		if d.ClientSubnet == nil {
			d.ClientSubnet = &AddrPrefix{}
		}
		return d.ClientSubnet.Set(value)
	default:
		return errors.New("unknown field: " + change)
	}
	return nil

}





type DNSClientOptions struct {
	Strategy         DomainStrategy `json:"strategy,omitempty"`
	DisableCache     bool           `json:"disable_cache,omitempty"`
	DisableExpire    bool           `json:"disable_expire,omitempty"`
	IndependentCache bool           `json:"independent_cache,omitempty"`
	ClientSubnet     *AddrPrefix    `json:"client_subnet,omitempty"`
}

type DNSFakeIPOptions struct {
	Enabled    bool          `json:"enabled,omitempty"`
	Inet4Range *netip.Prefix `json:"inet4_range,omitempty"`
	Inet6Range *netip.Prefix `json:"inet6_range,omitempty"`
}
