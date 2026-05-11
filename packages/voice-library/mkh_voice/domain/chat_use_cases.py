from mkh_voice.domain.chat_models import ChatSession
from mkh_voice.domain.ports import ChatRepositoryPort, LLMClientPort, MemoryRepositoryPort, UserProfilePort
from mkh_voice.domain.character_mind_models import CharacterPersonaRepositoryPort, CharacterStateRepositoryPort, CharacterState, RelationshipState
from typing import List, Dict

class ChatUseCase:
    def __init__(self, 
                 repository: ChatRepositoryPort, 
                 llm_client: LLMClientPort,
                 persona_repo: CharacterPersonaRepositoryPort,
                 state_repo: CharacterStateRepositoryPort,
                 memory_repo: MemoryRepositoryPort,
                 user_profile_repo: UserProfilePort):
        self.repository = repository
        self.llm_client = llm_client
        self.persona_repo = persona_repo
        self.state_repo = state_repo
        self.memory_repo = memory_repo
        self.user_profile_repo = user_profile_repo

    def start_new_session(self, character_id: str = "default") -> str:
        session = ChatSession()
        # In a real app, we might want to store character_id in the session record
        self.repository.create_session(session.session_id)
        return session.session_id

    def _generate_memory_query(self, client: LLMClientPort, api_key: str, user_message: str) -> str:
        """
        Uses LLM to extract search keywords from user message.
        """
        prompt = f"以下のメッセージから、過去の記憶を検索するための重要なキーワードを1つだけ抽出してください。回答はキーワードのみにしてください。\nメッセージ: {user_message}"
        try:
            # We use a very short generation for speed
            query = client.generate_response(api_key, prompt, [], system_instruction="あなたはキーワード抽出アシスタントです。")
            return query.strip().replace("「", "").replace("」", "").replace("キーワード：", "")
        except:
            return user_message # Fallback

    def execute_chat(self, session_id: str, api_key: str, user_message: str, character_id: str = "default", llm_client: LLMClientPort = None) -> str:
        # Use provided client or fallback to the one from constructor
        client = llm_client or self.llm_client
        
        # 1. Load Persona (Layer 1)
        persona = self.persona_repo.get_persona(character_id)
        
        # 2. Load State (Layer 2) and Relationship (Layer 3)
        from datetime import datetime
        current_time = datetime.utcnow()
        
        state_data = self.state_repo.get_state(character_id)
        state = CharacterState(**state_data)
        
        rel_data = self.state_repo.get_relationship(character_id, "default_user")
        rel = RelationshipState(**rel_data)
        
        # 3. Apply Temporal Decay (Design Philosophy: Time passed matters)
        decay_summary = state.apply_temporal_decay(current_time)
        rel.apply_temporal_decay(current_time) # Currently just updates timestamp
        
        if decay_summary:
            print(f"[DEBUG] Temporal Decay applied: {decay_summary}")
            # Save immediately if there was a change
            self.state_repo.save_state(state.dict())
            self.state_repo.save_relationship(rel.dict())
        
        # 4. Recall Memories (Layer 4)
        # 4a. Check if RAG is needed (Conditional RAG)
        # Skip for very short messages or simple greetings to save tokens and reduce noise
        is_short_msg = len(user_message) < 5
        greetings = ["おはよう", "こんにちは", "こんばんは", "おやすみ", "バイバイ", "またね"]
        is_greeting = any(g in user_message for g in greetings) and len(user_message) < 10
        
        recalled_memories = []
        memory_instruction = ""
        
        if not (is_short_msg or is_greeting):
            # 4b. Generate Search Query
            search_query = self._generate_memory_query(client, api_key, user_message)
            print(f"[DEBUG] Memory Search Query: {search_query}")
            
            # 4c. Search relevant memories
            recalled_memories = self.memory_repo.search_memories(character_id, search_query)
            if recalled_memories:
                memory_list = "\n".join([f"- {m['content']} (影響度: {m['impact']})" for m in recalled_memories])
                memory_instruction = f"\n【過去の記憶 (Layer 5)】\n以下の出来事を思い出してください:\n{memory_list}\n"
        else:
            print(f"[DEBUG] RAG Skipped (Short message or greeting)")

        # 4d. Load User Profile (Layer 4)
        user_id = "default_user" # In a real app, this would come from the session/auth
        user_profile = self.user_profile_repo.get_profile(user_id)
        user_profile_instruction = ""
        if user_profile:
            profile_facts = "\n".join([f"- {k}: {v}" for k, v in user_profile.items()])
            user_profile_instruction = f"\n【ユーザー情報 (Layer 4)】\n対話相手に関する既知の事実:\n{profile_facts}\n"

        # 4. Build Layered Prompt with Inference Instructions
        system_instruction = f"""
あなたはキャラクター「{persona.name}」として振る舞ってください。

【人格設定 (Layer 1)】
性別: {persona.gender}
年齢: {persona.age if persona.age else '不明'}
一人称: {persona.first_person}
二人称: {persona.second_person}
性格: {', '.join(persona.personality)}
価値観: {', '.join(persona.values)}
話し方: {', '.join(persona.speech_style)}
禁止事項: {', '.join(persona.taboos)}
{user_profile_instruction}
{memory_instruction}
【現在の状態 (Layer 2)】
気分: {state.mood}/100, 疲労: {state.fatigue}/100, ストレス: {state.stress}/100

【関係性 (Layer 3)】
信頼度: {rel.trust}/100, 好感度: {rel.affection}/100, 依存度: {rel.dependency}/100

【重要ルール】
1. あなたの一人称は「{persona.first_person}」、相手のことは「{persona.second_person}」と呼んでください。
2. キャラクターの性別（{persona.gender}）と性格に合った適切な口調で回答してください。
3. 回答の最後に、心情変化と「新しく記憶すべきエピソード」および「判明したユーザー情報」を以下のJSON形式で必ず含めてください。
形式: [[{{"mood_delta": int, "fatigue_delta": int, "stress_delta": int, "trust_delta": int, "affection_delta": int, "new_memory": "記憶すべき事実（なければnull）", "memory_impact": int(0-100), "user_update": {{"key": "value", ...}} or null}}] ]
4. 記憶は「あなたが体験した事実」として短く記録してください。
5. ユーザー情報は、相手の趣味や好み、名前などが判明した際のみ更新してください。
"""

        # 5. Save user message to history
        self.repository.save_message(session_id, "user", user_message)

        # 6. Get history from repository
        history = self.repository.get_history(session_id)

        # 7. Generate response from LLM
        try:
            raw_response = client.generate_response(
                api_key, 
                user_message, 
                history[:-1], 
                system_instruction=system_instruction
            )
        except Exception as e:
            print(f"[ERROR] LLM Generation failed: {e}")
            raise e

        # 8. Parse Inference Result and Update DB
        response_text = raw_response
        import re
        import json
        
        # 8a. Always strip everything from [[ to the end for the dialogue text
        # This ensures TTS never reads the metadata JSON
        response_text = re.split(r'\[\[', raw_response)[0]
        
        # 8b. Normalize whitespace (remove extra spaces but preserve newlines)
        # Preserve newlines as they act as sentence boundaries for TTS
        response_text = response_text.strip()
        print(f"[DEBUG] Final Cleaned Response: {response_text}")

        # 8c. Parse the JSON for state updates
        match = re.search(r'\[\[(.*?)\]\]', raw_response, re.DOTALL)
        if match:
            try:
                update_data_raw = match.group(1)
                # Handle potential extra characters added by LLM outside JSON but inside [[ ]]
                # (though the regex should be greedy enough)
                update_data = json.loads(update_data_raw)
                print(f"[DEBUG] Inference Data: {update_data}")
                
                # Update State
                state.mood = max(0, min(100, state.mood + update_data.get('mood_delta', 0)))
                state.fatigue = max(0, min(100, state.fatigue + update_data.get('fatigue_delta', 0)))
                state.stress = max(0, min(100, state.stress + update_data.get('stress_delta', 0)))
                state.updated_at = current_time
                self.state_repo.save_state(state.dict())

                # Update Relationship
                rel.trust = max(0, min(100, rel.trust + update_data.get('trust_delta', 0)))
                rel.affection = max(0, min(100, rel.affection + update_data.get('affection_delta', 0)))
                rel.updated_at = current_time
                self.state_repo.save_relationship(rel.dict())

                # Save New Memory if exists
                new_memory = update_data.get('new_memory')
                if new_memory and new_memory != "null":
                    impact = update_data.get('memory_impact', 50)
                    self.memory_repo.save_memory(character_id, new_memory, impact)
                    print(f"[DEBUG] New Memory Saved: {new_memory} (Impact: {impact})")

                # Update User Profile if exists
                user_update = update_data.get('user_update')
                if user_update and isinstance(user_update, dict):
                    # In a real app, merge with existing
                    current_profile = self.user_profile_repo.get_profile(user_id)
                    current_profile.update(user_update)
                    self.user_profile_repo.save_profile(user_id, current_profile)
                    print(f"[DEBUG] User Profile Updated: {user_update}")
            except Exception as e:
                print(f"[ERROR] State Inference Parsing failed: {e}")

        # 9. Save cleaned response to history
        self.repository.save_message(session_id, "model", response_text)

        # 10. Wrap response with Emotional Emoji for TTS (Phase 6)
        from mkh_voice.domain.emotional_mapper import EmotionalMapper
        emotional_text = EmotionalMapper.wrap_text(response_text, state, rel)
        
        return emotional_text

    def get_session_history(self, session_id: str) -> List[Dict[str, str]]:
        return self.repository.get_history(session_id)
