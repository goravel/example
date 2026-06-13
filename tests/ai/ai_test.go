package ai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	openaifacades "github.com/goravel/openai/facades"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

type AITestSuite struct {
	suite.Suite
	tests.TestCase
}

type aiProviderCase struct {
	name string
	env  string
}

type aiTestAgent struct {
	instructions string
	messages     []contractsai.Message
	middleware   []contractsai.Middleware
	tools        []contractsai.Tool
}

type aiRewriteMiddleware struct{}

type aiStaticTool struct {
	called bool
}

func TestAITestSuite(t *testing.T) {
	suite.Run(t, &AITestSuite{})
}

func (s *AITestSuite) TestFacadeResolves() {
	s.NotNil(facades.AI())
}

func (s *AITestSuite) TestPromptStreamMiddlewareAndHistory() {
	for _, provider := range aiProviderCases() {
		s.Run(provider.name, func() {
			s.requireProvider(provider)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			initial := []contractsai.Message{{Role: contractsai.RoleAssistant, Content: "Initial assistant context."}}
			conversation, err := facades.AI().WithContext(ctx).Agent(&aiTestAgent{
				instructions: "Follow the user's requested exact token responses.",
				messages:     initial,
			}, frameworkai.WithProvider(provider.name), frameworkai.WithModel(aiTextModel(provider.name)), frameworkai.WithMiddleware(&aiRewriteMiddleware{}))
			s.Require().NoError(err)

			input := "middleware rewrite this request"
			response, err := conversation.Prompt(input)
			s.Require().NoError(err)
			s.Require().NotNil(response)
			s.containsToken(response.Text(), "AI_MIDDLEWARE_OK")

			var callbackText string
			response.Then(func(response contractsai.AgentResponse) {
				callbackText = response.Text()
			})
			s.Equal(response.Text(), callbackText)

			messages := conversation.Messages()
			s.Require().Len(messages, 3)
			s.Equal(initial[0], messages[0])
			s.Equal(contractsai.Message{Role: contractsai.RoleUser, Content: input}, messages[1])
			s.Equal(contractsai.RoleAssistant, messages[2].Role)

			conversation.Reset()
			s.Equal(initial, conversation.Messages())

			stream, err := conversation.Stream("Reply with the exact token AI_STREAM_OK and no other text.")
			s.Require().NoError(err)

			var streamResponse contractsai.AgentResponse
			stream.Then(func(response contractsai.AgentResponse) {
				streamResponse = response
			})

			var events []contractsai.StreamEvent
			err = stream.Each(func(event contractsai.StreamEvent) error {
				events = append(events, event)
				return nil
			})
			s.Require().NoError(err)
			s.Require().NotNil(streamResponse)
			s.containsToken(streamResponse.Text(), "AI_STREAM_OK")
			s.True(hasStreamEvent(events, contractsai.StreamEventTypeDone))
		})
	}
}

func (s *AITestSuite) TestTools() {
	for _, provider := range aiProviderCases() {
		s.Run(provider.name, func() {
			s.requireProvider(provider)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			tool := &aiStaticTool{}
			conversation, err := facades.AI().WithContext(ctx).Agent(&aiTestAgent{
				instructions: "Call lookup_goravel_ai_test_answer before answering. After the tool result is available, reply with exactly the tool result and no other text.",
				tools:        []contractsai.Tool{tool},
			}, frameworkai.WithProvider(provider.name))
			s.Require().NoError(err)

			response, err := conversation.Prompt("What is the Goravel AI integration test answer?")
			s.Require().NoError(err)
			s.Require().True(tool.called)
			s.containsToken(response.Text(), "SKY_BLUE")

			messages := conversation.Messages()
			s.True(hasMessageRole(messages, contractsai.RoleToolResult))
		})
	}
}

