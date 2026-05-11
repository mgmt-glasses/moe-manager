from typing import Optional
from pydantic import BaseModel, Field

class VoicePreset(BaseModel):
    """Represents a Character or specific VoiceDesign preset."""
    id: str
    name: str
    caption: str
    reference_audio_path: Optional[str] = Field(None, description="Path to the saved WAV file used for Voice Cloning.")
    speaker_kv_scale: Optional[float] = Field(None, description="Extra speaker K/V scaling for stronger speaker identity.")
    cfg_scale_speaker: Optional[float] = Field(None, description="CFG scale for speaker conditioning.")

class VoiceRequest(BaseModel):
    """Client request format for voice generation."""
    text: str = Field(..., description="The raw text containing dialog and optional emoji tags.")
    preset_id: Optional[str] = Field(None, description="The ID of the saved preset to use.")
    override_caption: Optional[str] = Field(None, description="A raw caption string to use instead of a saved preset. Useful for testing.")
    
class VoiceResponse(BaseModel):
    """The conceptual response from the domain (used internally)."""
    audio_bytes: bytes
    media_type: str = "audio/wav"

