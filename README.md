# goDNSoverHTTP
This is a proxy used for DNS over HTTP. It provides:
1. a ClientProxy listen DNS query on port 53 and transport it to another address through HTTP.Then unpack response to DNS response and send it back to port 53
2. a ServerProxy listen port 80 and foward DNS query to a resolver(can be loopback). Then response the query packed in HTTP.

These proxies are written in GO version 1.4 and depends on github.com/miekg/dns (https://github.com/miekg/dns). And the ClientProxy is fully compatible with original FCI proxy in C.(should use a -support_version option)