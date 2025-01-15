package route

import (
	"sync"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
)

func (r *Router) AddNewRule(rule adapter.Rule) {
	r.rules = append(r.rules, rule)

}

func (r *Router) AddOutboundUser(user, outboundtag string) {
	r.RemoveAllRule(user)
	
	for _, rule := range r.rules {
		if ruler, ok := rule.(Botrule);ok {
			if ruler.Outbound() == outboundtag {
				ruler.AddUser(user)
				out, loaded := r.outboundByTag[outboundtag]
				if loaded {
					if out.Type() == C.TypeDirect {
						r.directusermap.Store(user, struct{}{})
					}
				}
				break
			}
		}
	}

}


func (r *Router) RemoveAllRule(user string) {
	for _, rule := range r.rules {
		if ruler, ok := rule.(Botrule); ok {
			ruler.RemoveUser(user)
		}
		
	}
	r.directusermap.Delete(user)

}

func (r *Router) SkipSniff(user string ) bool {
	_, ok := r.directusermap.Load(user)
	return ok
}



type Botrule interface {
	AddUser(user string)
	RemoveUser(user string)
	Outbound() string
}


type RuleForBot struct {
	//users *sync.Map
	usermap map[string]bool
	mu *sync.RWMutex
	outbound string
}


func NewBotRule(out string) adapter.Rule {
	return &RuleForBot{
		//users: &sync.Map{},
		mu: &sync.RWMutex{},
		usermap: map[string]bool{},
		outbound: out,

	}
}


func (r *RuleForBot) Start() error {return nil}
func (r *RuleForBot) Close() error {return nil}


func (r *RuleForBot) Match(metadata *adapter.InboundContext) bool {
	r.mu.RLock()
	_, ok := r.usermap[metadata.User]
	r.mu.RUnlock()
	return ok
}

func (r *RuleForBot) AddUser(user string) {

	r.mu.Lock()
	r.usermap[user] = true
	r.mu.Unlock()
}

func (r *RuleForBot) RemoveUser(user string) {
	r.mu.Lock()
	delete(r.usermap, user)
	r.mu.Unlock()
}


func (r *RuleForBot) String() string { 
	return "rule only for connected bot " + r.outbound
}


func (r *RuleForBot) Type() string { return C.RuleTypeDefault }
func (r *RuleForBot) UpdateGeosite() error {return nil }
func (r *RuleForBot) Outbound() string {return r.outbound }
