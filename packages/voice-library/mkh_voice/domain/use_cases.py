from typing import Optional
from .models import VoiceRequest, VoicePreset
from .ports import SynthesizerPort, PresetRepositoryPort, AudioStoragePort

class PresetManagementUseCase:
    """
    Handles CRUD operations for VoicePresets (Characters).
    """
    def __init__(self, repository: PresetRepositoryPort, audio_storage: AudioStoragePort = None):
        self._repository = repository
        self._audio_storage = audio_storage

    def list_presets(self) -> list[VoicePreset]:
        return self._repository.list_all()

    def get_preset(self, preset_id: str) -> Optional[VoicePreset]:
        return self._repository.get(preset_id)

    def save_preset(self, preset: VoicePreset) -> None:
        self._repository.save(preset)

    def delete_preset(self, preset_id: str) -> None:
        self._repository.delete(preset_id)

    def save_reference_audio(self, preset_id: str, audio_bytes: bytes) -> str:
        """
        Saves the provided audio bytes as the reference audio for the given preset.
        Returns the path to the saved audio file.
        """
        if not self._audio_storage:
            raise RuntimeError("AudioStoragePort is not configured.")
            
        preset = self.get_preset(preset_id)
        if not preset:
            raise ValueError(f"Preset with id {preset_id} not found.")
            
        # Delegate saving to the injected storage port
        file_path = self._audio_storage.save_audio(f"{preset_id}_reference", audio_bytes)
            
        preset.reference_audio_path = file_path
        self.save_preset(preset)
        return file_path

    def append_reference_audio(self, preset_id: str, audio_bytes: bytes) -> str:
        """
        Appends the provided audio bytes to the existing reference audio for the given preset.
        """
        if not self._audio_storage:
            raise RuntimeError("AudioStoragePort is not configured.")
            
        preset = self.get_preset(preset_id)
        if not preset:
            raise ValueError(f"Preset with id {preset_id} not found.")
            
        # Delegate appending to the injected storage port
        file_path = self._audio_storage.append_audio(f"{preset_id}_reference", audio_bytes)
            
        preset.reference_audio_path = file_path
        self.save_preset(preset)
        return file_path

    def clear_reference_audio(self, preset_id: str) -> None:
        """
        Clears the reference audio path for the given preset.
        """
        preset = self.get_preset(preset_id)
        if not preset:
            raise ValueError(f"Preset with id {preset_id} not found.")
            
        preset.reference_audio_path = None
        self.save_preset(preset)

class VoiceGenerationUseCase:
    """
    Orchestrates the process of voice generation.
    Completely decoupled from HTTP and specific ML frameworks.
    """
    def __init__(self, synthesizer: SynthesizerPort, preset_repo: PresetRepositoryPort):
        self._synthesizer = synthesizer
        self._preset_repo = preset_repo
        
    def execute(self, request: VoiceRequest) -> bytes:
        """
        The main business logic flow.
        1. Determine the caption (from override_caption, preset_id, or default).
        2. Resolve reference audio if preset_id is provided.
        3. Pass the raw text (containing emojis) and caption to the synthesizer.
        4. Return generated audio bytes.
        """
        # Determine caption and reference audio
        caption = "標準的な女性の話者" # Default
        ref_wav_path = None
        sampling_params = {}
        
        if request.override_caption:
            caption = request.override_caption
        
        if request.preset_id:
            preset = self._preset_repo.get(request.preset_id)
            if preset:
                if not request.override_caption:
                    caption = preset.caption
                ref_wav_path = preset.reference_audio_path
                
                # Extract sampling parameters if they exist
                if hasattr(preset, "speaker_kv_scale") and preset.speaker_kv_scale is not None:
                    sampling_params["speaker_kv_scale"] = preset.speaker_kv_scale
                if hasattr(preset, "cfg_scale_speaker") and preset.cfg_scale_speaker is not None:
                    sampling_params["cfg_scale_speaker"] = preset.cfg_scale_speaker
                
        # Irodori-TTS base model natively supports emoji-driven style control,
        # so we pass the raw text without stripping the emojis.
        audio_bytes = self._synthesizer.synthesize(
            text=request.text, 
            caption=caption,
            ref_wav_path=ref_wav_path,
            sampling_params=sampling_params
        )
        return audio_bytes

