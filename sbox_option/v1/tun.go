package sboxoption

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json"
)

type TunInboundOptions struct {
	InterfaceName          string                 `json:"interface_name,omitempty"`
	MTU                    uint32                 `json:"mtu,omitempty"`
	GSO                    bool                   `json:"gso,omitempty"`
	Address                Listable[netip.Prefix] `json:"address,omitempty"`
	AutoRoute              bool                   `json:"auto_route,omitempty"`
	IPRoute2TableIndex     int                    `json:"iproute2_table_index,omitempty"`
	IPRoute2RuleIndex      int                    `json:"iproute2_rule_index,omitempty"`
	AutoRedirect           bool                   `json:"auto_redirect,omitempty"`
	AutoRedirectInputMark  FwMark                 `json:"auto_redirect_input_mark,omitempty"`
	AutoRedirectOutputMark FwMark                 `json:"auto_redirect_output_mark,omitempty"`
	StrictRoute            bool                   `json:"strict_route,omitempty"`
	RouteAddress           Listable[netip.Prefix] `json:"route_address,omitempty"`
	RouteAddressSet        Listable[string]       `json:"route_address_set,omitempty"`
	RouteExcludeAddress    Listable[netip.Prefix] `json:"route_exclude_address,omitempty"`
	RouteExcludeAddressSet Listable[string]       `json:"route_exclude_address_set,omitempty"`
	IncludeInterface       Listable[string]       `json:"include_interface,omitempty"`
	ExcludeInterface       Listable[string]       `json:"exclude_interface,omitempty"`
	IncludeUID             Listable[uint32]       `json:"include_uid,omitempty"`
	IncludeUIDRange        Listable[string]       `json:"include_uid_range,omitempty"`
	ExcludeUID             Listable[uint32]       `json:"exclude_uid,omitempty"`
	ExcludeUIDRange        Listable[string]       `json:"exclude_uid_range,omitempty"`
	IncludeAndroidUser     Listable[int]          `json:"include_android_user,omitempty"`
	IncludePackage         Listable[string]       `json:"include_package,omitempty"`
	ExcludePackage         Listable[string]       `json:"exclude_package,omitempty"`
	EndpointIndependentNat bool                   `json:"endpoint_independent_nat,omitempty"`
	UDPTimeout             UDPTimeoutCompat       `json:"udp_timeout,omitempty"`
	Stack                  string                 `json:"stack,omitempty"`
	Platform               *TunPlatformOptions    `json:"platform,omitempty"`
	InboundOptions

	// Deprecated: merged to Address
	Inet4Address Listable[netip.Prefix] `json:"inet4_address,omitempty"`
	// Deprecated: merged to Address
	Inet6Address Listable[netip.Prefix] `json:"inet6_address,omitempty"`
	// Deprecated: merged to RouteAddress
	Inet4RouteAddress Listable[netip.Prefix] `json:"inet4_route_address,omitempty"`
	// Deprecated: merged to RouteAddress
	Inet6RouteAddress Listable[netip.Prefix] `json:"inet6_route_address,omitempty"`
	// Deprecated: merged to RouteExcludeAddress
	Inet4RouteExcludeAddress Listable[netip.Prefix] `json:"inet4_route_exclude_address,omitempty"`
	// Deprecated: merged to RouteExcludeAddress
	Inet6RouteExcludeAddress Listable[netip.Prefix] `json:"inet6_route_exclude_address,omitempty"`
}
func (o *TunInboundOptions) Changer(changer, value string) error {
    var err error

    if o == nil {
        return errors.New("TunInboundOptions object is not available")
    }

    switch changer {
    case "interface_name":
        o.InterfaceName = value

    case "mtu":
        var mtu uint64
        if mtu, err = strconv.ParseUint(value, 10, 32); err != nil {
            return err
        }
        o.MTU = uint32(mtu)

    case "gso":
        if o.GSO, err = strconv.ParseBool(value); err != nil {
            return err
        }

    case "auto_route":
        if o.AutoRoute, err = strconv.ParseBool(value); err != nil {
            return err
        }

    case "iproute2_table_index":
        var index int64
        if index, err = strconv.ParseInt(value, 10, 32); err != nil {
            return err
        }
        o.IPRoute2TableIndex = int(index)

    case "iproute2_rule_index":
        var index int64
        if index, err = strconv.ParseInt(value, 10, 32); err != nil {
            return err
        }
        o.IPRoute2RuleIndex = int(index)

    case "auto_redirect":
        if o.AutoRedirect, err = strconv.ParseBool(value); err != nil {
            return err
        }

    case "strict_route":
        if o.StrictRoute, err = strconv.ParseBool(value); err != nil {
            return err
        }

    case "stack":
        o.Stack = value

    case "include_interface":
        o.IncludeInterface = Listable[string](strings.Split(value, ","))
        if len(o.IncludeInterface) == 0 {
            return errors.New("include_interface cannot be empty")
        }

    case "exclude_interface":
        o.ExcludeInterface = Listable[string](strings.Split(value, ","))
        if len(o.ExcludeInterface) == 0 {
            return errors.New("exclude_interface cannot be empty")
        }

    case "route_address":
		ips := strings.Split(value, ",")
		if len(ips) == 0 {
    	    return errors.New("route_exclude_address cannot be empty")
    	}
		o.RouteAddress = Listable[netip.Prefix]{}
		for _, d := range ips {
			prefix, err := netip.ParsePrefix(d)
			if err != nil {
				return errors.New("prefix parsing err " + d + " " + err.Error())
			}
			o.RouteAddress = append(o.RouteAddress, prefix)
		}

    case "route_address_set":
        o.RouteAddressSet = Listable[string](strings.Split(value, ","))
        if len(o.RouteAddressSet) == 0 {
            return errors.New("route_address_set cannot be empty")
        }

    case "route_exclude_address":
		
    	ips := strings.Split(value, ",")
		if len(ips) == 0 {
    	    return errors.New("route_exclude_address cannot be empty")
    	}
		o.RouteExcludeAddress = Listable[netip.Prefix]{}
		for _, d := range ips {
			prefix, err := netip.ParsePrefix(d)
			if err != nil {
				return errors.New("prefix parsing err " + d + " " + err.Error())
			}
			o.RouteExcludeAddress = append(o.RouteExcludeAddress, prefix)
		}
    	

    case "route_exclude_address_set":
        o.RouteExcludeAddressSet = Listable[string](strings.Split(value, ","))
        if len(o.RouteExcludeAddressSet) == 0 {
            return errors.New("route_exclude_address_set cannot be empty")
        }

    case "include_uid":
        uidStrings := strings.Split(value, ",")
        var uids []uint32
        for _, uidStr := range uidStrings {
            uid, err := strconv.ParseUint(uidStr, 10, 32)
            if err != nil {
                return fmt.Errorf("include_uid contains invalid uint32 value: %v", err)
            }
            uids = append(uids, uint32(uid))
        }
        if len(uids) == 0 {
            return errors.New("include_uid cannot be empty")
        }
        o.IncludeUID = Listable[uint32](uids)

    case "include_uid_range":
        o.IncludeUIDRange = Listable[string](strings.Split(value, ","))
        if len(o.IncludeUIDRange) == 0 {
            return errors.New("include_uid_range cannot be empty")
        }

    case "exclude_uid":
        uidStrings := strings.Split(value, ",")
        var uids []uint32
        for _, uidStr := range uidStrings {
            uid, err := strconv.ParseUint(uidStr, 10, 32)
            if err != nil {
                return fmt.Errorf("exclude_uid contains invalid uint32 value: %v", err)
            }
            uids = append(uids, uint32(uid))
        }
        if len(uids) == 0 {
            return errors.New("exclude_uid cannot be empty")
        }
        o.ExcludeUID = Listable[uint32](uids)

    case "exclude_uid_range":
        o.ExcludeUIDRange = Listable[string](strings.Split(value, ","))
        if len(o.ExcludeUIDRange) == 0 {
            return errors.New("exclude_uid_range cannot be empty")
        }

    case "include_android_user":
        userStrings := strings.Split(value, ",")
        var users []int
        for _, userStr := range userStrings {
            user, err := strconv.Atoi(userStr)
            if err != nil {
                return fmt.Errorf("include_android_user contains invalid int value: %v", err)
            }
            users = append(users, user)
        }
        if len(users) == 0 {
            return errors.New("include_android_user cannot be empty")
        }
        o.IncludeAndroidUser = Listable[int](users)

    case "include_package":
        o.IncludePackage = Listable[string](strings.Split(value, ","))
        if len(o.IncludePackage) == 0 {
            return errors.New("include_package cannot be empty")
        }

    case "exclude_package":
        o.ExcludePackage = Listable[string](strings.Split(value, ","))
        if len(o.ExcludePackage) == 0 {
            return errors.New("exclude_package cannot be empty")
        }

    case "endpoint_independent_nat":
        if o.EndpointIndependentNat, err = strconv.ParseBool(value); err != nil {
            return err
        }

    default:
        return fmt.Errorf("unsupported changer: %s", changer)
    }

    return nil
}


type FwMark uint32

func (f FwMark) MarshalJSON() ([]byte, error) {
	return json.Marshal(F.ToString("0x", strconv.FormatUint(uint64(f), 16)))
}

func (f *FwMark) UnmarshalJSON(bytes []byte) error {
	var stringValue string
	err := json.Unmarshal(bytes, &stringValue)
	if err != nil {
		if rawErr := json.Unmarshal(bytes, (*uint32)(f)); rawErr == nil {
			return nil
		}
		return E.Cause(err, "invalid number or string mark")
	}
	intValue, err := strconv.ParseUint(stringValue, 0, 32)
	if err != nil {
		return err
	}
	*f = FwMark(intValue)
	return nil
}
