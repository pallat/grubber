package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

var port string
var logger *log.Logger
var f *os.File

var config struct {
	Ignores []string `toml:"ignores"`
	Match   []string `toml:"match"`
}

func init() {
	var err error
	f, err = os.OpenFile("trace.log", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(f, "logger: ", log.Lshortfile)
	flag.StringVar(&port, "port", ":8080", "-port=:8080")
}

func main() {
	flag.Parse()
	defer f.Close()

	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir("cache", os.ModeDir|0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	fmt.Println("omr api proxy active", port)
	log.Fatal(http.ListenAndServe(port, http.HandlerFunc(ServeHTTP)))
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	stateis := "body"

	save := r.Body
	for _, ignore := range config.Ignores {
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

	for _, match := range config.Match.URI {
		if match == r.RequestURI {
			state = r.RequestURI
			stateis = "uri"
		}
	}

	h := myReverseProxy(state)

	b, err = ioutil.ReadFile("cache/" + hash(state))
	if err == nil {
		w.Write(b)
		log.Println("buffer")
		return
	}

	logger.Println("state:", stateis, "uri:", r.RequestURI, "hash:", hash(state))

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
	return "_" + base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
