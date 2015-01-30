package main

import (
	"bytes"
	"crypto/tls"
	. "fmt"
	"html"
	"io/ioutil"
	"net/http"
)

const (
	STATUS = "https://localhost:1024/"
	USER   = "https://localhost:1024/user"
	FILE   = "https://localhost:1024/file"
)

var client http.Client
var tlsconfig = &tls.Config{
	InsecureSkipVerify: true,
	ServerName:         "localhost",
}

func main() {
	client = http.Client{Transport: &http.Transport{TLSClientConfig: tlsconfig}}

	ServerStatus()

	users := []string{RegisterUser(), RegisterUser(), RegisterUser()}
	ServerStatus()
	for _, v := range users {
		UserStatus(v)
		ForgetUser(v)
		ServerStatus()
		for _, v := range users {
			UserStatus(v)
		}
	}

	user := RegisterUser()
	files := []string{"A", "BC", "DEF", "GHIJ"}

	filename := func(i int) string {
		return Sprint("file_", i)
	}
	for i, v := range files {
		StoreFile(user, filename(i), v)
	}
	ListFiles(user)
	UserStatus(user)

	for i, _ := range files {
		f := filename(i)
		RetrieveFile(user, f)
		DeleteFile(user, f)
		ListFiles(user)
	}
	UserStatus(user)
	ForgetUser(user)
	ServerStatus()
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

func ServerStatus() {
	printResponse(Do("GET", STATUS))
}

func RegisterUser() (v string) {
	return printResponse(Do("POST", USER))
}

func UserStatus(id string) {
	printResponse(Do("GET", USER, id))
}

func ForgetUser(id string) {
	printResponse(Do("DELETE", USER, id))
}

func ListFiles(id string) {
	printResponse(Do("GET", FILE, id))
}

func StoreFile(id, tag, content string) {
	printResponse(Do("POST", FILE, id, tag, content))
}

func RetrieveFile(id, tag string) {
	printResponse(Do("GET", FILE, id, tag))
}

func DeleteFile(id, tag string) {
	printResponse(Do("DELETE", FILE, id, tag))
}

func Do(m, r string, p ...string) (b []byte, e error) {
	var res *http.Response
	req := NewRequest(m, r, p...)
	if res, e = client.Do(req); e == nil {
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
