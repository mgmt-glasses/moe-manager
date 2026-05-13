from datetime import date
from .models import DailySummary
from .ports import StatisticsQueryPort

class StatisticsUseCase:
    def __init__(self, query: StatisticsQueryPort):
        self._query = query

    def get_today(self, user_id: str) -> DailySummary:
        # TODO: implement
        pass

    def get_weekly(self, user_id: str) -> list[DailySummary]:
        # TODO: implement
        pass
