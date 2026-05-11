from typing import Protocol, Optional
from mkh_voice.domain.models import VoicePreset

class PresetRepositoryPort(Protocol):
    """
    Outbound Port (Interface) for persisting VoicePreset (Character) data.
    """
    def save(self, preset: VoicePreset) -> None:
        pass
        
    def get(self, preset_id: str) -> Optional[VoicePreset]:
        pass
        
    def list_all(self) -> list[VoicePreset]:
        pass
        
    def delete(self, preset_id: str) -> None:
        pass

class AudioStoragePort(Protocol):
    """
    Outbound Port (Interface) for persisting audio files.
    Abstracts away the file system or cloud storage.
    """
    def save_audio(self, identifier: str, audio_bytes: bytes) -> str:
        """
        Saves the audio and returns a URI or path that can be used to retrieve it later.
        """
        pass

    def append_audio(self, identifier: str, audio_bytes: bytes) -> str:
        """
        Appends new audio bytes to an existing audio file identified by the identifier.
        If the file doesn't exist, it should behave like save_audio.
        Returns the path to the updated file.
        """
        pass

class SynthesizerPort(Protocol):
    """
    Outbound Port (Interface) for the TTS Engine.
    Any implementation (PyTorch, CoreML, Mock) must satisfy this protocol.
    """
    def synthesize(self, text: str, caption: str, ref_wav_path: Optional[str] = None, sampling_params: Optional[dict] = None) -> bytes:
        """
        Synthesizes speech from text and voice design caption.
        Optionally uses a reference audio file for Voice Cloning.
        Returns the audio data as raw bytes (e.g., WAV format).
        """
        pass

class ChatRepositoryPort(Protocol):
    """
    Outbound Port for persisting chat sessions and messages.
    """
    def create_session(self, session_id: str) -> None:
        pass
        
    def save_message(self, session_id: str, role: str, content: str) -> None:
        pass
        
    def get_history(self, session_id: str) -> list:
        pass

class LLMClientPort(Protocol):
    """
    Outbound Port for interacting with LLMs like Gemini.
    """
    def generate_response(self, api_key: str, prompt: str, history: list, system_instruction: str = None) -> str:
        pass

class MemoryRepositoryPort(Protocol):
    """
    Outbound Port for persisting and retrieving episodic memories.
    """
    def save_memory(self, character_id: str, content: str, emotional_impact: int) -> None:
        pass

    def search_memories(self, character_id: str, query: str, limit: int = 5) -> list:
        """
        Searches memories with emotional impact and time-based decay.
        """
        pass

class UserProfilePort(Protocol):
    """
    Outbound Port for persisting user-specific facts (Layer 4).
    """
    def get_profile(self, user_id: str) -> dict:
        pass

    def save_profile(self, user_id: str, profile_data: dict) -> None:
        pass
