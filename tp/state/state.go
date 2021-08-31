package state

import (
	"bytes"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"healthcare-system-sawtooth/crypto"
	"healthcare-system-sawtooth/tp/storage"
	"healthcare-system-sawtooth/tp/user"
)

type AddressType uint8

var (
	// User namespace address type
	AddressTypeUser  AddressType = 0
	AddressTypeGroup AddressType = 1
)

var (
	Namespace      = crypto.SHA512HexFromBytes([]byte("Healthcare"))[:6]
	UserNamespace  = crypto.SHA256HexFromBytes([]byte("User"))[:4]
	GroupNamespace = crypto.SHA256HexFromBytes([]byte("Group"))[:4]
)

// Storage state struct
type StorageState struct {
	context    *processor.Context
	userCache  map[string][]byte
	groupCache map[string][]byte
	seaCache   map[string][]byte
}

// Creates new storage state struct
func NewSeaStorageState(context *processor.Context) *StorageState {
	return &StorageState{
		context:    context,
		userCache:  make(map[string][]byte),
		groupCache: make(map[string][]byte),
		seaCache:   make(map[string][]byte),
	}
}

// Gets user data stored on the blockchain
func (sss *StorageState) GetUser(address string) (*user.User, error) {
	userBytes, ok := sss.userCache[address]
	if ok {
		return user.UserFromBytes(userBytes)
	}
	results, err := sss.context.GetState([]string{address})
	if err != nil {
		return nil, err
	}
	if len(results[address]) > 0 {
		sss.userCache[address] = results[address]
		return user.UserFromBytes(results[address])
	}
	return nil, &processor.InvalidTransactionError{Msg: "user doesn't exists"}
}

// Creates new user data
func (sss *StorageState) CreateUser(username string, publicKey string) error {
	address := MakeAddress(AddressTypeUser, username, publicKey)
	_, ok := sss.userCache[address]
	if ok {
		return &processor.InvalidTransactionError{Msg: "user exists"}
	}
	results, err := sss.context.GetState([]string{address})
	if err != nil {
		return err
	}
	if len(results[address]) > 0 {
		return &processor.InvalidTransactionError{Msg: "user exists"}
	}
	return sss.saveUser(user.GenerateUser(username, publicKey), address)
}

func (sss *StorageState) saveUser(u *user.User, address string) error {
	uBytes := u.ToBytes()
	addresses, err := sss.context.SetState(map[string][]byte{
		address: uBytes,
	})
	if err != nil {
		return err
	}
	if len(addresses) == 0 {
		return &processor.InternalError{Msg: "No addresses in set response"}
	}
	sss.userCache[address] = uBytes
	return nil
}

func (sss *StorageState) GetGroup(address string) (*user.Group, error) {
	groupBytes, ok := sss.groupCache[address]
	if ok {
		return user.GroupFromBytes(groupBytes)
	}
	results, err := sss.context.GetState([]string{address})
	if err != nil {
		return nil, err
	}
	if len(results[address]) > 0 {
		sss.seaCache[address] = results[address]
		return user.GroupFromBytes(results[address])
	}
	return nil, &processor.InvalidTransactionError{Msg: "group doesn't exists"}
}

func (sss *StorageState) CreateGroup(groupName, leader, key string) error {
	address := MakeAddress(AddressTypeGroup, groupName, "")
	_, ok := sss.groupCache[address]
	if ok {
		return &processor.InvalidTransactionError{Msg: "group exists"}
	}
	results, err := sss.context.GetState([]string{address})
	if err != nil {
		return err
	}
	if len(results[address]) > 0 {
		return &processor.InvalidTransactionError{Msg: "group exists"}
	}
	return sss.saveGroup(user.GenerateGroup(groupName, leader), address)
}

func (sss *StorageState) saveGroup(g *user.Group, address string) error {
	gBytes := g.ToBytes()
	addresses, err := sss.context.SetState(map[string][]byte{
		address: gBytes,
	})
	if err != nil {
		return err
	}
	if len(addresses) > 0 {
		return &processor.InternalError{Msg: "No addresses in set response"}
	}
	sss.groupCache[address] = gBytes
	return nil
}

func (sss *StorageState) CreateUserData(username, publicKey string, info storage.DataInfo) error {
	address := MakeAddress(AddressTypeUser, username, publicKey)
	u, err := sss.GetUser(address)
	if err != nil {
		return err
	}
	err = u.Root.CreateData(info)
	if err != nil {
		return &processor.InvalidTransactionError{Msg: err.Error()}
	}
	return sss.saveUser(u, address)
}

func MakeAddress(addressType AddressType, name, publicKey string) string {
	switch addressType {
	case AddressTypeUser:
		return Namespace + UserNamespace + crypto.SHA512HexFromBytes(bytes.Join([][]byte{[]byte(name), crypto.HexToBytes(publicKey)}, []byte{}))[:60]
	case AddressTypeGroup:
		return Namespace + GroupNamespace + crypto.SHA512HexFromBytes([]byte(name))[:60]
	default:
		return ""
	}
}
