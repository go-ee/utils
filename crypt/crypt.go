package crypt

import "golang.org/x/crypto/bcrypt"

func Hash(str string) (ret string, err error) {
	var encrypted []byte
	if encrypted, err = bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost); err == nil {
		ret = string(encrypted)
	}
	return
}

func HashAndEquals(str string, hashed string) (ret bool) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(str))
	return err == nil
}
