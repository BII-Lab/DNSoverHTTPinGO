package main

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)
import "dns-master"

func TestDoDNSquery(t *testing.T) {
	domainTestArray := [4]string{"", "google.com", "tisf.net", "baiduo.abc"}
	transportStrArray := []string{"UDP", "TCP", "XXX"}
	testinput := [1]string{"8.8.8.8:53"}
	var inputslice []string = testinput[:]
	var testMsg dns.Msg
	testMsg.SetQuestion(domainTestArray[1]+".", dns.TypeA)
	//test wrong transport input
	testResult, err := DoDNSquery(testMsg, transportStrArray[2], inputslice, 5000)
	if testResult != nil && err != nil {
		t.Error("do not pass wrong transport input!")
	}
	//test tcp
	testResult, err = DoDNSquery(testMsg, transportStrArray[1], inputslice, 5000)
	if err != nil || testResult == nil {
		t.Error("do not pass nomal tcp test")
		t.Error(err)
	}
	//test udp
	testResult, err = DoDNSquery(testMsg, transportStrArray[0], inputslice, 5000)
	if err != nil || testResult == nil {
		t.Error("do not pass nomal udp test")
		t.Error(err)
	}
}

type testHTTPWriter struct {
	testResult    *http.Response
	testResultStr *string
}

func (this testHTTPWriter) Header() http.Header {
	return this.testResult.Header
}
func (this testHTTPWriter) Write(b []byte) (int, error) {
	*this.testResultStr = CToGoString(b)
	return len(b), nil
}

func (this testHTTPWriter) WriteHeader(int) {
	return
}
func TestServerHTTP(t *testing.T) {
	testStr := [1]string{"8.8.8.8:53"}
	var inputslice []string = testStr[:]
	testProxy := Server{
		SERVERS: inputslice,
		timeout: 1000,
	}
	var testWriter testHTTPWriter
	testWriter.testResult = new(http.Response)
	testWriter.testResult.Header = make(map[string][]string)
	testWriter.testResultStr = new(string)

	dnsRequest := new(dns.Msg)
	dnsRequest.SetQuestion("baiduo.com.", dns.TypeA)
	requestBytes, _ := dnsRequest.Pack()
	postBytesReader := bytes.NewReader(requestBytes)
	req, _ := http.NewRequest("POST", "http://127.0.0.1", postBytesReader)
	req.Header.Add("X-Proxy-DNS-Transport", "UDP")
	req.Header.Add("Content-Type", "application/X-DNSoverHTTP")
	testProxy.ServeHTTP(testWriter, req)
	if strings.Contains(*testWriter.testResultStr, "Server Error:") {
		t.Error("not passed UDP test, server ERROR")
	}
	req.Header.Set("X-Proxy-DNS-Transport", "TCP")
	testProxy.ServeHTTP(testWriter, req)
	if strings.Contains(*testWriter.testResultStr, "Server Error:") {
		t.Error("not passed TCP test, server ERROR")
	}
	req.Header.Set("X-Proxy-DNS-Transport", "XXX")
	testProxy.ServeHTTP(testWriter, req)
	if !strings.Contains(*testWriter.testResultStr, "Server Error:") {
		t.Error("no server ERROR when transport input invalid")
	}
	req.Header.Set("X-Proxy-DNS-Transport", "UDP")
	req.Header.Set("Content-Type", "text/html")
	if !strings.Contains(*testWriter.testResultStr, "Server Error:") {
		t.Error("no server ERROR when content-type unmatch")
	}

}
func CToGoString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}
