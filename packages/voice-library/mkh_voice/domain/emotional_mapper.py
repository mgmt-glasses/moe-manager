from mkh_voice.domain.character_mind_models import CharacterState, RelationshipState

class EmotionalMapper:
    """
    Service to map character mind states to Irodori-TTS emotional emojis.
    Emojis are placed at the beginning of the text to control voice emotion.
    """
    
    @staticmethod
    def get_emoji(state: CharacterState, rel: RelationshipState) -> str:
        # Priority 1: High Stress / Bad Mood (Negative Emotions)
        if state.stress > 70 or state.mood < 20:
            return "💢" # Angry / Extremely annoyed
            
        if state.mood < 40:
            return "😭" # Sad / Depressed
            
        # Priority 2: High Fatigue (Sleepy/Tired)
        if state.fatigue > 70:
            return "🥱" # Tired / Yawning
            
        # Priority 3: High Affection/Trust (Positive/Soft Emotions)
        if rel.affection > 80:
            return "🥰" # Love / Very high affection
            
        if rel.trust > 70 or rel.affection > 50:
            return "😊" # Friendly / Smile
            
        # Default: Neutral
        return "" # Normal voice

    @staticmethod
    def wrap_text(text: str, state: CharacterState, rel: RelationshipState) -> str:
        """Adds emotional emoji to the start of the text."""
        emoji = EmotionalMapper.get_emoji(state, rel)
        if emoji:
            return f"{emoji}{text}"
        return text
