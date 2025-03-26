package server

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/dhaifley/game2d/client"
	"github.com/dhaifley/game2d/errors"
	"github.com/dhaifley/game2d/logger"
	"github.com/dhaifley/game2d/request"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Prompt values represent a single AI prompt and response.
type Prompt struct {
	Prompt   request.FieldString `bson:"prompt"   json:"prompt"   yaml:"prompt"`
	Response request.FieldString `bson:"response" json:"response" yaml:"response"`
	Thinking request.FieldString `bson:"thinking" json:"thinking" yaml:"thinking"`
}

// Prompts values contain the AI prompt data for a game.
type Prompts struct {
	Current Prompt              `bson:"current" json:"current" yaml:"current"`
	History []Prompt            `bson:"history" json:"history" yaml:"history"`
	Error   request.FieldString `bson:"error"   json:"error"   yaml:"error"`
	GameID  request.FieldString `bson:"game_id" json:"game_id" yaml:"game_id"`
}

// Copy creates a copy of the Prompts struct.
func (p *Prompts) Copy() *Prompts {
	if p == nil {
		return nil
	}

	return &Prompts{
		Current: p.Current,
		History: p.History,
		Error:   p.Error,
		GameID:  p.GameID,
	}
}

// promptsToFieldJSON converts a Prompts struct to a FieldJSON value.
func promptsToFieldJSON(p *Prompts) (request.FieldJSON, error) {
	if p == nil {
		return request.FieldJSON{}, nil
	}

	b, err := json.Marshal(p)
	if err != nil {
		return request.FieldJSON{}, err
	}

	m := make(map[string]any)
	if err := json.Unmarshal(b, &m); err != nil {
		return request.FieldJSON{}, err
	}

	return request.FieldJSON{
		Set:   true,
		Valid: true,
		Value: m,
	}, nil
}

// promptsFromFieldJSON converts a FieldJSON value to a Prompts struct.
func promptsFromFieldJSON(f request.FieldJSON) (*Prompts, error) {
	if !f.Set || !f.Valid || f.Value == nil {
		return nil, nil
	}

	b, err := json.Marshal(f.Value)
	if err != nil {
		return nil, err
	}

	p := &Prompts{}
	if err := json.Unmarshal(b, p); err != nil {
		return nil, err
	}

	return p, nil
}

// sendPrompt sends a prompt to the AI service and updates the game state with
// the response. It is called as a goroutine to run the the background, and will
// block until the prompt is complete.
func (s *Server) sendPrompt(ctx context.Context, g *Game, prompts *Prompts) {
	if g == nil {
		return
	}

	defer s.removePrompt(g.ID.Value)

	if prompts == nil {
		return
	}

	updateGame := func(g *Game) {
		if _, err := s.updateGame(ctx, g); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to update game with prompt result",
				"error", err,
				"game_id", g.ID.Value,
				"prompts", prompts)
		}
	}

	p := s.getPrompter()
	if p == nil {
		prompts.Error = request.FieldString{
			Set: true, Valid: true, Value: "prompter not found",
		}

		g.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusError,
		}

		s.log.Log(ctx, logger.LvlError,
			"unable to get prompter for game",
			"error", "prompter not found",
			"game_id", g.ID.Value,
			"prompts", prompts)

		updateGame(g)

		return
	}

	err := p.Prompt(ctx, prompts, g)
	if err != nil {
		prompts.Error = request.FieldString{
			Set: true, Valid: true, Value: err.Error(),
		}

		prompts.Current.Response.Value += "\nError: " + err.Error() + "\n"

		g.Prompts, err = promptsToFieldJSON(prompts)
		if err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to encode prompt response for game state",
				"error", err,
				"game_id", g.ID.Value,
				"prompts", prompts)
		}

		g.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusError,
		}

		updateGame(g)
	}
}

// updateGamePrompts periodically updates pending game prompts.
func (s *Server) updateGamePrompts(ctx context.Context,
) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context) {
		tick := time.NewTimer(time.Second)

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				accounts, err := s.getAllAccounts(ctx)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to get accounts to import games",
						"error", err)

					break
				}

				var wg sync.WaitGroup

				for _, aID := range accounts {
					wg.Add(1)

					go func(ctx context.Context, accountID string) {
						ctx = context.WithValue(ctx, request.CtxKeyAccountID,
							accountID)
						ctx = context.WithValue(ctx, request.CtxKeyUserID,
							request.SystemUser)
						ctx = context.WithValue(ctx, request.CtxKeyScopes,
							request.ScopeSuperuser)

						if n, err := s.updatePrompts(ctx); err != nil {
							s.log.Log(ctx, logger.LvlError,
								"unable to update game prompts",
								"error", err)
						} else if n > 0 {
							s.log.Log(ctx, logger.LvlInfo,
								"updated game prompt timeouts",
								"account_id", accountID,
								"updated", n)
						}

						wg.Done()
					}(ctx, aID)
				}

				wg.Wait()
			}

			tick = time.NewTimer(time.Minute)
		}
	}(ctx)

	return cancel
}

