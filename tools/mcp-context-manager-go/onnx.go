package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	ort "github.com/yalue/onnxruntime_go"
)

// EmbeddingDim is the output dimension for all-MiniLM-L6-v2.
const EmbeddingDim = 384

// MaxTokens is the maximum input length for the model.
const MaxTokens = 128

// EmbedProvider indicates which backend produced the embedding.
type EmbedProvider string

const (
	ProviderONNX   EmbedProvider = "onnx"
	ProviderOpenAI EmbedProvider = "openai"
	ProviderNone   EmbedProvider = "none"
)

// onnxSession holds the singleton ONNX model session.
type onnxSession struct {
	session     *ort.AdvancedSession
	vocab       map[string]int64
	inputIDs    *ort.Tensor[int64]
	attMask     *ort.Tensor[int64]
	tokenType   *ort.Tensor[int64]
	output      *ort.Tensor[float32]
	mu          sync.Mutex
}

var (
	globalONNX     *onnxSession
	onnxInitOnce   sync.Once
	onnxInitErr    error
	onnxEnvInited  bool
)

// initONNXSession loads the ONNX model and vocabulary.
// Model must be at $AGK_MODELS_DIR/all-MiniLM-L6-v2.onnx (or ~/.agk/models/).
// Vocab must be at $AGK_MODELS_DIR/vocab.txt (WordPiece vocabulary).
func initONNXSession() (*onnxSession, error) {
	onnxInitOnce.Do(func() {
		modelsDir := os.Getenv("AGK_MODELS_DIR")
		if modelsDir == "" {
			home, _ := os.UserHomeDir()
			modelsDir = filepath.Join(home, ".agk", "models")
		}

		modelPath := filepath.Join(modelsDir, "all-MiniLM-L6-v2.onnx")
		vocabPath := filepath.Join(modelsDir, "vocab.txt")

		// Check files exist
		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			onnxInitErr = fmt.Errorf("ONNX model not found at %s — set AGK_MODELS_DIR or run: agk model download", modelPath)
			return
		}
		if _, err := os.Stat(vocabPath); os.IsNotExist(err) {
			onnxInitErr = fmt.Errorf("vocab.txt not found at %s", vocabPath)
			return
		}

		// Load vocabulary
		vocab, err := loadVocab(vocabPath)
		if err != nil {
			onnxInitErr = fmt.Errorf("failed to load vocab: %w", err)
			return
		}

		// Set shared library path if provided
		if libPath := os.Getenv("ONNXRUNTIME_LIB"); libPath != "" {
			ort.SetSharedLibraryPath(libPath)
		}

		// Initialize ONNX runtime environment (idempotent guard)
		if !onnxEnvInited {
			if err := ort.InitializeEnvironment(); err != nil {
				onnxInitErr = fmt.Errorf("ONNX runtime init failed: %w", err)
				return
			}
			onnxEnvInited = true
		}

		// Pre-allocate tensors for single-text inference
		seqLen := int64(MaxTokens)
		shape := ort.NewShape(1, seqLen)

		inputIDs, err := ort.NewEmptyTensor[int64](shape)
		if err != nil {
			onnxInitErr = fmt.Errorf("input_ids tensor: %w", err)
			return
		}

		attMask, err := ort.NewEmptyTensor[int64](shape)
		if err != nil {
			inputIDs.Destroy()
			onnxInitErr = fmt.Errorf("attention_mask tensor: %w", err)
			return
		}

		tokenType, err := ort.NewEmptyTensor[int64](shape)
		if err != nil {
			inputIDs.Destroy()
			attMask.Destroy()
			onnxInitErr = fmt.Errorf("token_type_ids tensor: %w", err)
			return
		}

		outputShape := ort.NewShape(1, seqLen, int64(EmbeddingDim))
		output, err := ort.NewEmptyTensor[float32](outputShape)
		if err != nil {
			inputIDs.Destroy()
			attMask.Destroy()
			tokenType.Destroy()
			onnxInitErr = fmt.Errorf("output tensor: %w", err)
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
			onnxInitErr = fmt.Errorf("ONNX session creation failed: %w", err)
			return
		}

		globalONNX = &onnxSession{
			session:   session,
			vocab:     vocab,
			inputIDs:  inputIDs,
			attMask:   attMask,
			tokenType: tokenType,
			output:    output,
		}
	})

	return globalONNX, onnxInitErr
}

