package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
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
