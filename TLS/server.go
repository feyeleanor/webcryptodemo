package main

import (
	"crypto/rand"
	"encoding/base32"
	. "fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	. "time"

	"github.com/julienschmidt/httprouter"
)

type FileStore map[string]string

type user struct {
	ID         string
	Registered Time
	FileStore
}

func (u *user) Files() (r int) {
	return len(u.FileStore)
}

type UserDirectory map[string]user

func (u *UserDirectory) NewUserToken() string {
	b := make([]byte, 30)
	if _, e := rand.Read(b); e != nil {
		panic(Sprintf("rand.Read failed: %v", e))
	}
	return base32.StdEncoding.EncodeToString(b)
}

func (u *UserDirectory) CreateUser() (t string) {
	t = u.NewUserToken()
	for _, ok := server.UserDirectory[t]; ok; _, ok = server.UserDirectory[t] {
		t = u.NewUserToken()
	}
	(*u)[t] = user{t, Now(), make(FileStore)}
	return
}

type FileServer struct {
	Started Time
	Address string
	*httprouter.Router
	UserDirectory
	Requests int
}

func (s *FileServer) ListenAndServeTLS(c, k string) {
	s.Started = Now()
	http.ListenAndServeTLS(s.Address, c, k, s.Router)
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

func NewFileServer(a string) *FileServer {
	return &FileServer{Address: a, UserDirectory: make(UserDirectory), Router: httprouter.New()}
}

var templates = template.Must(template.ParseFiles("server_status.html", "user_status.html", "list_files.html"))
var server = NewFileServer("localhost:1024")

func main() {
	server.GET("/", ServerStatus)

	server.POST("/user", RegisterUser)
	server.GET("/user/:id", UserStatus)
	server.DELETE("/user/:id", ForgetUser)

	server.GET("/file/:id", ListFiles)
	server.POST("/file/:id/:filename", StoreFile)
	server.GET("/file/:id/:filename", RetrieveFile)
	server.DELETE("/file/:id/:filename", DeleteFile)

	server.ListenAndServeTLS("cert.pem", "key.pem")
}

func ServerStatus(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderTemplate(w, "server_status", server)
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
