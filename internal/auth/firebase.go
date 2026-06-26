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
)

const firebaseCertURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

var (
	ErrMissingProjectID = errors.New("firebase project id is required")
	ErrInvalidClaims    = errors.New("invalid firebase token claims")
)

type FirebaseVerifier struct {
	projectID string
	certURL   string
	client    *http.Client
	now       func() time.Time

	mu      sync.Mutex
	certs   map[string]*rsa.PublicKey
	expires time.Time
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
	v.mu.Lock()
	key, ok := v.certs[kid]
	fresh := v.now().Before(v.expires)
	v.mu.Unlock()

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

	if err := v.refreshCerts(ctx); err != nil {
		return nil, err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	key, ok = v.certs[kid]
	if !ok {
		return nil, fmt.Errorf("%w: unknown kid", ErrInvalidToken)
	}
	return key, nil
}

func (v *FirebaseVerifier) refreshCerts(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.certURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch firebase certs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch firebase certs: status %d", resp.StatusCode)
	}

	var pemCerts map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&pemCerts); err != nil {
		return fmt.Errorf("decode firebase certs: %w", err)
	}

	certs := make(map[string]*rsa.PublicKey, len(pemCerts))
	for kid, certPEM := range pemCerts {
		block, _ := pem.Decode([]byte(certPEM))
		if block == nil {
			return fmt.Errorf("decode firebase cert %s: missing pem block", kid)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("parse firebase cert %s: %w", kid, err)
		}
		key, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("parse firebase cert %s: public key is not rsa", kid)
		}
		certs[kid] = key
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.certs = certs
	v.expires = v.now().Add(cacheMaxAge(resp.Header.Get("Cache-Control")))
	return nil
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
