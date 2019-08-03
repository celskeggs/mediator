package webclient

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

type fakeWriter struct {
	Headers   http.Header
	Collector bytes.Buffer
	Code      int
}

func (w *fakeWriter) Header() http.Header {
	return w.Headers
}

func (w *fakeWriter) Write(b []byte) (int, error) {
	return w.Collector.Write(b)
}

func (w *fakeWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
}

func TestCreateMux(t *testing.T) {
	sm, err := CreateMux(nil)
	assert.NoError(t, err)
	req, err := http.NewRequest("GET", "/style.css", nil)
	assert.NoError(t, err)
	wr := fakeWriter{
		Headers: make(http.Header),
	}
	if sm == nil {
		assert.Fail(t, "mux cannot be nil")
		return
	}
	sm.ServeHTTP(&wr, req)
	assert.Equal(t, 200, wr.Code)
	assert.Equal(t, "#test", strings.Split(wr.Collector.String(), " ")[0])
	assert.Equal(t, "text/css", strings.Split(wr.Headers["Content-Type"][0], ";")[0])
}
