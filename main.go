package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

var saved = map[string]string{}

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

			save, r.Body, err = drainBody(r.Body)
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
			}
			r.Body = save

			saved[m.state] = string(b)

			log.Println("saved")
			fmt.Printf("%#v\n\n", saved)

			return nil
		},
	}

	if store, ok := saved[m.state]; ok {
		w.Write([]byte(store))
		log.Println("buffer")
		return
	}

	log.Println("proxy")
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
