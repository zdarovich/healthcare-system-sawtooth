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

// Test_User_Shares_Other_User_Gets_Data_100_times benchmarks data sharing by user A. Also, it tests shared data access by user B. It does it 100 times.
func Test_User_Shares_Other_User_Gets_Data_100_times(t *testing.T) {
	stats := tachymeter.New(&tachymeter.Config{Size: requestSamples100})
	testKeyPath := "resources/keys"

	var memoryUsed int
	testUsersClients := make(map[string]string)
	for i := 0; i < requestSamples100; i++ {
		randName := uuid.New().String()
		lib.GenerateKey(randName, testKeyPath)
		privKeyPath := path.Join(testKeyPath, randName+".priv")
		cli, err := user.NewUserClient(randName, privKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		testUsersClients[randName] = privKeyPath
		err = cli.UserRegister()
		if err != nil {
			t.Fatal(err)
		}
	}
	creator := uuid.New().String()
	lib.GenerateKey(creator, testKeyPath)
	privKeyPath := path.Join(testKeyPath, creator+".priv")
	cli, err := user.NewUserClient(creator, privKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UserRegister()
	if err != nil {
		t.Fatal(err)
	}
	dataName := uuid.New().String()
	randInt := rand.Intn(100)
	if randInt <= 0 {
		randInt = 20
	}
	data := RandStringRunes(randInt)
	dataInfo, err := cli.CreatePatientData(dataName, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	var success, fails int
	for userName, _ := range testUsersClients {
		start := time.Now()
		err = cli.ShareData(dataInfo.Hash, userName)
		if err != nil {
			t.Error(err)
			fails++
			continue
		}
		memoryUsed += len(data)
		stats.AddTime(time.Since(start))
		success++
	}
	t.Log("User shares data benchmark \n")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Logf("throughput(bytes/second): %f \n", float64(memoryUsed)/stats.Calc().Time.Cumulative.Seconds())
	t.Logf("memory (bytes): %d \n", memoryUsed)
	t.Logf("avg memory per transaction (bytes): %d \n", memoryUsed/requestSamples100)
	t.Log(stats.Calc())

	success, fails = 0, 0
	memoryUsed = 0
	stats.Reset()
	for name, privKeyPath := range testUsersClients {
		cli, err := user.NewUserClient(name, privKeyPath)
		if err != nil {
			t.Fatal(err)
		}

		start := time.Now()
		sharedData, err := cli.ListSharedPatientData(creator)
		if err != nil {
			t.Error(err)
			fails++
			continue
		}
		if len(sharedData) == 0 {
			t.Error("no shared data was found")
			fails++
			continue
		}
		_, data, err = cli.GetSharedPatientData(sharedData[0].GetHash(), creator)
		if err != nil {
			t.Error(err)
			fails++
			continue
		}
		memoryUsed += len(data)
		stats.AddTime(time.Since(start))
		success++
	}
	t.Log("Other user gets shared data benchmark")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Logf("throughput(bytes/second): %f \n", float64(memoryUsed)/stats.Calc().Time.Cumulative.Seconds())
	t.Logf("memory (bytes): %d \n", memoryUsed)
	t.Logf("avg memory per transaction (bytes): %d \n", memoryUsed/requestSamples100)
	t.Log(stats.Calc())
}
