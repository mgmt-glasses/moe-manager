package auth

import (
	"context"
	"strings"
)

// BypassVerifier はローカル開発・テスト専用の TokenVerifier。
//
// ID トークンの署名検証を一切行わず、受け取った Bearer トークン文字列を
// そのまま UID として扱う。フロントが Firebase ログインを実装する前でも、
// `Authorization: Bearer <userId>` を送るだけで認証必須 API を叩けるようにする。
//
// RequirePathUser はこの UID とパスの {userId} を照合するため、トークンに
// userId を載せれば従来どおり所有権チェックも機能する。
//
// 環境変数 AUTH_BYPASS=true のときだけ FirebaseVerifier の代わりに注入される。
// 本番では AUTH_BYPASS を設定せず、必ず FirebaseVerifier を使うこと。
type BypassVerifier struct{}

// maxBypassUIDLen は uid として受け付ける最大長。FirebaseVerifier が
// subject に課す上限（128）に揃え、長大 uid による DB 側の想定外を避ける。
const maxBypassUIDLen = 128

// VerifyIDToken はトークン文字列を UID とみなして返す。
// 空トークンと上限超過の uid を拒否する。
func (BypassVerifier) VerifyIDToken(_ context.Context, rawToken string) (User, error) {
	uid := strings.TrimSpace(rawToken)
	if uid == "" || len(uid) > maxBypassUIDLen {
		return User{}, ErrInvalidToken
	}
	return User{UID: uid}, nil
}
