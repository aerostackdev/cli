package agent

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/openai"
)

type Agent struct {
	llm   llms.Model
	pkg   *pkg.Store
	debug bool
}

func NewAgent(store *pkg.Store, debug bool) (*Agent, error) {
	// Provider selection: Azure OpenAI (same as API) -> Anthropic -> OpenAI -> Aerostack backend
	// Uses AZURE_OPENAI_* when set (matches packages/api .dev.vars and wrangler.toml)
	// Aerostack backend: when no user keys but aerostack login credentials exist
	var llm llms.Model
	var err error

	if endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT"); endpoint != "" && os.Getenv("AZURE_OPENAI_API_KEY") != "" {
		model := os.Getenv("AZURE_OPENAI_DEPLOYMENT_ASSISTANT")
		if model == "" {
			model = os.Getenv("AZURE_OPENAI_DEPLOYMENT_VISION")
		}
		if model == "" {
			model = "gpt-4o-mini"
		}
		apiVersion := os.Getenv("AZURE_OPENAI_API_VERSION")
		if apiVersion == "" {
			apiVersion = "2024-04-01-preview"
		}
		llm, err = openai.New(
			openai.WithToken(os.Getenv("AZURE_OPENAI_API_KEY")),
			openai.WithBaseURL(endpoint),
			openai.WithModel(model),
			openai.WithEmbeddingModel(model),
			openai.WithAPIType(openai.APITypeAzure),
			openai.WithAPIVersion(apiVersion),
		)
	} else if os.Getenv("ANTHROPIC_API_KEY") != "" {
		llm, err = anthropic.New()
	} else if os.Getenv("OPENAI_API_KEY") != "" {
		llm, err = openai.New()
	} else if apiKey := getAerostackAPIKey(); apiKey != "" {
		llm = NewBackendLLM(api.BaseURL(), apiKey)
	} else {
		return nil, fmt.Errorf("no API key found. Set AZURE_OPENAI_ENDPOINT + AZURE_OPENAI_API_KEY (same as API), or ANTHROPIC_API_KEY, or OPENAI_API_KEY, or run 'aerostack login'")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM: %w", err)
	}

	return &Agent{
		llm:   llm,
		pkg:   store,
		debug: debug,
	}, nil
}

func getAerostackAPIKey() string {
	return credentials.GetAPIKey()
}
