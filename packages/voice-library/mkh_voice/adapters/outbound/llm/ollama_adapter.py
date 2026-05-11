import httpx
import json
from mkh_voice.domain.ports import LLMClientPort
from typing import List, Dict

class OllamaAdapter(LLMClientPort):
    def __init__(self, model_name: str = "gemma2", base_url: str = "http://localhost:11434"):
        self.model_name = model_name
        self.base_url = f"{base_url}/api/chat"

    def generate_response(self, api_key: str, prompt: str, history: List[Dict[str, str]], system_instruction: str = None) -> str:
        """
        Sends the prompt and history to local Ollama API.
        The api_key is ignored for local LLM but kept for interface compatibility.
        """
        messages = []
        
        # Add system instruction if provided
        if system_instruction:
            messages.append({"role": "system", "content": system_instruction})
            
        # Convert history format
        for h in history:
            messages.append({
                "role": "user" if h["role"] == "user" else "assistant",
                "content": h["content"]
            })
            
        # Add current prompt
        messages.append({"role": "user", "content": prompt})
            
        payload = {
            "model": self.model_name,
            "messages": messages,
            "stream": False,
            "options": {
                "temperature": 0.7
            }
        }
        
        try:
            with httpx.Client(timeout=60.0) as client:
                response = client.post(self.base_url, json=payload)
                response.raise_for_status()
                data = response.json()
                return data["message"]["content"]
        except Exception as e:
            print(f"[ERROR] Ollama request failed: {e}")
            raise e
