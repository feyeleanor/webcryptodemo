package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type AESKey struct {
	Key []byte
	IV  []byte
	b   cipher.Block
}

//	A new AESKey is always created with a random IV
func NewAESKey(v interface{}) (r *AESKey) {
	r = &AESKey{IV: make([]byte, aes.BlockSize)}
	switch v := v.(type) {
	case int:
		switch v {
		case 128:
			r.Key = make([]byte, 16)
		case 192:
			r.Key = make([]byte, 24)
		case 256:
			r.Key = make([]byte, 32)
		}
		rand.Read(r.Key)
	case []byte:
		r.Key = v
	case string:
		r.Key = []byte(v)
	case AESKey:
		r.Key = v.Key
	}
	r.NewIV()
	r.b, _ = aes.NewCipher(r.Key)
	return
}

func (k *AESKey) NewIV() {
	rand.Read(k.IV)
}

func (k *AESKey) writeIV(w io.Writer, f func()) (e error) {
	if _, e = w.Write(k.IV); e == nil {
		f()
	}
	return
}

func (k *AESKey) readIV(r io.Reader, f func()) (e error) {
	if _, e = r.Read(k.IV); e == nil {
		f()
	}
	return
}

func (k *AESKey) EncryptCFB(w io.Writer, v interface{}) (e error) {
	return k.writeIV(w, func() {
		s := &cipher.StreamWriter{S: cipher.NewCFBEncrypter(k.b, k.IV), W: w}
		switch v := v.(type) {
		case string:
			_, e = s.Write([]byte(v))
		case []byte:
			_, e = s.Write(v)
		case func(*cipher.StreamWriter):
			v(s)
		}
	})
}

func (k *AESKey) DecryptCFB(r io.Reader, v interface{}) (e error) {
	return k.readIV(r, func() {
		s := &cipher.StreamReader{S: cipher.NewCFBDecrypter(k.b, k.IV), R: r}
		switch v := v.(type) {
		case []byte:
			_, e = s.Read(v)
		case func(*cipher.StreamReader):
			v(s)
		}
	})
}

type CFBStream struct {
	*AESKey
}

func NewCFBStream(v interface{}) *CFBStream {
	return &CFBStream{AESKey: NewAESKey(v)}
}

func (s *CFBStream) Write(w io.Writer, v interface{}) (e error) {
	s.NewIV()
	return s.writeIV(w, func() {
		sw := &cipher.StreamWriter{S: cipher.NewCFBEncrypter(s.b, s.IV), W: w}
		switch v := v.(type) {
		case string:
			_, e = sw.Write([]byte(v))
		case []byte:
			_, e = sw.Write(v)
		case func(*cipher.StreamWriter):
			v(sw)
		}
	})
}

func (s *CFBStream) Read(r io.Reader, v interface{}) (e error) {
	return s.readIV(r, func() {
		sr := &cipher.StreamReader{S: cipher.NewCFBDecrypter(s.b, s.IV), R: r}
		switch v := v.(type) {
		case []byte:
			_, e = sr.Read(v)
		case func(*cipher.StreamReader):
			v(sr)
		}
	})
}
