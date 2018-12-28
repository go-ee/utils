package vault

import (
	vaultapi "github.com/hashicorp/vault/api"
	"fmt"
	"strings"
	"errors"
	"github.com/eugeis/gee/as"
)

type VaultClient struct {
	appName string
	client  *vaultapi.Client
	logical *vaultapi.Logical
}

func NewVaultClient(appName string, token string, address string) (ret *VaultClient, err error) {
	client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if err == nil {
		client.SetToken(token)
		if len(address) > 0 {
			client.SetAddress(address)
		}
		ret = &VaultClient{appName: appName, client: client, logical: client.Logical()}
	}
	return
}

func (o *VaultClient) fillAccessData(name string, security *as.Security) (err error) {
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

	var vault *VaultClient
	vault, err = NewVaultClient(appName, vaultToken, vaultAddress)
	if err == nil {
		err = vault.fillAccessData(name, security)
	}
	return
}
