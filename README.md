# load balancer


### goals
 - implement a fully customizable reverse proxy that reads from a config file (preferably a toml)
 - can use quite a few different strategies like RoundRobin, Random, and should be easy to add more.
 - handles ssl termination
 - handles http/2
 - try to support as many proxy specific headers?
    - like certain caching headers
    - check priortization on http2 connections
    -handle downstream downgrades to http/1.1
 - rate limiting

 ### notes
- is it possible to make it so only deps of a specific strategy need to be installed
- just realized all this requires my endpoint urls to be different. should the upstream servers be required to be on different ports?
- in process healthcheck.


### nice to have
- 0 dependency setup (idk if i wanna write a TOML parser tbh, or a custom logger)
- dashboard to see all the incoming requests and where they have been routed to. and whats healthy vs whats not
- http3? might not be possible since we are using a tcp server?
- performance enhancements
  - like buffered log writing?
  - pre-creating goroutine pools on server startup.
