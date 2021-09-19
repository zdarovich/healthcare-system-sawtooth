package test

import (
	"github.com/sirupsen/logrus"
	"healthcare-system-sawtooth/client/lib"
	"io/ioutil"
	"math/rand"
	"time"
)

var (
	requestSamples100 = 100
	letterRunes       = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init() {
	rand.Seed(time.Now().UnixNano())
	lib.Logger = logrus.New()
	lib.Logger.Out = ioutil.Discard
	lib.MongoDbUrl = lib.DefaultMongoDbUrl
	lib.ValidatorURL = lib.DefaultValidatorURL
	lib.TPURL = lib.DefaultTPURL
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
