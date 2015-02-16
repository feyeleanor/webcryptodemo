package main

import (
	"crypto/rand"
	"encoding/base32"
	. "fmt"
	. "time"
)

type FileStore map[string]string

type user struct {
	*AESKey
	ID         string
	Registered Time
	FileStore
}

func (u *user) PrintableKey() (r string) {
	r = base32.StdEncoding.EncodeToString([]byte(u.Key))
	return
}

func (u *user) Files() (r int) {
	return len(u.FileStore)
}

type UserDirectory map[string]*user

func (u *UserDirectory) NewUserToken() string {
	b := make([]byte, 30)
	if _, e := rand.Read(b); e != nil {
		panic(Sprintf("rand.Read failed: %v", e))
	}
	return base32.StdEncoding.EncodeToString(b)
}

func (u *UserDirectory) CreateUser(k []byte) (t string) {
	t = u.NewUserToken()
	for _, ok := server.UserDirectory[t]; ok; _, ok = server.UserDirectory[t] {
		t = u.NewUserToken()
	}
	(*u)[t] = &user{AESKey: NewAESKey(k), ID: t, Registered: Now(), FileStore: make(FileStore)}
	return
}
