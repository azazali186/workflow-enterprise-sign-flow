// Package httpx adapts net/http handlers (swagger, prometheus) to Hertz.
package httpx

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/cloudwego/hertz/pkg/app"
)

// writer bridges an http.ResponseWriter into the Hertz response buffer.
type writer struct {
	status int
	header http.Header
	body   bytes.Buffer
}

func (w *writer) Header() http.Header         { return w.header }
func (w *writer) Write(b []byte) (int, error) { return w.body.Write(b) }
func (w *writer) WriteHeader(s int)           { w.status = s }

// Adapt wraps a net/http handler as a Hertz handler.
func Adapt(h http.Handler) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path, err := url.PathUnescape(string(c.Request.URI().Path()))
		if err != nil {
			path = string(c.Request.URI().Path())
		}
		query := string(c.Request.URI().QueryString())
		requestURI := path
		if query != "" {
			requestURI += "?" + query
		}
		req, err := http.NewRequestWithContext(
			ctx,
			string(c.Request.Method()),
			requestURI,
			bytes.NewReader(c.Request.Body()),
		)
		if err != nil {
			c.SetStatusCode(http.StatusBadRequest)
			c.SetBodyString("bad request")
			return
		}
		// http.NewRequest leaves RequestURI empty for relative URLs; net/http
		// handlers (http-swagger parses it) rely on it being populated.
		req.RequestURI = requestURI
		req.Header = make(http.Header)
		c.Request.Header.VisitAll(func(k, v []byte) { req.Header.Add(string(k), string(v)) })
		req.Host = string(c.Request.Host())
		rw := &writer{status: http.StatusOK, header: make(http.Header)}
		h.ServeHTTP(rw, req)
		c.Response.SetStatusCode(rw.status)
		for k, vs := range rw.header {
			for _, v := range vs {
				c.Response.Header.Add(k, v)
			}
		}
		c.Response.SetBody(rw.body.Bytes())
	}
}
