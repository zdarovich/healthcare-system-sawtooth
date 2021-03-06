package lib

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/messaging"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/batch_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/client_event_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/events_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_receipt_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/validator_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"github.com/pebbe/zmq4"
	"github.com/sirupsen/logrus"
	tpCrypto "healthcare-system-sawtooth/crypto"
	tpPayload "healthcare-system-sawtooth/tp/payload"
	tpState "healthcare-system-sawtooth/tp/state"
)

// The Category of ClientFramework.
const (
	ClientCategoryUser = true
)

// ClientFramework provides SeaStorage base operations for both user and sea.
type ClientFramework struct {
	Name       string // The name of user.
	Category   bool   // The category of client framework.
	PrivKeyHex []byte
	signer     *signing.Signer
	zmqConn    *messaging.ZmqConnection
	corrID     string
	waiting    bool
	signal     chan bool
	State      chan []byte
}

// NewClientFramework is the construct for ClientFramework.
func NewClientFramework(name string, category bool, keyFile string) (*ClientFramework, error) {
	if name == "" {
		return nil, errors.New("need a valid name")
	}
	if keyFile == "" {
		return nil, errors.New("need a valid key")
	}
	// Read private key file
	privateKeyHex, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %v", err)
	}
	// Get private key object
	privateKey := signing.NewSecp256k1PrivateKey(tpCrypto.HexToBytes(string(privateKeyHex)))
	cryptoFactory := signing.NewCryptoFactory(signing.NewSecp256k1Context())
	signer := cryptoFactory.NewSigner(privateKey)
	cf := &ClientFramework{
		Name:       name,
		Category:   category,
		signer:     signer,
		PrivKeyHex: privateKeyHex,
		signal:     make(chan bool),
		State:      make(chan []byte),
	}
	err = cf.generateZmqConnection()
	if err != nil {
		return nil, err
	}
	err = cf.WatchingForState()
	if err != nil {
		return nil, err
	}
	go cf.subscribeHandler()
	return cf, nil
}

// Close is the deconstruct for ClientFramework.
func (cf *ClientFramework) Close() {
	err := cf.unsubscribeEvents(cf.corrID)
	if err != nil {
		Logger.WithFields(logrus.Fields{
			"correlationID": cf.corrID,
		}).Errorf("failed to unsubscribe events: %v", err)
	}
	close(cf.signal)
	cf.zmqConn.Close()
}

// Register user. Create user in the blockchain.
func (cf *ClientFramework) Register(name string) error {
	var seaStoragePayload tpPayload.StoragePayload
	seaStoragePayload.Action = tpPayload.CreateUser
	seaStoragePayload.Target = []string{name}
	cf.Name = name
	return cf.SendTransactionAndWaiting([]tpPayload.StoragePayload{seaStoragePayload}, []string{cf.GetAddress()}, []string{cf.GetAddress()})
}

// GetData returns the data of user.
func (cf *ClientFramework) GetData() ([]byte, error) {
	return GetStateData(cf.GetAddress())
}

// GetAddress returns the address of user.
func (cf *ClientFramework) GetAddress() string {
	return tpState.MakeAddress(tpState.AddressTypeUser, cf.Name, cf.signer.GetPublicKey().AsHex())
}

// GetPublicKey returns the public key of user.
func (cf *ClientFramework) GetPublicKey() string {
	return cf.signer.GetPublicKey().AsHex()
}

// Whoami display the information of user.
func (cf *ClientFramework) Whoami() {
	fmt.Println("User name: " + cf.Name)
	fmt.Println("Public key: " + cf.signer.GetPublicKey().AsHex())
	fmt.Println("Sawtooth address: " + cf.GetAddress())
}

// DecryptDataKey returns the key decrypted by user's private key.
// If the error is not nil, it will return.
func (cf *ClientFramework) DecryptDataKey(key string) ([]byte, error) {
	return tpCrypto.Decryption(string(cf.PrivKeyHex), key)
}

// DecryptDataKey returns the key encrypted by user's public key.
func (cf *ClientFramework) EncryptDataKey(publicKey, key string) ([]byte, error) {
	return tpCrypto.Encryption(publicKey, key)
}

// GetStatus returns the status of batch.
func (cf *ClientFramework) getStatus(batchID string, wait int64) (map[string]interface{}, error) {
	// API to call
	apiSuffix := fmt.Sprintf("%s?id=%s&wait=%d", BatchStatusAPI, batchID, wait)
	response, err := sendRequestByAPISuffix(apiSuffix, []byte{}, "")
	if err != nil {
		return nil, err
	}

	entry := response["data"].([]interface{})[0].(map[string]interface{})
	return entry, nil
}

