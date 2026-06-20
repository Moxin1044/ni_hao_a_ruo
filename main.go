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
		KeywordsThreshold: 0.1,
	}
	return cfg
}

// detectFromFile 从 WAV 文件中检测关键词
func detectFromFile(spotter *sherpa.KeywordSpotter, wavFile string) {
	fmt.Printf("\n[处理文件] %s\n", wavFile)
	wave := sherpa.ReadWave(wavFile)
	if wave == nil {
		fmt.Printf("  错误: 无法读取 WAV 文件: %s\n", wavFile)
		return
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
}

func main() {
	modelDir := flag.String("model-dir", "sherpa-onnx-kws-zipformer-zh-en-3M-2025-12-20", "模型目录路径")
	keywordsFile := flag.String("keywords-file", "keywords.txt", "关键词文件路径")
	wavFile := flag.String("wav-file", "", "单个 WAV 文件路径(可选)")
	testWavsDir := flag.String("test-wavs", "", "测试 WAV 文件目录(可选,默认使用模型自带的 test_wavs)")
	numThreads := flag.Int("num-threads", 2, "推理线程数")
	debug := flag.Int("debug", 0, "是否显示模型加载调试信息 (0 或 1)")
	flag.Parse()

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

	fmt.Println("========================================")
	fmt.Println("  Sherpa-ONNX KWS 关键词检测演示 (Go)")
	fmt.Println("========================================")
	fmt.Printf("模型目录: %s\n", *modelDir)
	fmt.Printf("关键词文件: %s\n", *keywordsFile)
	fmt.Printf("线程数: %d\n", *numThreads)

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
	if *wavFile != "" {
		// 处理单个文件
		detectFromFile(spotter, *wavFile)
	} else {
		// 处理测试目录中的所有 WAV 文件
		testDir := *testWavsDir
		if testDir == "" {
			testDir = filepath.Join(*modelDir, "test_wavs")
		}

		files, err := filepath.Glob(filepath.Join(testDir, "*.wav"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无法查找 WAV 文件: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Printf("未找到 WAV 文件: %s\n", testDir)
			os.Exit(0)
		}

		fmt.Printf("\n[开始检测, 共 %d 个 WAV 文件]\n", len(files))
		for _, f := range files {
			detectFromFile(spotter, f)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("  检测完成")
	fmt.Println("========================================")
}
