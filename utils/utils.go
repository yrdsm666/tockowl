package utils

import (
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func LogOnError(err error, msg string, log *logrus.Entry) {
	if err != nil {
		log.WithError(err).Warn(msg)
	}
}

func ChangeWD2ProjectRoot(relativePath string) {
	_, filename, _, _ := runtime.Caller(1)
	dir := path.Join(path.Dir(filename), relativePath)
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
