package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/rsa"
	. "fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	STATUS = "http://localhost:1024/"
	KEY    = "http://localhost:1024/key"
	USER   = "http://localhost:1024/user"
	FILE   = "http://localhost:1024/file"
)

var PublicKey *rsa.PublicKey

func main() {
	PublicKey = GetServerKey()
	u, k := RegisterUser()
	UserStatus(k, u)

	f := "this is a test file"
	StoreFile(k, u, "test", f)
	UserStatus(k, u)
	RetrieveFile(k, u, "test")
	k = GenerateAESKey(256)
	StoreKey(u, k)
	RetrieveFile(k, u, "test")
	//	UserStatus(k, u)
	//	StoreFile(k, u, "test", f)
	//	if rf, e := RetrieveFile(k, u, "test"); e == nil {
	//		if b, e := ioutil.ReadAll(rf); e == nil {
	//			Println("Test file corrupted:", string(b))
	//		} else {
	//			Println(e)
	//		}
	//	}

	//	k2 := GenerateAESKey(256)
	//	Println("k2 =", k2)
	//	StoreKey(u, k2)
	//	UserStatus(k, u)
	//	UserStatus(k2, u)
	/*
		f := "this is a test file"
		StoreFile(k2, u, "test", f)
		UserStatus(k2, u)
		ListFiles(k2, u)
		if rf, e := RetrieveFile(k2, u, "test"); e == nil {
			switch b, e := ioutil.ReadAll(rf); {
			case e != nil:
				Println(e)
			case string(b) != f:
				Println("Test file corrupted:", string(b))
			}
		}
	*/
}

func printResponse(b []byte, e error) (v string) {
	if e == nil {
		v = string(b)
		Println(v)
	} else {
		Println(e)
	}
	return
}

func printEncryptedResponse(k, b []byte, e error) string {
	if e == nil {
		e = DecryptAES(bytes.NewBuffer(b), k, func(s *cipher.StreamReader) {
			b, e = ioutil.ReadAll(s)
		})
	}
	return printResponse(b, e)
}

func ServerStatus() {
	printResponse(Do("GET", STATUS))
}

func GetServerKey() (v *rsa.PublicKey) {
	if b, e := Do("GET", KEY); e == nil {
		if k, e := LoadPublicKey(string(b)); e == nil {
			v = k.(*rsa.PublicKey)
		} else {
			panic(e)
		}
	}
	return
}

func RegisterUser() (u string, k []byte) {
	k = GenerateAESKey(256)
	if key, e := EncryptRSA(PublicKey, []byte(k), []byte("REGISTER")); e == nil {
		if v, e := Do("POST", USER, string(key)); e == nil {
			u = printEncryptedResponse(k, v, e)
		}
	}
	return
}

func StoreKey(id string, key []byte) {
	if k, e := EncryptRSA(PublicKey, key, []byte(id)); e == nil {
		v, e := Do("POST", KEY, id, string(k))
		printEncryptedResponse(key, v, e)
	}
}

func UserStatus(key []byte, id string) {
	r, e := Do("GET", USER, id)
	printEncryptedResponse(key, r, e)
}

func ListFiles(key []byte, id string) {
	r, e := Do("GET", FILE, id)
	printEncryptedResponse(key, r, e)
}

func StoreFile(key []byte, id, tag, content string) {
	Println("StoreFile:", content)
	r, e := DoEncrypted(key, "POST", FILE, id, tag, content)
	printResponse(r, e)
}

func RetrieveFile(key []byte, id, tag string) (f io.Reader, e error) {
	r, e := Do("GET", FILE, id, tag)
	printEncryptedResponse(key, r, e)
	f = bytes.NewBuffer(r)
	return
}

func Do(m, r string, p ...string) (b []byte, e error) {
	do(NewRequest(m, r, p...), func(res *http.Response) {
		b, e = ioutil.ReadAll(res.Body)
	})
	return
}

func DoEncrypted(k []byte, m, r string, p ...string) (b []byte, e error) {
	do(NewEncryptedRequest(k, m, r, p...), func(res *http.Response) {
		DecryptAES(res.Body, k, func(s *cipher.StreamReader) {
			b, e = ioutil.ReadAll(s)
		})
	})
	return
}

func do(req *http.Request, f func(*http.Response)) {
	if res, e := http.DefaultClient.Do(req); e == nil {
		Printf("%v %v --> %v\n", req.Method, req.URL, res.Status)
		f(res)
	} else {
		Println(e)
	}
	return
}

func NewRequest(m, r string, p ...string) (req *http.Request) {
	switch m {
	case "POST":
		switch l := len(p); l {
		case 0:
			req, _ = http.NewRequest("POST", Resource(r), new(bytes.Buffer))
		case 1:
			req, _ = http.NewRequest("POST", Resource(r), bytes.NewBufferString(p[0]))
		default:
			req, _ = http.NewRequest("POST", Resource(r, p[:l-1]...), bytes.NewBufferString(p[l-1]))
		}
	default:
		req, _ = http.NewRequest(m, Resource(r, p...), nil)
	}
	return
}

func NewEncryptedRequest(k []byte, m, r string, p ...string) (req *http.Request) {
	switch m {
	case "POST":
		switch l := len(p); l {
		case 0:
			req, _ = http.NewRequest("POST", Resource(r), new(bytes.Buffer))
		case 1:
			b := new(bytes.Buffer)
			EncryptAES(b, k, func(s *cipher.StreamWriter) {
				s.Write([]byte(p[0]))
			})
			req, _ = http.NewRequest("POST", Resource(r), b)
		default:
			b := new(bytes.Buffer)
			EncryptAES(b, k, func(s *cipher.StreamWriter) {
				s.Write([]byte(p[l-1]))
			})
			req, _ = http.NewRequest("POST", Resource(r, p[:l-1]...), b)
		}
	default:
		req, _ = http.NewRequest(m, Resource(r, p...), nil)
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
