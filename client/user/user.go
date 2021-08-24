package user

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"healthcare-system-sawtooth/client/db"
	"healthcare-system-sawtooth/tp/storage"

	"github.com/sirupsen/logrus"
	"healthcare-system-sawtooth/client/crypto"
	"healthcare-system-sawtooth/client/lib"
	tpCrypto "healthcare-system-sawtooth/crypto"
	tpPayload "healthcare-system-sawtooth/tp/payload"
	tpUser "healthcare-system-sawtooth/tp/user"
)

// Client provides the platform for user storing files in P2P network.
type Client struct {
	User         *tpUser.User
	PWD          string
	lastQueryEnd string
	QueryCache   map[string]*tpUser.User
	*lib.ClientFramework
}

// NewUserClient is the construct for User's Client.
func NewUserClient(name, keyFile string) (*Client, error) {
	c, err := lib.NewClientFramework(name, lib.ClientCategoryUser, keyFile)
	if err != nil {
		return nil, err
	}
	var u *tpUser.User
	userBytes, _ := c.GetData()
	if userBytes != nil {
		lib.Logger.WithField("username", name).Info("user login success")
		u, err = tpUser.UserFromBytes(userBytes)
		if err != nil {
			u = nil
			lib.Logger.Error(err)
		}
	}

	cli := &Client{
		User:            u,
		ClientFramework: c,
		QueryCache:      make(map[string]*tpUser.User),
	}
	go func() {
		var data []byte
		for {
			data = <-cli.State
			u, err := tpUser.UserFromBytes(data)
			if err != nil {
				lib.Logger.Errorf("failed to sync: %v", err)
			} else {
				cli.User = u
			}
		}
	}()
	return cli, nil
}

// Sync get user's data from blockchain.
func (c *Client) Sync() error {
	userBytes, err := c.GetData()
	if err != nil {
		return err
	}
	u, err := tpUser.UserFromBytes(userBytes)
	if err != nil {
		return err
	}
	c.User = u
	return nil
}

// UserRegister register user in the blockchain.
func (c *Client) UserRegister() error {
	err := c.Register(c.Name)
	if err != nil {
		return err
	}
	lib.Logger.WithFields(logrus.Fields{
		"name":       c.Name,
		"public key": c.GetPublicKey(),
		"address":    c.GetAddress(),
	}).Info("user register success")
	return c.Sync()
}

// GetSize returns the total size of files stored in P2P network.
func (c *Client) GetSize() int64 {
	return c.User.Root.Repo.Size
}

// CreatePatientData create new data of the source.
// After sending transaction, upload data into MongoDB.
func (c *Client) CreatePatientData(name, data string) error {
	hash, err := crypto.CalDataHash(data)
	if err != nil {
		return err
	}
	// Check Destination Path exists
	_, err = c.User.Root.GetData(hash, c.User.Name)
	if err != nil {
		return err
	}
	keyAES := tpCrypto.GenerateRandomAESKey(lib.AESKeySize)
	info, err := crypto.GenerateDataInfo(name, data, c.GetPublicKey(), c.User.Name, tpCrypto.BytesToHex(keyAES))
	if err != nil {
		return err
	}
	err = c.User.Root.CreateData(info)
	if err != nil {
		return err
	}
	addresses := []string{c.GetAddress()}
	err = c.SendTransactionAndWaiting([]tpPayload.SeaStoragePayload{{
		Action:   tpPayload.UserCreateData,
		Name:     c.Name,
		DataInfo: info,
	}}, addresses, addresses)
	return nil
}

func (c *Client) ListPatientData() ([]storage.INode, error) {
	return c.User.Root.Repo.INodes, nil
}

func (c *Client) GetPatientData(hash string) (string, error) {
	// Check Destination Path exists
	di, err := c.User.Root.GetData(hash, c.User.Name)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	d, err := db.GetByHash(ctx, hash)
	if err != nil {
		return "", err
	}
	_, out, err := crypto.DecryptData([]byte(d.Payload), []byte(di.Key))
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (c *Client) ListSharedPatientData(username string) ([]storage.INode, error) {
	_, user, err := c.GetUser(username)
	if err != nil {
		return nil, err
	}
	return user.Root.Repo.INodes, nil
}

func (c *Client) GetSharedPatientData(hash, username string) (string, error) {
	_, user, err := c.GetUser(username)
	if err != nil {
		return "", err
	}
	di, err := user.Root.GetData(hash, c.User.Name)
	if err != nil {
		return "", err
	}
	keyAES, err := c.DecryptFileKey(di.Key)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	d, err := db.GetByHash(ctx, hash)
	if err != nil {
		return "", err
	}
	_, out, err := crypto.DecryptData([]byte(d.Payload), keyAES)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (c *Client) ShareData(hash, username string) error {
	di, err := c.User.Root.GetData(hash, c.User.Name)
	if err != nil {
		return err
	}
	addresses := []string{c.GetAddress()}

	_, user, err := c.GetUser(username)
	keyAES, err := c.DecryptFileKey(di.Key)
	if err != nil {
		fmt.Println("failed to decrypt file key:", err)
		return err
	}

	info, err := crypto.GenerateSharedDataInfo(di.Name, user.PublicKey, user.Name, di.Hash, tpCrypto.BytesToHex(keyAES), di.Size)
	if err != nil {
		return err
	}

	err = c.User.Root.CreateData(info)
	if err != nil {
		return err
	}
	return c.SendTransactionAndWaiting([]tpPayload.SeaStoragePayload{{
		Action:   tpPayload.UserCreateData,
		Name:     c.Name,
		DataInfo: info,
	}}, addresses, addresses)
}

func (c *Client) GetUser(username string) (string, *tpUser.User, error) {
	for a, u := range c.QueryCache {
		if u.Name == username {
			return a, u, nil
		}
	}
	err := c.ListUsers()
	if err != nil {
		return "", nil, errors.New("failed to get user")
	}
	for a, u := range c.QueryCache {
		if u.Name == username {
			return a, u, nil
		}
	}
	return "", nil, errors.New("no such user")
}

// ListUsers get the query cache for list shared files.
func (c *Client) ListUsers() error {
	users, err := lib.ListUsers(c.lastQueryEnd, lib.DefaultQueryLimit+1)
	if err != nil {
		return err
	}
	for k := range c.QueryCache {
		delete(c.QueryCache, k)
	}
	for i := 1; i < len(users); i++ {
		m := users[i].(map[string]interface{})
		userBytes, err := base64.StdEncoding.DecodeString(m["data"].(string))
		if err != nil {
			continue
		}
		u, err := tpUser.UserFromBytes(userBytes)
		if err != nil {
			continue
		}
		c.QueryCache[m["address"].(string)] = u
	}
	return nil
}

// check user whether in the query cache.
// If exists, it will return directly.
// Else it will get user's data from blockchain.
func (c *Client) checkUser(addr string) (*tpUser.User, error) {
	u, ok := c.QueryCache[addr]
	if !ok {
		userBytes, err := lib.GetStateData(addr)
		if err != nil {
			return nil, err
		}
		u, err = tpUser.UserFromBytes(userBytes)
		if err != nil {
			return nil, err
		}
		c.QueryCache[addr] = u
	}
	return u, nil
}
