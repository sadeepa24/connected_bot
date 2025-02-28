package server

import (
	"fmt"
	"net"
)


type RangeCheck interface {
	Contains(ip net.IP) bool
}

// this is lightweight cidr range checker just for this project
// it's only support ipv4 ( all telegram ip rangers are ipv4 )
// after initiating first time cannot add rangers only checkings ip
type CIDRRange struct {
	roots []*IPNode
}

type IPNode struct {
	mask      byte
	maskedVal byte
	children  []*IPNode
	maskSize int
	cbyte byte
}

type SimpleRangeCheck struct {
	nets []*net.IPNet
}

func (s *SimpleRangeCheck) Contains(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, s := range s.nets {
		if s.Contains(ip) {
			return true
		}
	}
	return false
}



func NewCIDRRange(cidrList []string) (RangeCheck, error) {
	
	if len(cidrList) > 15 {
		cr := &CIDRRange{}
		for _, cidr := range cidrList {
			_, network, err := net.ParseCIDR(cidr)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR: %v", cidr)
			}
			cr.addCIDR(network)
		}
		return cr, nil
	} else {
		cr := &SimpleRangeCheck{}
		for _, cidr := range cidrList {
			_, network, err := net.ParseCIDR(cidr)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR: %v", cidr)
			}
			cr.nets = append(cr.nets, network)
		}
		return cr, nil

	}
}

func (c *CIDRRange) addCIDR(network *net.IPNet) {
	ip := network.IP.To4()
	//maskSize, _ := network.Mask.Size()

	root := c.findOrCreateRootNode(ip[0], network.Mask[0])
	root.addRange2(1, network)
	// if maskSize%8 > 0 || maskSize/8 > 1 {
		
	// }
	
}

func (c *CIDRRange) findOrCreateRootNode(byteVal, maskByte byte) *IPNode {
	maskedVal := byteVal & maskByte
	for _, root := range c.roots {
		if root.maskedVal == maskedVal {
			return root
		}
	}
	node := &IPNode{
		mask:      maskByte,
		maskedVal: maskedVal,
	}
	c.roots = append(c.roots, node)
	return node
}

// func (c *IPNode) addRange(byteIndex int, network *net.IPNet) {
// 	if byteIndex == 4 {
// 		return
// 	}
// 	maskSize, _ := network.Mask.Size()

// 	if byteIndex+1 - maskSize/8 < c.cuurentLow  {
// 		c.cuurentLow = byteIndex+1 - maskSize/8
// 	} else {
// 		return
// 	}


// 	msval := network.IP[byteIndex]&network.Mask[byteIndex]
// 	for _, ch := range c.children {
// 		if ch.maskedVal == msval {
// 			ch.addRange(byteIndex+1, network)
// 			return
// 		}
// 	}
// 	child := &IPNode{
// 		mask: network.Mask[byteIndex],
// 		maskedVal: msval,
// 	}
// 	c.children = append(c.children, child)
	
	
// 	if maskSize%8 > 0 {
// 		child.addRange(byteIndex+1, network)
// 	}

// 	if maskSize/8 == byteIndex+1 {
// 		
func (c *IPNode) addRange2(byteIndex int, network *net.IPNet) {
	if byteIndex == 4 {
		return
	}
	maskSize, _ := network.Mask.Size()
	
	if ((byteIndex+1)*8 - maskSize >= 8) {
		return
	}
	msval := network.IP[byteIndex]&network.Mask[byteIndex]
	for i, ch := range c.children {
		if network.IP[byteIndex] == ch.cbyte {
			if ch.maskSize <  (maskSize - byteIndex*2) {
				return
			} else {
				c.children[i] = &IPNode{
					mask: network.Mask[byteIndex],
					maskedVal: msval,
					maskSize: maskSize - byteIndex*2,
					cbyte: ch.cbyte,
				}
				//replace current
			}
		}

		if ch.maskedVal == msval {
			ch.addRange2(byteIndex+1, network)
			return
		}
	}

	ctbitmasksize := 8
	if maskSize - byteIndex*2 < 8 {
		ctbitmasksize = maskSize - byteIndex*2
	}

	

	child := &IPNode{
		mask: network.Mask[byteIndex],
		maskedVal: msval,
		maskSize: ctbitmasksize,
		cbyte: network.IP[byteIndex],
	}
	c.children = append(c.children, child)
	child.addRange2(byteIndex+1, network)
	
}





func (c *CIDRRange) Contains(ip net.IP) bool {
	if ip == nil {
		return false
	}
	ip = ip.To4()
	if ip == nil {
		return false
	}
	for _, root := range c.roots {
		if root.isValid(ip, 0) {
			return true
		}
	}
	return false
}

func (n *IPNode) isValid(ip net.IP, byteIndex int) bool {
	fmt.Println(byteIndex)
	
	if ip[byteIndex]&n.mask != n.maskedVal {
		return false
	}

	if byteIndex == 3 || len(n.children) == 0 {
		return true
	}
	for _, child := range n.children {
		if child.isValid(ip, byteIndex+1) {
			return true
		}
	}
	return false
}

type UnknownRemote string

func (e UnknownRemote) Error() string   { return "unknown remote " + string(e) }
func (e UnknownRemote) Timeout() bool   { return false }
func (e UnknownRemote) Temporary() bool { return true }