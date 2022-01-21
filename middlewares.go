package http

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kas2000/logger"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"time"
)

type Endpoint func(w http.ResponseWriter, r *http.Request) Response

func JWT(next Endpoint, verifyKey *rsa.PublicKey) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) Response {
		fmt.Println(r.URL.String())
		if r.URL.String() == "/token" || r.URL.String() == "/authenticate" {
			return next(w, r)
		} else {
			t := r.Header.Get("Authorization")
			s := strings.Split(t, " ")
			accessToken := s[1]
			fmt.Println(accessToken)

			token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					return nil, err
				}
				return verifyKey, nil
			})

			if !token.Valid {
				return Unauthorized(111, "token is not valid", "JWT middleware")
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				fmt.Println(claims)
			} else {
				return Unauthorized(112, err.Error(), "JWT middleware")
			}

			response := next(w, r)

			return response
		}
	}
}

func Json(next Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d := next(w, r)
		if d == nil {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		for k, v := range d.Headers() {
			w.Header().Set(k, v)
		}

		statusCode := d.StatusCode()
		if statusCode == 302 || statusCode == 301 {
			http.Redirect(w, r, d.Response().(string), statusCode)
			return
		}

		w.WriteHeader(d.StatusCode())
		err := json.NewEncoder(w).Encode(d.Response())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func Logging(next Endpoint, log logger.Logger) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) Response {
		t1 := time.Now()
		rec := httptest.NewRecorder()
		log.Debug(r.Method + " " + r.URL.String() + " " + r.Header.Get("X-Request-Id"))
		response := next(w, r)

		dumpResp, _ := httputil.DumpResponse(rec.Result(), true)
		dumpReq, _ := httputil.DumpRequest(r, true)

		// we copy the captured response headers to our new response
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}

		//grab the captured response body
		//data := rec.Body.Bytes()
		//w.WriteHeader(rec.Result().StatusCode)
		//_, _ = w.Write(data)
		dur := time.Since(t1)
		log.Debug(string(dumpReq) + "\n" + string(dumpResp) + " " + dur.String())

		return response
	}
}