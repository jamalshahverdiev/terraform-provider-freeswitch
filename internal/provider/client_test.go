package provider

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestClientInsecureTLS proves the `insecure` flag actually disables TLS
// verification, and that without it a self-signed endpoint is rejected (so the
// flag is doing real work, not a no-op).
func TestClientInsecureTLS(t *testing.T) {
	// httptest TLS server uses its own self-signed cert that no system CA trusts.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	ctx := context.Background()

	t.Run("insecure=true connects", func(t *testing.T) {
		c, err := NewClient(srv.URL, "tok", "", true)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if err := c.do(ctx, http.MethodGet, "/", nil, nil); err != nil {
			t.Fatalf("insecure=true should connect, got: %v", err)
		}
	})

	t.Run("insecure=false rejects self-signed", func(t *testing.T) {
		c, err := NewClient(srv.URL, "tok", "", false)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		err = c.do(ctx, http.MethodGet, "/", nil, nil)
		if err == nil {
			t.Fatal("insecure=false must reject an untrusted cert, but it connected")
		}
		if !strings.Contains(err.Error(), "x509") && !strings.Contains(err.Error(), "certificate") {
			t.Fatalf("expected a TLS verification error, got: %v", err)
		}
	})

	t.Run("ca_cert_file trusts the server", func(t *testing.T) {
		// Write the test server's own cert out and trust it explicitly.
		caPath := filepath.Join(t.TempDir(), "ca.crt")
		pem := certPEM(t, srv)
		if err := os.WriteFile(caPath, pem, 0o600); err != nil {
			t.Fatal(err)
		}
		c, err := NewClient(srv.URL, "tok", caPath, false)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if err := c.do(ctx, http.MethodGet, "/", nil, nil); err != nil {
			t.Fatalf("explicit CA should connect, got: %v", err)
		}
	})

	t.Run("bad ca_cert_file errors clearly", func(t *testing.T) {
		bad := filepath.Join(t.TempDir(), "bad.crt")
		_ = os.WriteFile(bad, []byte("not a cert"), 0o600)
		if _, err := NewClient(srv.URL, "tok", bad, false); err == nil {
			t.Fatal("expected an error for a CA file with no certs")
		}
	})
}

func certPEM(t *testing.T, srv *httptest.Server) []byte {
	t.Helper()
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
}
