package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	. "fmt"
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
	Println(k.PublicKey)
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

func EncryptAES(m, k string) (o []byte, e error) {
	if o, e = Quantise(m); e == nil {
		var b cipher.Block
		if b, e = aes.NewCipher([]byte(k)); e == nil {
			var iv []byte
			if iv, e = IV(); e == nil {
				c := cipher.NewCBCEncrypter(b, iv)
				c.CryptBlocks(o, o)
				o = append(iv, o...)
			}
		}
	}
	return
}

func Quantise(m string) (b []byte, e error) {
	b = append(b, m...)
	if p := len(b) % aes.BlockSize; p != 0 {
		p = aes.BlockSize - p
		//  this is insecure and inflexible as we're padding with NUL!
		b = append(b, make([]byte, p)...)
	}
	return
}

func IV() (b []byte, e error) {
	b = make([]byte, aes.BlockSize)
	_, e = rand.Read(b)
	return
}

func DecryptAES(m []byte, k string) (r string, e error) {
	var b cipher.Block
	if b, e = aes.NewCipher([]byte(k)); e == nil {
		var iv []byte
		iv, m = Unpack(m)
		c := cipher.NewCBCDecrypter(b, iv)
		c.CryptBlocks(m, m)
		r = Dequantise(m)
	}
	return
}

func Dequantise(m []byte) string {
	var i int
	for i = len(m) - 1; i > 0 && m[i] == 0; i-- {
	}
	return string(m[:i+1])
}

func Unpack(m []byte) (iv, r []byte) {
	return m[:aes.BlockSize], m[aes.BlockSize:]
}
