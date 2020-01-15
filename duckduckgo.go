// Package duckduckgo provides helper functions to work with DuckDuckGo search engine
// this package is designed with iterative refinement (from "How to Design Programs")
package duckduckgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
)

var URLHTML = "https://duckduckgo.com/html"

func GetNoJS(cli *http.Client, word string, page int, opts ...func(*http.Request) *http.Request) ([]byte, error) {
	return getV3(cli, word, page, opts...)
}

func getV3(cli *http.Client, w string, n int, opts ...func(*http.Request) *http.Request) ([]byte, error) {
	queryURL, err := makeQueryNoJSURL(w, n)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	for _, o := range opts {
		r = o(r)
	}

	resp, err := cli.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, BadCode(resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func makeQueryNoJSURL(w string, n int) (string, error) {
	u, err := url.Parse(URLHTML)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Add("q", w)
	q.Add("v", "l")
	q.Add("api", "/d.js")
	q.Add("o", "json")

	// page params
	switch n {
	case 0:
	case 1:
		q.Add("s", "30")
		q.Add("dc", "31")
	default:
		s := 30 + 50*(n-1)
		q.Add("s", strconv.Itoa(s))
		q.Add("dc", strconv.Itoa(s+1))
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Lang sets language query option
// for example Lang("en_us"))
func Lang(v string) func(*http.Request) *http.Request {
	return func(r *http.Request) *http.Request {
		q := r.URL.Query()
		q.Add("kl", v)
		r.URL.RawQuery = q.Encode()
		return r
	}
}

// TODO: use golang.org/x/net/html instead of github.com/anaskhan96/soup
func ParseBody(body []byte) ([]string, error) {
	r := soup.HTMLParse(string(body))
	if r.Error != nil {
		return nil, r.Error
	}

	elems := r.FindAllStrict("a", "class", "result__url")
	var urls []string
	for _, a := range elems {
		if v := a.Attrs()["href"]; v != "" {
			urls = append(urls, v)
		}
	}
	return urls, nil
}

var (
	URL   = "https://duckduckgo.com"
	URLJS = URL + "/d.js"
)

type Query struct {
	Text string
	// Page     int
	Location string
}

func GetDJS(cli *http.Client, query Query, token string) ([]string, error) {
	body, err := GetDJSAPIBody(cli, query, token)
	if err != nil {
		return nil, err
	}

	return ParseJSLinks(body)
}

func GetDJSAPIBody(cli *http.Client, query Query, token string) ([]byte, error) {
	var err error
	if token == "" {
		if token, err = FetchToken(cli, query.Text); err != nil {
			return nil, err
		}
	}

	req, _ := http.NewRequest("GET", URLJS, nil)
	q := req.URL.Query()
	q.Set("q", query.Text)
	q.Set("l", "ru-ru")
	q.Set("vqd", token)
	q.Set("t", "A")
	req.URL.RawQuery = q.Encode()

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, BadCode(resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

// inspired by "https://l-lin.github.io/post/2019/2019-04-02-download_ddg_image_file_with_go/"
func FetchToken(cli *http.Client, word string) (string, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", fmt.Errorf("Could not build the request for URL %s. Error was %s", URL, err)
	}
	q := req.URL.Query()
	q.Add("q", word)
	req.URL.RawQuery = q.Encode()
	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", BadCode(resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Could not read the response body from duckduckgo. Error was %s", err)
	}

	content := string(body)

	r := regexp.MustCompile("vqd=([\\d-]+)")
	token := strings.Replace(r.FindString(content), "vqd=", "", -1)

	return token, nil
}

func ParseJSLinks(data []byte) ([]string, error) {
	pieces := bytes.Split(data, []byte("if (nrn) nrn('d',"))
	if len(pieces) == 1 {
		return nil, errors.New("not splitted")
	}
	data = bytes.TrimRight(pieces[1], ");")

	var urls []struct {
		URL string `json:"u"`
	}

	if err := json.Unmarshal(data, &urls); err != nil {
		return nil, err
	}

	var res []string
	for _, u := range urls {
		if u.URL != "" {
			res = append(res, u.URL)
		}
	}

	return res, nil
}
