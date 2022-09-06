package vault

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ee/utils/as"
	vaultapi "github.com/hashicorp/vault/api"
)

type Client struct {
	appName string
	client  *vaultapi.Client
	logical *vaultapi.Logical
}

func NewClient(appName string, token string, address string) (ret *Client, err error) {
	client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if err == nil {
		client.SetToken(token)
		if len(address) > 0 {
			client.SetAddress(address)
		}
		ret = &Client{appName: appName, client: client, logical: client.Logical()}
	}
	return
}

func (o *Client) fillAccessData(name string, security *as.Security) (err error) {
	var secret *vaultapi.Secret
	basePath := fmt.Sprintf("secret/%v/%v", o.appName, strings.ToLower(name))
	for key, item := range security.Access {
		vaultPath := fmt.Sprintf("%v/%v", basePath, key)
		secret, err = o.logical.Read(vaultPath)
		if err != nil {
			break
		}
		if secret != nil {
			item.User = secret.Data["user"].(string)
			item.Password = secret.Data["password"].(string)
			security.Access[key] = item
		} else {
			err = errors.New(fmt.Sprintf("No access data in Vault for '%key", vaultPath))
		}
	}
	return
}

func BuildAccessFinderFromVault(appName string, vaultToken string, vaultAddress string, name string, keys []string) (ret as.AccessFinder, err error) {
	security := as.FillAccessKeys(keys, &as.Security{})
	ret = security

	var vault *Client
	vault, err = NewClient(appName, vaultToken, vaultAddress)
	if err == nil {
		err = vault.fillAccessData(name, security)
	}
	return
}
