from fastapi import APIRouter, Header, HTTPException, Depends
from pydantic import BaseModel
from typing import List, Optional
from mkh_voice.domain.chat_use_cases import ChatUseCase
from mkh_voice.adapters.outbound.repositories.sqlite_chat_repository import SQLiteChatRepository
from mkh_voice.adapters.outbound.repositories.json_persona_repository import JsonPersonaRepository
from mkh_voice.adapters.outbound.llm.ollama_adapter import OllamaAdapter
from mkh_voice.adapters.outbound.llm.gemini_adapter import GeminiAIAdapter
from mkh_voice.adapters.outbound.llm.openai_adapter import OpenAIAdapter

router = APIRouter(prefix="/chat", tags=["chat"])

# Singletons (initialized once) - following LESSON-002
_repo = SQLiteChatRepository()
_persona_repo = JsonPersonaRepository()

# Prepare all available adapters
_ollama = OllamaAdapter(model_name="gemma2") 
_gemini = GeminiAIAdapter()
_openai = OpenAIAdapter(model_name="gpt-4o")

# Default use case initialized with Ollama as baseline
_use_case = ChatUseCase(_repo, _ollama, _persona_repo, _repo, _repo, _repo)

# Dependencies
def get_chat_use_case():
    return _use_case

class ChatRequest(BaseModel):
    message: str

class SessionResponse(BaseModel):
    session_id: str

class MessageResponse(BaseModel):
    role: str
    content: str

@router.post("/sessions", response_model=SessionResponse)
async def create_session(use_case: ChatUseCase = Depends(get_chat_use_case)):
    session_id = use_case.start_new_session()
    return {"session_id": session_id}

@router.post("/sessions/{session_id}/messages")
async def send_message(
    session_id: str, 
    request: ChatRequest, 
    character_id: str = "default",
    x_gemini_api_key: Optional[str] = Header(None),
    x_openai_api_key: Optional[str] = Header(None),
    use_case: ChatUseCase = Depends(get_chat_use_case)
):
    try:
        # Dynamic Provider Selection
        api_key = None
        llm_client = None
        
        if x_openai_api_key:
            api_key = x_openai_api_key
            llm_client = _openai
        elif x_gemini_api_key:
            api_key = x_gemini_api_key
            llm_client = _gemini
        else:
            # Fallback to Ollama (default in _use_case)
            api_key = ""
            llm_client = _ollama
            
        response_text = use_case.execute_chat(
            session_id, 
            api_key, 
            request.message, 
            character_id,
            llm_client=llm_client
        )
        return {"content": response_text}
    except Exception as e:
        print(f"[ERROR] Chat execution failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/characters/{character_id}/state")
async def get_state(
    character_id: str,
    use_case: ChatUseCase = Depends(get_chat_use_case)
):
    state = use_case.state_repo.get_state(character_id)
    rel = use_case.state_repo.get_relationship(character_id, "default_user")
    return {
        "state": state,
        "relationship": rel
    }

@router.get("/characters/{character_id}/persona")
async def get_persona(
    character_id: str,
    use_case: ChatUseCase = Depends(get_chat_use_case)
):
    return use_case.persona_repo.get_persona(character_id)

from mkh_voice.domain.character_mind_models import CharacterPersona
@router.post("/characters/{character_id}/persona")
async def update_persona(
    character_id: str,
    persona: CharacterPersona,
    use_case: ChatUseCase = Depends(get_chat_use_case)
):
    use_case.persona_repo.save_persona(persona)
    return {"status": "saved"}

@router.get("/sessions/{session_id}/history", response_model=List[MessageResponse])
async def get_history(session_id: str, use_case: ChatUseCase = Depends(get_chat_use_case)):
    return use_case.get_session_history(session_id)
