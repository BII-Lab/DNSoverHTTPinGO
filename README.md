# goDNSoverHTTP

Introduction
------------

This proxy implementation is a follow－up works of DNSoverHTTP（https://github.com/BII-Lab/DNSoverHTTP）. Compared with the previous version, this works provides:

1. a go version end point which can be totally compatible with the FastCGI－version server with -support_version option. 

2. a go version Proxy server which optimized parallelizing with event-based lib. It should be noticed that this server is only compatible with the client in same repository without -support-version option

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

Server Installation
-------------------

The server proxy will need a working name server configuration on your server. The server should be reachable by UDP and TCP, and you should have a clear ICMP path to it, as well as full MTU (1500 octets or larger) and the ability to receive fragmented UDP (to make EDNS0 usable.)

1.compile ServerProxy.
2.make sure you have a working resolver.
3.run the ServerProxy as ./ServerProxy -proxy="[your resolver ip address]"

Client Installation
-------------------

The ClientProxy will listen on the port assigned(defaut port is 53). And it must also be told where are which type proxy service to connect to. If you want to connect to the FCGI proxy server(https://github.com/BII-Lab/DNSoverHTTP/tree/master/proxy_dns_fcgi) your need add a -support_version option. Both domain name or ip address for server proxy is acceptable. If you use a domain name, you need to set a resolver's IP address as a start point.

1. compile ClientProxy.
2. If you want to redirect all you normal DNS traffic to the proxy, configure your /etc/resolv.conf. Set nameserver to 127.0.0.1.(optional)
3.run ClientProxy. Example ./ClientProxy -proxy="[240c:f:1:11::66]" -support_version.

Testing
-------

Make sure you have a working "dig" command. If you started your client side
dns_proxy service on 127.0.0.1, then you should be able to say:

	dig @127.0.0.1 www.vix.su aaaa

and get a result back.
