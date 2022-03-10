package http

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kas2000/logger"
	"go.uber.org/zap"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"time"
)

type Endpoint func(w http.ResponseWriter, r *http.Request) Response

func JWT(next Endpoint, verifyKey *rsa.PublicKey) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) Response {
		fmt.Println(r.URL.String())
		if r.URL.String() == "/token" || r.URL.String() == "/authenticate" || r.URL.String() == "/dummy"{
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
		start := time.Now()

		// Read the content
		var bodyBytes []byte
		var bodyString string

		var contentType string
		contentType, _, _ = mime.ParseMediaType(r.Header.Get("Content-Type"))
		if contentType != "" && !strings.HasPrefix(contentType, "multipart/") || contentType == "" {
			if r.Body != nil {
				bodyBytes, _ = ioutil.ReadAll(r.Body)
			}
			// Restore the io.ReadCloser to its original state
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			// Use the content
			bodyString = string(bodyBytes)
		} else {
			bodyString = "file uploading - logs rejected"
		}
		response := next(w, r)

		if response == nil {
			return nil
		}
		dBytes, _ := json.Marshal(response.Response())

		if response.StatusCode()%200 < 100 || response.StatusCode()%300 < 100 {
			log.Warn(LogRequest(r), zap.Any(" body= ",  bodyString),
									zap.Any(" duration= ", time.Since(start)),
									zap.Any(" status= ",  response.StatusCode()),
									zap.Any(" ", string(dBytes)),
									)
		} else {
			log.Debug(LogRequest(r), zap.Any(" body= ",  bodyString),
				zap.Any(" duration= ", time.Since(start)),
				zap.Any(" status= ",  response.StatusCode()),
				zap.Any(" ", string(dBytes)),)
		}

		return response
	}
}

func LogRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	urlPath := fmt.Sprintf("%v %v", r.Method, r.URL)
	request = append(request, urlPath)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		_ = r.ParseForm()
		request = append(request, " ")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, " ") + " "
}