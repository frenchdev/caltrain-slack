package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	main()
	os.Exit(m.Run())
}

func TestNext(t *testing.T) {
	resp := httptest.NewRecorder()

	uri := "http://localhost:5001/next/NB/Redwood%20City%20Caltrain"

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}
	//t.Logf("req: %s", req.)
	http.DefaultServeMux.ServeHTTP(resp, req)
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fail()
	} else {
		if strings.Contains(string(p), "Error") {
			t.Errorf("header response shouldn't return error: %s", p)
		} else if !strings.Contains(string(p), `expected result`) {
			t.Errorf("header response doen't match:\n%s", p)
		}
	}
}
