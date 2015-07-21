package main

import (
	"bytes"
	"dns-master"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// flag whether we want to emit debug output
var DEBUG bool = false

// called for debug output
func _D(fmt string, v ...interface{}) {
	if DEBUG {
		log.Printf(fmt, v...)
	}
}
func (this ClientProxy) getServerIP() error {
	var dns_servers []string
	dnsClient := new(dns.Client)
	if dnsClient == nil {
		return errors.New("Can not new dns Client")
	}
	dnsClient.WriteTimeout = this.timeout
	dnsClient.ReadTimeout = this.timeout
	for _, serverstring := range this.SERVERS {
		ipaddress := net.ParseIP(serverstring)
		if ipaddress != nil {
			dns_servers = append(dns_servers, serverstring)
		} else {
			dnsRequest := new(dns.Msg)
			dnsRequest.SetQuestion("domain", dns.TypeA)
			dnsResponse, _, err := dnsClient.Exchange(dnsRequest, this.DNS_SERVERS[rand.Intn(len(this.DNS_SERVERS))])
			if err == nil {
				for i := 0; i < len(dnsResponse.Answer); i++ {
					dns_servers = append(dns_servers, dnsResponse.Answer[i].String())
				}
			} else {
				return err
			}
			dnsRequest.SetQuestion("domain", dns.TypeAAAA)
			dnsResponse, _, err = dnsClient.Exchange(dnsRequest, this.DNS_SERVERS[0])
			if err == nil {
				for i := 0; i < len(dnsResponse.Answer); i++ {
					dns_servers = append(dns_servers, "["+dnsResponse.Answer[i].String()+"]")
				}
			} else {
				return err
			}
		}
	}
	this.SERVERS = dns_servers
	return nil
}

func (this ClientProxy) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	request_bytes, err := request.Pack() //I am not sure it is better to pack directly or using a pointer
	if err != nil {
		SRVFAIL(w, request)
		_D("error in packing request, error message: %s", err)
		return
	}
	ServerInput := this.SERVERS[rand.Intn(len(this.SERVERS))]
	ipaddress := net.ParseIP(ServerInput)
	var ServerInputurl string
	if ipaddress.To4() != nil {
		ServerInputurl = "http://" + ServerInput
	} else {
		ServerInputurl = "http://[" + ServerInput + "]"
	}

	postBytesReader := bytes.NewReader(request_bytes)
	if this.C_version {
		ServerInputurl = ServerInputurl + "/proxy_dns"
	}
	req, err := http.NewRequest("POST", ServerInputurl, postBytesReader) //need add random here in future
	if err != nil {
		SRVFAIL(w, request)
		_D("error in creating HTTP request, error message: %s", err)
		return
	}
	if this.C_version {
		req.Header.Add("Host", ServerInput)
		req.Header.Add("Accept: ", "application/octet-stream")
		req.Header.Add("Content-Type: ", "application/octet-stream")
		if this.TransPro == UDPcode {
			req.Header.Add("Proxy-DNS-Transport", "UDP")
		} else if this.TransPro == TCPcode {
			req.Header.Add("Proxy-DNS-Transport", "TCP")
		}
	} else {
		if this.TransPro == UDPcode {
			req.Header.Add("X-Proxy-DNS-Transport", "udp")
		} else if this.TransPro == TCPcode {
			req.Header.Add("X-Proxy-DNS-Transport", "tcp")
		}
		req.Header.Add("Content-Type", "application/X-DNSoverHTTP")
	}
	resp, err := http.DefaultClient.Do(req)
	//	defer resp.Body.Close()
	if err != nil {
		SRVFAIL(w, request)
		_D("error in HTTP post request, error message: %s", err)
		return
	}
	var requestBody []byte
	requestBody, err = ioutil.ReadAll(resp.Body)
	//	nRead, err := resp.Body.Read(requestBody)
	if err != nil {
		// these need to be separate checks, otherwise you will get a nil-reference
		// when you print the error message below!
		SRVFAIL(w, request)
		_D("error in reading HTTP response, error message: %s", err)
		return
	}
	//I not sure whether I should return server fail directly
	//I just found there is a bug here. Body.Read can not read all the contents out, I don't know how to solve it.
	if len(requestBody) < (int)(resp.ContentLength) {
		SRVFAIL(w, request)
		_D("fail to read all HTTP content")
		return
	}
	var DNSreponse dns.Msg
	err = DNSreponse.Unpack(requestBody)
	if err != nil {
		SRVFAIL(w, request)
		_D("error in packing HTTP response to DNS, error message: %s", err)
		return
	}
	err = w.WriteMsg(&DNSreponse)
	if err != nil {
		_D("error in sending DNS response back, error message: %s", err)
		return
	}
}

