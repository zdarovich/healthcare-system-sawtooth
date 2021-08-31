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

func Test_User_Shares_Other_User_Gets_Data_100_times(t *testing.T) {
	stats := tachymeter.New(&tachymeter.Config{Size: requestSamples100})
	testKeyPath := "resources/keys"

	testUsersClients := make(map[string]*user.Client)
	for i := 0; i < requestSamples100; i++ {
		randName := uuid.New().String()
		lib.GenerateKey(randName, testKeyPath)
		privKeyPath := path.Join(testKeyPath, randName+".priv")
		cli, err := user.NewUserClient(randName, privKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		testUsersClients[randName] = cli
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
	data := RandStringRunes(rand.Intn(100))
	dataInfo, err := cli.CreatePatientData(dataName, data)
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
		stats.AddTime(time.Since(start))
		success++
	}
	t.Log("User shares data benchmark \n")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Log(stats.Calc())

	success, fails = 0, 0
	stats.Reset()
	for _, cli := range testUsersClients {
		start := time.Now()
		_, err = cli.GetSharedPatientData(dataInfo.Hash, creator)
		if err != nil {
			t.Error(err)
			fails++
			continue
		}
		stats.AddTime(time.Since(start))
		success++
	}
	t.Log("Other user gets shared data benchmark")
	t.Logf("succes rate: %f%% \n", float64(success)/float64(requestSamples100)*100)
	t.Logf("fail rate: %f%% \n", float64(fails)/float64(requestSamples100)*100)
	t.Log(stats.Calc())
}
