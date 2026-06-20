// hi_aura - Sherpa-ONNX 关键词检测程序(单音频检测)
//
// 专门用于检测单个 WAV 音频文件中的关键词:
//   - 你好阿若 / aura / hi阿若 / hi aura
//   - 欧若 / 你好欧若
//   - 欧ro / 欧ra / 欧ru
//   - 啊ro / 啊ra
//   - hi欧ro / hi欧ra / hi欧ru / hi啊ro / hi啊ra
//
// 用法:
//   hi_aura.exe <wav_file>              # 检测指定 WAV 文件
//   hi_aura.exe                         # 检测 test_wavs 目录下所有 WAV
//   hi_aura.exe --keywords <file>       # 指定关键词文件
//   hi_aura.exe --model-dir <dir>       # 指定模型目录
//   hi_aura.exe --debug 1              # 显示调试信息
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	sherpa "github.com/k2-fsa/sherpa-onnx-go-windows"
)

// buildKWSConfig 构建关键词检测器配置
func buildKWSConfig(modelDir, keywordsFile string, numThreads int, debug int) *sherpa.KeywordSpotterConfig {
	cfg := &sherpa.KeywordSpotterConfig{
		FeatConfig: sherpa.FeatureConfig{
			SampleRate: 16000,
			FeatureDim: 80,
		},
		ModelConfig: sherpa.OnlineModelConfig{
			Transducer: sherpa.OnlineTransducerModelConfig{
				Encoder: filepath.Join(modelDir, "encoder-epoch-13-avg-2-chunk-16-left-64.int8.onnx"),
				Decoder: filepath.Join(modelDir, "decoder-epoch-13-avg-2-chunk-16-left-64.onnx"),
				Joiner:  filepath.Join(modelDir, "joiner-epoch-13-avg-2-chunk-16-left-64.int8.onnx"),
			},
			Tokens:     filepath.Join(modelDir, "tokens.txt"),
			NumThreads: numThreads,
			Provider:   "cpu",
			Debug:      debug,
		},
		MaxActivePaths:    4,
		KeywordsFile:      keywordsFile,
		KeywordsScore:     1.5,
		KeywordsThreshold: 0.05,
	}
	return cfg
}

// detectFromFile 从单个 WAV 文件中检测关键词
func detectFromFile(spotter *sherpa.KeywordSpotter, wavFile string) []string {
	fmt.Printf("\n[处理文件] %s\n", wavFile)

	wave := sherpa.ReadWave(wavFile)
	if wave == nil {
		fmt.Printf("  错误: 无法读取 WAV 文件: %s\n", wavFile)
		return nil
	}
	fmt.Printf("  采样率: %d Hz, 样本数: %d, 时长: %.2f 秒\n",
		wave.SampleRate, len(wave.Samples), float32(len(wave.Samples))/float32(wave.SampleRate))

	stream := sherpa.NewKeywordStream(spotter)
	defer sherpa.DeleteOnlineStream(stream)

	// 模拟流式输入: 每次送入 0.1 秒的音频
	chunkSize := wave.SampleRate / 10 // 100ms
	samples := wave.Samples
	detectedKeywords := []string{}

	for i := 0; i < len(samples); i += chunkSize {
		end := i + chunkSize
		if end > len(samples) {
			end = len(samples)
		}
		chunk := samples[i:end]
		stream.AcceptWaveform(wave.SampleRate, chunk)

		// 解码
		for spotter.IsReady(stream) {
			spotter.Decode(stream)
		}

		// 获取结果
		result := spotter.GetResult(stream)
		if result.Keyword != "" {
			fmt.Printf("  [检测到关键词] %s\n", result.Keyword)
			detectedKeywords = append(detectedKeywords, result.Keyword)
			spotter.Reset(stream)
		}
	}

	// 输入完成,刷新缓冲区
	stream.InputFinished()
	for spotter.IsReady(stream) {
		spotter.Decode(stream)
	}
	result := spotter.GetResult(stream)
	if result.Keyword != "" {
		fmt.Printf("  [检测到关键词] %s\n", result.Keyword)
		detectedKeywords = append(detectedKeywords, result.Keyword)
	}

	if len(detectedKeywords) == 0 {
		fmt.Printf("  未检测到关键词\n")
	} else {
		fmt.Printf("  共检测到 %d 个关键词: %v\n", len(detectedKeywords), detectedKeywords)
	}
	return detectedKeywords
}

func main() {
	modelDir := flag.String("model-dir", "model", "模型目录路径")
	keywordsFile := flag.String("keywords", "keywords.txt", "关键词文件路径")
	numThreads := flag.Int("num-threads", 2, "推理线程数")
	debug := flag.Int("debug", 0, "是否显示模型加载调试信息 (0 或 1)")
	flag.Parse()

	// 获取 WAV 文件路径(位置参数)
	args := flag.Args()

	fmt.Println("========================================")
	fmt.Println("  hi_aura - Sherpa-ONNX 关键词检测")
	fmt.Println("========================================")
	fmt.Printf("模型目录: %s\n", *modelDir)
	fmt.Printf("关键词文件: %s\n", *keywordsFile)
	fmt.Printf("线程数: %d\n", *numThreads)

	// 检查模型目录
	if _, err := os.Stat(*modelDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 模型目录不存在: %s\n", *modelDir)
		os.Exit(1)
	}

	// 检查关键词文件
	if _, err := os.Stat(*keywordsFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 关键词文件不存在: %s\n", *keywordsFile)
		os.Exit(1)
	}

	// 读取并显示关键词
	fmt.Println("\n[关键词列表]")
	content, err := os.ReadFile(*keywordsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法读取关键词文件: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(content))

	// 创建关键词检测器
	fmt.Println("[加载模型中...]")
	startTime := time.Now()
	config := buildKWSConfig(*modelDir, *keywordsFile, *numThreads, *debug)
	spotter := sherpa.NewKeywordSpotter(config)
	if spotter == nil {
		fmt.Fprintf(os.Stderr, "错误: 无法创建关键词检测器\n")
		os.Exit(1)
	}
	defer sherpa.DeleteKeywordSpotter(spotter)
	fmt.Printf("[模型加载完成, 耗时: %.2f 秒]\n", time.Since(startTime).Seconds())

	// 处理音频文件
	if len(args) > 0 {
		// 检测指定的单个 WAV 文件
		wavFile := args[0]
		if _, err := os.Stat(wavFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "错误: WAV 文件不存在: %s\n", wavFile)
			os.Exit(1)
		}
		detectFromFile(spotter, wavFile)
	} else {
		// 检测 test_wavs 目录下所有 WAV 文件
		testDir := "test_wavs"
		files, err := filepath.Glob(filepath.Join(testDir, "*.wav"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无法查找 WAV 文件: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Printf("未找到 WAV 文件: %s\n", testDir)
			fmt.Println("用法: hi_aura.exe <wav_file>  或将 WAV 文件放入 test_wavs 目录")
			os.Exit(0)
		}

		fmt.Printf("\n[开始检测, 共 %d 个 WAV 文件]\n", len(files))
		totalDetected := 0
		for _, f := range files {
			keywords := detectFromFile(spotter, f)
			totalDetected += len(keywords)
		}
		fmt.Printf("\n[总计检测到 %d 个关键词]\n", totalDetected)
	}

	fmt.Println("\n========================================")
	fmt.Println("  检测完成")
	fmt.Println("========================================")
}
