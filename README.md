#DNSoverHTTPinGO

Introduction
------------
This proxy implementation is a follow－up works of DNSoverHTTP（https://github.com/BII-Lab/DNSoverHTTP）written in C. Compared with the previous version, this works provides:

1. A go version client proxy and server proxy which is indepentdent of web server(nginx or apache). That is to say the server proxy will listen to 80/443 and handle the http connect itself based on Golang lib.
2. This Golang version using the event-based method resoloves the performance problem of threads on the proxy_dns_fcgi side in C version.
3. The Go version client/server is implemented without resource name and with different content tpye. For compatible consideration, Go client can connect to FastCGI－version server with -support_version option. If you connect the go version proxy server, do not use -support_version option.
4. In FastCGI-verion, the client should be told where to connect for its DNS proxy service with the "-s arg". It will cause a kind of *loop problem* if arg is a url like "http://proxy-dns.vix.su" and GW server happens to  use 127.0.0.1 as one of its /etc/resolv.conf nameservers. The Golang version correct it by using a speciall DNS request for the domain in the url firstly.

Construction
------------

To compile the code, make sure your have install golang 1.4 version and  already compiled go dns lib written by miekg(https://github.com/miekg/dns).You can find introduction's in miekg's github.To simply get and compile miekg's package in golang, just run:

	go get github.com/miekg/dns
	go build github.com/miekg/dns

Then you can compile the code in this repository by:

	go get github.com/BII-Lab/DNSoverHTTPinGO/
	go build github.com/BII-Lab/DNSoverHTTPinGO/ClientProxy
	go build github.com/BII-Lab/DNSoverHTTPinGO/ServerProxy

Server Installation
-------------------

The server proxy will need a working name server configuration on your server. The server should be reachable by UDP and TCP, and you should have a clear ICMP path to it, as well as full MTU (1500 octets or larger) and the ability to receive fragmented UDP (to make EDNS0 usable.)

1.compile ServerProxy.
	
	go build github.com/BII-Lab/DNSoverHTTPinGO/ServerProxy

2.make sure you have a working resolver.
3.run the ServerProxy as （listion to port 80 currently） 
	
	./ServerProxy -proxy "[your upper resolver's ip address]"
4. For more help information, you can use -h option
	
	./ServeProxy -h

Client Installation
-------------------

The ClientProxy will listen on the port assigned(defaut port is 53). And it must also be told where are which type proxy service to connect to. If you want to connect to the FCGI proxy server(https://github.com/BII-Lab/DNSoverHTTP/tree/master/proxy_dns_fcgi) your need add a -support_version option. Both domain name or ip address for server proxy is acceptable. If you use a domain name, you need to set a resolver's IP address as a start point.

1. compile ClientProxy.
	
	go build github.com/BII-Lab/DNSoverHTTPinGO/ClientProxy
	
2. If you want to redirect all you normal DNS traffic to the proxy, configure your /etc/resolv.conf. Set nameserver to 127.0.0.1.(optional)
	
3.run ClientProxy to connect the ServerProxy. 
	
	./ClientProxy -proxy "the domain or address of ServerProxy"

4. Note that If you want to use domain as the arg for -proxy, the code will use a default resolver 114.114.114.114:53 for the resolution of the ServerProxy domain. You can also specify the resolver with --dns_server with "you prefered dns resolver ip:port".

	./ClientProxy -proxy "the domain or address of ServerProxy" -dns_server "8.8.8.8:53"
	
5. For more help information, you can use -h option
	
	./ClientProxy -h

Testing
-------

Make sure you have a working "dig" command. If you started your client side dns_proxy service on 127.0.0.1, then you should be able to say:

	dig @127.0.0.1 www.yeti-dns.org aaaa

and get a result back.
