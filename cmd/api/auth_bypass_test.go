package main

import "testing"

func TestAuthBypassDecision(t *testing.T) {
	tests := []struct {
		name           string
		env            map[string]string
		wantBypass     bool
		wantOnCloudRun bool
		wantFatal      bool
	}{
		{
			name:           "未設定はバイパスしない",
			env:            map[string]string{},
			wantBypass:     false,
			wantOnCloudRun: false,
			wantFatal:      false,
		},
		{
			name:           "AUTH_BYPASS=false はバイパスしない",
			env:            map[string]string{"AUTH_BYPASS": "false"},
			wantBypass:     false,
			wantOnCloudRun: false,
			wantFatal:      false,
		},
		{
			name:           "ローカル（K_SERVICE なし）はバイパス許可",
			env:            map[string]string{"AUTH_BYPASS": "true"},
			wantBypass:     true,
			wantOnCloudRun: false,
			wantFatal:      false,
		},
		{
			name:           "本番 Cloud Run（オプトインなし）は起動拒否",
			env:            map[string]string{"AUTH_BYPASS": "true", "K_SERVICE": "moe-manager-api"},
			wantBypass:     false,
			wantOnCloudRun: true,
			wantFatal:      true,
		},
		{
			name:           "デモ Cloud Run（オプトインあり）はバイパス許可",
			env:            map[string]string{"AUTH_BYPASS": "true", "K_SERVICE": "moe-manager-demo", "AUTH_BYPASS_ALLOW_CLOUD_RUN": "true"},
			wantBypass:     true,
			wantOnCloudRun: true,
			wantFatal:      false,
		},
		{
			name:           "オプトインが true 以外は拒否",
			env:            map[string]string{"AUTH_BYPASS": "true", "K_SERVICE": "moe-manager-api", "AUTH_BYPASS_ALLOW_CLOUD_RUN": "1"},
			wantBypass:     false,
			wantOnCloudRun: true,
			wantFatal:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string { return tc.env[key] }
			bypass, onCloudRun, fatalReason := authBypassDecision(getenv)
			if bypass != tc.wantBypass {
				t.Errorf("bypass = %v, want %v", bypass, tc.wantBypass)
			}
			if onCloudRun != tc.wantOnCloudRun {
				t.Errorf("onCloudRun = %v, want %v", onCloudRun, tc.wantOnCloudRun)
			}
			if (fatalReason != "") != tc.wantFatal {
				t.Errorf("fatalReason = %q, wantFatal %v", fatalReason, tc.wantFatal)
			}
		})
	}
}
