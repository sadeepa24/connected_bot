package route

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sagernet/sing-box/adapter"
	F "github.com/sagernet/sing/common/format"
)

var _ RuleItem = (*AuthUserItem)(nil)

type AuthUserItem struct {
	users   []string
	userMap map[string]bool
	userSyncmap *sync.Map
}

func NewAuthUserItem(users []string) *AuthUserItem {
	userMap := make(map[string]bool)

	authitem := &AuthUserItem{
		users:   users,
		userMap: userMap,
		userSyncmap: &sync.Map{},
	}
	
	for _, protocol := range users {
		userMap[protocol] = true
		authitem.userSyncmap.Store(protocol, true)
	}

	return authitem
}

func (r *AuthUserItem) Match(metadata *adapter.InboundContext) bool {
	return r.userMap[metadata.User]
}

func (r *AuthUserItem) String() string {
	if len(r.users) == 1 {
		return F.ToString("auth_user=", r.users[0])
	}
	return F.ToString("auth_user=[", strings.Join(r.users, " "), "]")
}

func (r *AuthUserItem) Adduser(user string) error {
	r.userSyncmap.Store(user, true)
	return nil
} 

func  (r *AuthUserItem) RemoveUser(user string) error {
	_, loaded := r.userSyncmap.LoadAndDelete(user)
	if !loaded {
		return fmt.Errorf("no found")
	}
	return nil
}


type UserAdd interface {
	Adduser(string) error
	RemoveUser(string) error
}



// type SafeMap struct {
//     m  map[string]int64
//     mu sync.RWMutex
// }
// func NewSafeMap() *SafeMap {
//     return &SafeMap{
// 		m: make(map[string]int64)}
// }

// func (s *SafeMap) Load(key string) (int64, bool) {
//     s.mu.RLock()
//     defer s.mu.RUnlock()
//     value, ok := s.m[key]
//     return value, ok
// }

// func (s *SafeMap) Store(key string, value int64) {
//     s.mu.Lock()
//     defer s.mu.Unlock()
//     s.m[key] = value
// }

// func (s *SafeMap) Delete(key string) {
//     s.mu.Lock()
//     defer s.mu.Unlock()
//     delete(s.m, key)
// }

// func (s *SafeMap) Range(f func(key string, value int64) bool) {
//     s.mu.RLock()
//     defer s.mu.RUnlock()
//     for k, v := range s.m {
//         if !f(k, v) {
//             break
//         }
//     }
// }