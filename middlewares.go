package http

import (
	"github.com/kas2000/logger"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"
)

func Logging(next http.HandlerFunc, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		rec := httptest.NewRecorder()
		log.Debug(r.Method + " " + r.URL.String() + " " + r.Header.Get("X-Request-Id"))
		next.ServeHTTP(w, r)

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
	}
}