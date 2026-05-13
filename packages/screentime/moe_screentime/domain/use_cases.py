from datetime import date
from .models import ScreenTimeRecord
from .ports import ScreenTimeRepositoryPort

class ScreenTimeUseCase:
    def __init__(self, repository: ScreenTimeRepositoryPort):
        self._repository = repository

    def record(self, record: ScreenTimeRecord) -> None:
        # TODO: implement
        pass

    def get_today(self, user_id: str) -> ScreenTimeRecord:
        # TODO: implement
        pass

    def get_weekly_summary(self, user_id: str) -> list[ScreenTimeRecord]:
        # TODO: implement
        pass