func (s *AITestSuite) TestAttachmentsAndProviderFiles() {
	for _, provider := range aiProviderCases() {
		s.Run(provider.name, func() {
			s.requireProvider(provider)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			conversation, err := facades.AI().WithContext(ctx).Agent(&aiTestAgent{
				instructions: "Answer using attached documents when they are provided.",
			}, frameworkai.WithProvider(provider.name))
			s.Require().NoError(err)

			attachment := frameworkai.DocumentFromString("The required code word is ATTACHMENT_OK.", frameworkai.WithMimeType("text/plain"))
			response, err := conversation.Prompt("Read the attached document and reply with the code word only.", frameworkai.WithAttachments(attachment))
			s.Require().NoError(err)
			s.containsToken(response.Text(), "ATTACHMENT_OK")

			file := frameworkai.DocumentFromString("Stored provider file content.", frameworkai.WithMimeType("text/plain"))
			uploaded, err := file.Put(ctx, frameworkai.WithProvider(provider.name))
			s.Require().NoError(err)
			s.Require().NotEmpty(uploaded.ID())

			stored := frameworkai.DocumentFromID(uploaded.ID())
			s.T().Cleanup(func() {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cleanupCancel()
				_ = stored.Delete(cleanupCtx, frameworkai.WithProvider(provider.name))
			})

			resolved, err := stored.Get(ctx, frameworkai.WithProvider(provider.name))
			s.Require().NoError(err)
			content, err := resolved.Content(ctx)
			s.Require().NoError(err)
			s.Contains(string(content), "Stored provider file content.")

			s.NoError(stored.Delete(ctx, frameworkai.WithProvider(provider.name)))
		})
	}
}

func (s *AITestSuite) TestMediaRequests() {
	for _, provider := range []aiProviderCase{{name: "openai", env: "OPENAI_API_KEY"}, {name: "gemini", env: "GEMINI_API_KEY"}} {
		s.Run(provider.name, func() {
			s.requireProvider(provider)

			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()

			imageResponse, err := facades.AI().WithContext(ctx).Image("A simple red square icon with no text.").Provider(provider.name).Square().Generate()
			s.Require().NoError(err)
			imageContent, err := imageResponse.Content()
			s.Require().NoError(err)
			s.NotEmpty(imageContent)
			s.Contains(imageResponse.MimeType(), "image/")

			imagePath := "ai-tests/" + provider.name + "-image.png"
			s.T().Cleanup(func() {
				_ = facades.Storage().Delete(imagePath)
			})
			storedImagePath, err := imageResponse.StoreAs(imagePath)
			s.Require().NoError(err)
			s.Equal(imagePath, storedImagePath)
			s.True(facades.Storage().Exists(imagePath))

			audioResponse, err := facades.AI().WithContext(ctx).Audio("Say the words Goravel integration test clearly.").Provider(provider.name).Timeout(90 * time.Second).Generate()
			s.Require().NoError(err)
			audioContent, err := audioResponse.Content()
			s.Require().NoError(err)
			s.NotEmpty(audioContent)
			s.Contains(audioResponse.MimeType(), "audio/")

			audioPath := "ai-tests/" + provider.name + "-audio.mp3"
			s.T().Cleanup(func() {
				_ = facades.Storage().Delete(audioPath)
			})
			storedAudioPath, err := audioResponse.StoreAs(audioPath)
			s.Require().NoError(err)
			s.Equal(audioPath, storedAudioPath)
			s.True(facades.Storage().Exists(audioPath))

			transcriptionResponse, err := facades.AI().WithContext(ctx).Transcription(
				frameworkai.DocumentFromByte(audioContent, frameworkai.WithMimeType(audioResponse.MimeType())),
			).Provider(provider.name).Language("en").Timeout(90 * time.Second).Generate()
			s.Require().NoError(err)
			s.NotEmpty(transcriptionResponse.Text())
		})
	}
}

