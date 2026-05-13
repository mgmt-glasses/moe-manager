from pydantic import BaseModel
from datetime import date

class ScreenTimeRecord(BaseModel):
    record_id: str
    user_id: str
    date: date
    entertainment_minutes: int
    target_minutes: int

    @property
    def diff_minutes(self) -> int:
        return self.entertainment_minutes - self.target_minutes
