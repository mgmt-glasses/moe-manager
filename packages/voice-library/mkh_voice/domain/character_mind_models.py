from datetime import datetime
from typing import List, Optional
from pydantic import BaseModel, Field

class CharacterPersona(BaseModel):
    """
    Layer 1: Core Persona (Immutable Personality)
    Represents the fundamental traits of a character used for LLM instructions.
    """
    id: str = Field(..., description="Same ID as VoicePreset to link them.")
    name: str
    gender: str = Field("unknown", description="Gender (e.g., Female, Male, Non-binary)")
    age: Optional[int] = Field(None, description="Age of the character")
    first_person: str = Field("私", description="How the character refers to themselves (e.g., 私, 僕, 俺)")
    second_person: str = Field("あなた", description="How the character refers to the user (e.g., あなた, 君, お前)")
    personality: List[str] = Field(default_factory=list, description="Core traits (e.g., Introverted, Possessive)")
    values: List[str] = Field(default_factory=list, description="Fundamental values (e.g., Values promises)")
    speech_style: List[str] = Field(default_factory=list, description="Guidelines for speaking (e.g., Short sentences, Sarcastic)")
    taboos: List[str] = Field(default_factory=list, description="Things the character must never do/say.")

class CharacterPersonaRepositoryPort:
    """Port for managing persona data (SSOT)."""
    def get_persona(self, persona_id: str) -> CharacterPersona:
        raise NotImplementedError

    def save_persona(self, persona: CharacterPersona) -> None:
        raise NotImplementedError

    def list_personas(self) -> List[CharacterPersona]:
        raise NotImplementedError

class CharacterState(BaseModel):
    """
    Layer 2: Character State (Dynamic Mood)
    """
    character_id: str
    mood: int = Field(50, ge=0, le=100)
    fatigue: int = Field(0, ge=0, le=100)
    stress: int = Field(0, ge=0, le=100)
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    def apply_temporal_decay(self, current_time: datetime) -> dict:
        """
        Applies decay based on time passed since updated_at.
        Returns a summary of changes.
        """
        elapsed_seconds = (current_time - self.updated_at).total_seconds()
        if elapsed_seconds <= 0:
            return {}

        hours = elapsed_seconds / 3600
        
        # Recovery rates
        fatigue_recovery = int(hours * 10) # 10/hour
        stress_recovery = int(hours * 5)   # 5/hour
        
        # Mood convergence to 50
        mood_adjustment = 0
        if self.mood > 50:
            mood_adjustment = -int(hours * 5) # -5/hour
            new_mood = max(50, self.mood + mood_adjustment)
        elif self.mood < 50:
            mood_adjustment = int(hours * 5)  # +5/hour
            new_mood = min(50, self.mood + mood_adjustment)
        else:
            new_mood = 50

        orig_mood, orig_fatigue, orig_stress = self.mood, self.fatigue, self.stress
        
        self.mood = new_mood
        self.fatigue = max(0, self.fatigue - fatigue_recovery)
        self.stress = max(0, self.stress - stress_recovery)
        self.updated_at = current_time

        return {
            "elapsed_hours": round(hours, 2),
            "mood_change": self.mood - orig_mood,
            "fatigue_change": self.fatigue - orig_fatigue,
            "stress_change": self.stress - orig_stress
        }

class RelationshipState(BaseModel):
    """
    Layer 3: Relationship State (Trust/Affection)
    """
    character_id: str
    user_id: str = "default_user"
    trust: int = Field(0, ge=0, le=100)
    affection: int = Field(0, ge=0, le=100)
    dependency: int = Field(0, ge=0, le=100)
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    def apply_temporal_decay(self, current_time: datetime) -> dict:
        """
        Relationships are more stable but might slightly decay over very long periods if ignored.
        For now, we just update the timestamp.
        """
        self.updated_at = current_time
        return {}

class CharacterStateRepositoryPort:
    """Port for persisting dynamic states."""
    def get_state(self, character_id: str) -> CharacterState:
        raise NotImplementedError
    def save_state(self, state: CharacterState) -> None:
        raise NotImplementedError
    def get_relationship(self, character_id: str, user_id: str) -> RelationshipState:
        raise NotImplementedError
    def save_relationship(self, rel: RelationshipState) -> None:
        raise NotImplementedError

class StateUpdate(BaseModel):
    """
    Delta values inferred from conversation.
    """
    mood_delta: int = 0
    fatigue_delta: int = 0
    stress_delta: int = 0
    trust_delta: int = 0
    affection_delta: int = 0
