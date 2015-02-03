package main

import (
	"crypto/cipher"
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

func (s *FileServer) SendEncrypted(w http.ResponseWriter, id string, f func(*cipher.StreamWriter, *user)) {
	s.Requests++
	if u, ok := s.UserDirectory[id]; ok {
		if len(u.Key) > 0 {
			if e := EncryptAES(w, u.Key, func(s *cipher.StreamWriter) {
				f(s, u)
			}); e != nil {
				http.Error(w, e.Error(), http.StatusNotAcceptable)
			}
		} else {
			http.Error(w, "Encryption Key Missing", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Not Authorised", http.StatusUnauthorized)
	}
}

func (s *FileServer) ReceiveEncrypted(w http.ResponseWriter, r *http.Request, id string, f func(*cipher.StreamReader, *user)) {
	s.Requests++
	if u, ok := s.UserDirectory[id]; ok {
		if len(u.Key) > 0 {
			if e := DecryptAES(r.Body, u.Key, func(s *cipher.StreamReader) {
				f(s, u)
			}); e != nil {
				http.Error(w, e.Error(), http.StatusNotAcceptable)
			}
		} else {
			http.Error(w, "Encryption Key Missing", http.StatusNotAcceptable)
		}
	} else {
		http.Error(w, "Not Authorised", http.StatusUnauthorized)
	}

	/*	s.RequiresAuthorisation(w, id, func(u *user) {
			if len(u.Key) > 0 {
				if e := DecryptAES(r.Body, u.Key, func(s *cipher.StreamReader) {
					f(s, u)
				}); e != nil {
					http.Error(w, e.Error(), http.StatusNotAcceptable)
				}
			} else {
				http.Error(w, "Encryption Key Missing", http.StatusNotAcceptable)
			}
		})
	*/
}

func (s *FileServer) RequiresAuthorisation(w http.ResponseWriter, id string, f func(*user)) {
	s.Requests++
	if u, ok := s.UserDirectory[id]; ok {
		f(u)
	} else {
		http.Error(w, "Not Authorised", http.StatusUnauthorized)
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
