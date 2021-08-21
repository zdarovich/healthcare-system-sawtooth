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
	pl, err := payload.SeaStoragePayloadFromBytes(request.GetPayload())
	if err != nil {
		return err
	}
	st := state.NewSeaStorageState(context)

	logger.Debugf("SeaStorage txn %v: user %v: payload: Name='%v', Action='%v', Target='%v'", request.Signature, user, pl.Name, pl.Action, pl.Target)

	switch pl.Action {
	// Base Action
	case payload.CreateUser:
		if len(pl.Target) != 1 || pl.Target[0] == "" {
			return &processor.InvalidTransactionError{Msg: "username is nil"}
		}
		return st.CreateUser(pl.Target[0], user)
	case payload.CreateGroup:
		if len(pl.Target) != 1 || pl.Target[0] == "" {
			return &processor.InvalidTransactionError{Msg: "group name is nil"}
		}
		return st.CreateGroup(pl.Target[0], state.MakeAddress(state.AddressTypeUser, pl.Name, user), pl.Key)
	case payload.CreateSea:
		if len(pl.Target) != 1 || pl.Target[0] == "" {
			return &processor.InvalidTransactionError{Msg: "sea name is nil"}
		}
		return st.CreateSea(pl.Target[0], user)

	case payload.UserCreateFile:
		return st.UserCreateFile(pl.Name, user, pl.PWD, pl.FileInfo)
	case payload.UserUpdateFileData:
		return st.UserUpdateFileData(pl.Name, user, pl.PWD, pl.FileInfo)
	case payload.UserUpdateFileKey:
		return st.UserUpdateFileKey(pl.Name, user, pl.PWD, pl.FileInfo)
	case payload.UserPublishKey:
		if len(pl.Target) != 1 || pl.Target[0] == "" {
			return &processor.InvalidTransactionError{Msg: "the index of key is nil"}
		}
		return st.UserPublishKey(pl.Name, user, pl.Target[0], pl.Key)
	case payload.UserShare:
		if len(pl.Target) != 2 || pl.Target[0] == "" || pl.Target[1] == "" {
			return &processor.InvalidTransactionError{Msg: "the name of file or directory is nil"}
		}
		return st.UserShareFiles(pl.Name, user, pl.PWD, pl.Target[0], pl.Target[1])
	case payload.SeaStoreFile:
		return st.SeaStoreFile(pl.Name, user, pl.UserOperations)
	case payload.SeaConfirmOperations:
		return st.SeaConfirmOperations(pl.Name, user, pl.SeaOperations)

	default:
		return &processor.InvalidTransactionError{Msg: fmt.Sprint("Invalid Action: ", pl.Action)}
	}
}
