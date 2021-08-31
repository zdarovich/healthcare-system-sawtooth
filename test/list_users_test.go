package test

import (
	"fmt"
	"github.com/google/uuid"
	"healthcare-system-sawtooth/client/lib"
	"healthcare-system-sawtooth/client/user"
	"path"
	"testing"
)

func Test_ListUsers(t *testing.T) {
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
	err = cli.ListUsers()
	if err != nil {
		t.Fatal(err)
	}
	for addr, u := range cli.QueryCache {
		fmt.Println(addr)
		fmt.Println(u)
	}
}
