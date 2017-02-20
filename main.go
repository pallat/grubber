package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", &my{}))
}

type my struct {
	state string
	h     http.Handler
}

func (m *my) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	save := r.Body

	save, r.Body, err = drainBody(r.Body)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	r.Body = save
	m.state = string(b)

	m.h = &httputil.ReverseProxy{
		Director: func(r *http.Request) {
		},
		ModifyResponse: func(r *http.Response) error {
			var err error
			save := r.Body

			if r.StatusCode > 299 {
				log.Println("error status", r.Status)
				return nil
			}

			save, r.Body, err = drainBody(r.Body)
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
			}
			r.Body = save

			err = ioutil.WriteFile(hash(m.state), b, 0644)
			if err != nil {
				log.Println(err)
			}

			log.Println("saved")
			return nil
		},
	}

	b, err = ioutil.ReadFile(hash(m.state))
	if err == nil {
		w.Write(b)
		log.Println("buffer")
		return
	}

	m.h.ServeHTTP(w, r)
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func hash(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
