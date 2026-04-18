package handlers

import (
	"bytes"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// wantsMarkdown reports whether the request's Accept header prefers text/markdown.
// Per the "Markdown for Agents" spec, HTML remains the default; markdown is
// returned only when explicitly requested.
func wantsMarkdown(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	for _, part := range strings.Split(accept, ",") {
		mt := strings.TrimSpace(part)
		if i := strings.Index(mt, ";"); i >= 0 {
			mt = strings.TrimSpace(mt[:i])
		}
		if strings.EqualFold(mt, "text/markdown") {
			return true
		}
	}
	return false
}

// markdownCaptureWriter buffers an HTML response so it can be converted to
// markdown after the handler completes. Status and non-Content-Type headers
// pass through to the underlying ResponseWriter; the body is withheld until
// Flush converts and writes it.
type markdownCaptureWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	status      int
	wroteHeader bool
}

func (w *markdownCaptureWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
	w.ResponseWriter.Header().Del("Content-Length")
}

func (w *markdownCaptureWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.buf.Write(b)
}

// approximateTokens returns a rough token count. A common heuristic is
// ~4 characters per token for English prose.
func approximateTokens(s string) int {
	if s == "" {
		return 0
	}
	n := len([]rune(s)) / 4
	if n < 1 {
		n = 1
	}
	return n
}

// MarkdownNegotiationMiddleware converts HTML responses to markdown when the
// client sends Accept: text/markdown. Other requests pass through unchanged.
// Responds with Content-Type: text/markdown and x-markdown-tokens for a rough
// token estimate of the converted body.
func MarkdownNegotiationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !wantsMarkdown(r) {
			next.ServeHTTP(w, r)
			return
		}

		cap := &markdownCaptureWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(cap, r)

		ct := w.Header().Get("Content-Type")
		if ct != "" && !strings.Contains(ct, "text/html") {
			// Not HTML (e.g. an error page emitted via http.Error as text/plain);
			// forward the buffered bytes untouched.
			if _, err := w.Write(cap.buf.Bytes()); err != nil {
				slog.Error("failed to write passthrough body", "error", err)
			}
			return
		}

		md, err := htmltomarkdown.ConvertString(cap.buf.String())
		if err != nil {
			slog.Error("failed to convert HTML to markdown", "error", err)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if _, werr := w.Write(cap.buf.Bytes()); werr != nil {
				slog.Error("failed to write fallback HTML body", "error", werr)
			}
			return
		}

		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("X-Markdown-Tokens", strconv.Itoa(approximateTokens(md)))
		w.Header().Set("Vary", appendVary(w.Header().Get("Vary"), "Accept"))
		if _, err := w.Write([]byte(md)); err != nil {
			slog.Error("failed to write markdown body", "error", err)
		}
	})
}

// appendVary adds a field to an existing Vary header value if missing.
func appendVary(existing, field string) string {
	if existing == "" {
		return field
	}
	for _, part := range strings.Split(existing, ",") {
		if strings.EqualFold(strings.TrimSpace(part), field) {
			return existing
		}
	}
	return existing + ", " + field
}
