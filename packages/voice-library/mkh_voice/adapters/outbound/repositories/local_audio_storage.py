import os
import io
import soundfile as sf
import numpy as np
from mkh_voice.domain.ports import AudioStoragePort

class LocalAudioStorage(AudioStoragePort):
    """
    Adapter for storing audio files locally on the file system.
    """
    def __init__(self, base_dir: str = "data/references"):
        self.base_dir = base_dir
        os.makedirs(self.base_dir, exist_ok=True)

    def save_audio(self, identifier: str, audio_bytes: bytes) -> str:
        """
        Saves the audio to a local file.
        Returns the path relative to the current working directory.
        """
        file_path = os.path.join(self.base_dir, f"{identifier}.wav")
        with open(file_path, "wb") as f:
            f.write(audio_bytes)
        return file_path

    def append_audio(self, identifier: str, audio_bytes: bytes) -> str:
        """
        Appends new audio bytes to an existing WAV file.
        """
        file_path = os.path.join(self.base_dir, f"{identifier}.wav")
        if not os.path.exists(file_path):
            return self.save_audio(identifier, audio_bytes)
            
        # Read existing audio
        data_existing, samplerate = sf.read(file_path)
        
        # Read new audio from bytes
        new_io = io.BytesIO(audio_bytes)
        data_new, _ = sf.read(new_io)
        
        # Concatenate (NumPy arrays)
        combined = np.concatenate([data_existing, data_new])
        
        # Write back to the same file
        sf.write(file_path, combined, samplerate, format='WAV', subtype='PCM_16')
        return file_path
