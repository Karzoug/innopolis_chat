package main

import (
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func newLogger() (logger zerolog.Logger, closeFn func() error) {
	l := &lumberjack.Logger{
		Filename:   "app.log", // example: "/var/log/mychat/app.log"
		MaxSize:    1,         // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}

	return zerolog.New(l).With().Timestamp().Logger(), l.Close
}
