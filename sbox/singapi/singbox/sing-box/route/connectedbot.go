package route

import (
	"github.com/sagernet/sing-box/adapter"
)

func (r *abstractDefaultRule) Match(metadata *adapter.InboundContext) bool {
	
	
	if len(r.allItems) == 0 {
		return true
	}

	if len(r.sourceAddressItems) > 0 && !metadata.SourceAddressMatch {
		metadata.DidMatch = true
		for _, item := range r.sourceAddressItems {
			if item.Match(metadata) {
				metadata.SourceAddressMatch = true
				break
			}
		}
	}

	if len(r.sourcePortItems) > 0 && !metadata.SourcePortMatch {
		metadata.DidMatch = true
		for _, item := range r.sourcePortItems {
			if item.Match(metadata) {
				metadata.SourcePortMatch = true
				break
			}
		}
	}

	if len(r.destinationAddressItems) > 0 && !metadata.DestinationAddressMatch {
		metadata.DidMatch = true
		for _, item := range r.destinationAddressItems {
			if item.Match(metadata) {
				metadata.DestinationAddressMatch = true
				break
			}
		}
	}

	if !metadata.IgnoreDestinationIPCIDRMatch && len(r.destinationIPCIDRItems) > 0 && !metadata.DestinationAddressMatch {
		metadata.DidMatch = true
		for _, item := range r.destinationIPCIDRItems {
			if item.Match(metadata) {
				metadata.DestinationAddressMatch = true
				break
			}
		}
	}

	if len(r.destinationPortItems) > 0 && !metadata.DestinationPortMatch {
		metadata.DidMatch = true
		for _, item := range r.destinationPortItems {
			if item.Match(metadata) {
				metadata.DestinationPortMatch = true
				break
			}
		}
	}

	for _, item := range r.items {
		if _, isRuleSet := item.(*RuleSetItem); !isRuleSet {
			metadata.DidMatch = true
		}
		if !item.Match(metadata) {
			return r.invert
		}
	}

	if len(r.sourceAddressItems) > 0 && !metadata.SourceAddressMatch {
		return r.invert
	}

	if len(r.sourcePortItems) > 0 && !metadata.SourcePortMatch {
		return r.invert
	}

	if ((!metadata.IgnoreDestinationIPCIDRMatch && len(r.destinationIPCIDRItems) > 0) || len(r.destinationAddressItems) > 0) && !metadata.DestinationAddressMatch {
		return r.invert
	}

	if len(r.destinationPortItems) > 0 && !metadata.DestinationPortMatch {
		return r.invert
	}

	if !metadata.DidMatch {
		return true
	}
	return !r.invert
}

