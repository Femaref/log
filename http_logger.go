//Copyright 2016-current lg authors. All rights reserved.
//from https://github.com/pressly/lg/

package log

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pressly/chi/middleware"
)

// RequestLogger is a middleware for the github.com/Sirupsen/logrus to log requests.
// It is equipt to handle recovery in case of panics and record the stack trace
// with a panic log-level.
func RequestLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	httpLogger := &HTTPLogger{logger}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := httpLogger.NewLogEntry(r)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				t2 := time.Now()

				// Recover and record stack traces in case of a panic
				if rec := recover(); rec != nil {
					entry.Panic(rec, debug.Stack())
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				// Log the entry, the request is complete.
				entry.Write(ww.Status(), ww.BytesWritten(), t2.Sub(t1))
			}()

			r = r.WithContext(WithLogEntry(r.Context(), entry))
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

type HTTPLogger struct {
	Logger *logrus.Logger
}

func (l *HTTPLogger) NewLogEntry(r *http.Request) *HTTPLoggerEntry {
	entry := &HTTPLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
	logFields := logrus.Fields{}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.Logger = entry.Logger.WithFields(logFields)
	return entry
}

type HTTPLoggerEntry struct {
	Logger logrus.FieldLogger // field logger interface, created by RequestLogger
	Level  *logrus.Level      // intended log level to write when request finishes
}

func (l *HTTPLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elasped_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
	})

	if status >= 200 && status < 400 {
		l.Logger.Info("successful request")
	} else if status >= 400 && status < 600 {
		l.Logger.Warn("error while request")
	}

}

func (l *HTTPLoggerEntry) Panic(rec interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", rec),
	})
	panicLevel := logrus.PanicLevel
	l.Level = &panicLevel
}

// PrintPanics is a development middleware that preempts the request logger
// and prints a panic message and stack trace to stdout.
func PrintPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Printf("\nPANIC: %+v\n", rec)
				fmt.Printf("%s\n", debug.Stack())
			}
		}()
		next.ServeHTTP(w, r)
	})
}
