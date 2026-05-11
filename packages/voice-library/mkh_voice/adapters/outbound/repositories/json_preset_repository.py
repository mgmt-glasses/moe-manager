import json
import os
from typing import Optional

from mkh_voice.domain.models import VoicePreset
from mkh_voice.domain.ports import PresetRepositoryPort

class JsonPresetRepository(PresetRepositoryPort):
    def __init__(self, file_path: str = "data/characters.json"):
        self.file_path = file_path
        self._ensure_file_exists()

    def _ensure_file_exists(self):
        os.makedirs(os.path.dirname(self.file_path), exist_ok=True)
        if not os.path.exists(self.file_path):
            with open(self.file_path, "w", encoding="utf-8") as f:
                json.dump([], f)

    def _read_all(self) -> list[dict]:
        try:
            with open(self.file_path, "r", encoding="utf-8") as f:
                return json.load(f)
        except json.JSONDecodeError:
            return []

    def _write_all(self, data: list[dict]):
        with open(self.file_path, "w", encoding="utf-8") as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

    def save(self, preset: VoicePreset) -> None:
        data = self._read_all()
        # Update if exists
        updated = False
        for i, item in enumerate(data):
            if item.get("id") == preset.id:
                data[i] = preset.model_dump()
                updated = True
                break
        
        # Add if new
        if not updated:
            data.append(preset.model_dump())
            
        self._write_all(data)

    def get(self, preset_id: str) -> Optional[VoicePreset]:
        data = self._read_all()
        for item in data:
            if item.get("id") == preset_id:
                return VoicePreset(**item)
        return None

    def list_all(self) -> list[VoicePreset]:
        data = self._read_all()
        return [VoicePreset(**item) for item in data]

    def delete(self, preset_id: str) -> None:
        data = self._read_all()
        filtered_data = [item for item in data if item.get("id") != preset_id]
        self._write_all(filtered_data)
