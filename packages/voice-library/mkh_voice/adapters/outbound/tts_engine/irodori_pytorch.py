import os
import sys
import io
import time
import torch
import soundfile as sf
from typing import Optional, Any
from mkh_voice.domain.ports import SynthesizerPort

# Inject the submodule path into sys.path so we can import irodori_tts
SUBMODULE_PATH = os.path.join(os.path.dirname(__file__), "../../../engine/irodori_tts")
if SUBMODULE_PATH not in sys.path:
    sys.path.insert(0, SUBMODULE_PATH)

try:
    from irodori_tts.inference_runtime import (
        InferenceRuntime,
        RuntimeKey,
        SamplingRequest,
    )
except ImportError as e:
    print(f"Warning: Could not import Irodori-TTS: {e}")
    InferenceRuntime = None

BASE_MODEL_ID = "Aratako/Irodori-TTS-500M-v2"
VOICE_DESIGN_MODEL_ID = "Aratako/Irodori-TTS-500M-v2-VoiceDesign"

class IrodoriPyTorchAdapter(SynthesizerPort):
    """
    Adapter for Irodori-TTS-500M using PyTorch and MPS.
    Implements the SynthesizerPort with support for both Base and VoiceDesign models.
    """
    def __init__(self, default_model: str = VOICE_DESIGN_MODEL_ID, use_mps: bool = True):
        self.default_model = default_model
        self.use_mps = use_mps
        self.runtimes = {} # Cache for multiple models (e.g. base and voicedesign)
        
        # Pre-load the default model
        self._get_runtime(self.default_model)

    def _get_runtime(self, model_id: str) -> Optional[InferenceRuntime]:
        """
        Retrieves a runtime for the specified model_id, loading it if necessary.
        """
        if InferenceRuntime is None:
            print("[IrodoriPyTorchAdapter] Irodori-TTS module not found.")
            return None

        if model_id in self.runtimes:
            return self.runtimes[model_id]

        device = "mps" if self.use_mps and torch.backends.mps.is_available() else "cpu"
        print(f"[IrodoriPyTorchAdapter] Loading model {model_id} on {device}...")
        
        # Check if model_id is a local file, otherwise download from HF
        checkpoint_path = model_id
        if not os.path.isfile(checkpoint_path):
            from huggingface_hub import hf_hub_download
            try:
                print(f"[IrodoriPyTorchAdapter] Downloading model.safetensors from HF repo: {model_id}...")
                checkpoint_path = hf_hub_download(
                    repo_id=model_id,
                    filename="model.safetensors"
                )
            except Exception as e:
                print(f"[IrodoriPyTorchAdapter] Failed to find or download checkpoint {model_id}: {e}")
                return None

        # Use InferenceRuntime from the submodule
        key = RuntimeKey(
            checkpoint=checkpoint_path,
            model_device=device,
            codec_repo="Aratako/Semantic-DACVAE-Japanese-32dim",
            model_precision="fp32", # MPS currently prefers fp32 for stability
            codec_device=device,
            codec_precision="fp32",
        )

        try:
            runtime = InferenceRuntime.from_key(key)
            self.runtimes[model_id] = runtime
            print(f"[IrodoriPyTorchAdapter] Model {model_id} loaded successfully.")
            return runtime
        except Exception as e:
            print(f"[IrodoriPyTorchAdapter] Error loading runtime for {model_id}: {e}")
            return None

    def synthesize(self, text: str, caption: str, ref_wav_path: str = None, sampling_params: dict = None) -> bytes:
        """
        Executes TTS inference.
        If ref_wav_path is provided, uses the base model for optimal cloning.
        Otherwise uses the VoiceDesign model.
        """
        if sampling_params is None:
            sampling_params = {}

        # 1. Decide which model to use
        # If reference audio is provided, we MUST use the base model because 
        # VoiceDesign model has speaker branch disabled.
        target_model = BASE_MODEL_ID if ref_wav_path else VOICE_DESIGN_MODEL_ID
        
        print(f"[IrodoriPyTorchAdapter] Synthesizing -> Model: {target_model}, Text: '{text[:20]}...', RefWav: '{ref_wav_path}'")
        
        runtime = self._get_runtime(target_model)
        if runtime is None:
            print(f"[IrodoriPyTorchAdapter] Target runtime {target_model} not available. Falling back to dummy.")
            return self._generate_dummy_wav()

        # 2. Resolve relative path to absolute path
        resolved_ref_wav = None
        if ref_wav_path:
            project_root = os.path.abspath(os.path.join(SUBMODULE_PATH, "../../.."))
            resolved_ref_wav = os.path.abspath(os.path.join(project_root, ref_wav_path))

        # 3. Build SamplingRequest with parameters from preset
        req = SamplingRequest(
            text=text,
            caption=caption,
            ref_wav=resolved_ref_wav,
            no_ref=(resolved_ref_wav is None),
            
            # Apply custom parameters if provided, else use spec defaults
            speaker_kv_scale=sampling_params.get("speaker_kv_scale", None),
            cfg_scale_speaker=sampling_params.get("cfg_scale_speaker", 5.0),
            num_steps=sampling_params.get("num_steps", 40),
        )
        
        # 4. Run inference
        try:
            result = runtime.synthesize(req)
        except Exception as e:
            print(f"[IrodoriPyTorchAdapter] Inference error: {e}")
            return self._generate_dummy_wav()
        
        # 5. Convert tensor to WAV bytes
        audio_tensor = result.audio.cpu()
        sample_rate = result.sample_rate
        
        wav_io = io.BytesIO()
        sf.write(wav_io, audio_tensor.squeeze(0).numpy(), sample_rate, format='WAV', subtype='PCM_16')
        return wav_io.getvalue()

    def _generate_dummy_wav(self) -> bytes:
        import wave
        wav_io = io.BytesIO()
        with wave.open(wav_io, 'wb') as wav_file:
            wav_file.setnchannels(1)
            wav_file.setsampwidth(2)
            wav_file.setframerate(24000)
            wav_file.writeframes(b'\x00\x00' * 24000)
        return wav_io.getvalue()
