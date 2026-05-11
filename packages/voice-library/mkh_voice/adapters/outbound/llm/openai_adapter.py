import time
import logging
from typing import List, Dict, Optional
from openai import OpenAI, RateLimitError, APIError
from mkh_voice.domain.ports import LLMClientPort

logger = logging.getLogger(__name__)

class OpenAIAdapter(LLMClientPort):
    def __init__(self, model_name: str = "gpt-4o"):
        self.model_name = model_name
        # Note: The client is initialized once, but we will use the api_key passed in generate_response
        # This aligns with the current architecture where api_key is passed from the application layer.
        self._last_api_key = None
        self._client = None

    def _get_client(self, api_key: str) -> OpenAI:
        if self._client is None or self._last_api_key != api_key:
            self._client = OpenAI(api_key=api_key)
            self._last_api_key = api_key
        return self._client

    def generate_response(self, api_key: str, prompt: str, history: List[Dict[str, str]], system_instruction: str = None) -> str:
        """
        Sends the prompt and history to OpenAI API.
        Includes handling for RateLimitError (429) as per LESSON-002.
        """
        client = self._get_client(api_key)
        
        messages = []
        if system_instruction:
            messages.append({"role": "system", "content": system_instruction})
            
        for h in history:
            messages.append({
                "role": h["role"],
                "content": h["content"]
            })
            
        messages.append({"role": "user", "content": prompt})

        max_retries = 3
        retry_delay = 2 # seconds
        
        for attempt in range(max_retries):
            try:
                response = client.chat.completions.create(
                    model=self.model_name,
                    messages=messages,
                    temperature=0.7,
                )
                return response.choices[0].message.content
                
            except RateLimitError as e:
                logger.warning(f"Rate limit hit (attempt {attempt + 1}/{max_retries}): {e}")
                if attempt < max_retries - 1:
                    time.sleep(retry_delay * (attempt + 1))
                    continue
                else:
                    raise Exception("OpenAI APIのリクエスト制限（Quota）に達しました。しばらく待ってから再試行してください。") from e
                    
            except APIError as e:
                logger.error(f"OpenAI API error: {e}")
                raise Exception(f"OpenAI APIでエラーが発生しました: {e}") from e
                
            except Exception as e:
                logger.error(f"Unexpected error in OpenAIAdapter: {e}")
                raise e