func (s *AITestSuite) TestFailoverUsesBackupProvider() {
	backup, ok := firstAvailableProvider()
	if !ok {
		s.T().Skip("no AI provider API key is set")
	}

	const failingProvider = "openai_failover_test"
	facades.Config().Add("ai.providers."+failingProvider, map[string]any{
		"key": "test-key",
		"models": map[string]any{
			"text": map[string]any{
				"default": aiTextModel("openai"),
			},
		},
		"failover": map[string][]string{
			"provider_overloaded": {"connection refused", "dial tcp", "connect:"},
		},
		"url": "http://127.0.0.1:1",
		"via": func() (contractsai.Provider, error) {
			return openaifacades.OpenAI(failingProvider)
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	conversation, err := facades.AI().WithContext(ctx).Agent(&aiTestAgent{
		instructions: "Follow the user's requested exact token responses.",
	}, frameworkai.WithProvider(failingProvider, backup.name))
	s.Require().NoError(err)

	response, err := conversation.Prompt("Reply with the exact token AI_FAILOVER_OK and no other text.")
	s.Require().NoError(err)
	s.containsToken(response.Text(), "AI_FAILOVER_OK")
}

func (a *aiTestAgent) Instructions() string {
	return a.instructions
}

func (a *aiTestAgent) Messages() []contractsai.Message {
	return append([]contractsai.Message(nil), a.messages...)
}

func (a *aiTestAgent) Middleware() []contractsai.Middleware {
	return append([]contractsai.Middleware(nil), a.middleware...)
}

func (a *aiTestAgent) Tools() []contractsai.Tool {
	return append([]contractsai.Tool(nil), a.tools...)
}

func (m *aiRewriteMiddleware) Handle(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.AgentResponse, error) {
	if strings.Contains(prompt.Input, "middleware rewrite") {
		prompt.Input = "Reply with the exact token AI_MIDDLEWARE_OK and no other text."
	}

	return next(ctx, prompt)
}

func (t *aiStaticTool) Name() string {
	return "lookup_goravel_ai_test_answer"
}

func (t *aiStaticTool) Description() string {
	return "Returns the exact answer for the Goravel AI integration test."
}

func (t *aiStaticTool) Parameters() map[string]any {
	return nil
}

func (t *aiStaticTool) Execute(context.Context, map[string]any) (string, error) {
	t.called = true

	return "SKY_BLUE", nil
}

func aiProviderCases() []aiProviderCase {
	return []aiProviderCase{
		{name: "openai", env: "OPENAI_API_KEY"},
		{name: "anthropic", env: "ANTHROPIC_API_KEY"},
		{name: "gemini", env: "GEMINI_API_KEY"},
	}
}

func (s *AITestSuite) requireProvider(provider aiProviderCase) {
	if strings.TrimSpace(os.Getenv(provider.env)) != "" {
		return
	}
	if strings.TrimSpace(facades.Config().GetString("ai.providers."+provider.name+".key")) != "" {
		return
	}

	s.T().Skipf("%s is not set", provider.env)
}

func firstAvailableProvider() (aiProviderCase, bool) {
	for _, provider := range aiProviderCases() {
		if strings.TrimSpace(os.Getenv(provider.env)) != "" || strings.TrimSpace(facades.Config().GetString("ai.providers."+provider.name+".key")) != "" {
			return provider, true
		}
	}

	return aiProviderCase{}, false
}

func aiTextModel(provider string) string {
	model := facades.Config().GetString("ai.providers." + provider + ".models.text.default")
	if model != "" {
		return model
	}

	switch provider {
	case "anthropic":
		return "claude-sonnet-4-5"
	case "gemini":
		return "gemini-2.5-flash"
	default:
		return "gpt-5.4"
	}
}

func (s *AITestSuite) containsToken(text, token string) {
	normalizedText := strings.ToUpper(strings.TrimSpace(text))
	normalizedToken := strings.ToUpper(token)
	s.Contains(normalizedText, normalizedToken)
}

func hasStreamEvent(events []contractsai.StreamEvent, eventType contractsai.StreamEventType) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}

	return false
}

func hasMessageRole(messages []contractsai.Message, role contractsai.MessageRole) bool {
	for _, message := range messages {
		if message.Role == role {
			return true
		}
	}

	return false
}
