package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func GenerateAESKey(n int) (b []byte) {
	switch n {
	case 128:
		b = make([]byte, 16)
	case 192:
		b = make([]byte, 24)
	case 256:
		b = make([]byte, 32)
	}
	rand.Read(b)
	return
}

func GenerateIV() (b []byte, e error) {
	b = make([]byte, aes.BlockSize)
	if _, e = rand.Read(b); e != nil {
		panic(e)
	}
	return
}

func SendIV(w io.Writer, k []byte, f func([]byte)) {
	if iv, e := GenerateIV(); e == nil {
		if _, e = w.Write(iv); e == nil {
			f(iv)
		} else {
			fmt.Println(e)
		}
	}
}

func ReadIV(r io.Reader, f func([]byte)) {
	iv := make([]byte, aes.BlockSize)
	if _, e := r.Read(iv); e == nil {
		f(iv)
	} else {
		fmt.Println(e)
	}

}

func EncryptAES(w io.Writer, k []byte, f func(*cipher.StreamWriter)) (e error) {
	//	fmt.Println("EncryptAES: k =", []byte(k))
	var b cipher.Block
	if b, e = aes.NewCipher(k); e == nil {
		SendIV(w, k, func(iv []byte) {
			//			fmt.Println("EncryptAES: iv =", iv)
			f(&cipher.StreamWriter{S: cipher.NewCFBEncrypter(b, iv), W: w})
		})
	}
	return
}

func DecryptAES(r io.Reader, k []byte, f func(*cipher.StreamReader)) (e error) {
	//	fmt.Println("DecryptAES: k =", []byte(k))
	ReadIV(r, func(iv []byte) {
		//		fmt.Println("DecryptAES: iv =", iv)
		var b cipher.Block
		if b, e = aes.NewCipher([]byte(k)); e == nil {
			f(&cipher.StreamReader{S: cipher.NewCFBDecrypter(b, iv), R: r})
		} else {
			fmt.Println(e)
		}
	})
	return
}
