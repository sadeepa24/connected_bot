package builder

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing/common/json"
	"gopkg.in/yaml.v3"
)

type StoreItem struct {
	Name   string `json:"name" yaml:"name"`
	Type   string `json:"type" yaml:"type"`
	Wizard bool   `json:"wizard" yaml:"wizard"`
	Item   string `json:"item" yaml:"item"`
}

type Store struct {
	boilerPlates map[string]json.RawMessage
	wizardConf map[string]WizardConf
	
	WizIn []string
	WizOut []string
	WizEnd []string
	WizRrules []string
	WizDrules []string
	WizConf []string
}

const (
	ItemConf = "fullconfig"
	ItemIn = "inbound"
	ItemOut = "outbound"
	ItemEnd = "endpoint"
	ItemRrule = "routerule"
	ItemDrule = "dnsrule"
)

func Newstore(path string) (*Store, error) {
	
	if globalCtx == nil {
		globalCtx = box.Context(globalCtx, include.InboundRegistry(), include.OutboundRegistry(), include.EndpointRegistry())
	}
	
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	items :=  []StoreItem{}


	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(file, &items)
	} else {
		err = yaml.Unmarshal(file, &items)
	}
	
	if err != nil {
		return nil, err
	}

	str := &Store{
		boilerPlates: map[string]json.RawMessage{},
		wizardConf: map[string]WizardConf{},
	}

	for i := range items {
		if len(items[i].Item) == 0 {
			return nil, errors.New("item " + items[i].Name + " have zero length item remove it or add valid item")
		}
		switch items[i].Type {
		case ItemConf:
			str.WizConf = append(str.WizConf, items[i].Name)
		case ItemIn:
			str.WizIn = append(str.WizIn, items[i].Name)
		case ItemOut:
			str.WizOut = append(str.WizOut, items[i].Name)
		case ItemEnd:
			str.WizEnd = append(str.WizEnd, items[i].Name)
		case ItemDrule:
			str.WizDrules = append(str.WizDrules, items[i].Name)
		case ItemRrule:
			str.WizRrules = append(str.WizRrules, items[i].Name)
		default:
			return nil, errors.New("unknown store item type" + items[i].Type + " on "+items[i].Name )
		}
		if items[i].Wizard {
			str.wizardConf[items[i].Name] = NewWizardConf([]byte(items[i].Item))
		} else {
			str.boilerPlates[items[i].Name] = []byte(items[i].Item)
		}
	}
	return str, nil
}

func ExportConf[T any](s *Store, tag string, conec Connector) (T, error) {
	if raw, ok := s.boilerPlates[tag]; ok {
		return json.UnmarshalExtendedContext[T](globalCtx, raw)
	}
	if s == nil {
		var new T
		return new, errors.New("nil store")
	}
	buf, err := s.exportWizard(tag, conec)
	
	if err != nil {
		var new T
		return new, err
	}
	defer func() {
		buf.Reset()
		buf = nil
	}()
	return json.UnmarshalExtendedContext[T](globalCtx, buf.Bytes())
}
func (s *Store) exportWizard(tag string, conec Connector) (*bytes.Buffer, error) {
	pt, ok := s.wizardConf[tag]
	if !ok {
		return nil, errors.New("cannot find boiler plate with name" + tag)
	}
	buf := &bytes.Buffer{}
	err := pt.ExportTo(buf, conec)
	if err != nil {
		buf.Reset()
		buf = nil
		return nil, err
	}
	return buf, nil
}







type WizardConf [][]byte 

func (w WizardConf) ExportTo(wr io.Writer, conec Connector) error {
	if conec == nil || wr == nil {
		return Error{
			error: errors.New("nil writer or connector"),
		}
	}
	for i := range w {
		if i%2 == 0 {
			wr.Write(w[i])
			continue
		}
		val, err := conec.ReciveVal(string(w[i]))
		if err != nil {
			return err
		}
		wr.Write([]byte(val))
	}
	return nil
}
func NewWizardConf(conf []byte) WizardConf {
	var wizconf [][]byte
	last := 0
	from := 0
	for i := range conf {
		if i != 0  &&  conf[i] == '<' && conf[i+1] == '<' {
			new := make([]byte, i - last)
			for j := range new {
				new[j] = conf[j+last]
			}
			wizconf = append(wizconf, new)
			from = i+2
		}
		if i+1 < len(conf)-1 && conf[i] == '>' && conf[i+1] == '>'  {
			new := make([]byte, i - from)
			for j := range new {
				new[j] = conf[j+from]
			}
			wizconf = append(wizconf, new)
			last = i+2
			continue
		}
		if i == len(conf)-1 {
			new := make([]byte, i+1 - last)
			for j := range new {
				new[j] = conf[j+last]
			}
			wizconf = append(wizconf, new)
			break
		}
	}
	return wizconf
}