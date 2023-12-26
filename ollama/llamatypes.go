package ollama

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"time"
)

type StatusError struct {
	Status       string `json:"status,omitempty"`
	ErrorMessage string `json:"error"`
	StatusCode   int    `json:"code,omitempty"`
}

func (e StatusError) Error() string {
	switch {
	case e.Status != "" && e.ErrorMessage != "":
		return fmt.Sprintf("%s: %s", e.Status, e.ErrorMessage)
	case e.Status != "":
		return e.Status
	case e.ErrorMessage != "":
		return e.ErrorMessage
	default:
		// this should not happen
		return "something went wrong, please see the ollama server logs for details"
	}
}

type GenerateRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	System   string `json:"system"`
	Template string `json:"template"`
	Context  []int  `json:"context,omitempty"`
	Stream   *bool  `json:"stream"`

	Options Options `json:"options"`
}

type ImageData []byte

type Message struct {
	Role    string      `json:"role"` // one of ["system", "user", "assistant"]
	Content string      `json:"content"`
	Images  []ImageData `json:"images,omitempty"`
}

type ChatRequest struct {
	Model    string     `json:"model"`
	Messages []*Message `json:"messages"`
	Stream   *bool      `json:"stream,omitempty"`
	Format   string     `json:"format"`

	Options Options `json:"options"`
}

type Metrics struct {
	TotalDuration      time.Duration `json:"total_duration,omitempty"`
	LoadDuration       time.Duration `json:"load_duration,omitempty"`
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `json:"eval_count,omitempty"`
	EvalDuration       time.Duration `json:"eval_duration,omitempty"`
}

type EmbeddingRequest struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Options Options `json:"options"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type GenerateResponse struct {
	CreatedAt          time.Time     `json:"created_at"`
	Model              string        `json:"model"`
	Response           string        `json:"response"`
	Context            []int         `json:"context,omitempty"`
	TotalDuration      time.Duration `json:"total_duration,omitempty"`
	LoadDuration       time.Duration `json:"load_duration,omitempty"`
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `json:"eval_count,omitempty"`
	EvalDuration       time.Duration `json:"eval_duration,omitempty"`
	Done               bool          `json:"done"`
}

type ChatResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   *Message  `json:"message,omitempty"`

	Done bool `json:"done"`

	Metrics
}

func (r *GenerateResponse) Summary() {
	if r.TotalDuration > 0 {
		fmt.Fprintf(os.Stderr, "total duration:       %v\n", r.TotalDuration)
	}

	if r.LoadDuration > 0 {
		fmt.Fprintf(os.Stderr, "load duration:        %v\n", r.LoadDuration)
	}

	if r.PromptEvalCount > 0 {
		fmt.Fprintf(os.Stderr, "prompt eval count:    %d token(s)\n", r.PromptEvalCount)
	}

	if r.PromptEvalDuration > 0 {
		fmt.Fprintf(os.Stderr, "prompt eval duration: %s\n", r.PromptEvalDuration)
		fmt.Fprintf(os.Stderr, "prompt eval rate:     %.2f tokens/s\n",
			float64(r.PromptEvalCount)/r.PromptEvalDuration.Seconds())
	}

	if r.EvalCount > 0 {
		fmt.Fprintf(os.Stderr, "eval count:           %d token(s)\n", r.EvalCount)
	}

	if r.EvalDuration > 0 {
		fmt.Fprintf(os.Stderr, "eval duration:        %s\n", r.EvalDuration)
		fmt.Fprintf(os.Stderr, "eval rate:            %.2f tokens/s\n", float64(r.EvalCount)/r.EvalDuration.Seconds())
	}
}

type Runner struct {
	NumCtx             int     `json:"num_ctx,omitempty"`
	NumBatch           int     `json:"num_batch,omitempty"`
	NumGQA             int     `json:"num_gqa,omitempty"`
	NumGPU             int     `json:"num_gpu,omitempty"`
	MainGPU            int     `json:"main_gpu,omitempty"`
	NumThread          int     `json:"num_thread,omitempty"`
	RopeFrequencyBase  float32 `json:"rope_frequency_base,omitempty"`
	RopeFrequencyScale float32 `json:"rope_frequency_scale,omitempty"`
	LogitsAll          bool    `json:"logits_all,omitempty"`
	VocabOnly          bool    `json:"vocab_only,omitempty"`
	UseMMap            bool    `json:"use_mmap,omitempty"`
	UseMLock           bool    `json:"use_mlock,omitempty"`
	EmbeddingOnly      bool    `json:"embedding_only,omitempty"`
	UseNUMA            bool    `json:"numa,omitempty"`
	F16KV              bool    `json:"f16_kv,omitempty"`
	LowVRAM            bool    `json:"low_vram,omitempty"`
}

type Options struct {
	Stop []string `json:"stop,omitempty"`
	Runner
	RepeatLastN      int     `json:"repeat_last_n,omitempty"`
	Seed             int     `json:"seed,omitempty"`
	TopK             int     `json:"top_k,omitempty"`
	NumKeep          int     `json:"num_keep,omitempty"`
	Mirostat         int     `json:"mirostat,omitempty"`
	NumPredict       int     `json:"num_predict,omitempty"`
	Temperature      float32 `json:"temperature,omitempty"`
	TypicalP         float32 `json:"typical_p,omitempty"`
	RepeatPenalty    float32 `json:"repeat_penalty,omitempty"`
	PresencePenalty  float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`
	TFSZ             float32 `json:"tfs_z,omitempty"`
	MirostatTau      float32 `json:"mirostat_tau,omitempty"`
	MirostatEta      float32 `json:"mirostat_eta,omitempty"`
	TopP             float32 `json:"top_p,omitempty"`
	PenalizeNewline  bool    `json:"penalize_newline,omitempty"`
}

