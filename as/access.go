package as

import (
	"fmt"
	"errors"
)

type Access struct {
	User     string
	Password string
}

type AccessFinder interface {
	FindAccess(key string) (Access, error)
}

type Security struct {
	Access map[string]Access
}

func NewSecurity() *Security {
	return &Security{Access: make(map[string]Access)}
}

func (o *Security) FindAccess(key string) (ret Access, err error) {
	var ok bool
	ret, ok = o.Access[key]
	if !ok {
		err = errors.New(fmt.Sprintf("No access data found for '%v'", key))
	}
	return
}

func (o *Security) AddAccess(key string, user string, password string) (ret *Security) {
	ret = o
	o.Access[key] = Access{User: user, Password: password}
	return
}

func NewAccessFinderSingle(accessKey, user string, password string) AccessFinder {
	return &Security{Access: map[string]Access{accessKey: {User: user, Password: password}}}
}

func NewAccessFinderFromFile(securityFile string) (ret AccessFinder, err error) {
	security := &Security{}
	ret = security
	err = fillAccessData(security, securityFile)
	return
}

func FillAccessKeys(keys []string, security *Security) (ret *Security) {
	ret = security
	for _, item := range keys {
		if _, ok := security.Access[item]; !ok {
			security.Access[item] = Access{}
		}
	}
	return
}
