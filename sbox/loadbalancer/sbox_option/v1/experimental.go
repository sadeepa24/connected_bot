package sboxoption

import (
	"errors"
	"strconv"
	"strings"
)

type ExperimentalOptions struct {
	CacheFile *CacheFileOptions `json:"cache_file,omitempty"`
	ClashAPI  *ClashAPIOptions  `json:"clash_api,omitempty"`
	V2RayAPI  *V2RayAPIOptions  `json:"v2ray_api,omitempty"`
	Debug     *DebugOptions     `json:"debug,omitempty"`
}

type CacheFileOptions struct {
	Enabled     bool     `json:"enabled,omitempty"`
	Path        string   `json:"path,omitempty"`
	CacheID     string   `json:"cache_id,omitempty"`
	StoreFakeIP bool     `json:"store_fakeip,omitempty"`
	StoreRDRC   bool     `json:"store_rdrc,omitempty"`
	RDRCTimeout Duration `json:"rdrc_timeout,omitempty"`
}
func (c *CacheFileOptions) Changer(changer, value string) error {	
	switch changer {
	case "enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("invalid enabled: must be true or false")
		}
		c.Enabled = enabled
	case "path":
		c.Path = value
	case "cache_id":
		c.CacheID = value
	case "store_fakeip":
		storeFakeIP, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("invalid store_fakeip: must be true or false")
		}
		c.StoreFakeIP = storeFakeIP
	case "store_rdrc":
		storeRDRC, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("invalid store_rdrc: must be true or false")
		}
		c.StoreRDRC = storeRDRC
	case "rdrc_timeout":
		return c.RDRCTimeout.Set(value)
	default:
		return errors.New("unknown changer field")
	}
	return nil
}




type ClashAPIOptions struct {
	ExternalController               string           `json:"external_controller,omitempty"`
	ExternalUI                       string           `json:"external_ui,omitempty"`
	ExternalUIDownloadURL            string           `json:"external_ui_download_url,omitempty"`
	ExternalUIDownloadDetour         string           `json:"external_ui_download_detour,omitempty"`
	Secret                           string           `json:"secret,omitempty"`
	DefaultMode                      string           `json:"default_mode,omitempty"`
	ModeList                         []string         `json:"-"`
	AccessControlAllowOrigin         Listable[string] `json:"access_control_allow_origin,omitempty"`
	AccessControlAllowPrivateNetwork bool             `json:"access_control_allow_private_network,omitempty"`

	// Deprecated: migrated to global cache file
	CacheFile string `json:"cache_file,omitempty"`
	// Deprecated: migrated to global cache file
	CacheID string `json:"cache_id,omitempty"`
	// Deprecated: migrated to global cache file
	StoreMode bool `json:"store_mode,omitempty"`
	// Deprecated: migrated to global cache file
	StoreSelected bool `json:"store_selected,omitempty"`
	// Deprecated: migrated to global cache file
	StoreFakeIP bool `json:"store_fakeip,omitempty"`
}



func (c *ClashAPIOptions) Changer(changer, value string) error {


	
	switch changer {
	case "external_controller":
		c.ExternalController = value
	case "external_ui":
		c.ExternalUI = value
	case "external_ui_download_url":
		c.ExternalUIDownloadURL = value
	case "external_ui_download_detour":
		c.ExternalUIDownloadDetour = value
	case "secret":
		c.Secret = value
	case "default_mode":
		c.DefaultMode = value
	case "access_control_allow_origin":
		if value == "" {
			c.AccessControlAllowOrigin = nil
			break
		}
		list := strings.Split(value, ",")
		c.AccessControlAllowOrigin = Listable[string](list)
	case "access_control_allow_private_network":
		privateNetwork, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("invalid access_control_allow_private_network: must be true or false")
		}
		c.AccessControlAllowPrivateNetwork = privateNetwork
	case "cache_file":
		c.CacheFile = value
	case "cache_id":
		c.CacheID = value
	case "store_mode":
		storeMode, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("invalid store_mode: must be true or false")
		}
		c.StoreMode = storeMode
	case "store_selected":
		storeSelected, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("invalid store_selected: must be true or false")
		}
		c.StoreSelected = storeSelected
	case "store_fakeip":
		storeFakeIP, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("invalid store_fakeip: must be true or false")
		}
		c.StoreFakeIP = storeFakeIP
	default:
		return errors.New("unknown changer field")
	}
	return nil
}






type V2RayAPIOptions struct {
	Listen string                    `json:"listen,omitempty"`
	Stats  *V2RayStatsServiceOptions `json:"stats,omitempty"`
}

type V2RayStatsServiceOptions struct {
	Enabled   bool     `json:"enabled,omitempty"`
	Inbounds  []string `json:"inbounds,omitempty"`
	Outbounds []string `json:"outbounds,omitempty"`
	Users     []string `json:"users,omitempty"`
}