type options struct {
	ollamaServerURL     *url.URL
	model               string
	ollamaOptions       Options
	customModelTemplate string
	system              string
}

type Option func(*options)

// WithModel Set the model to use.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithSystem Set the system prompt. This is only valid if
// WithCustomTemplate is not set and the ollama model use
// .System in its model template OR if WithCustomTemplate
// is set using {{.System}}.
func WithSystemPrompt(p string) Option {
	return func(opts *options) {
		opts.system = p
	}
}

// WithCustomTemplate To override the templating done on Ollama model side.
func WithCustomTemplate(template string) Option {
	return func(opts *options) {
		opts.customModelTemplate = template
	}
}

// WithServerURL Set the URL of the ollama instance to use.
func WithServerURL(rawURL string) Option {
	return func(opts *options) {
		var err error
		opts.ollamaServerURL, err = url.Parse(rawURL)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// WithBackendUseNUMA Use NUMA optimization on certain systems.
func WithRunnerUseNUMA(numa bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.UseNUMA = numa
	}
}

// WithRunnerNumCtx Sets the size of the context window used to generate the next token (Default: 2048).
func WithRunnerNumCtx(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumCtx = num
	}
}

// WithRunnerNumKeep Specify the number of tokens from the initial prompt to retain when the model resets
// its internal context.
func WithRunnerNumKeep(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumKeep = num
	}
}

// WithRunnerNumBatch Set the batch size for prompt processing (default: 512).
func WithRunnerNumBatch(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumBatch = num
	}
}

// WithRunnerNumThread Set the number of threads to use during computation (default: auto).
func WithRunnerNumThread(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumThread = num
	}
}

// WithRunnerNumGQA The number of GQA groups in the transformer layer. Required for some models.
func WithRunnerNumGQA(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumGQA = num
	}
}

// WithRunnerNumGPU The number of layers to send to the GPU(s).
// On macOS it defaults to 1 to enable metal support, 0 to disable.
func WithRunnerNumGPU(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumGPU = num
	}
}

// WithRunnerMainGPU When using multiple GPUs this option controls which GPU is used for small tensors
// for which the overhead of splitting the computation across all GPUs is not worthwhile.
// The GPU in question will use slightly more VRAM to store a scratch buffer for temporary results.
// By default GPU 0 is used.
func WithRunnerMainGPU(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.MainGPU = num
	}
}

// WithRunnerLowVRAM Do not allocate a VRAM scratch buffer for holding temporary results.
// Reduces VRAM usage at the cost of performance, particularly prompt processing speed.
func WithRunnerLowVRAM(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.LowVRAM = val
	}
}

// WithRunnerF16KV If set to falsem, use 32-bit floats instead of 16-bit floats for memory key+value.
func WithRunnerF16KV(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.F16KV = val
	}
}

// WithRunnerLogitsAll Return logits for all tokens, not just the last token.
func WithRunnerLogitsAll(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.LogitsAll = val
	}
}

// WithRunnerVocabOnly Only load the vocabulary, no weights.
func WithRunnerVocabOnly(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.VocabOnly = val
	}
}

// WithRunnerUseMMap Set to false to not memory-map the model.
// By default, models are mapped into memory, which allows the system to load only the necessary parts
// of the model as needed.
func WithRunnerUseMMap(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.UseMMap = val
	}
}

// WithRunnerUseMLock Force system to keep model in RAM.
func WithRunnerUseMLock(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.UseMLock = val
	}
}

// WithRunnerEmbeddingOnly Only return the embbeding.
func WithRunnerEmbeddingOnly(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.EmbeddingOnly = val
	}
}

// WithRunnerRopeFrequencyBase RoPE base frequency (default: loaded from model).
func WithRunnerRopeFrequencyBase(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.RopeFrequencyBase = val
	}
}

// WithRunnerRopeFrequencyScale Rope frequency scaling factor (default: loaded from model).
func WithRunnerRopeFrequencyScale(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.RopeFrequencyScale = val
	}
}

// WithPredictTFSZ Tail free sampling is used to reduce the impact of less probable tokens from the output.
// A higher value (e.g., 2.0) will reduce the impact more, while a value of 1.0 disables this setting (default: 1).
func WithPredictTFSZ(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.TFSZ = val
	}
}

// WithPredictTypicalP Enable locally typical sampling with parameter p (default: 1.0, 1.0 = disabled).
func WithPredictTypicalP(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.TypicalP = val
	}
}

// WithPredictRepeatLastN Sets how far back for the model to look back to prevent repetition
// (Default: 64, 0 = disabled, -1 = num_ctx).
func WithPredictRepeatLastN(val int) Option {
	return func(opts *options) {
		opts.ollamaOptions.RepeatLastN = val
	}
}

// WithPredictMirostat Enable Mirostat sampling for controlling perplexity
// (default: 0, 0 = disabled, 1 = Mirostat, 2 = Mirostat 2.0).
func WithPredictMirostat(val int) Option {
	return func(opts *options) {
		opts.ollamaOptions.Mirostat = val
	}
}

// WithPredictMirostatTau Controls the balance between coherence and diversity of the output.
// A lower value will result in more focused and coherent text (Default: 5.0).
func WithPredictMirostatTau(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.MirostatTau = val
	}
}

// WithPredictMirostatEta Influences how quickly the algorithm responds to feedback from the generated text.
// A lower learning rate will result in slower adjustments, while a higher learning rate will make the
// algorithm more responsive (Default: 0.1).
func WithPredictMirostatEta(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.MirostatEta = val
	}
}

// WithPredictPenalizeNewline Penalize newline tokens when applying the repeat penalty (default: true).
func WithPredictPenalizeNewline(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.PenalizeNewline = val
	}
}
