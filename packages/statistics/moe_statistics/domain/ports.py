from typing import Protocol
from datetime import date
from .models import DailySummary

class StatisticsQueryPort(Protocol):
    def get_daily_summary(self, user_id: str, target_date: date) -> DailySummary: ...
    def get_weekly_summaries(self, user_id: str) -> list[DailySummary]: ...
