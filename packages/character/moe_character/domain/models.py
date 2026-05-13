from pydantic import BaseModel
from typing import Optional
from enum import Enum

class MBTIType(str, Enum):
    ISTJ = "ISTJ"
    ENFJ = "ENFJ"
    INTJ = "INTJ"
    ENTP = "ENTP"
    # 残り12タイプは後続Phaseで追加

class Character(BaseModel):
    character_id: str
    name: str
    mbti: MBTIType
    description: str
    speech_style: str
    voice_preset_id: str        # voice-library の VoicePreset ID と対応
    icon_path: Optional[str] = None
    voice_sample_path: Optional[str] = None
