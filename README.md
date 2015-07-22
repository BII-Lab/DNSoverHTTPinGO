# goDNSoverHTTP
Introduction
------------

This proxy is a future works of DNSoverHTTP.Compared with the old work, this wors provides:

1. a go version end point which can be totally compatible with the old server with -support_version option. 
2. a go version Proxy server which optimized parallelizing compares with the old server. It should be noticed that this server is only compatible with the client in same repository without -support-version option.

The great advantage to this approach is that HTTP usually makes it through
even the worst coffee shop or hotel room firewalls, since commerce may be at
stake. We also benefit from HTTP's persistent TCP connection pool concept,
which DNS on TCP/53 does not have. Lastly, HTTPS will work, giving privacy.

Construction
------------

To compile the code, make sure your have install golang 1.4 version and  already compiled go dns lib written by miekg(https://github.com/miekg/dns).

go get github.com/BII-Lab/DNSoverHTTPinGO/
go build github.com/BII-Lab/DNSoverHTTPinGO/ClientProxy
go build github.com/BII-Lab/DNSoverHTTPinGO/ServerProxy


