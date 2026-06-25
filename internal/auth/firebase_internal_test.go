package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newTestCertServer は kid -> PEM 証明書を返す httptest サーバーを立て、
// 受け取った GET リクエスト数を atomic カウンタで記録する。
func newTestCertServer(t *testing.T, kid string, key *rsa.PrivateKey) (*httptest.Server, *int64) {
	t.Helper()

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Unix(0, 0),
		NotAfter:     time.Unix(1<<31, 0),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	var calls int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{kid: string(certPEM)})
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

// キャッシュが有効な間は、未知の kid に対して証明書エンドポイントへ
// 再取得しに行かないことを保証する（未認証リクエストによる増幅を防ぐ）。
func TestPublicKeyDoesNotRefreshForUnknownKidWhileCacheFresh(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv, calls := newTestCertServer(t, "kid1", key)

	now := time.Unix(1000, 0)
	v := &FirebaseVerifier{
		projectID: "test-project",
		certURL:   srv.URL,
		client:    srv.Client(),
		now:       func() time.Time { return now },
		certs:     map[string]*rsa.PublicKey{},
	}
	ctx := context.Background()

	// 1 回目: キャッシュ未取得（expires はゼロ値）なので refresh が走る。
	if _, err := v.publicKey(ctx, "unknown"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for unknown kid, got %v", err)
	}
	if got := atomic.LoadInt64(calls); got != 1 {
		t.Fatalf("expected 1 cert fetch after first call, got %d", got)
	}

	// 2 回目: キャッシュは有効。未知の kid でも refresh してはいけない。
	if _, err := v.publicKey(ctx, "unknown"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for unknown kid, got %v", err)
	}
	if got := atomic.LoadInt64(calls); got != 1 {
		t.Fatalf("unknown kid triggered a cert refresh while cache fresh: %d fetches", got)
	}

	// 既知の kid はキャッシュから解決でき、追加の fetch を発生させない。
	if _, err := v.publicKey(ctx, "kid1"); err != nil {
		t.Fatalf("expected known kid to resolve, got %v", err)
	}
	if got := atomic.LoadInt64(calls); got != 1 {
		t.Fatalf("known kid triggered a cert refresh while cache fresh: %d fetches", got)
	}
}

// キャッシュ期限切れ後は、未知の kid で一度だけ refresh して解決を試みる。
func TestPublicKeyRefreshesAfterCacheExpiry(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv, calls := newTestCertServer(t, "kid1", key)

	now := time.Unix(1000, 0)
	v := &FirebaseVerifier{
		projectID: "test-project",
		certURL:   srv.URL,
		client:    srv.Client(),
		now:       func() time.Time { return now },
		certs:     map[string]*rsa.PublicKey{},
	}
	ctx := context.Background()

	if _, err := v.publicKey(ctx, "kid1"); err != nil {
		t.Fatalf("first lookup failed: %v", err)
	}
	if got := atomic.LoadInt64(calls); got != 1 {
		t.Fatalf("expected 1 fetch, got %d", got)
	}

	// キャッシュ期限（max-age=3600）を越えて時刻を進める。
	now = now.Add(2 * time.Hour)
	if _, err := v.publicKey(ctx, "kid1"); err != nil {
		t.Fatalf("lookup after expiry failed: %v", err)
	}
	if got := atomic.LoadInt64(calls); got != 2 {
		t.Fatalf("expected refresh after expiry (2 fetches), got %d", got)
	}
}
