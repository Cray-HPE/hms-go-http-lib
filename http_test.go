// Copyright 2020 Hewlett Packard Enterprise Development LP


package hmshttp

import (
	"testing"
	"log"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
	"encoding/json"
	"context"
	"time"
	"github.com/hashicorp/go-retryablehttp"
)

type pldData struct {
	Field1 string `json:"Field1"`
	Field2 int    `json:"Field2"`
}

var jpayload = `{"Field1":"MyField1","Field2":1234}`
var rsltData = pldData{Field1:"MyField1", Field2:1234}

func handlerFunc(w http.ResponseWriter, req *http.Request) {
	if (req.Method == "POST") {
		body,err := ioutil.ReadAll(req.Body)
		if (err != nil) {
			log.Printf("INTERNAL ERROR can't read req body.\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	} else if (req.Method == "GET") {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jpayload))
	}
}


func TestHMSHttp(t *testing.T) {
	req := NewHTTPRequest("http://localhost/test")
	req.Payload = []byte(jpayload)

	rstr := req.String()
	t.Logf("Request string: '%s'\n",rstr)

	tserv := httptest.NewServer(http.HandlerFunc(handlerFunc))
	req.FullURL = tserv.URL
	req.Method = "POST"

	t.Logf(" *** Testing POST *** \n")
	rpld,err := req.GetBodyForHTTPRequest()

	if (err != nil) {
		t.Errorf("ERROR from POST GetBodyForHTTPRequest(): %v\n",err)
	}

	ba,baerr := json.Marshal(rpld)
	if (baerr != nil) {
		t.Errorf("ERROR marshalling returned struct.\n")
	}

	if (string(ba) != jpayload) {
		t.Errorf("ERROR data miscompare, exp: '%s', got '%s'\n",
			jpayload,string(ba))
	}

	t.Logf(" *** Testing GET *** \n")
	hdrs := make(map[string]string)
	hdrs["httptest1"] = "httptest1/content"
	auth := &Auth{Username: "Groot", Password: "Baz"}

	req.Auth = auth
	req.Payload = nil
	req.Method = "GET"
	req.CustomHeaders = hdrs
	rpld,err = req.GetBodyForHTTPRequest()

	if (err != nil) {
		t.Errorf("ERROR from GET GetBodyForHTTPRequest(): %v\n",err)
	}

	ba,baerr = json.Marshal(rpld)
	if (baerr != nil) {
		t.Errorf("ERROR marshalling returned struct.\n")
	}

	if (string(ba) != jpayload) {
		t.Errorf("ERROR data miscompare, exp: '%s', got '%s'\n",
			jpayload,string(ba))
	}

	t.Logf(" *** Testing GET With Exp Status Codes, Should PASS *** \n")

	req.Payload = []byte("")
	req.Method = "GET"
	req.ExpectedStatusCodes = []int{200,204}
	rpld,err = req.GetBodyForHTTPRequest()

	if (err != nil) {
		t.Errorf("ERROR from GET GetBodyForHTTPRequest(): %v\n",err)
	}

	ba,baerr = json.Marshal(rpld)
	if (baerr != nil) {
		t.Errorf("ERROR marshalling returned struct.\n")
	}

	if (string(ba) != jpayload) {
		t.Errorf("ERROR data miscompare, exp: '%s', got '%s'\n",
			jpayload,string(ba))
	}

	t.Logf(" *** Testing GET With Exp Status Codes, Should FAIL *** \n")

	req.Payload = []byte("")
	req.Method = "GET"
	req.ExpectedStatusCodes = []int{400,404}
	rpld,err = req.GetBodyForHTTPRequest()

	if (err == nil) {
		t.Errorf("GetBodyForHTTPRequest() should fail!\n")
	}

	t.Logf(" *** Error tests ***\n")

	req.FullURL = ""
	_,_,err = req.DoHTTPAction()
	if (err == nil) {
		t.Errorf("DoHTTPAction() should fail with no URL!\n")
	}
	req.FullURL = "http://a.b.c/xyzzy"
	req.Client = nil
	_,_,err = req.DoHTTPAction()
	if (err == nil) {
		t.Errorf("DoHTTPAction() should fail with client nil!\n")
	}
}

func TestCAHttp(t *testing.T) {
	req,err := NewCAHTTPRequest("http://localhost/test","")
	if (err != nil) {
		t.Errorf("ERROR creating CA HTTP request: %v",err)
	}
	req.Payload = []byte(jpayload)

	rstr := req.String()
	t.Logf("Request string: '%s'\n",rstr)

	tserv := httptest.NewServer(http.HandlerFunc(handlerFunc))
	req.FullURL = tserv.URL
	req.Method = "POST"

	t.Logf(" *** Testing POST *** \n")
	rpld,err := req.GetBodyForHTTPRequest()

	if (err != nil) {
		t.Errorf("ERROR from POST GetBodyForHTTPRequest(): %v\n",err)
	}

	ba,baerr := json.Marshal(rpld)
	if (baerr != nil) {
		t.Errorf("ERROR marshalling returned struct.\n")
	}

	if (string(ba) != jpayload) {
		t.Errorf("ERROR data miscompare, exp: '%s', got '%s'\n",
			jpayload,string(ba))
	}
}

func TestCAHttp2(t *testing.T) {
	t.Log("Client Pair, secure client, ca bundle.")

	req,err := NewCAHTTPRequest("http://localhost/test","./test_cabundle.crt")
	if (err != nil) {
		t.Errorf("ERROR creating CA HTTP request: %v",err)
	}
	req.Payload = []byte(jpayload)

	rstr := req.String()
	t.Logf("Request string: '%s'\n",rstr)

	tserv := httptest.NewTLSServer(http.HandlerFunc(handlerFunc))
	req.FullURL = tserv.URL
	req.Method = "POST"

	t.Logf(" *** Testing POST *** \n")
	req.MaxRetryCount = 2
	req.MaxRetryWait = 2
	rpld,err := req.GetBodyForHTTPRequest()

	if (err != nil) {
		t.Errorf("ERROR from POST GetBodyForHTTPRequest(): %v\n",err)
	}

	ba,baerr := json.Marshal(rpld)
	if (baerr != nil) {
		t.Errorf("ERROR marshalling returned struct.\n")
	}

	if (string(ba) != jpayload) {
		t.Errorf("ERROR data miscompare, exp: '%s', got '%s'\n",
			jpayload,string(ba))
	}

	//Try client pair with secure client == nil.

	t.Log("Client Pair, nil secure client.")

	req.TLSClientPair.SecureClient = nil
	rpld,err = req.GetBodyForHTTPRequest()
	if (err != nil) {
		t.Errorf("ERROR from POST GetBodyForHTTPRequest(): %v\n",err)
	}

	ba,baerr = json.Marshal(rpld)
	if (baerr != nil) {
		t.Errorf("ERROR marshalling returned struct.\n")
	}

	if (string(ba) != jpayload) {
		t.Errorf("ERROR data miscompare, exp: '%s', got '%s'\n",
			jpayload,string(ba))
	}

	//Try URL which won't work.

	t.Log("Bad URL, should fail.")
	req.FullURL = "http://blah.zay.com"
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	req.Context = ctx
	defer cancel()
	rpld,err = req.GetBodyForHTTPRequest()
	if (err == nil) {
		t.Errorf("Bad URL operation should have failed, did not.")
	}
}


//This tests backward-compatible mode of using only a retryablehttp client.

func TestManualHttp(t *testing.T) {
	httpClient := retryablehttp.NewClient()
	ctx, cancel := context.WithCancel(context.Background())

	req := HTTPRequest{
				Client: httpClient,
				Context: ctx,
				Timeout: 10 * time.Second,
				ExpectedStatusCodes: []int{http.StatusOK},
				ContentType: "application/json",
				CustomHeaders: make(map[string]string),
	}

	defer cancel()
	req.CustomHeaders["HMS-Service"] = "smd-loader"
	req.Payload = []byte(jpayload)

	rstr := req.String()
	t.Logf("Request string: '%s'\n",rstr)

	tserv := httptest.NewServer(http.HandlerFunc(handlerFunc))
	req.FullURL = tserv.URL
	req.Method = "POST"

	t.Logf(" *** Testing POST *** \n")
	rpld,err := req.GetBodyForHTTPRequest()

	if (err != nil) {
		t.Errorf("ERROR from POST GetBodyForHTTPRequest(): %v\n",err)
	}

	ba,baerr := json.Marshal(rpld)
	if (baerr != nil) {
		t.Errorf("ERROR marshalling returned struct.\n")
	}

	if (string(ba) != jpayload) {
		t.Errorf("ERROR data miscompare, exp: '%s', got '%s'\n",
			jpayload,string(ba))
	}
}

