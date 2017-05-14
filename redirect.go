//Copyright 2016-current lg authors. All rights reserved.
//from https://github.com/pressly/lg/

package log

import (
	stdlog "log"

	"github.com/Sirupsen/logrus"
)

func RedirectStdlogOutput(logger *logrus.Logger) {
	// Redirect standard logger
	stdlog.SetOutput(&logRedirectWriter{logger})
	stdlog.SetFlags(0)
}

// Proxy writer for any packages using the standard log.Println() stuff
type logRedirectWriter struct {
	Logger *logrus.Logger
}

func (l *logRedirectWriter) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		l.Logger.Infof("%s", p[:len(p)-1])
	}
	return len(p), nil
}
