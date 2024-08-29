package websocketratelimit

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func init() {
	caddy.RegisterModule(WebSocketRateLimit{})
}

type WebSocketRateLimit struct {
	UpByteRate     int64 `json:"up_byte_rate,omitempty"`
	DownByteRate   int64 `json:"down_byte_rate,omitempty"`
	UpBurstLimit   int64 `json:"up_burst_limit,omitempty"`
	DownBurstLimit int64 `json:"down_burst_limit,omitempty"`
	TimeWindow     int64 `json:"time_window,omitempty"`
	connections    sync.Map
	logger         *zap.Logger
}

type connectionInfo struct {
	uploadLimiter   *rate.Limiter
	downloadLimiter *rate.Limiter
}

func (WebSocketRateLimit) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.websocket_rate_limit",
		New: func() caddy.Module { return new(WebSocketRateLimit) },
	}
}

// ServeHTTP implements the caddyhttp.MiddlewareHandler interface.
// func (b BidirectionalRateLimitedWebSocketProxy) ServeHTTP(w http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
func (m *WebSocketRateLimit) ServeHTTP(w http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	if !isWebSocketRequest(req) {
		return next.ServeHTTP(w, req)
	}

	m.logger.Debug("[+]isWebSocketRequest", zap.Int64("up_byte_rate", m.UpByteRate), zap.Int64("up_burst_limit", m.UpBurstLimit), zap.Int64("down_byte_rate", m.DownByteRate), zap.Int64("down_burst_limit", m.DownBurstLimit))
	// Create a rate-limited connection
	limitedResponseWriter := &limitedResponseWriter{
		rw:              w,
		uploadLimiter:   nil,
		downloadLimiter: nil,
		ctx:             req.Context(),
		logger:          m.logger,
	}
	if m.isUpLimitEnable() {
		limitedResponseWriter.uploadLimiter = rate.NewLimiter(rate.Limit(m.UpByteRate/m.TimeWindow), int(m.UpBurstLimit))
	}
	if m.isDownLimitEnable() {
		limitedResponseWriter.downloadLimiter = rate.NewLimiter(rate.Limit(m.DownByteRate/m.TimeWindow), int(m.DownBurstLimit))
	}

	err := next.ServeHTTP(limitedResponseWriter, req)

	return err
}

type limitedResponseWriter struct {
	rw              http.ResponseWriter
	uploadLimiter   *rate.Limiter
	downloadLimiter *rate.Limiter
	ctx             context.Context
	logger          *zap.Logger
}

// Write implements http.ResponseWriter.
func (r *limitedResponseWriter) Write(buf []byte) (int, error) {
	r.logger.Debug("[+]Writing...", zap.Int("buffer_size", len(buf)))
	n, err := r.rw.Write(buf)
	r.logger.Debug("[+]Write", zap.Int("n", n))
	return n, err
}

// Header implements http.ResponseWriter.
func (r *limitedResponseWriter) Header() http.Header {
	r.logger.Debug("[+]Header", zap.Any("header", r.rw.Header()))
	return r.rw.Header()
}

// WriteHeader implements http.ResponseWriter.
func (r *limitedResponseWriter) WriteHeader(statusCode int) {
	r.logger.Debug("[+]WriteHeader", zap.Int("statusCode", statusCode))
	r.rw.WriteHeader(statusCode)
}

func (r *limitedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.logger.Debug("[+]Hijack")
	hijacker := http.NewResponseController(r.rw)
	conn, brw, hijackErr := hijacker.Hijack()
	if hijackErr != nil {
		return nil, nil, hijackErr
	}

	limitConn := &bidirectionalRateLimitedConn{
		conn:            conn,
		uploadLimiter:   r.uploadLimiter,
		downloadLimiter: r.downloadLimiter,
		ctx:             r.ctx,
		logger:          r.logger,
	}

	return limitConn, brw, nil
}

type bidirectionalRateLimitedConn struct {
	conn            net.Conn
	uploadLimiter   *rate.Limiter
	downloadLimiter *rate.Limiter
	ctx             context.Context
	logger          *zap.Logger
}

// Write implements net.Conn.
func (b *bidirectionalRateLimitedConn) Write(buf []byte) (n int, err error) {
	b.logger.Debug("[+]Writing...",
		zap.Int("buffer_size", len(buf)),
		zap.Any("download_limiter_enabled", b.downloadLimiter != nil),
	)

	if b.downloadLimiter != nil {
		if err := b.downloadLimiter.WaitN(b.ctx, len(buf)); err != nil {
			return 0, err
		}
	}
	n, err = b.conn.Write(buf)
	b.logger.Debug("[+]Write", zap.Int("n", n))
	return n, err
}

// Read implements net.Conn.
func (b *bidirectionalRateLimitedConn) Read(buf []byte) (n int, err error) {
	b.logger.Debug("[+]Reading...",
		zap.Int("buffer_size", len(buf)),
		zap.Any("enable_upload_limiter", b.uploadLimiter != nil),
	)
	n, err = b.conn.Read(buf)
	b.logger.Debug("[+]Read", zap.Int("n", n))
	if err != nil {
		return n, err
	}
	if b.uploadLimiter != nil {
		err = b.uploadLimiter.WaitN(b.ctx, n)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

// Close implements net.Conn.
func (b *bidirectionalRateLimitedConn) Close() error {
	b.logger.Debug("[+]Close")
	return b.conn.Close()
}

// LocalAddr implements net.Conn.
func (b *bidirectionalRateLimitedConn) LocalAddr() net.Addr {
	b.logger.Debug("[+]LocalAddr")
	return b.conn.LocalAddr()
}

// RemoteAddr implements net.Conn.
func (b *bidirectionalRateLimitedConn) RemoteAddr() net.Addr {
	b.logger.Debug("[+]RemoteAddr")
	return b.conn.RemoteAddr()
}

// SetDeadline implements net.Conn.
func (b *bidirectionalRateLimitedConn) SetDeadline(t time.Time) error {
	b.logger.Debug("[+]SetDeadline", zap.Time("t", t))
	return b.conn.SetDeadline(t)
}

// SetReadDeadline implements net.Conn.
func (b *bidirectionalRateLimitedConn) SetReadDeadline(t time.Time) error {
	b.logger.Debug("[+]SetReadDeadline", zap.Time("t", t))
	return b.conn.SetReadDeadline(t)
}

// SetWriteDeadline implements net.Conn.
func (b *bidirectionalRateLimitedConn) SetWriteDeadline(t time.Time) error {
	b.logger.Debug("[+]SetWriteDeadline", zap.Time("t", t))
	return b.conn.SetWriteDeadline(t)
}

func isWebSocketRequest(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket" && r.Header.Get("Connection") == "Upgrade"
}

func (m *WebSocketRateLimit) isUpLimitEnable() bool {
	return m.UpBurstLimit > 0
}
func (m *WebSocketRateLimit) isDownLimitEnable() bool {
	return m.DownBurstLimit > 0
}
