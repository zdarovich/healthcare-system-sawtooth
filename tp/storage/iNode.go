package storage

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"sync"
)

type INode interface {
	GetName() string
	GetSize() int64
	GetHash() string
	GetAddr() string
	GetKeys() []string
	ToBytes() []byte
	ToJson() string
	lock()
	unlock()
}

type Data struct {
	mutex    sync.Mutex
	Name     string
	Hash     string
	Size     int64
	KeyIndex string
	Addr     string
}

type Repo struct {
	mutex  sync.Mutex
	Name   string
	Hash   string
	Addr   string
	Size   int64
	INodes []INode
}

type INodeInfo struct {
	Hash string
	Size int64
}

func NewRepo(name string) *Repo {
	return &Repo{Name: name, Size: 0, INodes: make([]INode, 0), Hash: ""}
}

func (d *Repo) lock() {
	d.mutex.Lock()
}

func (d *Repo) unlock() {
	d.mutex.Unlock()
}

func (d *Repo) GetName() string {
	return d.Name
}

func (d *Repo) GetSize() int64 {
	return d.Size
}

func (d *Repo) GetAddr() string {
	return d.Addr
}

func (d *Repo) GetKeys() []string {
	keyIndexes := make([]string, 0)
	for _, iNode := range d.INodes {
		keyIndexes = append(keyIndexes, iNode.GetKeys()...)
	}
	return keyIndexes
}

func (d *Repo) ToBytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(d)
	return buf.Bytes()
}

func (d *Repo) ToJson() string {
	data, _ := json.MarshalIndent(d, "", "\t")
	return string(data)
}

func NewData(name string) *Data {
	return &Data{Name: name, Size: 0, Hash: ""}
}

func (d *Data) lock() {
	d.mutex.Lock()
}

func (d *Data) unlock() {
	d.mutex.Unlock()
}

func (d *Data) GetName() string {
	return d.Name
}

func (d *Data) GetSize() int64 {
	return d.Size
}

func (d *Data) GetHash() string {
	return d.Hash
}

func (d *Data) GetAddr() string {
	return d.Addr
}

func (d *Data) GetKeys() []string {
	return []string{d.KeyIndex}
}

func (d *Data) ToBytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(d)
	return buf.Bytes()
}

func (d *Data) ToJson() string {
	data, _ := json.MarshalIndent(d, "", "\t")
	return string(data)
}

func generateINodeInfos(iNodes []INode) []INodeInfo {
	var infos = make([]INodeInfo, len(iNodes))
	for i := 0; i < len(iNodes); i++ {
		infos[i].Hash = iNodes[i].GetHash()
		infos[i].Size = iNodes[i].GetSize()
	}
	return infos
}

func (r *Repo) CreateData(name, hash, keyIndex, addr string, size int64) error {

	for j := 0; j < len(r.INodes); j++ {
		if r.INodes[j].GetHash() == hash && r.INodes[j].GetAddr() == addr {
			return errors.New("data already exists")
		}
	}
	r.lock()
	defer r.unlock()
	data := NewData(name)
	data.Hash = hash
	data.Size = size
	data.KeyIndex = keyIndex
	data.Addr = addr
	r.INodes = append(r.INodes, data)
	return nil
}
func (d *Repo) checkDataExists(hash, addr string) (*Data, error) {

	for _, iNode := range d.INodes {
		switch iNode.(type) {
		case *Data:
			if iNode.GetHash() == hash && iNode.GetAddr() == addr {
				return iNode.(*Data), nil
			}
		}
	}
	return nil, nil
}

func RepoFromBytes(data []byte) (*Repo, error) {
	d := &Repo{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(d)
	return d, err
}

func DataFromBytes(data []byte) (*Data, error) {
	f := &Data{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(f)
	return f, err
}
