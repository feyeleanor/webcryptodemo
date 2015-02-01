package main

import (
	"crypto/cipher"
	. "fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

var templates = template.Must(
	template.ParseFiles(
		"server_status.txt", "server_status.html",
		"user_status.txt", "user_status.html",
		"list_files.txt", "list_files.html",
	),
)
var server = NewFileServer("localhost:1024")

func main() {
	server.GET("/", ServerStatus)

	server.GET("/key", PublicKey)
	server.POST("/key/:id", StoreKey)

	server.POST("/user", RegisterUser)
	server.GET("/user/:id", UserStatus)

	server.GET("/file/:id", ListFiles)
	server.POST("/file/:id/:filename", StoreFile)
	server.GET("/file/:id/:filename", RetrieveFile)

	server.ListenAndServe()
}

func ServerStatus(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderTemplate(w, "server_status", server)
}

func PublicKey(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	server.Requests++
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	Fprint(w, server.PEM)
}

func StoreKey(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	server.RequiresAuthorisation(w, id, func(u *user) {
		if b, e := ioutil.ReadAll(r.Body); e == nil {
			if b, e = DecryptRSA(server.PrivateKey, []byte(b), []byte(id)); e == nil {
				u.Key = string(b)
			} else {
				http.Error(w, e.Error(), http.StatusNotAcceptable)
			}
		} else {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	})
}

func RegisterUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	server.Requests++
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	Fprint(w, server.CreateUser())
}

func UserStatus(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	server.SendEncrypted(w, id, func(s *cipher.StreamWriter, u *user) {
		renderTemplate(s, "user_status", u)
	})
}

func ListFiles(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.SendEncrypted(w, p.ByName("id"), func(s *cipher.StreamWriter, u *user) {
		renderTemplate(w, "list_files", u)
	})
}

func RetrieveFile(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.SendEncrypted(w, p.ByName("id"), func(s *cipher.StreamWriter, u *user) {
		if file, ok := u.FileStore[p.ByName("filename")]; ok {
			Fprint(s, file)
		} else {
			http.NotFound(w, r)
		}
	})
}

func StoreFile(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.ReceiveEncrypted(w, r, p.ByName("id"), func(s *cipher.StreamReader, u *user) {
		if b, e := ioutil.ReadAll(s); e == nil {
			Println("received file:", string(b))
			u.FileStore[p.ByName("filename")] = string(b)
		} else {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	})
}

func renderTemplate(w io.Writer, t string, v interface{}) {
	if e := templates.ExecuteTemplate(os.Stderr, t+".txt", v); e != nil {
		Println(e)
	}
	if e := templates.ExecuteTemplate(w, t+".html", v); e != nil {
		Println(e)
	}
}
