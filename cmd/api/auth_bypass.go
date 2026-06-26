package main

import "fmt"

// authBypassDecision は環境変数から認証バイパスの可否を判定する。
//
// AUTH_BYPASS=true のときだけバイパスを検討する。Cloud Run（予約環境変数
// K_SERVICE が設定される）上では既定で拒否し、デモ・検証環境が
// AUTH_BYPASS_ALLOW_CLOUD_RUN=true を明示した場合に限って許可する。
// 本番 Cloud Run は両変数とも設定しないため自動的に拒否される（fail-safe）。
//
// onCloudRun は K_SERVICE の有無（Cloud Run 上で動いているか）を返し、
// 呼び出し側はログ文言の出し分けに使える。
// fatalReason が空でない場合、呼び出し側は起動を中止すべき（本番への誤混入）。
func authBypassDecision(getenv func(string) string) (bypass, onCloudRun bool, fatalReason string) {
	onCloudRun = getenv("K_SERVICE") != ""
	if getenv("AUTH_BYPASS") != "true" {
		return false, onCloudRun, ""
	}
	if onCloudRun && getenv("AUTH_BYPASS_ALLOW_CLOUD_RUN") != "true" {
		return false, onCloudRun, fmt.Sprintf(
			"AUTH_BYPASS=true は Cloud Run 環境（K_SERVICE=%q）では既定で拒否されます。"+
				"デモ・検証環境で有効化する場合のみ AUTH_BYPASS_ALLOW_CLOUD_RUN=true を明示してください。"+
				"本番では設定しないこと。",
			getenv("K_SERVICE"),
		)
	}
	return true, onCloudRun, ""
}