// SendTransaction send transactions by the batch.
func (cf *ClientFramework) SendTransaction(storagePayloads []tpPayload.StoragePayload, inputs, outputs []string) (map[string]interface{}, error) {
	var transactions []*transaction_pb2.Transaction

	for _, storagePayload := range storagePayloads {
		// Construct TransactionHeader
		rawTransactionHeader := transaction_pb2.TransactionHeader{
			SignerPublicKey:  cf.signer.GetPublicKey().AsHex(),
			FamilyName:       FamilyName,
			FamilyVersion:    FamilyVersion,
			Dependencies:     []string{},
			Nonce:            strconv.Itoa(rand.Int()),
			BatcherPublicKey: cf.signer.GetPublicKey().AsHex(),
			Inputs:           inputs,
			Outputs:          outputs,
			PayloadSha512:    tpCrypto.SHA512HexFromBytes(storagePayload.ToBytes()),
		}
		transactionHeader, err := proto.Marshal(&rawTransactionHeader)
		if err != nil {
			return nil, fmt.Errorf("unable to serialize transaction header: %v", err)
		}

		// Signature of TransactionHeader
		transactionHeaderSignature := hex.EncodeToString(cf.signer.Sign(transactionHeader))

		// Construct Transaction
		transaction := &transaction_pb2.Transaction{
			Header:          transactionHeader,
			HeaderSignature: transactionHeaderSignature,
			Payload:         storagePayload.ToBytes(),
		}

		transactions = append(transactions, transaction)
	}

	// Get BatchList
	rawBatchList, err := cf.createBatchList(transactions)
	if err != nil {
		return nil, fmt.Errorf("unable to construct batch list: %v", err)
	}
	batchList, err := proto.Marshal(&rawBatchList)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize batch list: %v", err)
	}

	return sendRequestByAPISuffix(BatchSubmitAPI, batchList, ContentTypeOctetStream)
}

// SendTransactionAndWaiting send transaction by the batch and waiting for the batches committed.
func (cf *ClientFramework) SendTransactionAndWaiting(seaStoragePayloads []tpPayload.StoragePayload, inputs, outputs []string) error {
	response, err := cf.SendTransaction(seaStoragePayloads, inputs, outputs)
	if err != nil {
		return err
	}
	for k, v := range response {
		Logger.Debugf("%s: %s \n", k, v)
	}
	return cf.WaitingForCommitted()
}

// create the list of batches.
func (cf *ClientFramework) createBatchList(transactions []*transaction_pb2.Transaction) (batch_pb2.BatchList, error) {
	// Get list of TransactionHeader signatures
	var transactionSignatures []string
	for _, transaction := range transactions {
		transactionSignatures = append(transactionSignatures, transaction.HeaderSignature)
	}

	// Construct BatchHeader
	rawBatchHeader := batch_pb2.BatchHeader{
		SignerPublicKey: cf.signer.GetPublicKey().AsHex(),
		TransactionIds:  transactionSignatures,
	}
	batchHeader, err := proto.Marshal(&rawBatchHeader)
	if err != nil {
		return batch_pb2.BatchList{}, fmt.Errorf("unable to serialize batch header: %v", err)
	}

	// Signature of BatchHeader
	batchHeaderSignature := hex.EncodeToString(cf.signer.Sign(batchHeader))

	// Construct Batch
	batch := batch_pb2.Batch{
		Header:          batchHeader,
		Transactions:    transactions,
		HeaderSignature: batchHeaderSignature,
	}

	// Construct BatchList
	return batch_pb2.BatchList{
		Batches: []*batch_pb2.Batch{&batch},
	}, nil
}

// WaitingForCommitted wait for batches committed.
// If timeout or batches invalid, it will return error.
func (cf *ClientFramework) WaitingForCommitted() error {
	cf.waiting = true
	defer func() { cf.waiting = false }()
	select {
	case <-cf.signal:
		return nil
	case <-time.After(DefaultWait):
		return errors.New("waiting for committed timeout")
	}
}

// WatchingForState waits for change of the state in the blockchain
func (cf *ClientFramework) WatchingForState() error {
	subscription := &events_pb2.EventSubscription{
		EventType: "sawtooth/state-delta",
		Filters: []*events_pb2.EventFilter{{
			Key:         "address",
			MatchString: cf.GetAddress(),
			FilterType:  events_pb2.EventFilter_SIMPLE_ANY,
		}},
	}
	corrID, err := cf.subscribeEvents([]*events_pb2.EventSubscription{subscription})
	if err != nil {
		return err
	}
	cf.corrID = corrID
	return nil
}

