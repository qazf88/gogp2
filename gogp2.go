package gogp2

import Log "github.com/qazf88/golog"

func Gogp2LogLevel(level int) {
	Log.LogLevel(level)
}

func Gogp2LogFormat(format int) {
	Log.LogFormat(format)
}

func Gogp2LogTimeFormat(format string) {
	Log.LogTimeFormat(format)
}
