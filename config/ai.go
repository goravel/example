package config

import (
	anthropicfacades "github.com/goravel/anthropic/facades"
	"github.com/goravel/framework/contracts/ai"
	geminifacades "github.com/goravel/gemini/facades"
	openaifacades "github.com/goravel/openai/facades"

	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("ai", map[string]any{
		"default": "openai",
		"providers": map[string]any{
			"openai": map[string]any{
				"key": config.Env("OPENAI_API_KEY", ""),
				"models": map[string]any{
					"text": map[string]any{
						"default":    config.Env("OPENAI_TEXT_MODEL", ""),
						"max_tokens": config.Env("OPENAI_MAX_TOKENS", 0),
					},
					"audio": map[string]any{
						"default": config.Env("OPENAI_AUDIO_MODEL", ""),
					},
					"transcription": map[string]any{
						"default": config.Env("OPENAI_TRANSCRIPTION_MODEL", ""),
					},
					"image": map[string]any{
						"default": config.Env("OPENAI_IMAGE_MODEL", ""),
					},
				},
				"failover": map[string][]string{},
				"url":      config.Env("OPENAI_BASE_URL", ""),
				"via": func() (ai.Provider, error) {
					return openaifacades.OpenAI("openai")
				},
			},
			"anthropic": map[string]any{
				"key": config.Env("ANTHROPIC_API_KEY", ""),
				"models": map[string]any{
					"text": map[string]any{
						"default":    config.Env("ANTHROPIC_TEXT_MODEL", ""),
						"max_tokens": config.Env("ANTHROPIC_MAX_TOKENS", 4096),
					},
				},
				"failover": map[string][]string{},
				"url":      config.Env("ANTHROPIC_BASE_URL", ""),
				"via": func() (ai.Provider, error) {
					return anthropicfacades.Anthropic("anthropic")
				},
			},
			"gemini": map[string]any{
				"key": config.Env("GEMINI_API_KEY", ""),
				"models": map[string]any{
					"text": map[string]any{
						"default": config.Env("GEMINI_TEXT_MODEL", ""),
					},
					"audio": map[string]any{
						"default": config.Env("GEMINI_AUDIO_MODEL", ""),
					},
					"transcription": map[string]any{
						"default": config.Env("GEMINI_TRANSCRIPTION_MODEL", ""),
					},
					"image": map[string]any{
						"default": config.Env("GEMINI_IMAGE_MODEL", ""),
					},
				},
				"failover": map[string][]string{},
				"url":      config.Env("GEMINI_BASE_URL", ""),
				"via": func() (ai.Provider, error) {
					return geminifacades.Gemini("gemini")
				},
			},
		},
	})
}
