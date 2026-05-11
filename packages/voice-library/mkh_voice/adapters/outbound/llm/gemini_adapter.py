from google import genai
from mkh_voice.domain.ports import LLMClientPort
from typing import List, Dict

class GeminiAIAdapter(LLMClientPort):
    def generate_response(self, api_key: str, prompt: str, history: List[Dict[str, str]], system_instruction: str = None) -> str:
        """
        Sends the prompt and history to Gemini API using the latest google.genai SDK.
        Optimized to use dedicated system_instruction.
        """
        client = genai.Client(api_key=api_key)
        
        # Convert history format to Gemini format
        gemini_contents = []
        for h in history:
            gemini_contents.append({
                "role": "user" if h["role"] == "user" else "model",
                "parts": [{"text": h["content"]}]
            })
            
        # Add current prompt
        gemini_contents.append({
            "role": "user",
            "parts": [{"text": prompt}]
        })
            
        response = client.models.generate_content(
            model="gemini-2.0-flash",
            contents=gemini_contents,
            config={
                "system_instruction": system_instruction
            } if system_instruction else None
        )
        
        return response.text