func SRVFAIL(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeServerFailure)
	w.WriteMsg(m)
}

type ClientProxy struct {
	ACCESS      []*net.IPNet
	SERVERS     []string
	s_len       int
	entries     int64
	max_entries int64
	NOW         int64
	giant       *sync.RWMutex
	timeout     time.Duration
	TransPro    int //specify for transmit protocol
	DNS_SERVERS []string
	C_version   bool
}

const UDPcode = 1
const TCPcode = 2

func main() {
	var (
		S_SERVERS       string
		S_LISTEN        string
		S_ACCESS        string
		timeout         int
		max_entries     int64
		expire_interval int64
		S_DNS_SERVERS   string
		Support_C       bool
	)
	flag.StringVar(&S_SERVERS, "proxy", "", "we proxy requests to those servers,input like http://biilab.cn") //Not sure use IP or URL, default server undefined
	flag.StringVar(&S_LISTEN, "listen", "[::]:53", "listen on (both tcp and udp)")
	flag.StringVar(&S_ACCESS, "access", "127.0.0.0/8,10.0.0.0/8", "allow those networks, use 0.0.0.0/0 to allow everything")
	flag.IntVar(&timeout, "timeout", 5, "timeout")
	flag.Int64Var(&expire_interval, "expire_interval", 300, "delete expired entries every N seconds")
	flag.BoolVar(&DEBUG, "debug", false, "enable/disable debug")
	flag.Int64Var(&max_entries, "max_cache_entries", 2000000, "max cache entries")
	flag.StringVar(&S_DNS_SERVERS, "dns_server", "114.114.114.114:53", "DNS server for initial server lookup")
	flag.BoolVar(&Support_C, "support_version", false, "Whether support Paul Vixie's C version")
	flag.Parse()
	servers := strings.Split(S_SERVERS, ",")
	dns_servers := strings.Split(S_DNS_SERVERS, ",")
	UDPproxyer := ClientProxy{
		giant:       new(sync.RWMutex),
		ACCESS:      make([]*net.IPNet, 0),
		SERVERS:     servers,
		s_len:       len(servers),
		NOW:         time.Now().UTC().Unix(),
		entries:     0,
		timeout:     time.Duration(timeout) * time.Second,
		max_entries: max_entries,
		TransPro:    UDPcode,
		DNS_SERVERS: dns_servers,
		C_version:   Support_C}
	TCPproxyer := ClientProxy{
		giant:       new(sync.RWMutex),
		ACCESS:      make([]*net.IPNet, 0),
		SERVERS:     servers,
		s_len:       len(servers),
		NOW:         time.Now().UTC().Unix(),
		entries:     0,
		timeout:     time.Duration(timeout) * time.Second,
		max_entries: max_entries,
		TransPro:    TCPcode,
		DNS_SERVERS: dns_servers,
		C_version:   Support_C}
	for _, mask := range strings.Split(S_ACCESS, ",") {
		_, cidr, err := net.ParseCIDR(mask)
		if err != nil {
			panic(err)
		}
		_D("added access for %s\n", mask)
		UDPproxyer.ACCESS = append(UDPproxyer.ACCESS, cidr)
		TCPproxyer.ACCESS = append(TCPproxyer.ACCESS, cidr)
	}
	err := UDPproxyer.getServerIP()
	if err != nil {
		_D("can not get server address")
		return
	}
	err = TCPproxyer.getServerIP()
	if err != nil {
		_D("can not get server address")
		return
	}
	for _, addr := range strings.Split(S_LISTEN, ",") {
		_D("listening @ %s\n", addr)
		go func() {
			if err := dns.ListenAndServe(addr, "udp", UDPproxyer); err != nil {
				log.Fatal(err)
			}
		}()

		go func() {
			if err := dns.ListenAndServe(addr, "tcp", TCPproxyer); err != nil {
				log.Fatal(err)
			}
		}()
	}
	for {
		UDPproxyer.NOW = time.Now().UTC().Unix()
		time.Sleep(time.Duration(1) * time.Second)
	}
}
