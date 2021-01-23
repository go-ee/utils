package net

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"os"
)

type RsaKeys struct {
	rsaFileName    string
	rsaPubFileName string
	rsaFile        string
	rsaPubFile     string

	private *rsa.PrivateKey
	public  *rsa.PublicKey

	keysFolder string
}

func (o *RsaKeys) Private() *rsa.PrivateKey {
	return o.private
}

func (o *RsaKeys) Public() *rsa.PublicKey {
	return o.public
}

func (o *RsaKeys) RsaFile() string {
	return o.rsaFile
}

func (o *RsaKeys) RsaPubFile() string {
	return o.rsaPubFile
}

func (o *RsaKeys) LoadOrCreate() (err error) {
	if !o.Exists() {
		err = o.Create()
	} else {
		var keyBytes []byte
		if keyBytes, err = ioutil.ReadFile(o.rsaFile); err == nil {
			o.private, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
		}

		if err != nil {
			return
		}

		if keyBytes, err = ioutil.ReadFile(o.rsaPubFile); err == nil {
			o.public, err = jwt.ParseRSAPublicKeyFromPEM(keyBytes)
		}
	}
	return
}

func (o *RsaKeys) Exists() (ret bool) {
	os.MkdirAll(o.keysFolder, 0700)

	_, err := os.Stat(o.rsaFile)
	if ret = !os.IsNotExist(err); ret {
		_, err := os.Stat(o.rsaPubFile)
		ret = !os.IsNotExist(err)
	}
	return
}

func (o *RsaKeys) Create() (err error) {

	if _, err = os.Stat(o.rsaFile); !os.IsNotExist(err) {
		return
	}

	bitSize := 4096

	// Generate RSA key.
	if o.private, err = rsa.GenerateKey(rand.Reader, bitSize); err != nil {
		return
	}

	// Extract public component.
	o.public = o.private.Public().(*rsa.PublicKey)

	// Encode private key to PKCS#1 ASN.1 PEM.
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(o.private),
		},
	)

	// Encode public key to PKCS#1 ASN.1 PEM.
	bytes, err := x509.MarshalPKIXPublicKey(o.public)
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: bytes,
		},
	)

	if err = os.MkdirAll(o.keysFolder, 0700); err != nil {
		return
	}

	if err = ioutil.WriteFile(o.rsaFile, keyPEM, 0700); err != nil {
		return
	}

	if err = ioutil.WriteFile(o.rsaPubFile, pubPEM, 0755); err != nil {
		return
	}
	return
}

func RsaKeysNew(keysFolder string, baseFileName string) *RsaKeys {
	rsaFileName := fmt.Sprintf("%v.rsa", baseFileName)
	rsaPubFileName := fmt.Sprintf("%v.rsa.pub", baseFileName)
	return &RsaKeys{
		rsaFileName:    rsaFileName,
		rsaPubFileName: rsaPubFileName,
		rsaFile:        fmt.Sprintf("%v/%v", keysFolder, rsaFileName),
		rsaPubFile:     fmt.Sprintf("%v/%v", keysFolder, rsaPubFileName),

		keysFolder: keysFolder,
	}
}
