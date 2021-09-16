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

// Test_User_Uploads_Gets_Data_100_times benchmarks data upload by user A. Also, it tests data access by user A. It does it 100 times.
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
	var memoryUsed int

	dataHashMap := make([]string, 0)
	var success, fails int
	for i := 0; i < requestSamples100; i++ {
		dataName := uuid.New().String()
		data := RandStringRunes(rand.Intn(100))

		start := time.Now()
		dataInfo, err := cli.CreatePatientData(dataName, data, 0)
		if err != nil {
			t.Error(err)
			fails++
			continue
		}
		memoryUsed += len(data)
		stats.AddTime(time.Since(start))
		dataHashMap = append(dataHashMap, dataInfo.Hash)
		success++
	}
	t.Log("User uploads data benchmark \n")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Logf("throughput(bytes/second): %f%% \n", float64(memoryUsed)/stats.Calc().Time.Cumulative.Seconds())
	t.Log(stats.Calc())

	memoryUsed = 0
	success, fails = 0, 0
	stats.Reset()
	for _, hash := range dataHashMap {
		start := time.Now()
		_, data, err := cli.GetPatientData(hash)
		if err != nil {
			t.Error(err)
			fails++
			continue
		}
		memoryUsed += len(data)
		stats.AddTime(time.Since(start))
		success++
	}
	t.Log("User gets own data benchmark")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Logf("throughput(bytes/second): %f%% \n", float64(memoryUsed)/stats.Calc().Time.Cumulative.Seconds())
	t.Log(stats.Calc())
}
