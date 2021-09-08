package user

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"healthcare-system-sawtooth/client/db/models"
	"healthcare-system-sawtooth/tp/storage"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"healthcare-system-sawtooth/client/crypto"
	"healthcare-system-sawtooth/client/lib"
	tpCrypto "healthcare-system-sawtooth/crypto"
	tpPayload "healthcare-system-sawtooth/tp/payload"
	tpUser "healthcare-system-sawtooth/tp/user"
)

// Client provides the platform for user storing data.
type Client struct {
	User         *tpUser.User
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
				lib.Logger.Infof("user state: %+v", u)
				for _, k := range u.Root.Keys.Keys {
					lib.Logger.Infof("keys state: %+v", k)
				}
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

// CreatePatientData create new data of the source.
// upload data into MongoDB, then send transaction.
func (c *Client) CreatePatientData(name, data string, accessType uint) (*storage.DataInfo, error) {
	err := c.Sync()
	if err != nil {
		return nil, err
	}
	keyAES := tpCrypto.GenerateRandomAESKey(lib.AESKeySize)
	info, err := crypto.GenerateDataInfo(name, data, c.GetPublicKey(), c.User.Name, tpCrypto.BytesToHex(keyAES), accessType, 0)
	if err != nil {
		return nil, err
	}
	err = c.User.Root.CreateData(info)
	if err != nil {
		return nil, err
	}
	addresses := []string{c.GetAddress()}
	err = c.SendTransactionAndWaiting([]tpPayload.StoragePayload{{
		Action:   tpPayload.UserCreateData,
		Name:     c.Name,
		DataInfo: info,
	}}, addresses, addresses)
	return &info, nil
}

// ListPatientData list all the data owned by the current user
func (c *Client) ListPatientData() ([]storage.INode, error) {
	err := c.Sync()
	if err != nil {
		return nil, err
	}
	var filtered []storage.INode
	for _, n := range c.User.Root.Repo.INodes {
		if n.GetAddr() != c.User.Name {
			continue
		}
		filtered = append(filtered, n)
	}
	return filtered, nil
}

// GetPatientData get the data owned by the current user by hash
func (c *Client) GetPatientData(hash string) (*storage.DataInfo, string, error) {
	err := c.Sync()
	if err != nil {
		return nil, "", err
	}
	// Check Destination Path exists
	di, err := c.User.Root.GetData(hash, c.User.Name)
	if err != nil {
		return nil, "", err
	}
	if di == nil {
		return nil, "", errors.New("data doesn't exist")
	}
	ctx := context.Background()
	d, err := models.GetDataByHashes(ctx, []string{hash})
	if err != nil {
		return nil, "", err
	}
	keyAes, err := c.DecryptDataKey(di.Key)
	if err != nil {
		fmt.Println("failed to decrypt file key:", err)
		return nil, "", err
	}
	_, out, err := crypto.DecryptData(tpCrypto.HexToBytes(d[0].Payload), keyAes)
	if err != nil {
		return nil, "", err
	}

	return di, string(out), nil
}

// ListSharedPatientData lists the data shared by the username
func (c *Client) ListSharedPatientData(username string) ([]storage.INode, error) {
	err := c.Sync()
	if err != nil {
		return nil, err
	}
	_, user, err := c.GetUser(username)
	if err != nil {
		return nil, err
	}

	var hashes []string
	nodeMap := make(map[string]storage.INode, 0)
	for _, n := range user.Root.Repo.INodes {
		if n.GetAddr() != c.User.Name {
			continue
		}
		hashes = append(hashes, n.GetHash())
		nodeMap[n.GetHash()] = n
	}
	ctx := context.Background()
	if len(hashes) == 0 {
		return nil, nil
	}
	datas, err := models.GetDataByHashes(ctx, hashes)
	if err != nil {
		return nil, err
	}
	var filtered []storage.INode
	for _, data := range datas {
		inode, ok := nodeMap[data.Hash]
		if !ok {
			continue
		}
		filtered = append(filtered, inode)
	}
	return filtered, nil
}

// GetSharedPatientData gets the data shared by hash and username
func (c *Client) GetSharedPatientData(hash, username string) (*storage.DataInfo, string, error) {
	err := c.Sync()
	if err != nil {
		return nil, "", err
	}
	_, user, err := c.GetUser(username)
	if err != nil {
		return nil, "", err
	}
	di, err := user.Root.GetData(hash, c.User.Name)
	if err != nil {
		return nil, "", err
	}
	if di == nil {
		return nil, "", errors.New("data doesn't exist")
	}
	keyAES, err := c.DecryptDataKey(di.Key)
	if err != nil {
		return nil, "", err
	}
	ctx := context.Background()
	d, err := models.GetDataByHashes(ctx, []string{hash})
	if err != nil {
		return nil, "", err
	}
	if d == nil {
		return nil, "", errors.New("data doesn't exist")
	}
	_, out, err := crypto.DecryptData(tpCrypto.HexToBytes(d[0].Payload), keyAES)
	if err != nil {
		return nil, "", err
	}

	return di, string(out), nil
}

// ShareData share the data owned by the current user
func (c *Client) ShareData(hash, usernameTo string) error {
	err := c.Sync()
	if err != nil {
		return err
	}
	di, data, err := c.GetPatientData(hash)
	if err != nil {
		fmt.Println("failed to get user:", err)
		return err
	}
	addresses := []string{c.GetAddress()}

	_, userTo, err := c.GetUser(usernameTo)
	if err != nil {
		fmt.Println("failed to get user:", err)
		return err
	}
	dataName := fmt.Sprintf("shared_by_%s_%s", c.Name, di.Name)
	keyAES := tpCrypto.GenerateRandomAESKey(lib.AESKeySize)
	info, err := crypto.GenerateDataInfo(dataName, data, userTo.PublicKey, userTo.Name, tpCrypto.BytesToHex(keyAES), di.AccessType, 0)
	if err != nil {
		return err
	}

	err = c.User.Root.CreateData(info)
	if err != nil {
		return err
	}
	lib.Logger.Infof("%s", info)
	return c.SendTransactionAndWaiting([]tpPayload.StoragePayload{{
		Action:   tpPayload.UserCreateData,
		Name:     c.Name,
		DataInfo: info,
	}}, addresses, addresses)
}

func (c *Client) OpenSharedDataToThirdParty(usernameFrom, usernameTo string, accessType int) error {
	err := c.Sync()
	if err != nil {
		return err
	}
	if accessType == 0 || accessType > 2 {
		return nil
	}
	sharedDataList, err := c.ListSharedPatientData(usernameFrom)
	if err != nil {
		return err
	}
	_, userTo, err := c.GetUser(usernameTo)
	if err != nil {
		return err
	}
	batches := make([]tpPayload.StoragePayload, 0)
	for _, sd := range sharedDataList {
		di, data, err := c.GetSharedPatientData(sd.GetHash(), usernameFrom)
		if err != nil {
			return err
		}
		// regular access type
		if accessType == 1 {
			if di.AccessType != 1 {
				continue
			}
			// critical access type
		} else if accessType == 2 {
			if di.AccessType == 0 {
				continue
			}
		}
		now := time.Now()
		expiration := now.Add(1 * time.Minute)
		dataName := fmt.Sprintf("shared_by_%s_%s", c.Name, di.Name)
		keyAES := tpCrypto.GenerateRandomAESKey(lib.AESKeySize)
		info, err := crypto.GenerateDataInfo(dataName, data, userTo.PublicKey, userTo.Name, tpCrypto.BytesToHex(keyAES), di.AccessType, expiration.Unix())
		if err != nil {
			return err
		}
		err = c.User.Root.CreateData(info)
		if err != nil {
			return err
		}
		batches = append(batches, tpPayload.StoragePayload{
			Action:   tpPayload.UserCreateData,
			Name:     c.Name,
			DataInfo: info,
		})
	}

	addresses := []string{c.GetAddress()}
	return c.SendTransactionAndWaiting(batches, addresses, addresses)
}

func (c *Client) OpenSharedDataToTrustedParty(usernameTo string) error {
	err := c.Sync()
	if err != nil {
		return err
	}
	sharedDataList, err := c.ListPatientData()
	if err != nil {
		return err
	}
	_, userTo, err := c.GetUser(usernameTo)
	if err != nil {
		return err
	}
	batches := make([]tpPayload.StoragePayload, 0)
	for _, sd := range sharedDataList {
		di, data, err := c.GetPatientData(sd.GetHash())
		if err != nil {
			return err
		}

		dataName := fmt.Sprintf("shared_by_%s_%s", c.Name, di.Name)
		keyAES := tpCrypto.GenerateRandomAESKey(lib.AESKeySize)
		info, err := crypto.GenerateDataInfo(dataName, data, userTo.PublicKey, userTo.Name, tpCrypto.BytesToHex(keyAES), di.AccessType, 0)
		if err != nil {
			return err
		}
		err = c.User.Root.CreateData(info)
		if err != nil {
			return err
		}
		batches = append(batches, tpPayload.StoragePayload{
			Action:   tpPayload.UserCreateData,
			Name:     c.Name,
			DataInfo: info,
		})
	}

	addresses := []string{c.GetAddress()}
	return c.SendTransactionAndWaiting(batches, addresses, addresses)
}

func (c *Client) RequestData(requestFrom, usernameFrom, accessTypeStr string) error {
	err := c.Sync()
	if err != nil {
		return err
	}
	_, userFrom, err := c.GetUser(usernameFrom)
	if err != nil {
		return err
	}
	accessType, err := strconv.Atoi(accessTypeStr)
	if err != nil {
		return err
	}
	if accessType < 0 || accessType > 2 {
		return errors.New("invalid access type")
	}
	nodeMap := make(map[string]storage.INode)
	for _, n := range userFrom.Root.Repo.INodes {
		if n.GetAddr() != requestFrom {
			continue
		}
		nodeMap[n.GetHash()] = n
	}

	var requests []*models.Request
	for hash, n := range nodeMap {
		oid := primitive.NewObjectID()
		requests = append(requests, &models.Request{
			OID:          &oid,
			Hash:         hash,
			Name:         n.GetName(),
			RequestFrom:  requestFrom,
			UsernameFrom: usernameFrom,
			UsernameTo:   c.Name,
			Status:       models.Unset,
			AccessType:   accessType,
		})
	}
	ctx := context.Background()
	if len(requests) == 0 {
		return nil
	}
	_, err = models.UpsertRequests(ctx, requests)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListRequests() ([]*models.Request, error) {
	ctx := context.Background()
	reqs, err := models.GetRequestsByRequestFrom(ctx, []string{c.Name})
	if err != nil {
		return nil, err
	}
	var filtered []*models.Request
	for _, r := range reqs {
		if r.Status == 1 || r.Status == 2 {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered, nil
}

func (c *Client) ProcessRequest(oidStr string, accept bool) error {
	err := c.Sync()
	if err != nil {
		return err
	}
	oid, err := primitive.ObjectIDFromHex(oidStr)
	if err != nil {
		return err
	}
	ctx := context.Background()
	req, err := models.GetRequestByOID(ctx, &oid)
	if err != nil {
		return err
	}
	if accept {
		req.Status = 1

		if req.UsernameFrom != c.Name {
			err := c.OpenSharedDataToThirdParty(req.UsernameFrom, req.UsernameTo, req.AccessType)
			if err != nil {
				return err
			}
		} else {
			err := c.OpenSharedDataToTrustedParty(req.UsernameTo)
			if err != nil {
				return err
			}
		}

	} else {
		req.Status = 2
	}
	_, err = models.UpsertRequests(ctx, []*models.Request{req})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) BatchUpload(path string) ([]error, error) {
	err := c.Sync()
	if err != nil {
		return nil, err
	}
	records, err := readCsvFile(path)
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, errors.New("no data in csv")
	}

	var trustedPartyColIdx, accessColIdx *int
	for id, colName := range records[0] {
		copiedId := id
		if colName == "trusted_party" {
			trustedPartyColIdx = &copiedId
		}
		if colName == "access_type" {
			accessColIdx = &copiedId
		}
		if trustedPartyColIdx != nil && accessColIdx != nil {
			break
		}
	}
	columnLen := len(records[0])
	var errs []error
	for id, row := range records[1:] {
		if len(row) != columnLen {
			errs = append(errs, errors.New(fmt.Sprintf("csv row %d is invalid: column name's length is not equal to row's column length", id+2)))
			continue
		}
		var accessType int
		if accessColIdx != nil {
			accessType, err = strconv.Atoi(row[*accessColIdx])
			if err != nil {
				errs = append(errs, errors.New(fmt.Sprintf("csv row %d col %d access type is wrong format", id+2, *accessColIdx+1)))
				continue
			}
			if accessType < 0 || accessType > 2 {
				errs = append(errs, errors.New(fmt.Sprintf("csv row %d col %d access type must be eqaul or greater than 0 and lesser than 3", id+2, *accessColIdx+1)))
				continue
			}
		}
		var trustedParties []string
		if trustedPartyColIdx != nil {
			trustedParties = strings.Split(row[*trustedPartyColIdx], " ")
		}
		for idcol, col := range row {
			columnName := records[0][idcol]
			if columnName == "access_type" || columnName == "trusted_party" {
				continue
			}
			if len(col) == 0 {
				errs = append(errs, errors.New(fmt.Sprintf("csv row %d col %d is empty", id+2, idcol+1)))
				continue
			}

			di, err := c.CreatePatientData(columnName, col, uint(accessType))
			if err != nil {
				errs = append(errs, errors.New(fmt.Sprintf("csv row %d col %d: failed to save data: %s", id+2, idcol+1, err)))
				continue
			}
			for _, tp := range trustedParties {
				err = c.ShareData(di.Hash, tp)
				if err != nil {
					errs = append(errs, errors.New(fmt.Sprintf("csv row %d col %d: failed to share with trusted party %s: %s", id+2, idcol+1, tp, err)))
					continue
				}
				fmt.Printf("sharing data with %s \n", tp)
			}
		}
	}

	return errs, nil
}

// GetUser get current user data
func (c *Client) GetUser(username string) (string, *tpUser.User, error) {
	err := c.Sync()
	if err != nil {
		return "", nil, err
	}
	err = c.ListUsers()
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
	err := c.Sync()
	if err != nil {
		return err
	}
	limit := 10000
	users, err := lib.ListUsers(c.lastQueryEnd, uint(limit))
	if err != nil {
		return err
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

func (c *Client) RemovedExpiredData() error {
	ctx := context.Background()
	now := time.Now().Unix()
	datas, err := models.GetDataByExpiration(ctx, now)
	if err != nil {
		return err
	}
	if len(datas) == 0 {
		return nil
	}
	var oids []*primitive.ObjectID
	for _, d := range datas {
		oids = append(oids, d.OID)
	}
	return models.DeleteDatasByOid(oids)
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

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to read input file %s %s", filePath, err))
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to parse file as CSV for %s %s", filePath, err))
	}

	return records, nil
}
