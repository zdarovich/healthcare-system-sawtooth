package test

import (
	"github.com/google/uuid"
	"github.com/jamiealquiza/tachymeter"
	"healthcare-system-sawtooth/client/lib"
	"healthcare-system-sawtooth/client/user"
	"path"
	"testing"
	"time"
)

func Test_Register_User_100_times(t *testing.T) {
	stats := tachymeter.New(&tachymeter.Config{Size: requestSamples100})
	testKeyPath := "resources/keys"

	var success, fails int
	for i := 0; i < requestSamples100; i++ {
		name := uuid.New().String()
		lib.GenerateKey(name, testKeyPath)
		privKeyPath := path.Join(testKeyPath, name+".priv")
		cli, err := user.NewUserClient(name, privKeyPath)
		if err != nil {
			fails++
			t.Error(err)
			continue
		}
		start := time.Now()
		err = cli.UserRegister()
		stats.AddTime(time.Since(start))
		if err != nil {
			fails++
			t.Error(err)
			continue
		}
		success++
	}
	t.Log("User identity register benchmark \n")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Log(stats.Calc())
}
