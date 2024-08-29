package websocketratelimit

import (
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	httpcaddyfile.RegisterHandlerDirective("websocket_rate_limit", parseCaddyfile)
}

// parseCaddyfile sets up a handler for rate-limiting from Caddyfile tokens. Syntax:
//
//     websocket_rate_limit
//         up_byte_rate <up_byte_rate>
//         up_burst_limit <up_burst_limit>
//         down_byte_rate <down_byte_rate>
//         down_burst_limit <down_burst_limit>
//         time_window <time_window>
// OR
//     websocket_rate_limit {
//         up_byte_rate <up_byte_rate>
//         up_burst_limit <up_burst_limit>
//         down_byte_rate <down_byte_rate>
//         down_burst_limit <down_burst_limit>
//         time_window <time_window>
//     }
//
// Parameters:
// - <up_byte_rate>: The request rate limit  specified in requests per time window.
// - <up_burst_limit>: The request rate limit specified in requests per time window.
// - <down_byte_rate>: The request rate limit specified in requests per time window.
// - <down_burst_limit>: The request rate limit specified in requests per time window.
// - <time_window>: The time window (in seconds) for the rate limit. Default is 60 seconds.

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	rl := new(WebSocketRateLimit)
	if err := rl.UnmarshalCaddyfile(h.Dispenser); err != nil {
		return nil, err
	}
	return rl, nil
}

func (m *WebSocketRateLimit) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	var (
		arg string
		err error
	)

	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "up_byte_rate":
				if !d.Args(&arg) {
					return d.ArgErr() // 错误处理
				}
				m.UpByteRate, err = strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return d.Errf("invalid up byte rate value '%s': %v", arg, err)
				}
			case "up_burst_limit":
				if !d.Args(&arg) {
					return d.ArgErr()
				}
				m.UpBurstLimit, err = strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return d.Errf("invalid up burst limit value '%s': %v", arg, err)
				}
			case "down_byte_rate":
				if !d.Args(&arg) {
					return d.ArgErr()
				}
				m.DownByteRate, err = strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return d.Errf("invalid down byte rate value '%s': %v", arg, err)
				}
			case "down_burst_limit":
				if !d.Args(&arg) {
					return d.ArgErr()
				}
				m.DownBurstLimit, err = strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return d.Errf("invalid down burst limit value '%s': %v", arg, err)
				}
			case "time_window":
				if !d.Args(&arg) {
					return d.ArgErr()
				}
				m.TimeWindow, err = strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return d.Errf("invalid time window value '%s': %v", arg, err)
				}
			default:
				return d.Errf("Unknown parameter: %s", d.Val())
			}
		}
	}
	return nil
}

func (m *WebSocketRateLimit) Provision(ctx caddy.Context) error {
	if m.TimeWindow == 0 {
		m.TimeWindow = 60
	}
	if m.UpBurstLimit == 0 {
		m.UpBurstLimit = 0
		m.UpByteRate = 0
		m.logger.Info("upstream rate limit is disable")
	}
	if m.DownBurstLimit == 0 {
		m.DownBurstLimit = 0
		m.DownByteRate = 0
		m.logger.Info("downstream rate limit is disable")
	}
	m.logger = ctx.Logger(m)
	m.logger.Info("Init WebSocketRateLimit",
		zap.Int64("up_byte_rate", m.UpByteRate),
		zap.Int64("up_burst_limit", m.UpBurstLimit),
		zap.Int64("down_byte_rate", m.DownByteRate),
		zap.Int64("down_burst_limit", m.DownBurstLimit),
		zap.Int64("time_window", m.TimeWindow),
	)
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*WebSocketRateLimit)(nil)
)
