package main

import (
	"fmt"
	"io/ioutil"
)

var Service *ServiceConnection
var LocalKey *AESKey

func main() {
	Service = new(ServiceConnection)
	Service.Connect("localhost:1024")
	fmt.Println(Service.Status())
	Service.CreateUser()
	fmt.Println(Service.UserStatus())

	f := "this is a test file"
	fmt.Println(Service.StoreFile("test", f))

	if rf := Service.RetrieveFile("test"); rf != nil {
		switch b, e := ioutil.ReadAll(rf); {
		case e != nil:
			fmt.Println(e)
		case string(b) != f:
			fmt.Println("Test file corrupted:", string(b))
		default:
			fmt.Println("file returned correctly:", string(b))
		}
	}

	ns := NewCipher()
	Service.StoreKey(string(ns.Key))
	fmt.Println(Service.UserStatus())
}
