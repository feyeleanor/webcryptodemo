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

	server.POST("/user", CreateUser)
	server.GET("/user/:id", UserStatus)

	server.GET("/file/:id", ListFiles)
	server.GET("/file/:id/:filename", RetrieveFile)
	server.POST("/file/:id/:filename", StoreFile)

	server.ListenAndServe()
}

func debugPrint(r *http.Request, s string) {
	Println(s)
	Println(r.Method, r.URL)
}

func ServerStatus(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderTemplate(w, "server_status", server)
}

func PublicKey(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	server.Requests++
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	Fprint(w, server.PEM)
}

func ReadRSA(r io.Reader, l []byte) (b []byte, e error) {
	if b, e = ioutil.ReadAll(r); e == nil {
		b, e = DecryptRSA(server.PrivateKey, []byte(b), l)
	}
	return
}

func CreateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.Requests++
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if b, e := ReadRSA(r.Body, []byte("REGISTER")); e == nil {
		server.SendEncrypted(w, server.CreateUser(b), func(s *cipher.StreamWriter, u *user) {
			s.Write([]byte(u.ID))
		})
	} else {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

func StoreKey(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.RequiresAuthorisation(w, p.ByName("id"), func(u *user) {
		Println("id =", u.ID)
		if b, e := ReadRSA(r.Body, []byte(u.ID)); e == nil {
			u.Key = b
			http.Error(w, "key updated", http.StatusAccepted)
			server.SendEncrypted(w, u.ID, func(s *cipher.StreamWriter, u *user) {
				renderTemplate(s, "user_status", u)
			})
		} else {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	})
}

func UserStatus(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.SendEncrypted(w, p.ByName("id"), func(s *cipher.StreamWriter, u *user) {
		renderTemplate(s, "user_status", u)
	})
}

func ListFiles(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.SendEncrypted(w, p.ByName("id"), func(s *cipher.StreamWriter, u *user) {
		renderTemplate(w, "list_files", u)
	})
}

func RetrieveFile(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.RequiresAuthorisation(w, p.ByName("id"), func(u *user) {
		if file, ok := u.FileStore[p.ByName("filename")]; ok {
			server.SendEncrypted(w, u.ID, func(s *cipher.StreamWriter, _ *user) {
				Fprint(s, file)
			})
		} else {
			http.NotFound(w, r)
		}
	})
}

func StoreFile(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	server.ReceiveEncrypted(w, r, p.ByName("id"), func(si *cipher.StreamReader, u *user) {
		if b, e := ioutil.ReadAll(si); e == nil {
			u.FileStore[p.ByName("filename")] = string(b)
			server.SendEncrypted(w, u.ID, func(so *cipher.StreamWriter, _ *user) {
				renderTemplate(so, "user_status", u)
			})
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