// getEmbeddingONNX generates an embedding using the local ONNX model.
func getEmbeddingONNX(text string) ([]float32, error) {
	sess, err := initONNXSession()
	if err != nil {
		return nil, err
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()

	// Tokenize
	tokenIDs := wordpieceTokenize(sess.vocab, text, MaxTokens)
	tokenCount := len(tokenIDs)

	// Fill input tensors
	ids := sess.inputIDs.GetData()
	mask := sess.attMask.GetData()
	types := sess.tokenType.GetData()

	for i := range ids {
		ids[i] = 0
		mask[i] = 0
		types[i] = 0
	}
	for i, id := range tokenIDs {
		ids[i] = id
		mask[i] = 1
	}

	// Run inference
	if err := sess.session.Run(); err != nil {
		return nil, fmt.Errorf("ONNX inference failed: %w", err)
	}

	// Mean pooling over non-padding tokens → 384-dim output
	outputData := sess.output.GetData()
	embedding := make([]float32, EmbeddingDim)

	for t := 0; t < tokenCount; t++ {
		offset := t * EmbeddingDim
		for d := 0; d < EmbeddingDim; d++ {
			embedding[d] += outputData[offset+d]
		}
	}

	// Average
	tc := float32(tokenCount)
	if tc == 0 {
		tc = 1
	}
	for d := 0; d < EmbeddingDim; d++ {
		embedding[d] /= tc
	}

	// L2 normalize
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

// getEmbeddingWithFallback implements the ONNX-first → OpenAI-fallback chain.
// Returns (embedding, provider, error).
func getEmbeddingWithFallback(text string) ([]float32, EmbedProvider, error) {
	// Try ONNX first
	emb, err := getEmbeddingONNX(text)
	if err == nil && len(emb) > 0 {
		return emb, ProviderONNX, nil
	}

	// Fallback to OpenAI
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, ProviderNone, nil // graceful degradation
	}

	client := openai.NewClient(apiKey)
	resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, ProviderNone, fmt.Errorf("OpenAI embedding fallback error: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, ProviderNone, fmt.Errorf("empty OpenAI embedding response")
	}
	return resp.Data[0].Embedding, ProviderOpenAI, nil
}

// ── WordPiece Tokenizer ─────────────────────────────────────────────────────

// loadVocab reads a vocab.txt file (one token per line) into a map.
func loadVocab(path string) (map[string]int64, error) {
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

// wordpieceTokenize performs basic WordPiece tokenization.
// Returns token IDs including [CLS] and [SEP], padded/truncated to maxLen.
func wordpieceTokenize(vocab map[string]int64, text string, maxLen int) []int64 {
	clsID, _ := vocab["[CLS]"]
	sepID, _ := vocab["[SEP]"]
	unkID, _ := vocab["[UNK]"]

	// Lowercase and split into words
	text = strings.ToLower(text)
	words := strings.Fields(text)

	var tokens []int64
	tokens = append(tokens, clsID)

	for _, word := range words {
		subTokens := wordpieceWord(vocab, word, unkID)
		for _, st := range subTokens {
			if len(tokens) >= maxLen-1 { // reserve space for [SEP]
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

// wordpieceWord breaks a single word into WordPiece sub-tokens.
func wordpieceWord(vocab map[string]int64, word string, unkID int64) []int64 {
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
			// Unknown token — skip one character
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

// ── Compatibility helpers ───────────────────────────────────────────────────

// marshalEmbedding serializes a float32 slice to JSON for DB storage.
func marshalEmbedding(emb []float32) string {
	data, _ := json.Marshal(emb)
	return string(data)
}

// unmarshalEmbedding deserializes a JSON string to float32 slice.
func unmarshalEmbedding(s string) []float32 {
	var emb []float32
	json.Unmarshal([]byte(s), &emb)
	return emb
}
