package log

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/femaref/reliable_conn"
	"io"
	"os"
)

type LoggingConfig struct {
	Host string `yaml:"host"`
	TLS bool `yaml:"tls"`
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

var Logger = logrus.New()
var Local = logrus.New()

func init() {
	// Output to stderr instead of stdout, could also be a file.
	Logger.Out = os.Stdout
	Local.Out = os.Stdout
	RedirectStdlogOutput(Logger)
	DefaultLogger = Logger
}

func Configure(appName string, cfg LoggingConfig) (io.Closer, error) {
	if cfg.Host != "" {
		conn, err := reliable_conn.Dial("tcp", cfg.Host)

		if err != nil {
			return nil, err
		}

		hook, err := logrus_logstash.NewHookWithFieldsAndConn(conn, appName, logrus.Fields{})
		if err != nil {
			return nil, err
		}
		hook.WithPrefix("_")

		if err != nil {
			return nil, err
		}
		Logger.Hooks.Add(hook)
	}
	return Setup(appName)
}

func Setup(appName string) (io.Closer, error) {
	f, err := os.OpenFile(fmt.Sprintf("%s.log", appName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Logger.Out = io.MultiWriter(f, Logger.Out)
		Local.Out = io.MultiWriter(f, Local.Out)
	}

	return f, err
}