// updatePrompts updates any games with status updating to status error if
// updated_at is older than the configured prompt timeout.
func (s *Server) updatePrompts(ctx context.Context) (int, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return 0, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	ts := time.Now().Add(s.cfg.ServerPromptTimeout() * -1).Unix()

	f := bson.M{
		"account_id": aID,
		"status":     request.StatusUpdating,
		"updated_at": bson.M{"$lt": ts},
	}

	pro := bson.M{"id": 1}

	cur, err := s.DB().Collection("games").Find(ctx, f,
		options.Find().SetProjection(pro))
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabase,
			"unable to get games to delete",
			"filter", f)
	}

	defer func() {
		if err := cur.Close(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to close cursor",
				"err", err)
		}
	}()

	n := 0

	ctx = context.WithValue(ctx, CtxKeyGameMinData, true)

	for cur.Next(ctx) {
		var g *Game

		if err := cur.Decode(&g); err != nil {
			return n, errors.Wrap(err, errors.ErrDatabase,
				"unable to decode game")
		}

		if g == nil {
			continue
		}

		g.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusError,
		}

		if _, err := s.updateGame(ctx, g); err != nil {
			return n, errors.Wrap(err, errors.ErrDatabase,
				"unable to update prompt timeout game",
				"game", g)
		}

		n++
	}

	if err := cur.Err(); err != nil {
		return n, errors.Wrap(err, errors.ErrDatabase,
			"unable to update prompt timeout games",
			"filter", f)
	}

	return n, nil
}

// Prompter values are able to send prompts to AI services.
type Prompter interface {
	Prompt(ctx context.Context,
		prompts *Prompts,
		state *Game,
	) error
}

// initPrompter initializes a prompter for use by the server.
func (s *Server) initPrompter(ctx context.Context) error {
	a, err := s.getAccount(ctx, "")
	if err != nil || a == nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to get account to initialize prompter")
	}

	if a.AIAPIKey.Value == "" {
		return errors.New(errors.ErrUnavailable,
			"the account does not have an AI API key")
	}

	maxTokens := int64(64000)
	if a.AIMaxTokens.Value > 0 {
		maxTokens = a.AIMaxTokens.Value
	}

	budgetTokens := int64(7000)
	if a.AIThinkingBudget.Value > 0 {
		budgetTokens = a.AIThinkingBudget.Value
	}

	s.getPrompter = func() Prompter {
		return NewAnthropicPrompter(s, a.AIAPIKey.Value,
			maxTokens, budgetTokens)
	}

	return nil
}

// anthropicPrompter values are able to send prompts to the Anthropic AI.
type anthropicPrompter struct {
	cli         *anthropic.Client
	s           *Server
	max, budget int64
}

// NewMockPrompter creates a new mock prompter with the given response, state
// and delay.
func NewAnthropicPrompter(s *Server,
	key string,
	maxTokens, budgetTokens int64,
) Prompter {
	cli := anthropic.NewClient(option.WithAPIKey(key))

	return &anthropicPrompter{
		s:      s,
		cli:    cli,
		max:    maxTokens,
		budget: budgetTokens,
	}
}

