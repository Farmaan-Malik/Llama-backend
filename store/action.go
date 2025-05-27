package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

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
	UserId string
	// QuestionsAsked   []string
	CorrectResponses int
	// Subject          string
	// Standard         string
}

type InititalPrompt struct {
	UserId   string
	Standard string
	Subject  string
}

func (s *Store) GetQuestion(a *Ask) error {
	ctx := context.Background()
	data, err := s.Redis.HGetAll(context.Background(), a.UserId).Result()
	if err != nil {
		fmt.Println("Error getting data from redis: ", err)
		return err
	}
	if len(data) == 0 {
		fmt.Println("No data found for user:", a.UserId)
		return errors.New(fmt.Sprint("No data found for user: ", a.UserId))
	}
	subject := data["subject"]
	questionsAsked := data["questionsAsked"]
	standard := data["standard"]
	fmt.Println("data: ", subject, questionsAsked, standard)
	prompt := fmt.Sprintf(`
You are an enthusiastic game show host for an educational quiz show, similar in style to "Who Wants to Be a Millionaire." Your job is to ask engaging multiple-choice questions to the user. The quiz should be based on the provided subject and suitable for the specified grade level.

IMPORTANT:
- Avoid any questions that are already included in the asked_questions list.
- Only return a well-formatted JSON object. Do not include markdown, code blocks, or commentary.
- The JSON must follow the structure below.

Inputs:
- subject: %s
- standard: %s
- asked_questions: %s

Output JSON structure:
{
  "dialogue": "An exciting welcome line from you, the host, leading into the question.",
  "question": "A unique and clear question related to the subject and suitable for the given standard.",
  "options": {
    "a": "Option A",
    "b": "Option B",
    "c": "Option C",
    "d": "Option D"
  },
  "correct": "b", // The correct option key
  "positive_dialogue": "What you say when the user gets it right!",
  "negative_dialogue": "What you say when the user gets it wrong!"
}

Your job:
- Generate a brand new question that is not in the asked_questions list.
- Ensure the content matches the subject and standard.
- Write in an engaging and conversational tone for the host's dialogue fields.
`, subject, standard, questionsAsked)

	llm, err := ollama.New(ollama.WithModel("llama3.2"))
	if err != nil {
		log.Fatal("Failed to initiate LLM model")
	}
	res, err := llm.Call(ctx, prompt, llms.WithTemperature(0.8))
	if err != nil {
		log.Fatal("Failed to call LLM model")
	}
	var r Question
	err = json.Unmarshal([]byte(res), &r)
	if err != nil {
		log.Fatalf("Failed to unmarshal res %v", err)
	}

	fmt.Println(r.Dialogue)
	fmt.Println(r.Question)
	fmt.Println(r.Options.A)
	fmt.Println(r.Options.B)
	fmt.Println(r.Options.C)
	fmt.Println(r.Options.D)
	fmt.Println(r.PositiveDialogue)
	fmt.Println(r.NegativeDialogue)
	return nil
}

func (s *Store) GetInitialData(i *InititalPrompt) {
	ctx := context.Background()
	jsonBytes, err := json.Marshal([]string{})
	if err != nil {
		fmt.Println("error marshalling questions: ", err)
	}
	jsonString := string(jsonBytes)
	cmd := s.Redis.HSet(ctx, i.UserId, map[string]any{
		"questionsAsked": jsonString,
		"standard":       i.Standard,
		"subject":        i.Subject,
	})
	boolCmd := s.Redis.Expire(ctx, i.UserId, 90*time.Minute)
	fmt.Println(cmd)
	fmt.Println(boolCmd)
}
