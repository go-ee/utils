package net

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"crypto/rsa"
	"os/user"
)

type UserCredentials struct {
	Username string
	Password string
}

type Response struct {
	Data string `json:"data"`
}

type AccountToken struct {
	Account interface{}
	Token   string `json:"token"`
}

type JwtController struct {
	privKeyPath   string //app.rsa, e.g. $ openssl genrsa -out app.rsa 1024
	pubKeyPath    string //app.rsa.pub, e.g $ openssl rsa -in app.rsa -pubout > app.rsa.pub
	useHttpCookie bool

	authenticate func(UserCredentials) (ret interface{}, err error)

	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
}

func NewJwtController(privKeyPath, pubKeyPath string, useHttpCookie bool,
	authenticator func(UserCredentials) (ret interface{}, err error)) *JwtController {
	return &JwtController{privKeyPath: privKeyPath, pubKeyPath: pubKeyPath, useHttpCookie: useHttpCookie,
		authenticate: authenticator}
}

func NewJwtControllerApp(appName string, authenticator func(UserCredentials) (ret interface{}, err error)) (ret *JwtController) {
	if usr, err := user.Current(); err == nil {
		ret = NewJwtController(
			fmt.Sprintf("%v/.rsa/%v.rsa", usr.HomeDir, appName),
			fmt.Sprintf("%v/.rsa/%v.rsa.pub", usr.HomeDir, appName), true,
			authenticator)
		if err := ret.Setup(); err != nil {
			panic(err)
		}
		return
	} else {
		panic(err)
	}
	return
}

func (o *JwtController) Setup() (err error) {
	var keyBytes []byte
	if keyBytes, err = ioutil.ReadFile(o.privKeyPath); err == nil {
		o.signKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	}

	if err != nil {
		return
	}

	if keyBytes, err = ioutil.ReadFile(o.pubKeyPath); err == nil {
		o.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	}
	return
}

func (o *JwtController) LoginHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user UserCredentials
		if err := Decode(&user, r); err != nil {
			ResponseResultErr(err, "Can't retrieve credentials", http.StatusForbidden, w)
			return
		}

		if account, err := o.authenticate(user); err != nil {
			ResponseResultErr(err, "Wrong credentials", http.StatusForbidden, w)
		} else {
			token := jwt.New(jwt.SigningMethodRS256)
			claims := make(jwt.MapClaims)
			claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
			claims["iat"] = time.Now().Unix()
			token.Claims = claims

			if tokenString, err := token.SignedString(o.signKey); err != nil {
				ResponseResultErr(err, "Error while signing the token", http.StatusInternalServerError, w)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				if o.useHttpCookie {
					expireCookie := time.Now().Add(time.Hour * 1)
					cookie := http.Cookie{Name: "Auth", Value: tokenString, Expires: expireCookie, HttpOnly: true}
					http.SetCookie(w, &cookie)
				}
				ResponseJson(AccountToken{Account: account, Token: tokenString}, w)
			}
		}
	})
}

func (o *JwtController) LogoutHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o.Logout(w)
	})
}

func (o *JwtController) ValidateTokenHandler(protected http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o.ValidateToken(w, r, protected)
	})
}

func (o *JwtController) ValidateToken(w http.ResponseWriter, r *http.Request, next http.Handler) {

	token, err := request.ParseFromRequest(r, o,
		func(token *jwt.Token) (interface{}, error) {
			return o.verifyKey, nil
		})

	if err == nil {
		if token.Valid {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Token is not valid")
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized access to this resource")
	}
}

func (o *JwtController) Logout(w http.ResponseWriter) {
	if o.useHttpCookie {
		deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now()}
		http.SetCookie(w, &deleteCookie)
	}
}

func (o *JwtController) ExtractToken(r *http.Request) (ret string, err error) {
	if o.useHttpCookie {
		var cookie *http.Cookie
		if cookie, err = r.Cookie("Auth"); err == nil {
			ret = cookie.Value
		}
	} else {
		ret, err = request.AuthorizationHeaderExtractor.ExtractToken(r)
	}
	return
}
