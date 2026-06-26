package main

import "testing"

func TestAuthBypassDecision(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantBypass bool
		wantFatal  bool
	}{
		{
			name:       "未設定はバイパスしない",
			env:        map[string]string{},
			wantBypass: false,
			wantFatal:  false,
		},
		{
			name:       "AUTH_BYPASS=false はバイパスしない",
			env:        map[string]string{"AUTH_BYPASS": "false"},
			wantBypass: false,
			wantFatal:  false,
		},
		{
			name:       "ローカル（K_SERVICE なし）はバイパス許可",
			env:        map[string]string{"AUTH_BYPASS": "true"},
			wantBypass: true,
			wantFatal:  false,
		},
		{
			name:       "本番 Cloud Run（オプトインなし）は起動拒否",
			env:        map[string]string{"AUTH_BYPASS": "true", "K_SERVICE": "moe-manager-api"},
			wantBypass: false,
			wantFatal:  true,
		},
		{
			name:       "デモ Cloud Run（オプトインあり）はバイパス許可",
			env:        map[string]string{"AUTH_BYPASS": "true", "K_SERVICE": "moe-manager-demo", "AUTH_BYPASS_ALLOW_CLOUD_RUN": "true"},
			wantBypass: true,
			wantFatal:  false,
		},
		{
			name:       "オプトインが true 以外は拒否",
			env:        map[string]string{"AUTH_BYPASS": "true", "K_SERVICE": "moe-manager-api", "AUTH_BYPASS_ALLOW_CLOUD_RUN": "1"},
			wantBypass: false,
			wantFatal:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string { return tc.env[key] }
			bypass, fatalReason := authBypassDecision(getenv)
			if bypass != tc.wantBypass {
				t.Errorf("bypass = %v, want %v", bypass, tc.wantBypass)
			}
			if (fatalReason != "") != tc.wantFatal {
				t.Errorf("fatalReason = %q, wantFatal %v", fatalReason, tc.wantFatal)
			}
		})
	}
}
