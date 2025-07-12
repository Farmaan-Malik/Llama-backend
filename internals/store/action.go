package store

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type ModelStore struct {
	Redis *redis.Client
}

type Question struct {
	Dialogue         string  `json:"dialogue"`
	PositiveDialogue string  `json:"positive_dialogue"`
	NegativeDialogue string  `json:"negative_dialogue"`
	Question         string  `json:"question"`
	Options          Options `json:"options"`
	Answer           string  `json:"answer"`
}

type Options struct {
	A string `json:"a"`
	B string `json:"b"`
	C string `json:"c"`
	D string `json:"d"`
}

type Ask struct {
	UserId           string `json:"user"`
	CorrectResponses int    `json:"correctResponses"`
}

type InititalPrompt struct {
	UserId   string `json:"user"`
	Standard string `json:"standard"`
	Subject  string `json:"subject"`
}

func (s *ModelStore) GetAllH(ctx context.Context, key string) (map[string]string, error) {
	return s.Redis.HGetAll(ctx, key).Result()
}

func (s *ModelStore) GetQuestion(w http.ResponseWriter, ctx context.Context, a *Ask) (*Question, error) {
	data, err := s.Redis.HGetAll(ctx, a.UserId).Result()
	if err != nil {
		fmt.Println("Redis error: ", err)
		return nil, fmt.Errorf("error getting data from redis: %s", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no data found for user: %s", a.UserId)
	}

	subject := data["subject"]
	grade := data["standard"]
	questionsAsked := data["questionsAsked"]

	var askedQuestions []string
	err = json.Unmarshal([]byte(questionsAsked), &askedQuestions)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling questionsAsked: %w", err)
	}

	askedText := ""
	for i, q := range askedQuestions {
		askedText += fmt.Sprintf("%d. %s\n", i+1, q)
	}

	prompt := fmt.Sprintf(`
You are Llama-sama — a thoughtful, caring, and wise teacher who believes every student has the potential to shine.
Your goal is to help them learn and grow through meaningful, supportive questions that spark curiosity and confidence.

Generate a multiple-choice question suitable for a student of class %s in the subject %s.

🟢 Guidelines:
- The question should be fully self-contained and **NOT require a reading passage, story, or external material**.
- The question should be **factual**, **direct**, and **clear** — NOT interpretive or abstract.
- Start with a warm, gentle introduction to ease the student into the question but if a question has been asked before, skip the dialogue.
- Do NOT include the student's grade or subject in your speech.
- Do NOT include options or answers — just the question itself.
- Do NOT include images, diagrams, or formatting — only plain text.
- IMPORTANT: Do NOT repeat any of these previously asked questions: %s
- It must be answerable in **1 to 3 words only**.

Keep your tone warm, encouraging, and simple.
`, grade, subject, askedText)

	llm, err := ollama.New(ollama.WithModel("llama-sama"), ollama.WithServerURL("http://localhost:11434"))
	if err != nil {
		return nil, fmt.Errorf("error initializing model: %s", err)
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
		llms.TextParts(llms.ChatMessageTypeHuman, "Generate"),
	}

	var questionBuffer string
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return nil, fmt.Errorf("streaming unsupported")
	}

	_, err = llm.GenerateContent(ctx, messages, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		text := string(chunk)
		questionBuffer += text
		fmt.Fprintf(w, "event: question\ndata: %s\n\n", text)
		flusher.Flush()
		return nil
	}))
	if err != nil {
		return nil, fmt.Errorf("error generating question: %s", err)
	}
	tryMetaExtraction := func(response string) (string, error) {
		metaStart := strings.Index(response, "{")
		metaEnd := strings.LastIndex(response, "}")
		if metaStart == -1 || metaEnd == -1 || metaEnd <= metaStart {
			return "", fmt.Errorf("model did not produce required metadata block")
		}
		return response[metaStart : metaEnd+1], nil
	}

	metaPrompt := fmt.Sprintf(`
For this question: %s,
Generate a JSON block with keys:
- "question": (the question itself),
- "options": (an object with keys "A", "B", "C", "D"),
- "answer": (correct option key e.g. "C").
- Only one of the options should be correct.
- Close it properly. In total there should be exactly 4 curly brackets.

Return ONLY the JSON. No intro or explanation.

{
  "options": {
    "A": "Option A",
    "B": "Option B",
    "C": "Option C",
    "D": "Option D"
  },
  "question": "question goes here",
  "answer": "C"
}
`, questionBuffer)

	response, err := llm.Call(ctx, metaPrompt)
	if err != nil {
		return nil, fmt.Errorf("error generating metadata: %s", err)
	}

	metaJSON, err := tryMetaExtraction(response)
	if err != nil {
		retryPrompt := fmt.Sprintf(`
For this question: %s,
Generate a strict JSON object with:
- "question": string,
- "options": { "A": "", "B": "", "C": "", "D": "" },
- "answer": correct option key.
- Only one of the options should be correct.
- Close it properly. In total there should be exactly 4 curly brackets.

	{
		"options": {
		"A": "Option",
		"B": "Option",
		"C": "Option",
		"D": "Option",
		},
		"answer":"Answer Key",
		"question": "Question",
	}


Return ONLY JSON, no extra text.
`, questionBuffer)
		response, err = llm.Call(ctx, retryPrompt)
		if err != nil {
			return nil, fmt.Errorf("metadata retry failed: %s", err)
		}
		metaJSON, err = tryMetaExtraction(response)
		if err != nil {
			fmt.Fprintf(w, "event: error\ndata: model did not produce valid metadata block after retry\n\n")
			flusher.Flush()
			return nil, fmt.Errorf("final metadata retry failed")
		}
	}

	compactMeta := strings.ReplaceAll(metaJSON, "\n", "")
	compactMeta = strings.ReplaceAll(compactMeta, "\r", "")
	fmt.Fprintf(w, "event: metadata\ndata: %s\n\n", compactMeta)
	fmt.Fprintf(w, "event: done\ndata: ok\n\n")
	flusher.Flush()

	var q Question
	err = json.Unmarshal([]byte(metaJSON), &q)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata: %s", err)
	}

	askedQuestions = append(askedQuestions, q.Question)
	updated, err := json.Marshal(askedQuestions)
	if err != nil {
		return nil, fmt.Errorf("error marshalling questionsAsked: %s", err)
	}

	if err := s.Redis.HSet(ctx, a.UserId, "questionsAsked", string(updated)).Err(); err != nil {
		return nil, fmt.Errorf("error saving asked questions: %s", err)
	}

	return &q, nil
}

func (s *ModelStore) GetInitialData(ctx context.Context, i *InititalPrompt) error {
	jsonBytes, err := json.Marshal([]string{})
	if err != nil {
		return fmt.Errorf("error marshalling question: %w", err)
	}
	jsonString := string(jsonBytes)

	fmt.Println("Initializing data for:", i)

	if err := s.Redis.Del(ctx, i.UserId).Err(); err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}
	cmd := s.Redis.HSet(ctx, i.UserId, map[string]any{
		"questionsAsked": jsonString,
		"standard":       i.Standard,
		"subject":        i.Subject,
	})
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("error saving initial session data: %w", err)
	}
	if err := s.Redis.Expire(ctx, i.UserId, 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("error setting expiration: %w", err)
	}

	return nil
}
