// Package search provides ONNX embedding support for code search.
package search

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	ort "github.com/yalue/onnxruntime_go"
)

// OnnxDim is the output dimension for all-MiniLM-L6-v2.
const OnnxDim = 384

// OnnxMaxTokens is the maximum input length for the model.
const OnnxMaxTokens = 128

// OnnxEmbedder wraps the ONNX model session for code embedding.
type OnnxEmbedder struct {
	session   *ort.AdvancedSession
	vocab     map[string]int64
	inputIDs  *ort.Tensor[int64]
	attMask   *ort.Tensor[int64]
	tokenType *ort.Tensor[int64]
	output    *ort.Tensor[float32]
	mu        sync.Mutex
}

var (
	globalOnnx     *OnnxEmbedder
	onnxOnce       sync.Once
	onnxErr        error
	onnxEnvStarted bool
)

// NewOnnxEmbedder returns the singleton ONNX embedder.
// Returns nil (no error) if the model or runtime is not available.
func NewOnnxEmbedder() (*OnnxEmbedder, error) {
	onnxOnce.Do(func() {
		modelsDir := os.Getenv("AGK_MODELS_DIR")
		if modelsDir == "" {
			home, _ := os.UserHomeDir()
			modelsDir = filepath.Join(home, ".agk", "models")
		}

		modelPath := filepath.Join(modelsDir, "all-MiniLM-L6-v2.onnx")
		vocabPath := filepath.Join(modelsDir, "vocab.txt")

		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			onnxErr = fmt.Errorf("ONNX model not found at %s", modelPath)
			return
		}
		if _, err := os.Stat(vocabPath); os.IsNotExist(err) {
			onnxErr = fmt.Errorf("vocab.txt not found at %s", vocabPath)
			return
		}

		vocab, err := loadOnnxVocab(vocabPath)
		if err != nil {
			onnxErr = fmt.Errorf("failed to load vocab: %w", err)
			return
		}

		if libPath := os.Getenv("ONNXRUNTIME_LIB"); libPath != "" {
			ort.SetSharedLibraryPath(libPath)
		}

		if !onnxEnvStarted {
			if err := ort.InitializeEnvironment(); err != nil {
				onnxErr = fmt.Errorf("ONNX runtime init failed: %w", err)
				return
			}
			onnxEnvStarted = true
		}

		seqLen := int64(OnnxMaxTokens)
		shape := ort.NewShape(1, seqLen)

		inputIDs, err := ort.NewEmptyTensor[int64](shape)
		if err != nil {
			onnxErr = err
			return
		}
		attMask, err := ort.NewEmptyTensor[int64](shape)
		if err != nil {
			inputIDs.Destroy()
			onnxErr = err
			return
		}
		tokenType, err := ort.NewEmptyTensor[int64](shape)
		if err != nil {
			inputIDs.Destroy()
			attMask.Destroy()
			onnxErr = err
			return
		}

		outputShape := ort.NewShape(1, seqLen, int64(OnnxDim))
		output, err := ort.NewEmptyTensor[float32](outputShape)
		if err != nil {
			inputIDs.Destroy()
			attMask.Destroy()
			tokenType.Destroy()
			onnxErr = err
			return
		}

		session, err := ort.NewAdvancedSession(modelPath,
			[]string{"input_ids", "attention_mask", "token_type_ids"},
			[]string{"last_hidden_state"},
			[]ort.Value{inputIDs, attMask, tokenType},
			[]ort.Value{output},
			nil,
		)
		if err != nil {
			inputIDs.Destroy()
			attMask.Destroy()
			tokenType.Destroy()
			output.Destroy()
			onnxErr = err
			return
		}

		globalOnnx = &OnnxEmbedder{
			session:   session,
			vocab:     vocab,
			inputIDs:  inputIDs,
			attMask:   attMask,
			tokenType: tokenType,
			output:    output,
		}
	})

	return globalOnnx, onnxErr
}

