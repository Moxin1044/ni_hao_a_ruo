"""Generate test WAV files for custom keywords using edge-tts."""
import asyncio
import os
import subprocess
import sys

import edge_tts
import imageio_ffmpeg
import soundfile as sf

# Keywords to generate
KEYWORDS = [
    ("你好阿若", "zh-CN-XiaoxiaoNeural", "test_ni_hao_aruo.wav"),
    ("aura", "en-US-AriaNeural", "test_aura.wav"),
    ("hi阿若", "zh-CN-XiaoxiaoNeural", "test_hi_aruo.wav"),
    ("hi aura", "en-US-AriaNeural", "test_hi_aura.wav"),
    ("欧若", "zh-CN-XiaoxiaoNeural", "test_ou_ruo.wav"),
    ("oh ro", "en-US-AriaNeural", "test_ou_ro.wav"),
    ("oh ra", "en-US-AriaNeural", "test_ou_ra.wav"),
    ("欧如", "zh-CN-XiaoxiaoNeural", "test_ou_ru.wav"),
    ("ah ro", "en-US-AriaNeural", "test_a_ro.wav"),
    ("ah ra", "en-US-AriaNeural", "test_a_ra.wav"),
]

# Sentences containing keywords for more realistic testing
SENTENCES = [
    ("你好阿若,今天天气怎么样?", "zh-CN-XiaoxiaoNeural", "test_sentence_ni_hao_aruo.wav"),
    ("The aura of the place was amazing.", "en-US-AriaNeural", "test_sentence_aura.wav"),
    ("hi阿若,你好吗?", "zh-CN-XiaoxiaoNeural", "test_sentence_hi_aruo.wav"),
    ("Hi aura, how are you today?", "en-US-AriaNeural", "test_sentence_hi_aura.wav"),
    ("欧若,你好吗?", "zh-CN-XiaoxiaoNeural", "test_sentence_ou_ruo.wav"),
]


async def generate_tts(text: str, voice: str, mp3_path: str):
    """Generate MP3 using edge-tts."""
    communicate = edge_tts.Communicate(text, voice)
    await communicate.save(mp3_path)
    print(f"  Generated MP3: {mp3_path}")


def mp3_to_wav(mp3_path: str, wav_path: str):
    """Convert MP3 to WAV (16kHz, mono, 16-bit) using ffmpeg."""
    ffmpeg = imageio_ffmpeg.get_ffmpeg_exe()
    cmd = [
        ffmpeg,
        "-y",  # overwrite
        "-i", mp3_path,
        "-ar", "16000",  # 16kHz
        "-ac", "1",  # mono
        "-sample_fmt", "s16",  # 16-bit
        wav_path,
    ]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"  Error converting {mp3_path}: {result.stderr}", file=sys.stderr)
        return False
    print(f"  Converted to WAV: {wav_path}")
    return True


def verify_wav(wav_path: str):
    """Verify WAV file format."""
    data, sr = sf.read(wav_path)
    duration = len(data) / sr
    print(f"  Verified: {wav_path} - SR={sr}, samples={len(data)}, duration={duration:.2f}s")
    return True


async def main():
    output_dir = "test_wavs_custom"
    os.makedirs(output_dir, exist_ok=True)

    print("=" * 50)
    print("  生成测试音频文件")
    print("=" * 50)

    # Generate keyword audio files
    print("\n[生成关键词音频]")
    for text, voice, filename in KEYWORDS:
        print(f"\n关键词: '{text}' (voice: {voice})")
        mp3_path = os.path.join(output_dir, filename.replace(".wav", ".mp3"))
        wav_path = os.path.join(output_dir, filename)
        await generate_tts(text, voice, mp3_path)
        if mp3_to_wav(mp3_path, wav_path):
            verify_wav(wav_path)
            os.remove(mp3_path)

    # Generate sentence audio files
    print("\n[生成句子音频]")
    for text, voice, filename in SENTENCES:
        print(f"\n句子: '{text}' (voice: {voice})")
        mp3_path = os.path.join(output_dir, filename.replace(".wav", ".mp3"))
        wav_path = os.path.join(output_dir, filename)
        await generate_tts(text, voice, mp3_path)
        if mp3_to_wav(mp3_path, wav_path):
            verify_wav(wav_path)
            os.remove(mp3_path)

    print("\n" + "=" * 50)
    print("  音频生成完成!")
    print(f"  输出目录: {output_dir}")
    print("=" * 50)


if __name__ == "__main__":
    asyncio.run(main())
