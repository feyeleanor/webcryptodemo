package main

import (
	"crypto/rsa"
	"io/ioutil"
	"net/http"
	. "time"

	"github.com/julienschmidt/httprouter"
)

type FileServer struct {
	PEM string
	*rsa.PrivateKey
	Started Time
	Address string
	*httprouter.Router
	UserDirectory
	Requests int
}

func (s *FileServer) ListenAndServe() {
	s.Started = Now()
	http.ListenAndServe(s.Address, s.Router)
}

func (s *FileServer) Users() int {
	return len(s.UserDirectory)
}

func (s *FileServer) Files(u ...string) (r int) {
	if len(u) == 0 {
		for _, v := range s.UserDirectory {
			r += len(v.FileStore)
		}
	} else {
		for _, k := range u {
			if f, ok := s.UserDirectory[k]; ok {
				r += f.Files()
			}
		}
	}
	return
}

func (s *FileServer) Now() Time {
	return Now()
}

func (s *FileServer) RequiresAuthorisation(w http.ResponseWriter, id string, f func(user)) {
	s.Requests++
	if u, ok := s.UserDirectory[id]; ok {
		f(u)
	} else {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func NewFileServer(a string) (f *FileServer) {
	if b, e := ioutil.ReadFile("key.pem"); e == nil {
		if k, e := LoadPrivateKey(b); e == nil {
			f = &FileServer{
				PrivateKey:    k,
				PEM:           PublicKeyAsPem(k),
				Address:       a,
				UserDirectory: make(UserDirectory),
				Router:        httprouter.New(),
			}
		} else {
			panic(e)
		}
	}
	return
}