// Prompt sends a prompt to the mock prompter and returns the response and
// state.
func (p *anthropicPrompter) Prompt(ctx context.Context,
	prompts *Prompts,
	game *Game,
) error {
	updateGame := func(g *Game, prompts *Prompts) error {
		var err error

		g.Prompts, err = promptsToFieldJSON(prompts)
		if err != nil {
			g.Status = request.FieldString{
				Set: true, Valid: true, Value: request.StatusError,
			}

			return errors.Wrap(err, errors.ErrDatabase,
				"unable to encode prompt response for game state",
				"error", err,
				"game_id", g.ID.Value,
				"prompts", prompts)
		}

		if _, err := p.s.updateGame(ctx, g); err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to update game with prompt result",
				"error", err,
				"game_id", g.ID.Value,
				"response", prompts.Current.Response.Value)
		}

		return nil
	}

	prompts.Current.Response = request.FieldString{
		Set: true, Valid: true, Value: "",
	}

	prompts.Current.Thinking = request.FieldString{
		Set: true, Valid: true, Value: "",
	}

	gameFile, err := client.SourceFS.ReadFile("game.go")
	if err != nil {
		return errors.Wrap(err, errors.ErrServer,
			"unable to read client source file",
			"file", "game.go")
	}

	imageFile, err := client.SourceFS.ReadFile("image.go")
	if err != nil {
		return errors.Wrap(err, errors.ErrServer,
			"unable to read client source file",
			"file", "image.go")
	}

	objectFile, err := client.SourceFS.ReadFile("object.go")
	if err != nil {
		return errors.Wrap(err, errors.ErrServer,
			"unable to read client source file",
			"file", "object.go")
	}

	scriptFile, err := client.SourceFS.ReadFile("script.go")
	if err != nil {
		return errors.Wrap(err, errors.ErrServer,
			"unable to read client source file",
			"file", "script.go")
	}

	gb, err := json.MarshalIndent(game, "", "  ")
	if err != nil {
		return errors.Wrap(err, errors.ErrServer,
			"unable to encode game for prompt",
			"game_id", game.ID.Value)
	}

	select {
	case <-ctx.Done():
		return errors.Context(ctx)
	default:
	}

	messages := []anthropic.MessageParam{}

	for _, m := range prompts.History {
		if m.Prompt.Set && m.Prompt.Valid {
			messages = append(messages, anthropic.NewUserMessage(
				anthropic.NewTextBlock(m.Prompt.Value)))
		}

		if m.Response.Set && m.Response.Valid {
			messages = append(messages, anthropic.NewAssistantMessage(
				anthropic.NewTextBlock(m.Response.Value)))
		}
	}

	messages = append(messages, anthropic.NewUserMessage(
		anthropic.NewTextBlock(prompts.Current.Prompt.Value)))
	messages = append(messages, anthropic.NewUserMessage(
		anthropic.NewTextBlock(
			"```current game definition\n"+
				string(gb)+"\n```\n")))

	count, err := p.cli.Messages.CountTokens(ctx,
		anthropic.MessageCountTokensParams{
			Model: anthropic.F(anthropic.ModelClaude3_7SonnetLatest),
			Thinking: anthropic.F(anthropic.ThinkingConfigParamUnion(
				&anthropic.ThinkingConfigEnabledParam{
					BudgetTokens: anthropic.F(p.budget),
					Type: anthropic.F(
						anthropic.ThinkingConfigEnabledTypeEnabled),
				})),
			Messages: anthropic.F(messages),
		})
	if err != nil {
		return errors.Wrap(err, errors.ErrServer,
			"unable to count tokens for prompt",
			"game_id", game.ID.Value,
			"prompt", prompts.Current.Prompt.Value)
	}

	prompts.Current.Thinking.Value += strconv.FormatInt(count.InputTokens, 10) +
		" tokens input\n\n"

	p.s.log.Log(ctx, logger.LvlDebug,
		"prompt token count",
		"game_id", game.ID.Value,
		"prompt", prompts.Current.Prompt.Value,
		"input_tokens", count.InputTokens)

	if err := updateGame(game, prompts); err != nil {
		return errors.Wrap(err, errors.ErrServer,
			"unable to update game with prompt token count",
			"game_id", game.ID.Value,
			"count", count.InputTokens)
	}

	select {
	case <-ctx.Done():
		return errors.Context(ctx)
	default:
	}

	stream := p.cli.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude3_7SonnetLatest),
		MaxTokens: anthropic.F(p.max),
		Thinking: anthropic.F(anthropic.ThinkingConfigParamUnion(
			&anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: anthropic.F(p.budget),
				Type: anthropic.F(
					anthropic.ThinkingConfigEnabledTypeEnabled),
			})),
		Messages: anthropic.F(messages),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(`You are an expert 2D game developer,
meticulously careful to avoid bugs and skilled working with limited resources.
You work with game2d, a framework which let's you express 2D games as game
definitions in JSON format. The game definitions include images, scripts and
objects which are used to define the game. The images are all SVG format images
encoded in base64. The client renders them into bitmap sprites. The scripts are
all written in Lua and are also encoded in base64. The total size of the JSON
game definition must not exceed the larger of 1 MB or your max tokens limit.

You will create one of these game definitions based on the user's prompt. And,
your response must include this game definition. The game definition must be
at the end of the response and must be immediately preceded by the text
"` + "```" + `game definition\n" and immediately followed by the text "\n` +
				"```" + `\n". The game definition "id" field must be a UUID and
can be random. In addition to the fields defined in the client, add a "name"
field to the game definition, containing the name of the game, a "description"
field to the game definition, containing the game controls and features, and add
an "icon" field containing a base64 encoded SVG image of an icon for the game.

The history of messages between you and the user has had the previous game
definitions replaced with the text "{{game definition}}". The current game
definition in the most recent user message is the last game definition in the
history.

The following text contains the Go source code of the game2d client, which is
used to run the game definitions.`),
			anthropic.NewTextBlock("```game.go\n" +
				string(gameFile) + "\n```\n"),
			anthropic.NewTextBlock("```image.go\n" +
				string(imageFile) + "\n```\n"),
			anthropic.NewTextBlock("```object.go\n" +
				string(objectFile) + "\n```\n"),
			anthropic.NewTextBlock("```script.go\n" +
				string(scriptFile) + "\n```\n"),
		}),
	})

	message := anthropic.Message{}

	for stream.Next() {
		select {
		case <-ctx.Done():
			return errors.Context(ctx)
		default:
		}

		e := stream.Current()
		message.Accumulate(e)

		switch delta := e.Delta.(type) {
		case anthropic.ContentBlockDeltaEventDelta:
			update := false

			if delta.Text != "" {
				prompts.Current.Thinking.Value += delta.Text

				update = true
			}

			if delta.Thinking != "" {
				prompts.Current.Thinking.Value += delta.Thinking

				update = true
			}

			if update {
				if err := updateGame(game, prompts); err != nil {
					return errors.Wrap(err, errors.ErrServer,
						"unable to update game with prompt delta",
						"game_id", game.ID.Value,
						"delta", delta)
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		return errors.Wrap(err, errors.ErrPrompt,
			"unable to get prompt response",
			"game_id", game.ID.Value,
			"prompt", prompts.Current.Prompt.Value)
	}

	if len(message.Content) == 0 {
		return errors.New(errors.ErrPrompt,
			"prompt response is empty",
			"prompt", prompts.Current.Prompt.Value)
	}

	msg := message.Content[len(message.Content)-1]

	index := strings.Index(msg.Text, "```game definition\n")
	if index > -1 {
		gs := msg.Text[index+len("```game definition\n"):]
		msg.Text = msg.Text[:index]

		index = strings.Index(gs, "```")
		if index == -1 {
			return errors.New(errors.ErrPrompt,
				"prompt response game definition is missing closing ```",
				"prompt", prompts.Current.Prompt.Value)
		}

		msg.Text += "{{game definition}}" + gs[index+3:]
		gs = gs[:index]

		var newGame *Game

		if err := json.Unmarshal([]byte(gs), &newGame); err != nil {
			return errors.Wrap(err, errors.ErrPrompt,
				"unable to decode game definition from prompt",
				"game_definition", gs,
				"prompt", prompts.Current.Prompt.Value)
		}

		if newGame == nil {
			return errors.New(errors.ErrPrompt,
				"prompt response game definition is empty",
				"prompt", prompts.Current.Prompt.Value)
		}

		newGame.AccountID = game.AccountID
		newGame.Debug = game.Debug
		newGame.Pause = game.Pause
		newGame.Public = game.Public
		newGame.ID = game.ID
		newGame.PreviousID = game.PreviousID
		newGame.Version = game.Version
		newGame.CreatedAt = game.CreatedAt
		newGame.CreatedBy = game.CreatedBy
		newGame.UpdatedAt = game.UpdatedAt
		newGame.UpdatedBy = game.UpdatedBy
		newGame.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusActive,
		}
		newGame.Source = game.Source
		newGame.CommitHash = game.CommitHash
		newGame.Tags = game.Tags
		newGame.Prompts = game.Prompts

		game = newGame
	}

	prompts.Current.Response.Value = msg.Text

	if err := updateGame(game, prompts); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to update game with prompt response",
			"game_id", game.ID.Value,
			"prompt", prompts.Current.Prompt.Value,
			"response", msg.Text)
	}

	return nil
}

