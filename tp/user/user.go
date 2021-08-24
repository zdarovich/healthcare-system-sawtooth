package user

import (
	"bytes"
	"encoding/gob"
	"healthcare-system-sawtooth/tp/storage"
)

type User struct {
	Name      string
	PublicKey string
	Groups    []string
	Root      *storage.Root
}

func NewUser(username, publicKey string, groups []string, root *storage.Root) *User {
	return &User{
		Name:      username,
		PublicKey: publicKey,
		Groups:    groups,
		Root:      root,
	}
}

func GenerateUser(username, publicKey string) *User {
	return NewUser(username, publicKey, make([]string, 0), storage.GenerateRoot())
}

func (u *User) VerifyPublicKey(publicKey string) bool {
	return publicKey == u.PublicKey
}

func (u *User) JoinGroup(group string) bool {
	for _, g := range u.Groups {
		if g == group {
			return false
		}
	}
	u.Groups = append(u.Groups, group)
	return true
}

func (u *User) LeaveGroup(group string) bool {
	for i, g := range u.Groups {
		if g == group {
			u.Groups = append(u.Groups[:i], u.Groups[i+1:]...)
			return true
		}
	}
	return false
}

func (u *User) IsInGroup(group string) bool {
	for _, g := range u.Groups {
		if g == group {
			return true
		}
	}
	return false
}

func (u *User) ToBytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(u)
	return buf.Bytes()
}

func UserFromBytes(data []byte) (*User, error) {
	buf := bytes.NewBuffer(data)
	u := &User{}
	dec := gob.NewDecoder(buf)
	err := dec.Decode(u)
	return u, err
}
