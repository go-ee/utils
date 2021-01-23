package net

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
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
	//app.rsa, e.g. $ openssl genrsa -out app.rsa 1024
	//app.rsa.pub, e.g $ openssl rsa -in app.rsa -pubout > app.rsa.pub

	appName       string // needed for cookie
	rsaKeys       *RsaKeys
	useHttpCookie bool

	authenticate func(UserCredentials) (ret interface{}, err error)
}

func NewJwtController(appName string, rsaKeys *RsaKeys, useHttpCookie bool,
	authenticator func(UserCredentials) (ret interface{}, err error)) *JwtController {

	return &JwtController{appName: appName, rsaKeys: rsaKeys,
		useHttpCookie: useHttpCookie, authenticate: authenticator}
}

func NewJwtControllerApp(certsFolder string, appName string,
	authenticator func(UserCredentials) (ret interface{}, err error)) (ret *JwtController, err error) {

	keyBaseName := strings.ToLower(appName)
	rsaKeys := RsaKeysNew(certsFolder, keyBaseName)

	ret = NewJwtController(
		appName,
		rsaKeys, true,
		authenticator)
	err = ret.Setup()
	return
}

func (o *JwtController) Setup() (err error) {
	err = o.rsaKeys.LoadOrCreate()
	return
}

func (o *JwtController) LoginHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user UserCredentials
		if err := Decode(&user, r); err != nil {
			ResponseResultErr(err, "can't retrieve credentials", nil, http.StatusForbidden, w)
			return
		}

		if account, err := o.authenticate(user); err != nil {
			ResponseResultErr(err, "wrong credentials", nil, http.StatusForbidden, w)
		} else {
			token := jwt.New(jwt.SigningMethodRS256)
			claims := make(jwt.MapClaims)
			claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
			claims["iat"] = time.Now().Unix()
			token.Claims = claims

			if tokenString, err := token.SignedString(o.rsaKeys.private); err != nil {
				ResponseResultErr(err, "wrror while signing the token", nil, http.StatusInternalServerError, w)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				if o.useHttpCookie {
					expireCookie := time.Now().Add(time.Hour * 1)
					cookie := http.Cookie{Name: "auth", Value: tokenString, Expires: expireCookie, HttpOnly: true}
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
			return o.rsaKeys.Public(), nil
		})

	if err == nil {
		if token.Valid {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "token is not valid")
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "unauthorized access to this resource")
	}
}

func (o *JwtController) Logout(w http.ResponseWriter) {
	if o.useHttpCookie {
		deleteCookie := http.Cookie{Name: o.appName, Value: "none", Expires: time.Now()}
		http.SetCookie(w, &deleteCookie)
	}
}

func (o *JwtController) ExtractToken(r *http.Request) (ret string, err error) {
	if o.useHttpCookie {
		var cookie *http.Cookie
		if cookie, err = r.Cookie(o.appName); err == nil {
			ret = cookie.Value
		}
	} else {
		ret, err = request.AuthorizationHeaderExtractor.ExtractToken(r)
	}
	return
}
