package main

import (
	. "fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var templates = template.Must(template.ParseFiles("server_status.html", "user_status.html", "list_files.html"))
var server = NewFileServer("localhost:1024")

func main() {
	server.GET("/", ServerStatus)

	server.GET("/key", PublicKey)
	server.POST("/key/:id", StoreKey)

	server.POST("/user", RegisterUser)
	server.GET("/user/:id", UserStatus)
	server.DELETE("/user/:id", ForgetUser)

	server.GET("/file/:id", ListFiles)
	server.POST("/file/:id/:filename", StoreFile)
	server.GET("/file/:id/:filename", RetrieveFile)
	server.DELETE("/file/:id/:filename", DeleteFile)

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
	server.RequiresAuthorisation(w, id, func(u user) {
		if b, e := ioutil.ReadAll(r.Body); e == nil {
			if b, e = DecryptRSA(server.PrivateKey, b, []byte(id)); e == nil {
				u.Key = string(b)
			} else {
				Println(e)
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
	server.RequiresAuthorisation(w, id, func(u user) {
		renderTemplate(w, "user_status", u)
	})
}

func ForgetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	server.RequiresAuthorisation(w, id, func(_ user) {
		delete(server.UserDirectory, id)
	})
}

func ListFiles(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.RequiresAuthorisation(w, p.ByName("id"), func(u user) {
		renderTemplate(w, "list_files", u)
	})
}

func RetrieveFile(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.RequiresAuthorisation(w, p.ByName("id"), func(u user) {
		if file, ok := u.FileStore[p.ByName("filename")]; ok {
			w.Header().Set("Content-Type", "text/plain")
			Fprint(w, file)
		} else {
			http.NotFound(w, r)
		}
	})
}

func StoreFile(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.RequiresAuthorisation(w, p.ByName("id"), func(u user) {
		if b, e := ioutil.ReadAll(r.Body); e == nil {
			u.FileStore[p.ByName("filename")] = string(b)
		} else {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	})
}

func DeleteFile(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.RequiresAuthorisation(w, p.ByName("id"), func(u user) {
		delete(u.FileStore, p.ByName("filename"))
	})
}

func renderTemplate(w http.ResponseWriter, t string, v interface{}) {
	if e := templates.ExecuteTemplate(w, t+".html", v); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}
