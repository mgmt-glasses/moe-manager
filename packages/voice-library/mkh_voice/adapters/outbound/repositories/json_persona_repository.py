import json
import os
from typing import List
from mkh_voice.domain.character_mind_models import CharacterPersona, CharacterPersonaRepositoryPort

class JsonPersonaRepository(CharacterPersonaRepositoryPort):
    def __init__(self, file_path: str = "data/personas.json"):
        self.file_path = file_path
        os.makedirs(os.path.dirname(file_path), exist_ok=True)
        if not os.path.exists(file_path):
            with open(file_path, 'w', encoding='utf-8') as f:
                json.dump([], f, ensure_ascii=False, indent=2)

    def _load_all(self) -> List[dict]:
        with open(self.file_path, 'r', encoding='utf-8') as f:
            return json.load(f)

    def list_personas(self) -> List[CharacterPersona]:
        data = self._load_all()
        return [CharacterPersona(**item) for item in data]

    def get_persona(self, persona_id: str) -> CharacterPersona:
        personas = self.list_personas()
        for p in personas:
            if p.id == persona_id:
                return p
        # Default or Empty persona if not found
        return CharacterPersona(id=persona_id, name="Unknown")

    def save_persona(self, persona: CharacterPersona) -> None:
        data = self._load_all()
        # Update or Append
        new_data = [item for item in data if item['id'] != persona.id]
        new_data.append(persona.dict())
        with open(self.file_path, 'w', encoding='utf-8') as f:
            json.dump(new_data, f, ensure_ascii=False, indent=2)
