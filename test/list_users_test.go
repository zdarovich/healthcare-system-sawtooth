package test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jamiealquiza/tachymeter"
	"healthcare-system-sawtooth/client/lib"
	"healthcare-system-sawtooth/client/user"
	"path"
	"testing"
	"time"
)

// Test_ListUsers benchmarks the listing request of all users on the blockchain
func Test_ListUsers(t *testing.T) {
	stats := tachymeter.New(&tachymeter.Config{Size: requestSamples100})
	testKeyPath := "resources/keys"
	name := uuid.New().String()
	lib.GenerateKey(name, testKeyPath)
	privKeyPath := path.Join(testKeyPath, name+".priv")
	cli, err := user.NewUserClient(name, privKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.Register(name)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	err = cli.ListUsers()
	if err != nil {
		t.Fatal(err)
	}
	stats.AddTime(time.Since(start))
	t.Log("List users benchmark \n")
	t.Log(stats.Calc())
	for addr, u := range cli.QueryCache {
		fmt.Println(addr)
		fmt.Println(u)
	}
}
