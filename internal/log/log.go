package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	defaultLogger = logrus.StandardLogger()
	grpcGo        = logrus.New()

	// Loggers is convenient when you want to apply configuration to all
	// loggers
	Loggers = []*logrus.Logger{defaultLogger, grpcGo}
)

func init() {
	// This ensures that any log statements that occur before
	// the configuration has been loaded will be written to
	// stdout instead of stderr
	for _, l := range Loggers {
		l.Out = os.Stdout
	}
}

// Configure sets the format and level on all loggers. It applies level
// mapping to the GrpcGo logger.
func Configure(format string, level string) {
	switch format {
	case "json":
		for _, l := range Loggers {
			l.Formatter = &logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.000Z"}
		}
	case "":
		// Just stick with the default
	default:
		logrus.WithField("format", format).Fatal("invalid logger format")
	}

	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrusLevel = logrus.InfoLevel
	}

	for _, l := range Loggers {
		if l == grpcGo {
			l.SetLevel(mapGrpcLogLevel(logrusLevel))
		} else {
			l.SetLevel(logrusLevel)
		}
	}
}

func mapGrpcLogLevel(level logrus.Level) logrus.Level {
	// grpc-go is too verbose at level 'info'. So when config.toml requests
	// level info, we tell grpc-go to log at 'warn' instead.
	if level == logrus.InfoLevel {
		return logrus.WarnLevel
	}

	return level
}

// Default is the default logrus logger
func Default() *logrus.Entry { return defaultLogger.WithField("pid", os.Getpid()) }

// GrpcGo is a dedicated logrus logger for the grpc-go library. We use it
// to control the library's chattiness.
func GrpcGo() *logrus.Entry { return grpcGo.WithField("pid", os.Getpid()) }