// mockPrompter is a mock implementation of the Prompter interface.
type mockPrompter struct {
	s     *Server
	res   string
	delay time.Duration
}

// NewMockPrompter creates a new mock prompter with the given response, state
// and delay.
func NewMockPrompter(s *Server,
	res string,
	delay time.Duration,
) Prompter {
	return &mockPrompter{
		s:     s,
		res:   res,
		delay: delay,
	}
}

// Prompt sends a prompt to the mock prompter and returns the response and
// state.
func (m *mockPrompter) Prompt(ctx context.Context,
	prompts *Prompts,
	game *Game,
) error {
	res := "The AI has responded."

	hp := prompts.Current

	hp.Thinking = request.FieldString{Set: true}

	prompts.History = append(prompts.History, hp)
	prompts.Current = Prompt{
		Prompt:   prompts.Current.Prompt,
		Response: request.FieldString{Set: true, Valid: true, Value: res},
	}

	prompts.GameID = game.ID

	ps, err := promptsToFieldJSON(prompts)
	if err != nil {
		return errors.Wrap(err, errors.ErrServer,
			"unable to encode game prompts",
			"game_id", game.ID.Value,
			"prompts", prompts)
	}

	game.Prompts = ps

	time.Sleep(m.delay)

	if _, err := m.s.updateGame(ctx, game); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to update game with prompt response",
			"game_id", game.ID.Value,
			"prompts", prompts)
	}

	return nil
}
