package duckduckgo

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
)

const txt = "hello world"

var wd, _ = os.Getwd()

func TestGetNoJSWorks(t *testing.T) {
	_, err := GetNoJS(http.DefaultClient, txt, 0)
	if c, ok := err.(BadCode); err != nil || (ok && c != 403) {
		t.Fatal(err)
	}
}

func TestMakeQueryURLv3(t *testing.T) {
	addq := "&v=l&api=/d.js&o=json"
	tests := []struct {
		w string
		n int
		s string
	}{
		{"star wars", 0, URL + "?q=star+wars" + addq},
		{"star wars", 1, URL + "?q=star+wars&s=30&dc=31" + addq},
		{"star wars", 2, URL + "?q=star+wars&s=80&dc=81" + addq},
		{"star wars", 3, URL + "?q=star+wars&s=130&dc=131" + addq},
	}

	for i, tt := range tests {
		got, err := makeQueryNoJSURL(tt.w, tt.n)
		if err != nil {
			t.Fatalf("test %d: %v", i, err)
		}

		g, _ := url.Parse(got)
		e, _ := url.Parse(tt.s)

		if !reflect.DeepEqual(g.Query(), e.Query()) {
			t.Errorf("test %d: expected %s, got %s\n", i, e.Query(), g.Query())
		}
	}
}

func TestGetDJS(t *testing.T) {
	links, err := GetDJS(http.DefaultClient, Query{Text: "Golang"}, "")
	if err != nil {
		t.Fatal(err)
	}

	for _, l := range links {
		fmt.Println(l)
	}
}

func TestFetchToken(t *testing.T) {
	token, err := FetchToken(http.DefaultClient, "Hello world")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
}

func TestGetDJSAPIWorks(t *testing.T) {
	TestFetchToken(t)
}

func TestParseJSLinks(t *testing.T) {
	f := func(n string) {
		b, err := ioutil.ReadFile(wd + "/testdata/" + n)
		if err != nil {
			log.Fatal(err)
		}

		links, err := ParseJSLinks(b)
		if err != nil {
			t.Fatal(err)
		}

		if len(links) == 0 {
			t.Fatal("0 links")
		}
	}

	tests := []string{"js_links0.js", "js_links1.js", "js_links2.js"}
	for _, tt := range tests {
		f(tt)
	}
}
