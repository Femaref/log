package log

import (
    "net/http"
    "io/ioutil"
    "strings"
    "bytes"
    "encoding/json"
)

func LogBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" {
				if r.ContentLength > 0 {
					body, err := ioutil.ReadAll(r.Body)
					if err != nil {
						RequestLog(r).Warnf("Could not read body")
						return
					}
					// restore the body so the following methods have the original behaviour
					r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
					if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
						var raw json.RawMessage
						err = json.Unmarshal(body, &raw)
						if err != nil {
							RequestLog(r).Warnf("Could not unmarshal json body")
							return
						}

						indented, _ := json.MarshalIndent(&raw, "", "  ")
						RequestLog(r).Infof(string(indented))
					} else {
						RequestLog(r).Infof(string(body))
					}

				}
			}
		}(w, r)
		next.ServeHTTP(w, r)
	})
}
