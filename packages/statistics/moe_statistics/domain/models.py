from pydantic import BaseModel
from datetime import date

class DailySummary(BaseModel):
    user_id: str
    date: date
    completed_tasks: int
    incomplete_tasks: int
    completion_rate: float          # 0.0 ~ 1.0
    entertainment_minutes: int
    target_minutes: int
    entertainment_diff: int         # 実績 - 目標（負ならOK）
