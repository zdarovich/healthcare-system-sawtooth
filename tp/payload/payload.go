package payload

import (
	"bytes"
	"encoding/gob"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"healthcare-system-sawtooth/tp/storage"
)

const _ = proto.ProtoPackageIsVersion3

// Common action
var (
	Unset      uint = 0
	CreateUser uint = 1
)

// User action
var (
	UserCreateData uint = 10
)

// Payload data model received by the transaction processor
type StoragePayload struct {
	Action   uint             `default:"Unset(0)"`
	Name     string           `default:""`
	Target   []string         `default:"nil"`
	Key      string           `default:""`
	DataInfo storage.DataInfo `default:"DataInfo{}"`
}

// Creates new payload data model
func StoragePayloadFromBytes(payloadData []byte) (*StoragePayload, error) {
	if payloadData == nil {
		return nil, &processor.InvalidTransactionError{Msg: "Must contain payload"}
	}
	pl := &StoragePayload{}
	buf := bytes.NewBuffer(payloadData)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(pl)
	return pl, err
}

// Converts new payload data model to bytes
func (ssp *StoragePayload) ToBytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(ssp)
	return buf.Bytes()
}
