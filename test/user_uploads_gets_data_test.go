package test

import (
	"github.com/google/uuid"
	"github.com/jamiealquiza/tachymeter"
	"healthcare-system-sawtooth/client/lib"
	"healthcare-system-sawtooth/client/user"
	"math/rand"
	"path"
	"testing"
	"time"
)

func Test_User_Uploads_Gets_Data_100_times(t *testing.T) {
	stats := tachymeter.New(&tachymeter.Config{Size: requestSamples100})
	testKeyPath := "resources/keys"
	name := uuid.New().String()
	lib.GenerateKey(name, testKeyPath)
	privKeyPath := path.Join(testKeyPath, name+".priv")
	cli, err := user.NewUserClient(name, privKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UserRegister()
	if err != nil {
		t.Fatal(err)
	}
	dataHashMap := make([]string, 0)
	var success, fails int
	for i := 0; i < requestSamples100; i++ {
		dataName := uuid.New().String()
		data := RandStringRunes(rand.Intn(100))

		start := time.Now()
		dataInfo, err := cli.CreatePatientData(dataName, data)
		if err != nil {
			t.Error(err)
			fails++
			continue
		}
		stats.AddTime(time.Since(start))
		dataHashMap = append(dataHashMap, dataInfo.Hash)
		success++
	}
	t.Log("User uploads data benchmark \n")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Log(stats.Calc())

	success, fails = 0, 0
	stats.Reset()
	for _, hash := range dataHashMap {
		start := time.Now()
		_, err := cli.GetPatientData(hash)
		if err != nil {
			t.Error(err)
			fails++
			continue
		}
		stats.AddTime(time.Since(start))
		success++
	}
	t.Log("User gets own data benchmark")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Log(stats.Calc())
}
