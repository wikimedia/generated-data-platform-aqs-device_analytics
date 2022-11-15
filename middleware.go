package main

import (
	"github.com/valyala/fasthttp"
)

// Set content type as application/json
func SetContentType(ctx *fasthttp.RequestCtx) bool {
	ctx.SetContentType("application/json")
	return true

}

// SecureHeadersMiddleware adds two basic security headers to each HTTP response
// X-XSS-Protection: 1; mode-block can help to prevent XSS attacks
// X-Frame-Options: deny can help to prevent clickjacking attacks
func SecureHeadersMiddleware(ctx *fasthttp.RequestCtx) bool {
	ctx.Response.Header.Set("X-XSS-Protection", "1; mode-block")
	ctx.Response.Header.Set("X-Frame-Options", "deny")
	return true

}
