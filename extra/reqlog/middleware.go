package reqlog

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/fatih/color"
	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/otel/trace"

	"github.com/golangwebLK/webrouter"
)

type middleware struct {
	enabled bool
	verbose bool
}

type Option func(m *middleware)

// WithEnabled enables/disables the middleware.
func WithEnabled(on bool) Option {
	return func(m *middleware) {
		m.enabled = on
	}
}

// WithVerbose configures the middleware to log all requests.
func WithVerbose(on bool) Option {
	return func(m *middleware) {
		m.verbose = on
	}
}

// WithEnv configures the middleware using the environment variable value.
// For example, WithEnv("BUNDEBUG"):
//   - BUNDEBUG=0 - disables the middleware.
//   - BUNDEBUG=1 - enables the middleware.
//   - BUNDEBUG=2 - enables the middleware and verbose mode.
func FromEnv(keys ...string) Option {
	if len(keys) == 0 {
		keys = []string{"BUNDEBUG"}
	}
	return func(m *middleware) {
		for _, key := range keys {
			if env, ok := os.LookupEnv(key); ok {
				m.enabled = env != "" && env != "0"
				m.verbose = env == "2"
				break
			}
		}
	}
}

func NewMiddleware(opts ...Option) bunlk.MiddlewareFunc {
	c := &middleware{
		enabled: true,
		verbose: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.Middleware
}

func (m *middleware) Middleware(next bunlk.HandlerFunc) bunlk.HandlerFunc {
	if !m.enabled {
		return next
	}

	return func(w http.ResponseWriter, req bunlk.Request) error {
		rec := NewResponseWriter(w)

		now := time.Now()
		err := next(rec.Wrapped, req)
		dur := time.Since(now)
		statusCode := rec.StatusCode()

		if !m.verbose && statusCode >= 200 && statusCode < 300 && err == nil {
			return nil
		}

		args := make([]interface{}, 0, 10)
		args = append(args,
			"",
			now.Format(" 15:04:05.000 "),
		)

		if spanCtx := trace.SpanContextFromContext(req.Context()); spanCtx.IsValid() {
			args = append(args, spanCtx.TraceID().String()+" ")
		}

		args = append(args,
			formatStatus(statusCode),
			fmt.Sprintf(" %10s ", dur.Round(time.Microsecond)),
			formatMethod(req.Method),
			req.URL.String(),
		)

		if err != nil {
			typ := reflect.TypeOf(err).String()
			args = append(args,
				"\t",
				color.New(color.BgRed).Sprintf(" %s ", typ+": "+err.Error()),
			)
		}

		log.Println(args...)

		return err
	}
}

//------------------------------------------------------------------------------

type ResponseWriter struct {
	Wrapped    http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	var rw ResponseWriter
	rw.Wrapped = httpsnoop.Wrap(w, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				if rw.statusCode == 0 {
					rw.statusCode = statusCode
				}
				next(statusCode)
			}
		},
	})
	return &rw
}

func (w *ResponseWriter) StatusCode() int {
	if w.statusCode != 0 {
		return w.statusCode
	}
	return http.StatusOK
}

//------------------------------------------------------------------------------

func formatStatus(code int) string {
	return statusColor(code).Sprintf(" %d ", code)
}

func statusColor(code int) *color.Color {
	switch {
	case code >= 200 && code < 300:
		return color.New(color.BgGreen, color.FgHiWhite)
	case code >= 300 && code < 400:
		return color.New(color.BgWhite, color.FgHiBlack)
	case code >= 400 && code < 500:
		return color.New(color.BgYellow, color.FgHiBlack)
	default:
		return color.New(color.BgRed, color.FgHiWhite)
	}
}

func formatMethod(method string) string {
	return methodColor(method).Sprintf(" %-7s ", method)
}

func methodColor(method string) *color.Color {
	switch method {
	case http.MethodGet:
		return color.New(color.BgBlue, color.FgHiWhite)
	case http.MethodPost:
		return color.New(color.BgGreen, color.FgHiWhite)
	case http.MethodPut:
		return color.New(color.BgYellow, color.FgHiBlack)
	case http.MethodDelete:
		return color.New(color.BgRed, color.FgHiWhite)
	case http.MethodPatch:
		return color.New(color.BgCyan, color.FgHiWhite)
	case http.MethodHead:
		return color.New(color.BgMagenta, color.FgHiWhite)
	default:
		return color.New(color.BgWhite, color.FgHiBlack)
	}
}
