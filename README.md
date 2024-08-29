# caddy-websocket-ratelimit

A websocket rate limiting module for Caddy 2

## Introduction

There has been no flow control plugin for websocket in Caddy, but in actual use, flow control of websocket is very necessary, especially when websocket is used as an API gateway.

Now, this plugin allows you to easily control the flow of websocket in Caddy.

We wrap the bidirectionalRateLimitedConn structure, which wraps the original net.Conn. Its function is to control the reading and writing of websocket based on the flow control algorithm of token bucket.

In addition, we create a new `limitedResponseWriter` structure, which wraps the original `http.ResponseWriter`. When other services call `http.NewResponseController(w).Hijack()` to hijack the connection, we return the `bidirectionalRateLimitedConn` structure, so that the reading and writing of websocket can be controlled.

The above approach is more consistent with the original implementation of Caddy. This implementation is now closer to Caddy's original WebSocket proxy implementation, while also retaining the rate limiting functionality we added. It ensures that all buffered data is handled correctly, which is important for certain WebSocket handshake scenarios.

To use this updated module, your Caddyfile configuration will remain unchanged: 


``` caddyfile
{
    order websocket_rate_limit before reverse_proxy
}

:8081 {
    websocket_rate_limit {
        up_byte_rate 2
        up_burst_limit 100
        down_byte_rate 2
        down_burst_limit 100
        time_window 60
    }
    reverse_proxy localhost:8080
}
```

Parameters:
 - <up_byte_rate>: The request rate limit  specified in requests per time window. 
 - <up_burst_limit>: The request rate limit specified in requests per time window. Set to 0 to disable the limit.
 - <down_byte_rate>: The request rate limit specified in requests per time window.
 - <down_burst_limit>: The request rate limit specified in requests per time window. Set to 0 to disable the limit. 
 - <time_window>: The time window (in seconds) for the rate limit. Default is 60 seconds.

 For more details, please refer to the sample directory. 
 
 
## Build

### Build From Github

```
xcaddy build --with github.com/imneov/caddy-websocket-ratelimit
```

### Build From Source
```
git clone https://github.com/imneov/caddy-websocket-ratelimit

cd caddy-websocket-ratelimit

xcaddy build --with github.com/imneov/caddy-websocket-ratelimit=./
```

### Check Plugin Module
```
$ ./caddy list-modules | grep rate
http.handlers.websocket_rate_limit
```


## Quick Start

1. Download sample

``` bash
git clone https://github.com/imneov/caddy-websocket-ratelimit
cd sample
```

2. Start Websocket Server

``` bash
go run server.go
```

A websocket server is running on port 8080, It will receive messages from client and echo to client.

3. Start Caddy with the plugin

Make sure you have caddy with the `caddy-websocket-ratelimit` plugin module

``` bash
$ caddy list-modules | grep rate
http.handlers.websocket_rate_limit
```

Start Caddy with sample caddyfile

``` bash
$ caddy run --config Caddyfile
```

the log level in this configuration is debug, do not use debug level logs in a production environment
4. Test Rate Limit

Start a websocket client to connect to the websocket server, and send messages to the server.

``` bash
$ curl -i http://localhost:8081
```

In the early stage of operation, tokens are consumed until the token bucket is empty.

After running for about 10 seconds, when caddy reads or writes data, it will return the error rate: Wait(...) exceeds limiter's burst ...

In the debug level log of caddy, you can see the following log:

```
024/08/29 02:23:53.904	DEBUG	http.handlers.reverse_proxy	streaming error	{"upstream": "localhost:8080", "duration": 0.002212792, "request": {"remote_ip": "::1", "remote_port": "53151", "client_ip": "::1", "proto": "HTTP/1.1", "method": "GET", "host": "localhost:8081", "uri": "/ws", "headers": {"X-Forwarded-For": ["::1"], "X-Forwarded-Proto": ["http"], "X-Forwarded-Host": ["localhost:8081"], "User-Agent": ["Go-http-client/1.1"], "Connection": ["Upgrade"], "Sec-Websocket-Key": ["DImXB/WyhrxPxVPPbkRKpw=="], "Sec-Websocket-Version": ["13"], "Upgrade": ["websocket"]}}, "error": "rate: Wait(n=13) exceeds limiter's burst 9"}
```

