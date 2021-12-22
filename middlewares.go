package http

import (
	"encoding/json"
	"github.com/kas2000/logger"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"
)

type Endpoint func(w http.ResponseWriter, r *http.Request) Response

func Logging(final Endpoint, log logger.Logger) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) Response {
		t1 := time.Now()
		rec := httptest.NewRecorder()
		log.Debug(r.Method + " " + r.URL.String() + " " + r.Header.Get("X-Request-Id"))
		response := final(w, r)

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