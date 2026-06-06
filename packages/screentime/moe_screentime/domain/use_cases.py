from datetime import date
from typing import Optional

from .models import ScreenTimeRecord, ScreenTimeAnalysisResult
from .ports import ScreenTimeRepositoryPort, ScreenTimeImageAnalyzerPort


class ScreenTimeUseCase:
    def __init__(self, repository: ScreenTimeRepositoryPort):
        self._repository = repository

    def record(self, record: ScreenTimeRecord) -> None:
        """娯楽時間を保存する。同一ユーザ・同一日付の場合は更新する。"""
        self._repository.save(record)

    def get_by_date(self, user_id: str, target_date: date) -> Optional[ScreenTimeRecord]:
        """特定日の娯楽時間を取得する。"""
        return self._repository.get_by_date(user_id, target_date)

    def get_today(self, user_id: str) -> Optional[ScreenTimeRecord]:
        """今日の娯楽時間を取得する。"""
        return self._repository.get_by_date(user_id, date.today())

    def get_by_range(self, user_id: str, start_date: date, end_date: date) -> list[ScreenTimeRecord]:
        """期間指定で娯楽時間の一覧を取得する。"""
        return self._repository.list_by_range(user_id, start_date, end_date)

    def get_weekly_summary(self, user_id: str) -> list[ScreenTimeRecord]:
        """直近7日分の娯楽時間を取得する。"""
        return self._repository.list_recent(user_id, days=7)


class ScreenTimeImageAnalysisUseCase:
    def __init__(self, analyzer: ScreenTimeImageAnalyzerPort):
        self._analyzer = analyzer

    def analyze(self, image_bytes: bytes, mime_type: str = "image/png") -> ScreenTimeAnalysisResult:
        """画像を解析してスクリーンタイムのカテゴリ別利用時間を抽出する"""
        return self._analyzer.analyze(image_bytes=image_bytes, mime_type=mime_type)
