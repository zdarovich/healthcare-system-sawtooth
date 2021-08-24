// Copyright © 2019 yellowsea <hh1271941291@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(&Data{})
	gob.Register(&Repo{})
}

// Root store information of files and Keys used to encryption.
// Store the information of private files in 'Home' directory.
// Store the information of shared files in 'Shared' directory.
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

// PublishKey publish the key encrypted by public key.
func (root *Root) PublishKey(publicKey, keyIndex, key string) error {
	return root.Keys.PublishKey(publicKey, keyIndex, key)
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

func (root *Root) List(p string) (iNodes []INodeInfo, err error) {

	return root.Repo.List(p)
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
