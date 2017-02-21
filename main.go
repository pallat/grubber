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
	"time"

	"os"

	"github.com/BurntSushi/toml"
)

var config struct{ Hosts []string }

func main() {
	_, err := toml.DecodeFile("ignore.toml", &config)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir("cache", os.ModeDir|0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(ServeHTTP)))
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	save := r.Body

	for _, ignore := range config.Hosts {
		if r.RequestURI == ignore {
			h := &httputil.ReverseProxy{
				Director: func(r *http.Request) {
				},
				ModifyResponse: func(r *http.Response) error {
					return nil
				},
			}
			h.ServeHTTP(w, r)
			log.Println("skip", ignore)
			return
		}
	}

	save, r.Body, err = drainBody(r.Body)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	r.Body = save

	state := string(b)
	h := myReverseProxy(state)

	b, err = ioutil.ReadFile("cache/" + hash(state))
	if err == nil {
		w.Write(b)
		log.Println("buffer")
		return
	}

	h.ServeHTTP(w, r)
}

func myReverseProxy(state string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
		},
		FlushInterval: 1 * time.Second,
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

			err = ioutil.WriteFile("cache/"+hash(state), b, 0644)
			if err != nil {
				log.Println(err)
			}

			log.Println("saved")
			return nil
		},
	}
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
