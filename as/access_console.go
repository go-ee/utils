package as

import (
	"bufio"
	"os"
	"fmt"
	"strings"
	"github.com/eugeis/gee/cfg"
	"github.com/howeyc/gopass"
)

func NewAccessFinderFromConsole(keys []string) (ret AccessFinder, err error) {
	security := FillAccessKeys(keys, NewSecurity())
	ret, err = fillAccessDataFromConsole(security)
	return
}

func NewAccessFinderFromConsoleSingle(key string, user string, password string) (ret AccessFinder, err error) {
	security := NewSecurity().AddAccess(key, user, password)
	if len(user) == 0 || len(password) == 0 {
		security, err = fillAccessDataFromConsole(security)
	}
	return security, err
}

func fillAccessDataFromConsole(security *Security) (ret *Security, err error) {
	ret = security
	reader := bufio.NewReader(os.Stdin)
	var text string
	var pw []byte
	for key, item := range security.Access {
		fmt.Printf("Enter access data for '%v'\n", key)

		if len(item.User) == 0 {
			fmt.Print("User: ")
			text, err = reader.ReadString('\n')
			if err != nil {
				break
			}
			item.User = strings.TrimSpace(text)
		}

		if len(item.Password) == 0 {
			fmt.Print("Password: ")
			pw, err = gopass.GetPasswdMasked()
			if err != nil {
				break
			}
			item.Password = string(pw)
		}
		security.Access[key] = item
	}
	return
}

func fillAccessData(security *Security, file string) (err error) {
	return cfg.UnmarshalFile(security, file)
}