### Full Log

```
❯ ./caddy run --config sample/Caddyfile
2024/08/29 02:23:41.747	INFO	using config from file	{"file": "sample/Caddyfile"}
2024/08/29 02:23:41.748	INFO	adapted config to JSON	{"adapter": "caddyfile"}
2024/08/29 02:23:41.750	INFO	admin	admin endpoint started	{"address": "localhost:2019", "enforce_origin": false, "origins": ["//localhost:2019", "//[::1]:2019", "//127.0.0.1:2019"]}
2024/08/29 02:23:41.750	INFO	tls.cache.maintenance	started background certificate maintenance	{"cache": "0x1400021e300"}
2024/08/29 02:23:41.750	DEBUG	http.auto_https	adjusted config	{"tls": {"automation":{"policies":[{}]}}, "http": {"servers":{"srv0":{"listen":[":8081"],"routes":[{"handle":[{"down_burst_limit":10000,"down_byte_rate":10000,"handler":"websocket_rate_limit","time_window":60,"up_burst_limit":100,"up_byte_rate":2},{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:8080"}]}]}],"automatic_https":{}}}}}
2024/08/29 02:23:41.750	INFO	http.handlers.websocket_rate_limit	Init WebSocketRateLimit	{"up_byte_rate": 2, "up_burst_limit": 100, "down_byte_rate": 10000, "down_burst_limit": 10000, "time_window": 60}
2024/08/29 02:23:41.751	DEBUG	http	starting server loop	{"address": "[::]:8081", "tls": false, "http3": false}
2024/08/29 02:23:41.751	INFO	http.log	server running	{"name": "srv0", "protocols": ["h1", "h2", "h3"]}
2024/08/29 02:23:41.752	INFO	autosaved config (load with --resume flag)	{"file": "/Users/neov/Library/Application Support/Caddy/autosave.json"}
2024/08/29 02:23:41.752	INFO	serving initial configuration
2024/08/29 02:23:41.765	INFO	tls	storage cleaning happened too recently; skipping for now	{"storage": "FileStorage:/Users/neov/Library/Application Support/Caddy", "instance": "3e5083b9-3d3a-463b-8e44-7b4464c67449", "try_again": "2024/08/30 02:23:41.765", "try_again_in": 86399.999999625}
2024/08/29 02:23:41.765	INFO	tls	finished cleaning storage units
2024/08/29 02:23:45.899	DEBUG	http.handlers.websocket_rate_limit	[+]isWebSocketRequest	{"up_byte_rate": 2, "up_burst_limit": 100, "down_byte_rate": 10000, "down_burst_limit": 10000}
2024/08/29 02:23:45.900	DEBUG	http.handlers.reverse_proxy	selected upstream	{"dial": "localhost:8080", "total_upstreams": 1}
2024/08/29 02:23:45.902	DEBUG	http.handlers.reverse_proxy	upstream roundtrip	{"upstream": "localhost:8080", "duration": 0.002212792, "request": {"remote_ip": "::1", "remote_port": "53151", "client_ip": "::1", "proto": "HTTP/1.1", "method": "GET", "host": "localhost:8081", "uri": "/ws", "headers": {"X-Forwarded-For": ["::1"], "X-Forwarded-Proto": ["http"], "X-Forwarded-Host": ["localhost:8081"], "User-Agent": ["Go-http-client/1.1"], "Connection": ["Upgrade"], "Sec-Websocket-Key": ["DImXB/WyhrxPxVPPbkRKpw=="], "Sec-Websocket-Version": ["13"], "Upgrade": ["websocket"]}}, "headers": {"Upgrade": ["websocket"], "Connection": ["Upgrade"], "Sec-Websocket-Accept": ["2V0r6IIiZ0UwFzLSPBBgvy3qqjg="]}, "status": 101}
2024/08/29 02:23:45.903	DEBUG	http.handlers.websocket_rate_limit	[+]Header	{"header": {"Server":["Caddy"]}}
2024/08/29 02:23:45.903	DEBUG	http.handlers.websocket_rate_limit	[+]WriteHeader	{"statusCode": 101}
2024/08/29 02:23:45.903	DEBUG	http.handlers.reverse_proxy	upgrading connection	{"upstream": "localhost:8080", "duration": 0.002212792, "request": {"remote_ip": "::1", "remote_port": "53151", "client_ip": "::1", "proto": "HTTP/1.1", "method": "GET", "host": "localhost:8081", "uri": "/ws", "headers": {"X-Forwarded-For": ["::1"], "X-Forwarded-Proto": ["http"], "X-Forwarded-Host": ["localhost:8081"], "User-Agent": ["Go-http-client/1.1"], "Connection": ["Upgrade"], "Sec-Websocket-Key": ["DImXB/WyhrxPxVPPbkRKpw=="], "Sec-Websocket-Version": ["13"], "Upgrade": ["websocket"]}}}
2024/08/29 02:23:45.903	DEBUG	http.handlers.websocket_rate_limit	[+]Hijack
2024/08/29 02:23:45.903	DEBUG	http.handlers.websocket_rate_limit	[+]Reading...	{"buffer_size": 32768, "enable_upload_limiter": true}
2024/08/29 02:23:46.904	DEBUG	http.handlers.websocket_rate_limit	[+]Read	{"n": 13}
2024/08/29 02:23:46.904	DEBUG	http.handlers.websocket_rate_limit	[+]Reading...	{"buffer_size": 32768, "enable_upload_limiter": true}
2024/08/29 02:23:46.905	DEBUG	http.handlers.websocket_rate_limit	[+]Writing...	{"buffer_size": 9, "download_limiter_enabled": true}
2024/08/29 02:23:46.905	DEBUG	http.handlers.websocket_rate_limit	[+]Write	{"n": 9}
2024/08/29 02:23:47.904	DEBUG	http.handlers.websocket_rate_limit	[+]Read	{"n": 13}
2024/08/29 02:23:47.904	DEBUG	http.handlers.websocket_rate_limit	[+]Reading...	{"buffer_size": 32768, "enable_upload_limiter": true}
2024/08/29 02:23:47.905	DEBUG	http.handlers.websocket_rate_limit	[+]Writing...	{"buffer_size": 9, "download_limiter_enabled": true}
2024/08/29 02:23:47.905	DEBUG	http.handlers.websocket_rate_limit	[+]Write	{"n": 9}
2024/08/29 02:23:48.905	DEBUG	http.handlers.websocket_rate_limit	[+]Read	{"n": 13}
2024/08/29 02:23:48.905	DEBUG	http.handlers.websocket_rate_limit	[+]Reading...	{"buffer_size": 32768, "enable_upload_limiter": true}
2024/08/29 02:23:48.905	DEBUG	http.handlers.websocket_rate_limit	[+]Writing...	{"buffer_size": 9, "download_limiter_enabled": true}
2024/08/29 02:23:48.905	DEBUG	http.handlers.websocket_rate_limit	[+]Write	{"n": 9}
2024/08/29 02:23:49.904	DEBUG	http.handlers.websocket_rate_limit	[+]Read	{"n": 13}
2024/08/29 02:23:49.904	DEBUG	http.handlers.websocket_rate_limit	[+]Reading...	{"buffer_size": 32768, "enable_upload_limiter": true}
2024/08/29 02:23:49.904	DEBUG	http.handlers.websocket_rate_limit	[+]Writing...	{"buffer_size": 9, "download_limiter_enabled": true}
2024/08/29 02:23:49.904	DEBUG	http.handlers.websocket_rate_limit	[+]Write	{"n": 9}
2024/08/29 02:23:50.904	DEBUG	http.handlers.websocket_rate_limit	[+]Read	{"n": 13}
2024/08/29 02:23:50.904	DEBUG	http.handlers.websocket_rate_limit	[+]Reading...	{"buffer_size": 32768, "enable_upload_limiter": true}
2024/08/29 02:23:50.905	DEBUG	http.handlers.websocket_rate_limit	[+]Writing...	{"buffer_size": 9, "download_limiter_enabled": true}
2024/08/29 02:23:50.905	DEBUG	http.handlers.websocket_rate_limit	[+]Write	{"n": 9}
2024/08/29 02:23:51.904	DEBUG	http.handlers.websocket_rate_limit	[+]Read	{"n": 13}
2024/08/29 02:23:51.904	DEBUG	http.handlers.websocket_rate_limit	[+]Reading...	{"buffer_size": 32768, "enable_upload_limiter": true}
2024/08/29 02:23:51.905	DEBUG	http.handlers.websocket_rate_limit	[+]Writing...	{"buffer_size": 9, "download_limiter_enabled": true}
2024/08/29 02:23:51.905	DEBUG	http.handlers.websocket_rate_limit	[+]Write	{"n": 9}
2024/08/29 02:23:52.903	DEBUG	http.handlers.websocket_rate_limit	[+]Read	{"n": 13}
2024/08/29 02:23:52.903	DEBUG	http.handlers.websocket_rate_limit	[+]Reading...	{"buffer_size": 32768, "enable_upload_limiter": true}
2024/08/29 02:23:52.904	DEBUG	http.handlers.websocket_rate_limit	[+]Writing...	{"buffer_size": 9, "download_limiter_enabled": true}
2024/08/29 02:23:52.904	DEBUG	http.handlers.websocket_rate_limit	[+]Write	{"n": 9}
2024/08/29 02:23:53.904	DEBUG	http.handlers.websocket_rate_limit	[+]Read	{"n": 13}
2024/08/29 02:23:53.904	DEBUG	http.handlers.reverse_proxy	streaming error	{"upstream": "localhost:8080", "duration": 0.002212792, "request": {"remote_ip": "::1", "remote_port": "53151", "client_ip": "::1", "proto": "HTTP/1.1", "method": "GET", "host": "localhost:8081", "uri": "/ws", "headers": {"X-Forwarded-For": ["::1"], "X-Forwarded-Proto": ["http"], "X-Forwarded-Host": ["localhost:8081"], "User-Agent": ["Go-http-client/1.1"], "Connection": ["Upgrade"], "Sec-Websocket-Key": ["DImXB/WyhrxPxVPPbkRKpw=="], "Sec-Websocket-Version": ["13"], "Upgrade": ["websocket"]}}, "error": "rate: Wait(n=13) exceeds limiter's burst 9"}
2024/08/29 02:23:53.904	DEBUG	http.handlers.websocket_rate_limit	[+]Close
2024/08/29 02:23:53.904	DEBUG	http.handlers.reverse_proxy	connection closed	{"upstream": "localhost:8080", "duration": 0.002212792, "request": {"remote_ip": "::1", "remote_port": "53151", "client_ip": "::1", "proto": "HTTP/1.1", "method": "GET", "host": "localhost:8081", "uri": "/ws", "headers": {"X-Forwarded-For": ["::1"], "X-Forwarded-Proto": ["http"], "X-Forwarded-Host": ["localhost:8081"], "User-Agent": ["Go-http-client/1.1"], "Connection": ["Upgrade"], "Sec-Websocket-Key": ["DImXB/WyhrxPxVPPbkRKpw=="], "Sec-Websocket-Version": ["13"], "Upgrade": ["websocket"]}}, "duration": 8.00145225}
```

