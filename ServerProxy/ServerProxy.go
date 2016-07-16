package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

type Server struct {
	ACCESS      []*net.IPNet
	SERVERS     []string
	s_len       int
	entries     int64
	max_entries int64
	NOW         int64
	giant       *sync.RWMutex
	timeout     time.Duration
	TransPro    int //specify for transmit protocol
}

const UDPcode = 1
const TCPcode = 2

func DoDNSquery(m dns.Msg, TransProString string, server []string, timeout time.Duration) (*dns.Msg, error) {
	dnsClient := new(dns.Client)
	if dnsClient == nil {
		return nil, errors.New("Cannot create DNS client")
	}

	dnsClient.ReadTimeout = timeout
	dnsClient.WriteTimeout = timeout
	if TransProString != "TCP" && TransProString != "UDP" {
		return nil, errors.New(fmt.Sprintf("Transport not TCP or UDP: %s", TransProString))
	}
	dnsClient.Net = strings.ToLower(TransProString)
	ServerStr := server[rand.Intn(len(server))]
	ServerAddr := net.ParseIP(ServerStr)
	if ServerAddr.To16() != nil {
		ServerStr = "[" + ServerStr + "]:53"
	} else if ServerAddr.To4() != nil {
		ServerStr = ServerStr + ":53"
	} else {
		return nil, errors.New(fmt.Sprintf("Invalid server address: %s", ServerStr))
	}
	dnsResponse, _, err := dnsClient.Exchange(&m, ServerStr)
	if err != nil {
		return nil, err
	}
	return dnsResponse, nil
}

// Process HTTP requests.
// "dns-wireformat" requests get proxied, others get read out of our answer
// directory.
func (this Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/.well-known/dns-wireformat" {
		this.tryDNSoverHTTP(w, r)
        } else {
		this.tryStaticHTTP(w, r)
	}
}

// XXX: the directory to read HTML from should be a command-line argument, not "html/"
func (this Server) tryStaticHTTP(w http.ResponseWriter, r *http.Request) {
	var path string
	if r.URL.Path == "/" {
		path = "html/index.html"
	} else {
		path = filepath.Clean("html/" + r.URL.Path)
	}
	if !strings.HasPrefix(path, "html/") {
		msg := fmt.Sprintf("Invalid URL: %s", r.URL.Path)
		http.Error(w, msg, 403)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		errcode := 400
		if err == os.ErrPermission {
			errcode = 401
		} else if err == os.ErrNotExist {
			errcode = 404
		}
		msg := fmt.Sprintf("Error retrieving '%s': %s", r.URL.Path, err)
		http.Error(w, msg, errcode)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	// not sure what to do with the error here
}

func (this Server) tryDNSoverHTTP(w http.ResponseWriter, r *http.Request) {
	TransProString := r.Header.Get("Proxy-DNS-Transport")
	if TransProString == "TCP" {
		this.TransPro = TCPcode
	} else if TransProString == "UDP" {
		this.TransPro = UDPcode
	} else {
		msg := fmt.Sprintf("Transport protocol not UDP or TCP: %s", this.TransPro)
		_D("%s", msg)
		http.Error(w, msg, 415)
		return
	}
	contentTypeStr := r.Header.Get("Content-Type")
	if contentTypeStr != "application/octet-stream" {
		msg := fmt.Sprintf("Unsupported content-type: %s", contentTypeStr)
		_D("%s", msg)
		http.Error(w, msg, 415)
		return
	}
	var requestBody []byte
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Error reading HTTP request: %s", err)
		_D("%s", msg)
		http.Error(w, msg, 400)
		return
	}
	if len(requestBody) < (int)(r.ContentLength) {
		msg := fmt.Sprintf("Error reading HTTP request: expected %d bytes but only read %d",
			(int)(r.ContentLength), len(requestBody))
		_D("%s", msg)
		http.Error(w, msg, 400)
		return
	}
	var dnsRequest dns.Msg
	err = dnsRequest.Unpack(requestBody)
	if err != nil {
		msg := fmt.Sprintf("Error unpacking DNS message: %s", err)
		_D("%s", msg)
		http.Error(w, msg, 400)
		return
	}
	dnsResponse, err := DoDNSquery(dnsRequest, TransProString, this.SERVERS, this.timeout)
	if err != nil {
		msg := fmt.Sprintf("Error querying DNS resolver: %s", err)
		_D("%s", msg)
		http.Error(w, msg, 500)
		return
	}
	if dnsResponse == nil {
		msg := "Error querying DNS resolver: no response"
		_D("%s", msg)
		http.Error(w, msg, 500)
		return
	}
	response_bytes, err := dnsResponse.Pack()
	if err != nil {
		msg := fmt.Sprintf("Error converting DNS message to bytes: %s", err)
		_D("%s", msg)
		http.Error(w, msg, 500)
		return
	}
	_, err = w.Write(response_bytes)
	if err != nil {
		msg := fmt.Sprintf("Error writing HTTP response: %s", err)
		_D("%s", msg)
		return
	}
}

func main() {
	var (
		S_SERVERS     string
		timeout       int
		max_entries   int64
		ACCESS        string
		ServeTLS      bool
		tls_cert_path string
		tls_key_path  string
	)
	flag.StringVar(&S_SERVERS, "proxy", "127.0.0.1", "we proxy requests to those servers") //Not sure use IP or URL, default server undefined
	flag.IntVar(&timeout, "timeout", 5, "timeout")
	flag.BoolVar(&DEBUG, "debug", false, "enable/disable debug")
	flag.Int64Var(&max_entries, "max_cache_entries", 2000000, "max cache entries")
	flag.StringVar(&ACCESS, "access", "0.0.0.0/0", "allow those networks, use 0.0.0.0/0 to allow everything")
	flag.BoolVar(&ServeTLS, "ServeTls", false, "whether serve TLS")
	flag.StringVar(&tls_cert_path, "certificate_path", "", "the path of server's certicate for TLS")
	flag.StringVar(&tls_key_path, "key_path", "", "the path of server's key for TLS")
	flag.Parse()
	servers := strings.Split(S_SERVERS, ",")
	proxyServer := Server{
		SERVERS:     servers,
		timeout:     time.Duration(timeout) * time.Second,
		max_entries: max_entries,
		ACCESS:      make([]*net.IPNet, 0)}
	for _, mask := range strings.Split(ACCESS, ",") {
		_, cidr, err := net.ParseCIDR(mask)
		if err != nil {
			panic(err)
		}
		_D("added access for %s\n", mask)
		proxyServer.ACCESS = append(proxyServer.ACCESS, cidr)
	}
	_D("start server HTTP")
	err := http.ListenAndServe(":80", proxyServer)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
		return
	}
	if ServeTLS {
		err := http.ListenAndServeTLS(":443", tls_cert_path, tls_key_path, proxyServer)
		if err != nil {
			log.Fatal("ListenAndServe:", err)
			return
		}
	}
	for {
		proxyServer.NOW = time.Now().UTC().Unix()
		time.Sleep(time.Duration(1) * time.Second)
	}
}
