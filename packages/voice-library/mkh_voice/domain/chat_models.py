from dataclasses import dataclass, field
from datetime import datetime
from typing import List, Optional
import uuid

@dataclass(frozen=True)
class ChatMessage:
    role: str  # "user" or "model"
    content: str
    timestamp: datetime = field(default_factory=datetime.now)

@dataclass
class ChatSession:
    session_id: str = field(default_factory=lambda: str(uuid.uuid4()))
    created_at: datetime = field(default_factory=datetime.now)
    messages: List[ChatMessage] = field(default_factory=list)

    def add_message(self, role: str, content: str):
        self.messages.append(ChatMessage(role=role, content=content))

@dataclass(frozen=True)
class EpisodeMemory:
    content: str
    emotional_impact: int  # 0-100
    timestamp: datetime = field(default_factory=datetime.now)
    id: str = field(default_factory=lambda: str(uuid.uuid4()))
