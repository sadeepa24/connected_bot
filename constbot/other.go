package constbot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// use this for bandwidth represtation
type Bwidth float64

// retuns Byte value
// 2MB => 2 * 1024 * 1024
func BwidthString(bwidth string) (Bwidth, error) {
	var numPart string
	for i, char := range bwidth {
		if char < '0' || char > '9' {
			numPart = bwidth[:i]
			bwidth = bwidth[i:]
			break
		}
	}
	value, err := strconv.Atoi(numPart)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %v", err)
	}
	switch strings.ToUpper(bwidth) {
	case "KB":
		return Bwidth(value * 1024), nil
	case "MB":
		return Bwidth(value * 1024 * 1024), nil
	case "GB":
		return Bwidth(value * 1024 * 1024 * 1024), nil
	case "TB":
		return Bwidth(value * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", bwidth)
	}
}

//return value as bytes,
//if no suffix found input will prosess as gb
func ParserBwidth(i string) (Bwidth, error) {
	ft, err := strconv.ParseFloat(i, 64)
	if err == nil {
		return Bwidth(ft).GbtoByte(), nil
	}
	i = strings.ReplaceAll(i, " ", "")
	i = strings.ToLower(i)

	if size, cut := strings.CutSuffix(i, "kb"); cut {
		ft, err := strconv.ParseFloat(size, 64)
		if err != nil {
			return 0, err
		}
		return Bwidth(ft).KbtoBYte(), nil
	}
	if size, cut := strings.CutSuffix(i, "mb"); cut {
		ft, err := strconv.ParseFloat(size, 64)
		if err != nil {
			return 0, err
		}
		return Bwidth(ft).MbtoBYte(), nil
	}
	if size, cut := strings.CutSuffix(i, "gb"); cut {
		ft, err := strconv.ParseFloat(size, 64)
		if err != nil {
			return 0, err
		}
		return Bwidth(ft).GbtoByte(), nil
	}
	return 0, errors.New("cannot parse")
}

func (b Bwidth) BytetoGB() Bwidth {
	return b / Bwidth(AsGB)
}

func (b Bwidth) MbtoBYte() Bwidth {
	return b * MBtoByte
}
func (b Bwidth) KbtoBYte() Bwidth {
	return b * KBtoByte
}
func (b Bwidth) GbtoByte() Bwidth {
	return b * GBtoByte
}


func (b Bwidth) BytetoMB() Bwidth {
	return b / Bwidth(AsMB)
}

func (b Bwidth) BytetoKB() Bwidth {
	return b / Bwidth(AsKB)
}

func (b Bwidth) Int() int {
	return int(b)
}

func (b Bwidth) Int64() int64 {
	return int64(b)
}

// carefull with this method
func (b Bwidth) Int32() int32 {
	return int32(b)
}

func (b Bwidth) String() string {
	return strconv.FormatFloat(b.Float64(), 'f', 2, 64)
}
func (b Bwidth) Float64() float64 {
	return float64(b)
}

// convert byte value to readble
func (b Bwidth) BToString() string {
	switch {
	case b < 1024 && b > -1024:
		return b.String() + " B"
	case b < 1024*1024 && b > -(1024*1024):
		return b.BytetoKB().String() + " KB"
	case b < 1024*1024*1024 && b > -(1024*1024*1024):
		return b.BytetoMB().String() + " MB"
	default:
		return b.BytetoGB().String() + " GB"
	}
}
