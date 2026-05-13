from pydantic import BaseModel
from datetime import datetime
from typing import Optional

class User(BaseModel):
    user_id: str
    username: str
    boss_name: str           # 社長としての呼び名
    selected_character_id: Optional[str] = None
    target_entertainment_minutes: int = 120
    created_at: datetime = datetime.utcnow()
