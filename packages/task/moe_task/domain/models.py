from pydantic import BaseModel
from datetime import datetime
from typing import Optional
from enum import Enum

class TaskStatus(str, Enum):
    INCOMPLETE = "incomplete"
    COMPLETE = "complete"

class Task(BaseModel):
    task_id: str
    user_id: str
    title: str
    status: TaskStatus = TaskStatus.INCOMPLETE
    created_at: datetime = datetime.utcnow()
    completed_at: Optional[datetime] = None
