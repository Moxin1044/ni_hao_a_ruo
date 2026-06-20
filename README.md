# hi_aura - Sherpa-ONNX 关键词检测程序

基于 [Sherpa-ONNX](https://github.com/k2-fsa/sherpa-onnx) 的中文/英文关键词唤醒(KWS)程序,使用 Go 语言开发,支持对单个或多个 WAV 音频文件进行关键词检测。

## 功能特性

- 支持中英文混合关键词检测
- 基于 Zipformer 流式模型,实时低延迟
- 单音频文件检测 / 批量检测两种模式
- 本地运行,无需联网
- 跨平台支持(Windows / Linux / macOS)

## 已支持的关键词

| 关键词 | 说明 |
|--------|------|
| 你好阿若 | 中文唤醒词 |
| aura | 英文唤醒词 |
| hi阿若 | 中英混合 |
| hi aura | 纯英文 |
| 欧若 | 中文 |
| 你好欧若 | 中文 |
| 欧ro / 欧ra / 欧ru | 中英混合 |
| 啊ro / 啊ra | 中英混合 |
| hi欧ro / hi欧ra / hi欧ru | 中英混合 |
| hi啊ro / hi啊ra | 中英混合 |

## 目录结构

```
hi_aura/
├── main.go                 # 主程序源码
├── go.mod                  # Go 模块定义
├── go.sum                   # 依赖校验
├── keywords.txt            # 关键词文件(音素/拼音序列)
├── keywords_raw.txt        # 关键词原始文本(用于生成 keywords.txt)
├── hi_aura.exe             # 编译后的可执行文件
├── onnxruntime.dll         # ONNX Runtime 动态库
├── sherpa-onnx-c-api.dll   # Sherpa-ONNX C API
├── sherpa-onnx-cxx-api.dll # Sherpa-ONNX C++ API
├── model/                  # 预训练模型
│   ├── encoder-epoch-13-avg-2-chunk-16-left-64.int8.onnx
│   ├── decoder-epoch-13-avg-2-chunk-16-left-64.onnx
│   ├── joiner-epoch-13-avg-2-chunk-16-left-64.int8.onnx
│   ├── tokens.txt          # 模型 token 表
│   └── en.phone            # 英文音素词典
└── work/
    └── test_wavs/          # 测试音频目录
        ├── test_ni_hao_aruo.wav
        ├── test_aura.wav
        └── ...
```

## 环境要求

### 运行环境
- Windows 10/11 (64位)
- 或 Linux / macOS(需替换对应的动态库)

### 开发环境
- Go 1.24+
- 依赖:`github.com/k2-fsa/sherpa-onnx-go-windows v1.13.2`

### 生成测试音频(可选)
- Python 3.8+
- 依赖包:`edge-tts`, `imageio-ffmpeg`, `soundfile`

### 生成关键词文件(可选)
- Python 3.8+
- 依赖包:`sentencepiece`, `pypinyin`
- Sherpa-ONNX 命令行工具 `sherpa-onnx-cli`

## 快速开始

### 1. 直接运行

```bash
# 进入项目目录
cd hi_aura

# 检测单个 WAV 文件
./hi_aura.exe work/test_wavs/test_ni_hao_aruo.wav

# 检测 test_wavs 目录下所有 WAV 文件
./hi_aura.exe
```

### 2. 编译运行

```bash
cd hi_aura
go build -o hi_aura.exe main.go
./hi_aura.exe work/test_wavs/test_aura.wav
```

### 3. 命令行参数

```
用法:
  hi_aura.exe [选项] [wav_file]

选项:
  -model-dir string    模型目录路径 (默认 "model")
  -keywords string     关键词文件路径 (默认 "keywords.txt")
  -num-threads int     推理线程数 (默认 2)
  -debug int           显示调试信息 (0 或 1, 默认 0)

示例:
  hi_aura.exe work/test_wavs/test_hi_aura.wav
  hi_aura.exe --model-dir model --keywords keywords.txt
  hi_aura.exe --debug 1 work/test_wavs/test_ni_hao_aruo.wav
  hi_aura.exe --num-threads 4
```

## 如何重新训练 / 新增关键词

本程序**不需要重新训练模型**,只需修改 `keywords.txt` 关键词文件即可新增或修改关键词。
预训练模型 `sherpa-onnx-kws-zipformer-zh-en-3M-2025-12-20` 已支持中英文识别,
通过关键词文件指定要检测的具体词即可。

### 关键词文件格式说明

`keywords.txt` 每行一个关键词,格式:

```
<音素/拼音序列> @<关键词名称>
```

示例:
```
n ǐ h ǎo ā r uò @你好阿若
AO1 R AH0 @aura
HH AY1 AO1 R AH0 @hi_aura
```

- **中文关键词**:使用带声调的拼音,音节之间用空格分隔
  - 例:`n ǐ h ǎo ā r uò` 表示 "你好阿若"
- **英文关键词**:使用 [ARPABET](https://en.wikipedia.org/wiki/ARPABET) 音素,大写
  - 例:`AO1 R AH0` 表示 "aura"
  - `HH AY1` 表示 "hi"
- **`@` 后面**:是关键词的显示名称,检测到时会输出这个名称

### 新增关键词步骤

#### 方法一:使用 sherpa-onnx-cli 自动生成(推荐)

**1. 安装 Sherpa-ONNX 命令行工具**

从 [Sherpa-ONNX Releases](https://github.com/k2-fsa/sherpa-onnx/releases) 下载对应平台的预编译版本,
解压后将 `sherpa-onnx-cli` (Windows 下为 `sherpa-onnx-cli.exe`) 加入 PATH。

**2. 安装 Python 依赖**

```bash
pip install sentencepiece pypinyin
```

**3. 编辑 `keywords_raw.txt`**

每行格式:`<关键词文本> @<显示名称>`

```
你好阿若 @你好阿若
AURA @aura
HI AURA @hi_aura
欧若 @欧若
你好欧若 @你好欧若
```

> 注意:
> - 英文单词必须**大写**,以匹配词典中的写法(如 `AURA` 而非 `aura`)
> - 中文直接写汉字即可,工具会自动转换为带声调的拼音
> - `@` 前后空格可有可无

**4. 生成 `keywords.txt`**

```bash
sherpa-onnx-cli text2token \
  --tokens model/tokens.txt \
  --lexicon model/en.phone \
  keywords_raw.txt \
  keywords.txt
```

参数说明:
- `--tokens`:模型目录下的 `tokens.txt`
- `--lexicon`:英文音素词典 `en.phone`(用于英文单词转音素)
- `keywords_raw.txt`:输入的原始关键词文本
- `keywords.txt`:输出的关键词文件

**5. 重新运行程序验证**

```bash
./hi_aura.exe work/test_wavs/test_xxx.wav
```

#### 方法二:手动编辑 keywords.txt

对于简单关键词,可以直接手动编写音素序列。

**中文关键词**:查询汉字拼音(带声调),用空格分隔

```
# 格式: 声母 韵母声调 声母 韵母声调 ...
n ǐ h ǎo @你好
```

拼音声调标注:
- 1声:`ˉ`(或不标)  2声:`ˊ`  3声:`ˇ`  4声:`ˋ`
- 实际 tokens.txt 中使用的是带数字的声调,如 `ǐ` `ǎ` `ò` 等

**英文关键词**:查阅 ARPABET 音素表

常用音素对照:
| 单词 | 音素 |
|------|------|
| hi | HH AY1 |
| aura | AO1 R AH0 |
| hello | HH AH0 L OW1 |
| hey | HH EY1 |

组合示例:
```
HH AY1 AO1 R AH0 @hi_aura
```

### 生成测试音频

使用 `work/generate_test_audio.py` 生成测试音频(基于 edge-tts):

```bash
cd work
pip install edge-tts imageio-ffmpeg soundfile
python generate_test_audio.py
```

编辑 `generate_test_audio.py` 中的 `KEYWORDS` 和 `SENTENCES` 列表可新增测试音频:

```python
KEYWORDS = [
    ("你好阿若", "zh-CN-XiaoxiaoNeural", "test_ni_hao_aruo.wav"),
    ("aura", "en-US-AriaNeural", "test_aura.wav"),
    # 新增关键词:
    ("新关键词", "zh-CN-XiaoxiaoNeural", "test_new_keyword.wav"),
]
```

语音选择建议:
- 中文关键词:使用 `zh-CN-XiaoxiaoNeural`
- 英文关键词:使用 `en-US-AriaNeural`
- 中英混合关键词:根据主要语言选择,或分别生成测试

### 调整检测灵敏度

如果关键词检测不到,或误检过多,可调整 `main.go` 中的参数:

```go
cfg := &sherpa.KeywordSpotterConfig{
    // ...
    KeywordsScore:     1.5,   // 关键词加分,越大越容易检测到(范围 0~5)
    KeywordsThreshold: 0.05,  // 检测阈值,越小越敏感(范围 -∞~+∞,通常 0~1)
    MaxActivePaths:    4,     // 搜索路径数,越大越准确但越慢
}
```

- **检测不到关键词** → 提高 `KeywordsScore`(如 2.0)或降低 `KeywordsThreshold`(如 0.01)
- **误检过多** → 降低 `KeywordsScore`(如 1.0)或提高 `KeywordsThreshold`(如 0.5)

修改后需重新编译:
```bash
go build -o hi_aura.exe main.go
```

## 常见问题

### Q1: 运行报错 "找不到 onnxruntime.dll"

确保以下 DLL 文件与 `hi_aura.exe` 在同一目录:
- `onnxruntime.dll`
- `sherpa-onnx-c-api.dll`
- `sherpa-onnx-cxx-api.dll`

### Q2: 运行报错 "模型目录不存在"

使用 `--model-dir` 指定正确的模型目录:
```bash
./hi_aura.exe --model-dir path/to/model work/test_wavs/test.wav
```

### Q3: 关键词检测不到

1. 检查 `keywords.txt` 中的音素序列是否正确
2. 检查音频格式:需为 16kHz、单声道、16-bit PCM WAV
3. 调整 `KeywordsScore` 和 `KeywordsThreshold` 参数
4. 使用 `--debug 1` 查看模型加载信息

### Q4: text2token 报错 "No module named 'sentencepiece'"

```bash
pip install sentencepiece pypinyin
```

### Q5: 英文关键词在 text2token 中被跳过

英文单词在 `keywords_raw.txt` 中必须**大写**,以匹配 `en.phone` 词典:
```
# 错误
aura @aura

# 正确
AURA @aura
```

### Q6: 中英混合关键词检测不准

中英混合关键词(如 "hi阿若")的声学特征可能与其他词相近,可尝试:
1. 调整音素序列,确保发音准确
2. 使用中文 TTS 生成测试音频(中文发音的 "hi" 更接近实际使用场景)
3. 调整 `KeywordsThreshold` 降低阈值

## 技术参考

- [Sherpa-ONNX 官方文档](https://k2-fsa.github.io/sherpa/onnx/)
- [Sherpa-ONNX GitHub](https://github.com/k2-fsa/sherpa-onnx)
- [预训练模型列表](https://k2-fsa.github.io/sherpa/onnx/pretrained_models/index.html)
- [关键词唤醒文档](https://k2-fsa.github.io/sherpa/onnx/kws/pretrained_models/index.html)
- [ARPABET 音素表](https://en.wikipedia.org/wiki/ARPABET)

## 许可证

本项目使用的 Sherpa-ONNX 采用 Apache 2.0 许可证。
预训练模型 `sherpa-onnx-kws-zipformer-zh-en-3M-2025-12-20` 来自 Sherpa-ONNX 项目。
