package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

type Encryptor struct {
	cipher.AEAD
}

func NewEncryptor(passphrase string) (ret *Encryptor, err error) {
	var block cipher.Block
	if block, err = aes.NewCipher([]byte(createHash(passphrase))); err == nil {
		var gcm cipher.AEAD
		if gcm, err = cipher.NewGCM(block); err == nil {
			ret = &Encryptor{AEAD: gcm}
		}
	}
	return
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (o *Encryptor) Encrypt(data []byte) (ret []byte, err error) {
	nonce := make([]byte, o.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}
	ret = o.Seal(nonce, nonce, data, nil)
	return
}

func (o *Encryptor) Decrypt(data []byte) (ret []byte, err error) {
	nonceSize := o.NonceSize()
	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	ret, err = o.Open(nil, nonce, cipherText, nil)
	return
}

func (o *Encryptor) EncryptFile(filename string, data []byte) (err error) {
	var f *os.File
	if f, err = os.Create(filename); err != nil {
		return
	}
	defer f.Close()
	var decryptedData []byte
	if decryptedData, err = o.Encrypt(data); err == nil {
		_, err = f.Write(decryptedData)
	}
	return
}

func (o *Encryptor) DecryptFile(filename string) (ret []byte, err error) {
	var data []byte
	if data, err = ioutil.ReadFile(filename); err == nil {
		ret, err = o.Decrypt(data)
	}
	return
}

func (o *Encryptor) EncryptInstance(v interface{}) (ret []byte, err error) {
	if jsonData, err := json.Marshal(v); err == nil {
		ret, err = o.Encrypt(jsonData)
	}
	return
}

func (o *Encryptor) DecryptInstance(v interface{}, data []byte) (err error) {
	var decrypted []byte
	if decrypted, err = o.Decrypt(data); err == nil {
		err = json.Unmarshal(decrypted, &v)
	}
	return
}
