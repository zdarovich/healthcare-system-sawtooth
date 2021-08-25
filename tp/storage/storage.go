package storage

import (
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(&Data{})
	gob.Register(&Repo{})
}

// Root store information of Data and Keys used to encryption.
type Root struct {
	Repo *Repo
	Keys *FileKeyMap
}

// FileInfo is the information of files for usage.
type DataInfo struct {
	Name string
	Size int64
	Hash string
	Key  string
	Addr string
}

// NewRoot is the construct for Root.
func NewRoot(repo *Repo, keyMap *FileKeyMap) *Root {
	return &Root{
		Repo: repo,
		Keys: keyMap,
	}
}

// NewFileInfo is the construct for FileInfo.
func NewDataInfo(name string, size int64, hash string, key string, addr string) *DataInfo {
	return &DataInfo{
		Name: name,
		Size: size,
		Hash: hash,
		Key:  key,
		Addr: addr,
	}
}

// GenerateRoot generate new root for usage.
func GenerateRoot() *Root {
	return NewRoot(NewRepo("home"), NewFileKeyMap())
}

// CreateFile generate file in the path and store its information.
func (root *Root) CreateData(info DataInfo) error {
	fileKeyIndex := root.Keys.AddKey(info.Key, true)
	err := root.Repo.CreateData(info.Name, info.Hash, fileKeyIndex, info.Addr, info.Size)
	if err != nil {
		return err
	}
	return nil
}

func (root *Root) GetData(hash, addr string) (data *DataInfo, err error) {
	f, err := root.Repo.checkDataExists(hash, addr)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	key := root.Keys.GetKey(f.KeyIndex)
	return NewDataInfo(f.Name, f.Size, f.Hash, key.Key, addr), nil
}

// ToBytes convert root to byte slice.
func (root *Root) ToBytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(root)
	return buf.Bytes()
}

// RootFromBytes convert root from byte slice.
func RootFromBytes(data []byte) (*Root, error) {
	root := &Root{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(root)
	return root, err
}
