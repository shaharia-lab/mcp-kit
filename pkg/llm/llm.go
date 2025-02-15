package llm

import (
	"github.com/shaharia-lab/goai"
	"github.com/sirupsen/logrus"
)

type LLM struct {
	provider    goai.LLMProvider
	logger      *logrus.Logger
	maxToken    int64
	topP        float64
	topK        int64
	temperature float64
}

func NewLLM(provider goai.LLMProvider, logger *logrus.Logger) *LLM {
	return &LLM{
		provider:    provider,
		logger:      logger,
		maxToken:    500,
		topP:        0.5,
		topK:        40,
		temperature: 0.5,
	}
}

func (l *LLM) UseTopP(topP float64) {
	l.topP = topP
}

func (l *LLM) UseMaxToken(maxToken int64) {
	l.maxToken = maxToken
}

func (l *LLM) UseTopK(topK int64) {
	l.topK = topK
}

func (l *LLM) UseTemperature(temperature float64) {
	l.temperature = temperature
}

func (l *LLM) GetResponse(messages []goai.LLMMessage) goai.LLMResponse {
	cfg := goai.LLMRequestConfig{
		MaxToken:    l.maxToken,
		TopP:        l.topP,
		Temperature: l.temperature,
		TopK:        l.topK,
	}

	response, err := l.provider.GetResponse(messages, cfg)
	if err != nil {
		return goai.LLMResponse{}
	}

	return response
}