// Embed generates a single embedding using the ONNX model.
func (e *OnnxEmbedder) Embed(text string) ([]float32, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	tokenIDs := onnxWordpieceTokenize(e.vocab, text, OnnxMaxTokens)
	tokenCount := len(tokenIDs)

	ids := e.inputIDs.GetData()
	mask := e.attMask.GetData()
	types := e.tokenType.GetData()

	for i := range ids {
		ids[i] = 0
		mask[i] = 0
		types[i] = 0
	}
	for i, id := range tokenIDs {
		ids[i] = id
		mask[i] = 1
	}

	if err := e.session.Run(); err != nil {
		return nil, fmt.Errorf("ONNX inference failed: %w", err)
	}

	outputData := e.output.GetData()
	embedding := make([]float32, OnnxDim)

	for t := 0; t < tokenCount; t++ {
		offset := t * OnnxDim
		for d := 0; d < OnnxDim; d++ {
			embedding[d] += outputData[offset+d]
		}
	}

	tc := float32(tokenCount)
	if tc == 0 {
		tc = 1
	}
	for d := 0; d < OnnxDim; d++ {
		embedding[d] /= tc
	}

	var norm float32
	for _, v := range embedding {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for d := range embedding {
			embedding[d] /= norm
		}
	}

	return embedding, nil
}

// EmbedBatch processes multiple texts sequentially.
func (e *OnnxEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		emb, err := e.Embed(text)
		if err != nil {
			continue // skip failures, don't block indexing
		}
		results[i] = emb
	}
	return results, nil
}

// ── Tokenizer helpers ───────────────────────────────────────────────────────

func loadOnnxVocab(path string) (map[string]int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	vocab := make(map[string]int64)
	for i, line := range strings.Split(string(data), "\n") {
		token := strings.TrimSpace(line)
		if token != "" {
			vocab[token] = int64(i)
		}
	}
	return vocab, nil
}

func onnxWordpieceTokenize(vocab map[string]int64, text string, maxLen int) []int64 {
	clsID := vocab["[CLS]"]
	sepID := vocab["[SEP]"]
	unkID := vocab["[UNK]"]

	text = strings.ToLower(text)
	words := strings.Fields(text)

	var tokens []int64
	tokens = append(tokens, clsID)

	for _, word := range words {
		subTokens := onnxWordpieceWord(vocab, word, unkID)
		for _, st := range subTokens {
			if len(tokens) >= maxLen-1 {
				break
			}
			tokens = append(tokens, st)
		}
		if len(tokens) >= maxLen-1 {
			break
		}
	}

	tokens = append(tokens, sepID)
	return tokens
}

func onnxWordpieceWord(vocab map[string]int64, word string, unkID int64) []int64 {
	if len(word) == 0 {
		return nil
	}

	var result []int64
	remaining := word
	isFirst := true

	for len(remaining) > 0 {
		bestLen := 0
		bestID := unkID

		for end := len(remaining); end > 0; end-- {
			candidate := remaining[:end]
			if !isFirst {
				candidate = "##" + candidate
			}
			if id, ok := vocab[candidate]; ok {
				bestLen = end
				bestID = id
				break
			}
		}

		if bestLen == 0 {
			result = append(result, unkID)
			remaining = remaining[1:]
		} else {
			result = append(result, bestID)
			remaining = remaining[bestLen:]
		}
		isFirst = false
	}

	return result
}

// ── Fallback: OpenAI embedding when ONNX is unavailable ─────────────────────

// OpenAIFallbackEmbed generates embeddings via OpenAI API.
// Returns nil with no error if OPENAI_API_KEY is absent.
func OpenAIFallbackEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, nil
	}

	client := openai.NewClient(apiKey)
	results := make([][]float32, len(texts))
	batchSize := 20

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]

		resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Input: batch,
			Model: openai.SmallEmbedding3,
		})
		if err != nil {
			return nil, fmt.Errorf("OpenAI embedding error at batch %d: %w", i/batchSize, err)
		}
		for j, d := range resp.Data {
			results[i+j] = d.Embedding
		}
	}

	return results, nil
}
