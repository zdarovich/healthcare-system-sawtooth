package handler

import (
	"fmt"
	"github.com/hyperledger/sawtooth-sdk-go/logging"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/processor_pb2"
	"healthcare-system-sawtooth/tp/payload"
	"healthcare-system-sawtooth/tp/state"
)

var logger = logging.Get()

type Handler struct {
	Name    string
	Version []string
}

func NewHandler(name string, version []string) *Handler {
	return &Handler{Name: name, Version: version}
}

func (h *Handler) FamilyName() string {
	return h.Name
}

func (h *Handler) FamilyVersions() []string {
	return h.Version
}

func (h *Handler) Namespaces() []string {
	return []string{string(state.Namespace)}
}

func (h *Handler) Apply(request *processor_pb2.TpProcessRequest, context *processor.Context) error {
	header := request.GetHeader()
	user := header.GetSignerPublicKey()
	pl, err := payload.StoragePayloadFromBytes(request.GetPayload())
	if err != nil {
		return err
	}
	st := state.NewSeaStorageState(context)

	logger.Debugf("Healthcare txn %v: user %v: payload: Name='%v', Action='%v', Target='%v', DataInfo='%v'", request.Signature, user, pl.Name, pl.Action, pl.DataInfo)

	switch pl.Action {
	// Base Action
	case payload.CreateUser:
		if len(pl.Target) != 1 || pl.Target[0] == "" {
			return &processor.InvalidTransactionError{Msg: "username is nil"}
		}
		return st.CreateUser(pl.Target[0], user)

	case payload.UserCreateData:
		return st.CreateUserData(pl.Name, user, pl.DataInfo)

	default:
		return &processor.InvalidTransactionError{Msg: fmt.Sprint("Invalid Action: ", pl.Action)}
	}
}
