from pydantic import BaseModel, Field
from datetime import date
from typing import List


class ScreenTimeCategoryUsage(BaseModel):
    category: str
    minutes: int


class ScreenTimeAnalysisResult(BaseModel):
    items: List[ScreenTimeCategoryUsage] = Field(default_factory=list, max_length=10)


class ScreenTimeRecord(BaseModel):
    record_id: str
    user_id: str
    date: date
    entertainment_minutes: int
    target_minutes: int
    categories: List[ScreenTimeCategoryUsage] = Field(default_factory=list)

    @property
    def diff_minutes(self) -> int:
        return self.entertainment_minutes - self.target_minutes
