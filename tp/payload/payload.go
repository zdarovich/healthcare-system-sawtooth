package payload

import (
	"bytes"
	"encoding/gob"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"healthcare-system-sawtooth/tp/sea"
	"healthcare-system-sawtooth/tp/storage"
	"healthcare-system-sawtooth/tp/user"
)

const _ = proto.ProtoPackageIsVersion3

// Common action
var (
	Unset      uint = 0
	CreateUser uint = 1
	CreateSea  uint = 2
)

// User action
var (
	UserCreateData uint = 10
)

type SeaStoragePayload struct {
	Action         uint             `default:"Unset(0)"`
	Name           string           `default:""`
	PWD            string           `default:"/"`
	Target         []string         `default:"nil"`
	Key            string           `default:""`
	DataInfo       storage.DataInfo `default:"DataInfo{}"`
	UserOperations []user.Operation `default:"nil"`
}

func NewSeaStoragePayload(action uint, name string, PWD string, target []string, key string, fileInfo storage.DataInfo, userOperations []user.Operation, seaOperations []sea.Operation) *SeaStoragePayload {
	return &SeaStoragePayload{
		Action:         action,
		Name:           name,
		PWD:            PWD,
		Target:         target,
		Key:            key,
		DataInfo:       fileInfo,
		UserOperations: userOperations,
	}
}

func SeaStoragePayloadFromBytes(payloadData []byte) (*SeaStoragePayload, error) {
	if payloadData == nil {
		return nil, &processor.InvalidTransactionError{Msg: "Must contain payload"}
	}
	pl := &SeaStoragePayload{}
	buf := bytes.NewBuffer(payloadData)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(pl)
	return pl, err
}

func (ssp *SeaStoragePayload) ToBytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(ssp)
	return buf.Bytes()
}
