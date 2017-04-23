package main

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"io"
	"log"
	"os"
)

var (
	hostname string
	response string
)

func init() {
	flag.StringVar(&hostname, "h", "", "-h=emailsvc-test.dtac.co.th:443")
	flag.StringVar(&response, "r", "", "-r=response.xml")
}

func main() {
	flag.Parse()
	if hostname == "" {
		log.Fatal("please give me a hostname")
	}
	if response == "" {
		log.Fatal("please give me a response file name")
	}

	os.Rename(response, hash(hostname))
}

func hash(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return "_" + base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
