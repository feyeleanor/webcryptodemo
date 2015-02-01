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
	f := "this is a test file"
	u := RegisterUser()
	k := GenerateAESKey(256)
	Println("k =", k)
	StoreKey(u, k)
	UserStatus(k, u)
	StoreFile(k, u, "test", f)
	if rf, e := RetrieveFile(k, u, "test"); e == nil {
		switch b, e := ioutil.ReadAll(rf); {
		case e != nil:
			Println(e)
		case string(b) != f:
			Println("Test file corrupted:", string(b))
		}
	}

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

func printEncryptedResponse(key string, b []byte, e error) {
	if e == nil {
		e = DecryptAES(bytes.NewBuffer(b), key, func(s *cipher.StreamReader) {
			b, e = ioutil.ReadAll(s)
		})
	}
	printResponse(b, e)
}

func ServerStatus() {
	printResponse(Do("GET", STATUS))
}

func GetServerKey() (v *rsa.PublicKey) {
	if k, e := LoadPublicKey(printResponse(Do("GET", KEY))); e == nil {
		v = k.(*rsa.PublicKey)
	} else {
		panic(e)
	}
	return
}

func GetIV() (v string) {
	return printResponse(Do("POST", USER))
}

func RegisterUser() string {
	return printResponse(Do("POST", USER))
}

func StoreKey(id, key string) {
	if k, e := EncryptRSA(PublicKey, []byte(key), []byte(id)); e == nil {
		printResponse(Do("POST", KEY, id, string(k)))
	} else {
		Println("StoreKey: public key encryption failed")
	}
}

func UserStatus(key, id string) {
	r, e := Do("GET", USER, id)
	printEncryptedResponse(key, r, e)
}

func ListFiles(key, id string) {
	r, e := Do("GET", FILE, id)
	printEncryptedResponse(key, r, e)
}

func StoreFile(key, id, tag, content string) {
	Println("StoreFile:", content)
	printResponse(Do("POST", FILE, id, tag, content))
}

func RetrieveFile(key, id, tag string) (f io.Reader, e error) {
	r, e := Do("GET", FILE, id, tag)
	printEncryptedResponse(key, r, e)
	f = bytes.NewBuffer(r)
	return
}

func Do(m, r string, p ...string) (b []byte, e error) {
	var res *http.Response
	req := NewRequest(m, r, p...)
	if res, e = http.DefaultClient.Do(req); e == nil {
		Printf("%v %v --> %v\n", req.Method, req.URL, res.Status)
		b, e = ioutil.ReadAll(res.Body)
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

func Resource(url string, p ...string) (r string) {
	r = url
	for _, v := range p {
		r = r + "/" + html.EscapeString(v)
	}
	return
}