// Subscribe to any state change events in the blockchain
func (cf *ClientFramework) subscribeEvents(subscriptions []*events_pb2.EventSubscription) (string, error) {
	// Construct the subscribeRequest
	subscribeRequest := &client_event_pb2.ClientEventsSubscribeRequest{
		Subscriptions: subscriptions,
	}
	requestBytes, err := proto.Marshal(subscribeRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal subscription subscribeRequest: %v", err)
	}
	corrID, err := cf.zmqConn.SendNewMsg(validator_pb2.Message_CLIENT_EVENTS_SUBSCRIBE_REQUEST, requestBytes)
	if err != nil {
		return "", fmt.Errorf("failed to send subscription message: %v", err)
	}
	// Received subscription response
	_, response, err := cf.zmqConn.RecvMsgWithId(corrID)
	if err != nil {
		return "", fmt.Errorf("failed to received subscribe event response: %v", err)
	}
	subscribeResponse := &client_event_pb2.ClientEventsSubscribeResponse{}
	err = proto.Unmarshal(response.Content, subscribeResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal subscribe response: %v", err)
	}
	if subscribeResponse.Status != client_event_pb2.ClientEventsSubscribeResponse_OK {
		return "", errors.New("failed to subscribe event")
	}
	return corrID, nil
}

// Unsubscribe from any state change events in the blockchain
func (cf *ClientFramework) unsubscribeEvents(corrID string) error {
	// Construct the UnsubscribeRequest
	unsubscribeRequest := &client_event_pb2.ClientEventsUnsubscribeRequest{}
	unsubscribeRequestBytes, err := proto.Marshal(unsubscribeRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscribe event: %v", err)
	}
	id, err := cf.zmqConn.SendNewMsg(validator_pb2.Message_CLIENT_EVENTS_UNSUBSCRIBE_REQUEST, unsubscribeRequestBytes)
	if err != nil {
		return fmt.Errorf("faield to send unsubscribe event message: %v", err)
	}
	// Received the unsubscription response
	_, response, err := cf.zmqConn.RecvMsgWithId(id)
	if err != nil {
		return fmt.Errorf("failed to received unsubcribe event response: %v", err)
	}
	unsubscribeResponse := &client_event_pb2.ClientEventsUnsubscribeResponse{}
	err = proto.Unmarshal(response.Content, unsubscribeResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal unsubscribe event response: %v", err)
	}
	if unsubscribeResponse.Status != client_event_pb2.ClientEventsUnsubscribeResponse_OK {
		return errors.New("failed to unsubscribe event")
	}
	return nil
}

// Subscribe handler
func (cf *ClientFramework) subscribeHandler() {
	for {
		_, message, err := cf.zmqConn.RecvMsg()
		if err != nil {
			Logger.Errorf("zmq failed to received message: %v", err)
			continue
		}
		if message.MessageType != validator_pb2.Message_CLIENT_EVENTS {
			continue
		}
		eventList := &events_pb2.EventList{}
		err = proto.Unmarshal(message.Content, eventList)
		if err != nil {
			Logger.WithFields(logrus.Fields{
				"message": message.String(),
			}).Error("failed unmarshal message")
			continue
		}
		for _, event := range eventList.Events {
			for _, attr := range event.Attributes {
				if attr.Key == "address" && attr.Value == cf.GetAddress() {
					if cf.waiting {
						cf.signal <- true
					}
					stateChangeList := &txn_receipt_pb2.StateChangeList{}
					err := proto.Unmarshal(event.Data, stateChangeList)
					if err != nil {
						Logger.Errorf("failed to unmarshal protobuf: %v", err)
						continue
					}

					for _, stateChange := range stateChangeList.StateChanges {
						if stateChange.Address == cf.GetAddress() {
							cf.State <- stateChange.Value
							break
						}
					}
					break
				}
			}
		}
	}
}

// Zmq connections for event subscription
func (cf *ClientFramework) generateZmqConnection() error {
	// Setup a connection to the validator
	ctx, err := zmq4.NewContext()
	if err != nil {
		return err
	}
	zmqConn, err := messaging.NewConnection(ctx, zmq4.DEALER, ValidatorURL, false)
	if err != nil {
		return err
	}
	cf.zmqConn = zmqConn
	return nil
}

// GenerateKey generate key pair (Secp256k1) and store them in the client path.
func GenerateKey(keyName string, keyPath string) {
	cont := signing.NewSecp256k1Context()
	pri := cont.NewRandomPrivateKey()
	pub := cont.GetPublicKey(pri)
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		err = os.MkdirAll(keyPath, 0755)
		if err != nil {
			panic(err)
		}
	}
	err := ioutil.WriteFile(path.Join(keyPath, keyName+".priv"), []byte(pri.AsHex()), 0600)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path.Join(keyPath, keyName+".pub"), []byte(pub.AsHex()), 0600)
	if err != nil {
		panic(err)
	}
}
