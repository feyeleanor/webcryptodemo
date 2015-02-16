package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/rsa"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	KEY  = "key"
	USER = "user"
	FILE = "file"
)

func NewCipher() (s *CFBStream) {
	return NewCFBStream(256)
}

type ServiceConnection struct {
	*rsa.PublicKey
	*CFBStream
	id      string
	address string
}

func (s *ServiceConnection) Connect(a string) {
	s.address = "http://" + a
	if b, e := s.Do("GET", KEY); e == nil {
		if k, e := LoadPublicKey(string(b)); e == nil {
			s.PublicKey = k.(*rsa.PublicKey)
		} else {
			panic(e)
		}
	}
}

func (s *ServiceConnection) CreateUser(a ...string) (r string) {
	if len(a) > 0 {
		s.Connect(a[0])
	}
	s.CFBStream = NewCipher()
	if m, e := EncryptRSA(Service.PublicKey, []byte(s.CFBStream.Key), []byte("REGISTER")); e == nil {
		if v, e := Service.Do("POST", USER, string(m)); e == nil {
			r = responseBody(v, e, s.CFBStream)
		}
	}
	s.id = r
	return
}

func (s *ServiceConnection) Status() string {
	return responseBody(Service.Do("GET"))
}

func (s *ServiceConnection) UserStatus() string {
	r, e := Service.Do("GET", USER, s.id)
	return responseBody(r, e, s.CFBStream)
}

func (s *ServiceConnection) StoreKey(k string) {
	nk := []byte(k)
	if m, e := EncryptRSA(Service.PublicKey, nk, []byte(s.id)); e == nil {
		responseBody(Service.Do("POST", KEY, s.id, string(m)))
		s.Key = nk
	}
	return
}

func (s *ServiceConnection) ListFiles() string {
	return responseBody(Service.Do("GET", FILE, s.id))
}

func (s *ServiceConnection) StoreFile(tag, content string) string {
	return responseBody(s.DoEncrypted("POST", FILE, s.id, tag, content))
}

func (s *ServiceConnection) RetrieveFile(tag string) (f io.Reader) {
	if b, e := s.DoEncrypted("GET", FILE, s.id, tag); e == nil {
		f = bytes.NewBuffer(b)
	}
	return
}

func (s *ServiceConnection) Do(m string, p ...string) (b []byte, e error) {
	do(s.Request(m, p...), func(res *http.Response) {
		b, e = ioutil.ReadAll(res.Body)
	})
	return
}

func (s *ServiceConnection) DoEncrypted(m string, p ...string) (b []byte, e error) {
	do(s.EncryptedRequest(m, p...), func(res *http.Response) {
		s.Read(res.Body, func(sr *cipher.StreamReader) {
			b, e = ioutil.ReadAll(sr)
		})
	})
	return
}

func responseBody(b []byte, e error, k ...*CFBStream) (v string) {
	if e == nil && len(k) > 0 {
		e = k[0].DecryptCFB(bytes.NewBuffer(b), func(s *cipher.StreamReader) {
			b, e = ioutil.ReadAll(s)
		})
	}
	if e == nil {
		v = string(b)
	}
	return
}

func (s *ServiceConnection) Request(m string, p ...string) (req *http.Request) {
	switch m {
	case "POST":
		switch l := len(p); l {
		case 0:
			req, _ = http.NewRequest("POST", Resource(s.address), new(bytes.Buffer))
		case 1:
			req, _ = http.NewRequest("POST", Resource(s.address), bytes.NewBufferString(p[0]))
		case 2:
			req, _ = http.NewRequest("POST", Resource(s.address, p[0]), bytes.NewBufferString(p[1]))
		default:
			req, _ = http.NewRequest("POST", Resource(s.address, p[:l-1]...), bytes.NewBufferString(p[l-1]))
		}
	default:
		req, _ = http.NewRequest(m, Resource(s.address, p...), nil)
	}
	return
}

func (s *ServiceConnection) EncryptedRequest(m string, p ...string) (req *http.Request) {
	switch m {
	case "POST":
		b := new(bytes.Buffer)
		switch l := len(p); l {
		case 0:
			req, _ = http.NewRequest("POST", Resource(s.address), b)
		case 1:
			s.Write(b, p[0])
			req, _ = http.NewRequest("POST", Resource(s.address), b)
		case 2:
			s.Write(b, p[1])
			req, _ = http.NewRequest("POST", Resource(s.address, p[0]), b)
		default:
			s.Write(b, p[l-1])
			req, _ = http.NewRequest("POST", Resource(s.address, p[:l-1]...), b)
		}
	default:
		req, _ = http.NewRequest(m, Resource(s.address, p...), nil)
	}
	return
}

func do(req *http.Request, f func(*http.Response)) {
	if res, e := http.DefaultClient.Do(req); e == nil {
		fmt.Printf("%v %v --> %v\n", req.Method, req.URL, res.Status)
		f(res)
	} else {
		fmt.Println(e)
	}
	return
}

func Resource(url string, p ...string) (r string) {
	r = url
	for _, v := range p {
		r = r + "/" + html.EscapeString(v)
	}
	return
}
