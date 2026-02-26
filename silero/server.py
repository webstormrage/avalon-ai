import torch
import soundfile as sf
import subprocess
import os
import time
import logging
from fastapi import FastAPI, HTTPException
from fastapi.responses import FileResponse
from pydantic import BaseModel

# ---------------- Logging ----------------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)s | %(message)s",
    datefmt="%H:%M:%S"
)

# ---------------- Config ----------------
MEDIA_DIR = "./media"
os.makedirs(MEDIA_DIR, exist_ok=True)

sample_rate = 48000

# ---------------- App ----------------
app = FastAPI(title="Silero TTS Service")

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
logging.info(f"Using device: {device}")

logging.info("Loading model...")
model, _ = torch.hub.load(
    repo_or_dir="snakers4/silero-models",
    model="silero_tts",
    language="ru",
    speaker="v3_1_ru"
)
model.to(device)
logging.info("Model loaded")

# ---------------- Schema ----------------
class TTSRequest(BaseModel):
    filename: str
    text: str
    speaker: str = "kseniya"
    speed: float = 1.0
    pitch: float = 0.0

# ---------------- 1️⃣ Generate & Save ----------------
@app.post("/tts")
def generate_tts(req: TTSRequest):

    start = time.perf_counter()

    filename = os.path.basename(req.filename)
    if not filename.endswith(".wav"):
        filename += ".wav"

    raw_path = os.path.join(MEDIA_DIR, f"raw_{filename}")
    final_path = os.path.join(MEDIA_DIR, filename)

    try:
        audio = model.apply_tts(
            req.text,
            speaker=req.speaker,
            sample_rate=sample_rate
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

    synth_time = time.perf_counter()
    sf.write(raw_path, audio, sample_rate)

    # SoX processing
    if req.speed != 1.0 or req.pitch != 0.0:

        pitch_cents = int(req.pitch * 100)
        cmd = ["sox", raw_path, final_path]

        if req.pitch != 0.0:
            cmd += ["pitch", str(pitch_cents)]

        if req.speed != 1.0:
            cmd += ["tempo", str(req.speed)]

        subprocess.run(cmd, check=True)
        os.remove(raw_path)
    else:
        os.rename(raw_path, final_path)

    total_time = time.perf_counter() - start

    logging.info(
        f"Saved: {final_path} | synth: {synth_time-start:.2f}s | total: {total_time:.2f}s"
    )

    return {
        "status": "ok",
        "file": filename,
        "synthesis_time": round(synth_time - start, 3),
        "total_time": round(total_time, 3)
    }

# ---------------- 2️⃣ Download file ----------------
@app.get("/media/{filename}")
def get_media(filename: str):

    filename = os.path.basename(filename)
    file_path = os.path.join(MEDIA_DIR, filename)

    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail="File not found")

    return FileResponse(file_path, media_type="audio/wav")
