import os
import sys
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, UploadFile, File
from fastapi.responses import StreamingResponse
from fastapi.staticfiles import StaticFiles
from io import BytesIO

# Strict Venv Check
if not (hasattr(sys, 'real_prefix') or (target := getattr(sys, 'base_prefix', sys.prefix)) != sys.prefix):
    print("Error: Virtual environment is not active. Please run 'source .venv/bin/activate'.", file=sys.stderr)

from mkh_voice.domain.models import VoiceRequest, VoicePreset
from mkh_voice.domain.use_cases import VoiceGenerationUseCase, PresetManagementUseCase
from mkh_voice.adapters.outbound.tts_engine.irodori_pytorch import IrodoriPyTorchAdapter
from mkh_voice.adapters.outbound.repositories.json_preset_repository import JsonPresetRepository
from mkh_voice.adapters.outbound.repositories.local_audio_storage import LocalAudioStorage
from mkh_voice.adapters.inbound.api import chat_router

voice_use_case: VoiceGenerationUseCase = None
preset_use_case: PresetManagementUseCase = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Setup: Load model on startup
    global voice_use_case, preset_use_case
    print("Initializing TTS Engine and Repositories...")
    adapter = IrodoriPyTorchAdapter(use_mps=True)
    repo = JsonPresetRepository()
    audio_storage = LocalAudioStorage()
    
    voice_use_case = VoiceGenerationUseCase(synthesizer=adapter, preset_repo=repo)
    preset_use_case = PresetManagementUseCase(repository=repo, audio_storage=audio_storage)
    
    # Ensure static directory exists
    os.makedirs("static", exist_ok=True)
    
    yield
    # Teardown
    voice_use_case = None
    preset_use_case = None

app = FastAPI(title="M.K.H. Voice Engine API", lifespan=lifespan)
app.include_router(chat_router.router)

@app.get("/health")
async def health_check():
    return {"status": "ok", "engine": "Irodori-TTS-500M (Hexagonal)"}

@app.get("/routes")
async def get_routes():
    return [{"path": route.path, "name": route.name, "methods": route.methods} for route in app.routes]

@app.get("/presets")
async def get_presets():
    if preset_use_case is None:
        raise HTTPException(status_code=503, detail="System not initialized")
    return preset_use_case.list_presets()

@app.post("/presets")
async def save_preset(preset: VoicePreset):
    if preset_use_case is None:
        raise HTTPException(status_code=503, detail="System not initialized")
    preset_use_case.save_preset(preset)
    return {"status": "saved"}

@app.delete("/presets/{preset_id}")
async def delete_preset(preset_id: str):
    if preset_use_case is None:
        raise HTTPException(status_code=503, detail="System not initialized")
    preset_use_case.delete_preset(preset_id)
    return {"status": "deleted"}

@app.post("/presets/{preset_id}/reference")
async def upload_reference_audio(preset_id: str, file: UploadFile = File(...)):
    if preset_use_case is None:
        raise HTTPException(status_code=503, detail="System not initialized")
    try:
        audio_bytes = await file.read()
        file_path = preset_use_case.save_reference_audio(preset_id, audio_bytes)
        return {"status": "saved", "reference_audio_path": file_path}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/presets/{preset_id}/reference/append")
async def append_reference_audio(preset_id: str, file: UploadFile = File(...)):
    if preset_use_case is None:
        raise HTTPException(status_code=503, detail="System not initialized")
    try:
        audio_bytes = await file.read()
        file_path = preset_use_case.append_reference_audio(preset_id, audio_bytes)
        return {"status": "appended", "reference_audio_path": file_path}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.delete("/presets/{preset_id}/reference")
async def clear_reference_audio(preset_id: str):
    if preset_use_case is None:
        raise HTTPException(status_code=503, detail="System not initialized")
    try:
        preset_use_case.clear_reference_audio(preset_id)
        return {"status": "cleared"}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/generate")
async def generate_voice(request: VoiceRequest):
    """
    Receives text with emojis, preset configuration, generates voice,
    and returns a binary audio stream directly without saving to disk.
    """
    if voice_use_case is None:
        raise HTTPException(status_code=503, detail="TTS Engine is not initialized")
    
    try:
        audio_bytes = voice_use_case.execute(request)
        return StreamingResponse(BytesIO(audio_bytes), media_type="audio/wav")
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# Mount reference audio files so they can be played in the UI
os.makedirs("data/references", exist_ok=True)
app.mount("/references", StaticFiles(directory="data/references"), name="references")

# Mount static files for the UI
app.mount("/", StaticFiles(directory="static", html=True), name="static")
