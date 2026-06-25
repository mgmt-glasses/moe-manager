from fastapi import FastAPI

# Inbound adapters (routers) from each module
from moe_user.adapters.inbound.api.user_router import router as user_router
from moe_character.adapters.inbound.api.character_router import router as character_router
from moe_task.adapters.inbound.api.task_router import router as task_router
from moe_statistics.adapters.inbound.api.statistics_router import router as statistics_router

# voice-library の chat/voice router はそのまま流用
from mkh_voice.adapters.inbound.api.chat_router import router as chat_router

app = FastAPI(title="Moe Manager API", version="0.1.0")

app.include_router(user_router)
app.include_router(character_router)
app.include_router(task_router)
app.include_router(statistics_router)
app.include_router(chat_router)

@app.get("/health")
async def health():
    return {"status": "ok"}
