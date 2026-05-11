from sqlalchemy import create_engine, Column, String, DateTime, Text, ForeignKey, Integer
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from datetime import datetime
import os
from mkh_voice.domain.ports import ChatRepositoryPort

Base = declarative_base()

class SessionModel(Base):
    __tablename__ = 'sessions'
    id = Column(String, primary_key=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    messages = relationship("MessageModel", back_populates="session", cascade="all, delete-orphan")

class MessageModel(Base):
    __tablename__ = 'messages'
    id = Column(String, primary_key=True) # UUID as string
    session_id = Column(String, ForeignKey('sessions.id'))
    role = Column(String)
    content = Column(Text)
    timestamp = Column(DateTime, default=datetime.utcnow)
    session = relationship("SessionModel", back_populates="messages")

class StateModel(Base):
    __tablename__ = 'character_states'
    character_id = Column(String, primary_key=True)
    mood = Column(Integer, default=50)
    fatigue = Column(Integer, default=0)
    stress = Column(Integer, default=0)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

class RelationshipModel(Base):
    __tablename__ = 'relationship_states'
    character_id = Column(String, primary_key=True)
    user_id = Column(String, primary_key=True, default="default_user")
    trust = Column(Integer, default=0)
    affection = Column(Integer, default=0)
    dependency = Column(Integer, default=0)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

class MemoryModel(Base):
    __tablename__ = 'character_memories'
    id = Column(String, primary_key=True)
    character_id = Column(String, index=True)
    content = Column(Text)
    emotional_impact = Column(Integer) # 0-100
    timestamp = Column(DateTime, default=datetime.utcnow)

class UserProfileModel(Base):
    __tablename__ = 'user_profiles'
    user_id = Column(String, primary_key=True)
    profile_json = Column(Text, default="{}")

from mkh_voice.domain.ports import MemoryRepositoryPort, UserProfilePort

class SQLiteChatRepository(ChatRepositoryPort, MemoryRepositoryPort, UserProfilePort):
    def __init__(self, db_path: str = "data/chat.db"):
        os.makedirs(os.path.dirname(db_path), exist_ok=True)
        self.engine = create_engine(f"sqlite:///{db_path}")
        try:
            Base.metadata.create_all(self.engine)
        except Exception as e:
            # Catch OperationalError if table already exists or other init issues
            print(f"[INFO] Database initialization note: {e}")
        self.Session = sessionmaker(bind=self.engine)

    def create_session(self, session_id: str) -> None:
        with self.Session() as session:
            # Check if exists (idempotent)
            exists = session.query(SessionModel).filter_by(id=session_id).first()
            if exists:
                return
            new_session = SessionModel(id=session_id)
            session.add(new_session)
            session.commit()

    def save_message(self, session_id: str, role: str, content: str) -> None:
        import uuid
        with self.Session() as session:
            new_message = MessageModel(
                id=str(uuid.uuid4()),
                session_id=session_id,
                role=role,
                content=content
            )
            session.add(new_message)
            session.commit()

    def get_history(self, session_id: str) -> list:
        with self.Session() as session:
            messages = session.query(MessageModel).filter_by(session_id=session_id).order_by(MessageModel.timestamp).all()
            return [{"role": m.role, "content": m.content} for m in messages]

    # Character Mind Methods
    def get_state(self, character_id: str) -> dict:
        with self.Session() as session:
            state = session.query(StateModel).filter_by(character_id=character_id).first()
            if not state:
                return {"character_id": character_id, "mood": 50, "fatigue": 0, "stress": 0, "updated_at": datetime.utcnow()}
            return {
                "character_id": state.character_id, 
                "mood": state.mood, 
                "fatigue": state.fatigue, 
                "stress": state.stress,
                "updated_at": state.updated_at
            }

    def save_state(self, state_data: dict) -> None:
        with self.Session() as session:
            state = session.query(StateModel).filter_by(character_id=state_data['character_id']).first()
            if not state:
                state = StateModel(character_id=state_data['character_id'])
                session.add(state)
            state.mood = state_data['mood']
            state.fatigue = state_data['fatigue']
            state.stress = state_data['stress']
            state.updated_at = state_data.get('updated_at', datetime.utcnow())
            session.commit()

    def get_relationship(self, character_id: str, user_id: str) -> dict:
        with self.Session() as session:
            rel = session.query(RelationshipModel).filter_by(character_id=character_id, user_id=user_id).first()
            if not rel:
                return {"character_id": character_id, "user_id": user_id, "trust": 0, "affection": 0, "dependency": 0, "updated_at": datetime.utcnow()}
            return {
                "character_id": rel.character_id, 
                "user_id": rel.user_id, 
                "trust": rel.trust, 
                "affection": rel.affection, 
                "dependency": rel.dependency,
                "updated_at": rel.updated_at
            }

    def save_relationship(self, rel_data: dict) -> None:
        with self.Session() as session:
            rel = session.query(RelationshipModel).filter_by(character_id=rel_data['character_id'], user_id=rel_data['user_id']).first()
            if not rel:
                rel = RelationshipModel(character_id=rel_data['character_id'], user_id=rel_data['user_id'])
                session.add(rel)
            rel.trust = rel_data['trust']
            rel.affection = rel_data['affection']
            rel.dependency = rel_data['dependency']
            rel.updated_at = rel_data.get('updated_at', datetime.utcnow())
            session.commit()

    # Episodic Memory Methods
    def save_memory(self, character_id: str, content: str, emotional_impact: int) -> None:
        import uuid
        with self.Session() as session:
            new_memory = MemoryModel(
                id=str(uuid.uuid4()),
                character_id=character_id,
                content=content,
                emotional_impact=emotional_impact
            )
            session.add(new_memory)
            session.commit()

    def search_memories(self, character_id: str, query: str, limit: int = 5) -> list:
        import math
        with self.Session() as session:
            # 1. Fetch candidates using keyword search
            # We fetch more than limit to allow re-ranking by decay
            memories = session.query(MemoryModel).filter(
                MemoryModel.character_id == character_id,
                MemoryModel.content.like(f"%{query}%")
            ).all()
            
            scored_memories = []
            now = datetime.utcnow()
            
            for m in memories:
                # Calculate time passed in days
                days_passed = (now - m.timestamp).total_seconds() / 86400.0
                
                # Human-like forgetting: Emotional memories persist longer
                # If impact=100, decay is very slow (0.01). If impact=0, decay is fast (0.5).
                decay_rate = 0.5 * (1.0 - (m.emotional_impact / 100.0)) + 0.01
                
                # Score = Emotional Impact * e^(-decay_rate * days_passed)
                # This means high-impact memories stay relevant longer even if old.
                score = m.emotional_impact * math.exp(-decay_rate * days_passed)
                
                scored_memories.append({
                    "content": m.content, 
                    "impact": m.emotional_impact, 
                    "timestamp": m.timestamp,
                    "score": score
                })
            
            # 2. Sort by computed score
            scored_memories.sort(key=lambda x: x['score'], reverse=True)
            
            return scored_memories[:limit]

    # User Profile Methods
    def get_profile(self, user_id: str) -> dict:
        import json
        with self.Session() as session:
            profile = session.query(UserProfileModel).filter_by(user_id=user_id).first()
            if not profile:
                return {}
            return json.loads(profile.profile_json)

    def save_profile(self, user_id: str, profile_data: dict) -> None:
        import json
        with self.Session() as session:
            profile = session.query(UserProfileModel).filter_by(user_id=user_id).first()
            if not profile:
                profile = UserProfileModel(user_id=user_id)
                session.add(profile)
            profile.profile_json = json.dumps(profile_data)
            session.commit()
