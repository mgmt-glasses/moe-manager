package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/sync/singleflight"
)

const firebaseCertURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

// refreshRetryGrace は cert 取得に失敗したとき、既存 cert を stale のまま使い続ける猶予。
// この間は再取得を間引き、cert エンドポイントの一時障害で全リクエストが
// 連続再試行・全ログイン不能になるのを防ぐ。
const refreshRetryGrace = time.Minute

var (
	ErrMissingProjectID = errors.New("firebase project id is required")
	ErrInvalidClaims    = errors.New("invalid firebase token claims")
)

type FirebaseVerifier struct {
	projectID string
	certURL   string
	client    *http.Client
	now       func() time.Time

	// sf は期限切れ時の cert 再取得を 1 本に集約する（thundering herd 防止）。
	sf singleflight.Group

	mu        sync.RWMutex
	certs     map[string]*rsa.PublicKey
	prevCerts map[string]*rsa.PublicKey // 直前世代。ローテーション overlap 中の旧 kid を引けるよう保持。
	expires   time.Time
}

func NewFirebaseVerifier(projectID string) (*FirebaseVerifier, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, ErrMissingProjectID
	}
	return &FirebaseVerifier{
		projectID: projectID,
		certURL:   firebaseCertURL,
		client:    &http.Client{Timeout: 10 * time.Second},
		now:       time.Now,
		certs:     map[string]*rsa.PublicKey{},
	}, nil
}

type firebaseClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func (v *FirebaseVerifier) VerifyIDToken(ctx context.Context, rawToken string) (User, error) {
	claims := &firebaseClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("%w: unexpected signing method", ErrInvalidToken)
		}
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("%w: missing kid", ErrInvalidToken)
		}
		return v.publicKey(ctx, kid)
	})
	if err != nil {
		return User{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !token.Valid {
		return User{}, ErrInvalidToken
	}
	if !claims.VerifyIssuer("https://securetoken.google.com/"+v.projectID, true) {
		return User{}, fmt.Errorf("%w: issuer", ErrInvalidClaims)
	}
	if !claims.VerifyAudience(v.projectID, true) {
		return User{}, fmt.Errorf("%w: audience", ErrInvalidClaims)
	}
	if claims.Subject == "" || len(claims.Subject) > 128 {
		return User{}, fmt.Errorf("%w: subject", ErrInvalidClaims)
	}
	// jwt/v4 の Valid() は exp 不在を許容する（VerifyExpiresAt の required=false）。
	// ID トークン検証の標準要件として、exp を持たないトークンは明示的に弾く。
	if claims.ExpiresAt == nil {
		return User{}, fmt.Errorf("%w: missing exp", ErrInvalidClaims)
	}

	return User{UID: claims.Subject, Email: claims.Email}, nil
}

func (v *FirebaseVerifier) publicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.lookupLocked(kid)
	fresh := v.now().Before(v.expires)
	v.mu.RUnlock()

	if ok && fresh {
		return key, nil
	}
	// キャッシュがまだ有効な間は、未知の kid を refresh せず即弾く。
	// 認証前の公開エンドポイントなので、攻撃者がランダムな kid を投げて
	// Google 証明書エンドポイントへの外部 HTTP を毎回誘発できないようにする。
	// Firebase の証明書は Cache-Control の max-age 内ではローテーションされない。
	if fresh {
		return nil, fmt.Errorf("%w: unknown kid", ErrInvalidToken)
	}

	// 期限切れ。並行リクエストは singleflight で 1 本の再取得に集約する。
	_, err, _ := v.sf.Do("refresh", func() (any, error) {
		return nil, v.refreshCerts(ctx)
	})

	v.mu.RLock()
	defer v.mu.RUnlock()
	if key, ok := v.lookupLocked(kid); ok {
		// 再取得が失敗していても、既存（stale）cert で kid を引けるなら通す。
		// Firebase は証明書をローテーション重複期間を設けて切り替えるため、
		// max-age 直後の stale cert でも正規トークンは検証できる。
		return key, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("%w: unknown kid", ErrInvalidToken)
}

// lookupLocked は現世代→直前世代の順に kid を引く。呼び出し側で RLock/Lock 済みであること。
func (v *FirebaseVerifier) lookupLocked(kid string) (*rsa.PublicKey, bool) {
	if k, ok := v.certs[kid]; ok {
		return k, true
	}
	k, ok := v.prevCerts[kid]
	return k, ok
}

func (v *FirebaseVerifier) refreshCerts(ctx context.Context) error {
	certs, maxAge, err := v.fetchCerts(ctx)
	if err != nil {
		v.mu.Lock()
		if len(v.certs) > 0 {
			// 既存 cert があるうちは stale 継続。短い猶予で再取得を間引き、
			// cert エンドポイント障害中の連続再試行・全ログイン不能を防ぐ。
			v.expires = v.now().Add(refreshRetryGrace)
		}
		v.mu.Unlock()
		return err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.prevCerts = v.certs // ローテーション overlap 用に直前世代を保持
	v.certs = certs
	v.expires = v.now().Add(maxAge)
	return nil
}

// fetchCerts は Firebase の公開証明書を取得・パースする副作用のない処理。
func (v *FirebaseVerifier) fetchCerts(ctx context.Context) (map[string]*rsa.PublicKey, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.certURL, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch firebase certs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("fetch firebase certs: status %d", resp.StatusCode)
	}

	var pemCerts map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&pemCerts); err != nil {
		return nil, 0, fmt.Errorf("decode firebase certs: %w", err)
	}

	certs := make(map[string]*rsa.PublicKey, len(pemCerts))
	for kid, certPEM := range pemCerts {
		block, _ := pem.Decode([]byte(certPEM))
		if block == nil {
			return nil, 0, fmt.Errorf("decode firebase cert %s: missing pem block", kid)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, 0, fmt.Errorf("parse firebase cert %s: %w", kid, err)
		}
		key, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, 0, fmt.Errorf("parse firebase cert %s: public key is not rsa", kid)
		}
		certs[kid] = key
	}

	return certs, cacheMaxAge(resp.Header.Get("Cache-Control")), nil
}

func cacheMaxAge(header string) time.Duration {
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if value, ok := strings.CutPrefix(part, "max-age="); ok {
			seconds, err := strconv.Atoi(value)
			if err == nil && seconds > 0 {
				return time.Duration(seconds) * time.Second
			}
		}
	}
	return time.Hour
}
