package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
)

func EncryptRSA(key *rsa.PublicKey, m, l []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha1.New(), rand.Reader, key, m, l)
}

func DecryptRSA(key *rsa.PrivateKey, m, l []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha1.New(), rand.Reader, key, m, l)
}

func LoadPrivateKey(b []byte) (r *rsa.PrivateKey, e error) {
	if block, _ := pem.Decode(b); block != nil {
		if block.Type == "RSA PRIVATE KEY" {
			r, e = x509.ParsePKCS1PrivateKey(block.Bytes)
		}
	}
	return
}

func PublicKeyAsPem(k *rsa.PrivateKey) (r string) {
	if pubkey, e := x509.MarshalPKIXPublicKey(&k.PublicKey); e == nil {
		r = string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey,
		}))
	} else {
		panic(e)
	}
	return
}

func LoadPublicKey(k string) (r interface{}, e error) {
	b, _ := pem.Decode([]byte(k))
	return x509.ParsePKIXPublicKey(b.Bytes)
}

func GenerateAESKey(n int) string {
	var b []byte
	switch n {
	case 128:
		b = make([]byte, 16)
	case 192:
		b = make([]byte, 24)
	case 256:
		b = make([]byte, 32)
	}
	rand.Read(b)
	return string(b)
}

func EncryptAES(w io.Writer, k string, f func(*cipher.StreamWriter)) (e error) {
	var b cipher.Block
	if b, e = aes.NewCipher([]byte(k)); e == nil {
		var iv []byte
		if iv, e = IV(); e == nil {
			fmt.Println("e1:", iv)
			f(&cipher.StreamWriter{S: cipher.NewOFB(b, iv), W: w})
			fmt.Println("e3:", iv)
		}
	}
	return
}

func IV() (b []byte, e error) {
	b = make([]byte, aes.BlockSize)
	_, e = rand.Read(b)
	fmt.Println("IV() =", string(b))
	return
}

func DecryptAES(r io.Reader, k string, f func(*cipher.StreamReader)) (e error) {
	var b cipher.Block
	if b, e = aes.NewCipher([]byte(k)); e == nil {
		iv := make([]byte, aes.BlockSize)
		if _, e = r.Read(iv); e == nil {
			f(&cipher.StreamReader{S: cipher.NewOFB(b, iv), R: r})
		}
	}
	return
}
