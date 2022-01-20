package http

import (
	"encoding/json"
	"fmt"
	"github.com/kas2000/logger"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"
	"strings"
)

type Endpoint func(w http.ResponseWriter, r *http.Request) Response

func JWT(next Endpoint) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) Response {
		t := r.Header.Get("Authorization")
		s := strings.Split(t, " ")
		accessToken := s[1]
		hmacSecret := []byte("Hello world")
		fmt.Println(accessToken)

		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ! ok {
				err := fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				return nil, err
			}
			return hmacSecret, nil
		})

		if !token.Valid {
			return Unauthorized(111, "token is not valid", "JWT middleware")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			fmt.Println(claims)
		} else {
			fmt.Println(err)
		}

		response := next(w, r)

		return response
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