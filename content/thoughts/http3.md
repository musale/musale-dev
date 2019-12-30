---
title: "HTTP/3 and it's underlying protocol QUIC"
date: 2019-11-26T02:55:15+03:00
lastmod: 2019-11-26T02:55:15+03:00
authors:
  - Martin Musale
tags:
  - HTTP/3
  - Networking
summary: "[HTTP/3 and it's underlying protocol QUIC](/thoughts/http3/) talks about the new HTTP/3 layer and the underlying protocol QUIC. It's from the very nice read by Daniel Haxx called [HTTP/3 Explained](https://daniel.haxx.se/http3-explained) – which you should read as it details the processes, how it works and what the future holds for HTPP/3 and QUIC."
---

## Introduction

> QUIC is a name, not an acronym. [^1]

![QUIC Logo](https://http3-explained.haxx.se/images/QUIC.png)

It is the underlying transport protocol on top of which HTTP3 is built. It is meant to address the challenges that were experienced by HTTP2 on TCP and TLS.

## A little about HTTP/2

Now, HTTP/1.X is pretty widespread and the default protocol for the web. However, IETF came up with HTTP/2 to address some challenges currently faced with HTTP/1. In doing so, they also encountered some shortcomings as they were not entirely doing away with HTTP/1 but solving some problems and maintaining the HTTP/1 standards and compatibility.

HTTP/2 is done over TCP. Typical browsers do many parallel transfers over a single connection and if a single packet is dropped, or lost in the network somewhere between two endpoints that speak HTTP/2, it means that the entire TCP connection is brought to a halt while the lost packet is re-transmitted and finds its way to the destination. So everything that comes after the drop, has to wait.

## Enter QUIC

The initial QUIC protocol was designed by Jim Roskind at Google and was initially implemented in 2012, announced publicly to the world in 2013 when Google's experimentation broadened.[^2]

### Features of QUIC

- It's implemented over UDP. This allows re-transmission of packets, congestion control etc. UDP packets will always get to the destination. On the other hand, many network setups block UDP traffic on other ports than 53 (used for DNS) while throttle such data in ways that makes QUIC perform worse than TCP based protocols.

- Guaranteed in-order delivery of streams. Each stream may reach the destination in a different order than the application sent it but each of it is reliably delivered to the application in order.

- Fast handshakes. It offers both 0-RTT and 1-RTT connection setups, meaning that at best QUIC needs no extra round-trips at all when setting up a new connection.

- The transport security used in QUIC is using TLS 1.3 [(RFC 8446)](https://tools.ietf.org/html/rfc8446) and there are never any unencrypted QUIC connections.
- The IETF QUIC protocol is a transport protocol, on top of which other application protocols can be used. The initial application layer protocol is HTTP/3 (h3). I can't wait to see what else can be done here.

## HTTP/3

[^3]HTTP is the first and primary protocol to transport over QUIC. It still maintains the same paradigms and concepts like before. In order to do HTTP over QUIC, changes were required and the results of this is what we now call HTTP/3. These changes include:

- In QUIC the streams are provided by the transport itself, while in HTTP/2 the streams were done within the HTTP layer.

- Due to the streams being independent of each other, the header compression protocol used for HTTP/2 could not be used without it causing a head of block situation.

- QUIC streams are slightly different than HTTP/2 streams. There's but a small fixed number (actually nine on December 18th, 2018!) of known frames in HTTP/3. The most important ones are probably:
  - HEADERS, that sends compressed HTTP headers
  - DATA, sends binary data contents
  - GOAWAY, please shutdown this connection [^4]

## HTTP/3 compared to HTTP/2

[^5]HTTP/3 is designed for QUIC, which is a transport protocol that handles streams by itself.

HTTP/2 is designed for TCP, and therefore handles streams in the HTTP layer.

### Similarities

The two protocols offer clients virtually identical feature sets.

- Both protocols offer streams
- Both protocols offer server push support
- Both protocols have header compression, and QPACK and HPACK are similar in design.
- Both protocols offer multiplexing over a single connection using streams
- Both protocols do prioritization on streams

### Differences

The differences are in the details and primarily there thanks to HTTP/3's use of QUIC:

- HTTP/3 has better and more likely to work early data support thanks to QUIC's 0-RTT handshakes, while TCP Fast Open and TLS usually sends less data and often faces problems.
- HTTP/3 has much faster handshakes thanks to QUIC vs TCP + TLS.
- HTTP/3 does not exist in an insecure or unencrypted version. HTTP/2 can be implemented and used without HTTPS - even if this is rare on the Internet.
- HTTP/2 can be negotiated directly in a TLS handshake with the ALPN extension, while HTTP/3 is over QUIC so it needs an `Alt-Svc:` header response first to inform the client about this fact.

[^1]: https://http3-explained.haxx.se/en/why-quic.html
[^2]: https://http3-explained.haxx.se/en/proc.html
[^3]: https://http3-explained.haxx.se/en/h3.html
[^4]: https://http3-explained.haxx.se/en/h3-streams.html
[^5]: https://http3-explained.haxx.se/en/h3-h2.html
