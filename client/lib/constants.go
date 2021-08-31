package lib

import (
	"crypto/aes"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// TPURL is the Hyperledger Sawtooth rest api url.
	TPURL string
	// ValidatorURL is the Hyperledger Sawtooth validator tcp url.
	ValidatorURL string
)

const (
	// Config Variable

	// FamilyName is the healthcare-system transaction identity.
	FamilyName string = "healthcare-system"
	// FamilyVersion is the version of SeaStorage's transaction.
	FamilyVersion string = "1.0"
	// DefaultWait is the waiting time for batch commits.
	DefaultWait = time.Minute
	// DefaultQueryLimit is the limit of state queries.
	DefaultQueryLimit uint = 20
	// DefaultConfigFilename is the config filename.
	DefaultConfigFilename string = "config"
	// PackageSize is the limit of each package's max size.
	PackageSize int64 = 128 * 1024 * 1024
	// ContentTypeOctetStream is the content type for request.
	ContentTypeOctetStream string = "application/octet-stream"
	// Mongo connection url
	MongoDbUrl string = "mongodb://localhost:27017"
	// Mongo db name
	MongoDbName string = "healthcare"

	// APIs

	// BatchSubmitAPI is the api for batch submission.
	BatchSubmitAPI string = "batches"
	// BatchStatusAPI is the api for getting batches' status.
	BatchStatusAPI string = "batch_statuses"
	// StateAPI is the api for getting data stored in the blockchain.
	StateAPI string = "state"

	// AES-CTR

	// AESKeySize is the size of AES key.
	AESKeySize int = 256
	// IvSize is the AES-CTR iv's size.
	IvSize = aes.BlockSize
)

var (
	// Logger provides log function.
	Logger *logrus.Logger
	// DefaultTPURL is the default Hyperledger Sawtooth rest api url.
	DefaultTPURL = "http://localhost:8008"
	// DefaultValidatorURL is the default Hyperledger Sawtooth validator tcp url.
	DefaultValidatorURL = "tcp://localhost:4004"
	// PrivateKeyFile is the path of private key.
	PrivateKeyFile string
	// DefaultKeyPath is the default path for key storing.
	DefaultKeyPath = "resources/keys"
	// DefaultPrivateKeyFile is the default path of private key.
	DefaultPrivateKeyFile string
	// DefaultConfigPath is the default path for config storing.
	DefaultConfigPath string
	// DefaultLogPath is the default path for log storing.
	DefaultLogPath string
)
