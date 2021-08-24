package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/goverview/libs"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var logger = logrus.New()

// InitLog init log
func InitLog(options *libs.Options) {
	logger.SetLevel(logrus.InfoLevel)
	logger = &logrus.Logger{
		Out: os.Stderr,
		//Out:   mwr,
		Level: logrus.InfoLevel,
		Formatter: &prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
		},
	}

	if !options.Debug && !options.Verbose {
		logger.SetOutput(ioutil.Discard)
		return
	}
	if options.Verbose == true {
		logger.SetOutput(os.Stdout)
	}
	if options.Debug == true {
		logger.SetLevel(logrus.DebugLevel)
	}
}

// PrintLine print seperate line
func PrintLine() {
	dash := color.HiWhiteString("-")
	fmt.Println(strings.Repeat(dash, 40))
}

// GoodF print good message
func GoodF(format string, args ...interface{}) {
	good := color.HiGreenString("[+]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

// BannerF print info message
func BannerF(format string, data string) {
	banner := fmt.Sprintf("%v%v%v ", color.WhiteString("["), color.BlueString(format), color.WhiteString("]"))
	fmt.Printf("%v%v\n", banner, color.HiGreenString(data))
}

// BlockF print info message
func BlockF(name string, data string) {
	banner := fmt.Sprintf("%v%v%v ", color.WhiteString("["), color.GreenString(name), color.WhiteString("]"))
	fmt.Printf(fmt.Sprintf("%v%v\n", banner, data))
}

// BadBlockF print info message
func BadBlockF(name string, data string) {
	banner := fmt.Sprintf("%v%v%v ", color.WhiteString("["), color.RedString(name), color.WhiteString("]"))
	fmt.Printf(fmt.Sprintf("%v%v\n", banner, data))
}

// InforF print info message
func InforF(format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...))
}

// ErrorF print good message
func ErrorF(format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...))
}

// WarningF print good message
func WarningF(format string, args ...interface{}) {
	good := color.YellowString("[!]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

// DebugF print debug message
func DebugF(format string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(format, args...))
}
