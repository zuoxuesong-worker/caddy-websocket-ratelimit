

# caddy-websocket-ratelimit

A websocket rate limiting module for Caddy 2


## 介绍

一直以来，caddy 下都没有关于 websocket 的流量控制插件，但是在实际使用中，websocket 的流量控制是很有必要的，尤其是在 websocket 作为 api 网关的场景下。

现在，这个插件可以让你在 caddy 中方便地控制 websocket 的流量。

我们包装了 bidirectionalRateLimitedConn 结构，它包装了原始的 net.Conn，它的作用是基于令牌桶的流量控制算法对 websocket 的读写进行控制。

另外，我们创建了一个新的 `limitedResponseWriter` 结构，它包装了原始的 `http.ResponseWriter`，当其他服务调用 `http.NewResponseController(w).Hijack()` 以劫持连接时，我们返回 `bidirectionalRateLimitedConn` 结构，这样就可以对 websocket 的读写进行控制了。

上述做法在与 Caddy 的原始实现更加一致。这个实现现在更接近 Caddy 的原始 WebSocket 代理实现，同时还保留了我们添加的速率限制功能。它确保了所有的缓冲数据都被正确处理，这对于某些 WebSocket 握手场景是很重要的。


要使用这个更新后的模块，你的 Caddyfile 配置将保持不变：

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
 - <up_burst_limit>: The request rate limit specified in requests per time window.
 - <down_byte_rate>: The request rate limit specified in requests per time window.
 - <down_burst_limit>: The request rate limit specified in requests per time window.
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

注意：此配置中日志级别为 debug，在生产环境中请勿使用 debug 级别日志



4. Test Rate Limit

Start a websocket client to connect to the websocket server, and send messages to the server.

``` bash
$ curl -i http://localhost:8081
```

运行前期一直在消耗令牌，直到令牌桶为空。

大约运行10s 左右，此时caddy 读取或者写入数据时，会返回 rate: Wait(...) exceeds limiter's burst ... 的错误。

在 caddy 的调试级别日志，可以看到如下日志：

```
024/08/29 02:23:53.904	DEBUG	http.handlers.reverse_proxy	streaming error	{"upstream": "localhost:8080", "duration": 0.002212792, "request": {"remote_ip": "::1", "remote_port": "53151", "client_ip": "::1", "proto": "HTTP/1.1", "method": "GET", "host": "localhost:8081", "uri": "/ws", "headers": {"X-Forwarded-For": ["::1"], "X-Forwarded-Proto": ["http"], "X-Forwarded-Host": ["localhost:8081"], "User-Agent": ["Go-http-client/1.1"], "Connection": ["Upgrade"], "Sec-Websocket-Key": ["DImXB/WyhrxPxVPPbkRKpw=="], "Sec-Websocket-Version": ["13"], "Upgrade": ["websocket"]}}, "error": "rate: Wait(n=13) exceeds limiter's burst 9"}
```


### 详细日志

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


## 参考

1. https://github.com/RussellLuo/caddy-ext/blob/master/ratelimit/caddyfile.go
2. https://github.com/caddyserver/caddy/blob/master/modules/caddyhttp/ratelimit/caddyfile.go