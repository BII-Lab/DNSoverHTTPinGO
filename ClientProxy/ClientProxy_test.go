package main

import (
	"testing"
)

import "net"
import "dns-master"
import "strings"

func TestSearchIP(t *testing.T) {
	domainTestArray := [4]string{"", "google.com", "tisf.net", "baiduo.abc"}
	versionTestArray := [7]int{-1, 0, 3, 4, 5, 6, 7}
	testinput := [1]string{"114.114.114.114:53"}
	var inputslice []string = testinput[:]
	//test when input nothing.
	testResult, err := searchServerIP(domainTestArray[0], versionTestArray[3], inputslice)
	t.Log("start testing SearchIP")
	if err == nil && len(testResult.Answer) != 0 {
		t.Error("not return nil when input nothing, test faills")
	}
	testResult, err = searchServerIP(domainTestArray[0], versionTestArray[5], inputslice)
	if err == nil && len(testResult.Answer) != 0 {
		t.Error("not return nil when input nothing, test faills")
	}
	//test when version input not 4 or 6
	for i := 0; i < len(versionTestArray); i++ {
		if versionTestArray[i] == 4 || versionTestArray[i] == 6 {
			continue
		}
		testResult, err = searchServerIP(domainTestArray[1], versionTestArray[i], inputslice)
		if err == nil || testResult != nil {
			t.Error("not return error when vesion is not 4 or 6 ")
		}
	}
	//test correct work
	testResult, err = searchServerIP(domainTestArray[1], versionTestArray[3], inputslice)
	if err != nil || len(testResult.Answer) == 0 {
		t.Log(err.Error())
		t.Error("do not get answer back when parametters are right")
	}
	testResult, err = searchServerIP(domainTestArray[1], versionTestArray[5], inputslice)
	if err != nil || len(testResult.Answer) == 0 {
		t.Log(err.Error())
		t.Error("do not get answer back when parametters are right")
	}
	//test domain without a or aaaa record
	testResult, err = searchServerIP(domainTestArray[2], versionTestArray[3], inputslice)
	if err == nil && len(testResult.Answer) != 0 {
		t.Error("return answer when domain without a or aaaa record")
	}

	testResult, err = searchServerIP(domainTestArray[2], versionTestArray[5], inputslice)
	if err == nil && len(testResult.Answer) != 0 {
		t.Error("return answer when domain without a or aaaa record")
	}
	//test wrong domain
	testResult, err = searchServerIP(domainTestArray[3], versionTestArray[3], inputslice)
	if err == nil && len(testResult.Answer) != 0 {
		t.Error("return answer when domain without a or aaaa record")
	}
	testResult, err = searchServerIP(domainTestArray[3], versionTestArray[5], inputslice)
	if err == nil && len(testResult.Answer) != 0 {
		t.Error("return answer when domain without a or aaaa record")
	}
	t.Log("SearchIP ended")
}

func TestGetServerIP(t *testing.T) {
	t.Log("start testing getServerIP")
	var testServer1 []string = make([]string, 1, 1)
	var testDNSServer []string = make([]string, 1, 1)
	testProxy := ClientProxy{
		SERVERS:     testServer1,
		DNS_SERVERS: testDNSServer}
	//test IPV4 address input
	testProxy.DNS_SERVERS[0] = "114.114.114.114:53"
	testProxy.SERVERS[0] = "192.168.121.133"
	err := testProxy.getServerIP()
	if err != nil {
		t.Error("return error when input IPV4 address")
	}
	if testProxy.SERVERS[0] != "192.168.121.133" {
		t.Error("do not parse IPV4 address correctly")
	}
	//test IPV6 address input
	testProxy.SERVERS[0] = "240c::6666"
	err = testProxy.getServerIP()
	if err != nil {
		t.Error("return error when input IPV6 address")
	}
	if testProxy.SERVERS[0] != "240c::6666" {
		t.Error("do not parse IPV6 address correctly")
	}
	//test domain input
	testProxy.SERVERS[0] = "example.com"
	err = testProxy.getServerIP()
	if err != nil {
		t.Error("return error when input correct domain")
	}
	if strings.EqualFold(testProxy.SERVERS[0], "127.0.0.1") {
		t.Error("don't pass example test")
		t.Log(testProxy.SERVERS[0])
	}
	// test invalid input
	testProxy.SERVERS[0] = "12312312"
	err = testProxy.getServerIP()
	if err == nil && len(testProxy.SERVERS) != 0 {
		t.Error("don't pass invalid input test")
		t.Log(testProxy.SERVERS[0])
	}
}

type testResponseWriter struct {
	writebackMsg *string
}

func (this testResponseWriter) LocalAddr() net.Addr {
	return nil
}
func (this testResponseWriter) RemoteAddr() net.Addr {
	return nil
}
func (this testResponseWriter) WriteMsg(m *dns.Msg) error {
	*(this.writebackMsg) = m.String()
	return nil
}
func (this testResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (this testResponseWriter) Close() error {
	return nil
}

// TsigStatus returns the status of the Tsig.
func (this testResponseWriter) TsigStatus() error {
	return nil
}
func (this testResponseWriter) TsigTimersOnly(b bool) {
	return
}
func (this testResponseWriter) Hijack() {
	return
}
func TestServeDNS(t *testing.T) {
	var testServer1 []string = make([]string, 1, 1)
	var testDNSServer []string = make([]string, 1, 1)
	testProxy := ClientProxy{
		SERVERS:     testServer1,
		DNS_SERVERS: testDNSServer}
	var testrequest dns.Msg
	testrequest.SetQuestion("123", dns.TypeA)
	var testRW testResponseWriter
	testRW.writebackMsg = new(string)
	testProxy.C_version = true
	testProxy.ServeDNS(testRW, &testrequest)
	t.Log(*(testRW.writebackMsg))
	testProxy.C_version = false
	testProxy.ServeDNS(testRW, &testrequest)
	t.Log(*(testRW.writebackMsg))
}
