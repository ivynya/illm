package ollama

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

var (
	ErrEmptyResponse       = errors.New("no response")
	ErrIncompleteEmbedding = errors.New("no all input got emmbedded")
)

// LLM is a ollama LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *Client
	options          options
}

// New creates a new ollama LLM implementation.
func New(opts ...Option) (*LLM, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	client, err := NewClient(o.ollamaServerURL)
	if err != nil {
		return nil, err
	}

	return &LLM{client: client, options: o}, nil
}

// Call Implement the call interface for LLM.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := o.Generate(ctx, []string{prompt}, []int{}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

// Generate implemente the generate interface for LLM.
func (o *LLM) Generate(ctx context.Context, prompts []string, chatContext []int, options ...llms.CallOption) ([]*llms.Generation, error) {
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, prompts)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Load back CallOptions as ollamaOptions
	ollamaOptions := o.options.ollamaOptions
	ollamaOptions.NumPredict = opts.MaxTokens
	ollamaOptions.Temperature = float32(opts.Temperature)
	ollamaOptions.Stop = opts.StopWords
	ollamaOptions.TopK = opts.TopK
	ollamaOptions.TopP = float32(opts.TopP)
	ollamaOptions.Seed = opts.Seed
	ollamaOptions.RepeatPenalty = float32(opts.RepetitionPenalty)
	ollamaOptions.FrequencyPenalty = float32(opts.FrequencyPenalty)
	ollamaOptions.PresencePenalty = float32(opts.PresencePenalty)

	// Override LLM model if set as llms.CallOption
	model := o.options.model
	if opts.Model != "" {
		model = opts.Model
	}

	generations := make([]*llms.Generation, 0, len(prompts))

	for _, prompt := range prompts {
		req := &GenerateRequest{
			Model:    model,
			System:   o.options.system,
			Prompt:   prompt,
			Template: o.options.customModelTemplate,
			Options:  ollamaOptions,
			Context:  chatContext,
			Stream:   func(b bool) *bool { return &b }(opts.StreamingFunc != nil),
		}

		var fn GenerateResponseFunc

		var output string
		fn = func(response GenerateResponse) error {
			if opts.StreamingFunc != nil {
				j, err := json.Marshal(response)
				if err != nil {
					return err
				}
				if err := opts.StreamingFunc(ctx, j); err != nil {
					return err
				}
			}
			output += response.Response
			return nil
		}

		err := o.client.Generate(ctx, req, fn)
		if err != nil {
			return []*llms.Generation{}, err
		}

		generations = append(generations, &llms.Generation{Text: output})
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	embeddings := [][]float32{}

	for _, input := range inputTexts {
		embedding, err := o.client.CreateEmbedding(ctx, &EmbeddingRequest{
			Prompt: input,
			Model:  o.options.model,
		})
		if err != nil {
			return nil, err
		}

		if len(embedding.Embedding) == 0 {
			return nil, ErrEmptyResponse
		}

		embeddings = append(embeddings, embedding.Embedding)
	}

	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrIncompleteEmbedding
	}

	return embeddings, nil
}

func (o *LLM) GetNumTokens(text string) int {
	return llms.CountTokens(o.options.model, text)
}
