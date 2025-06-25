package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

type Question struct {
	Dialogue         string  `json:"dialogue"`
	PositiveDialogue string  `json:"positive_dialogue"`
	NegativeDialogue string  `json:"negative_dialogue"`
	Question         string  `json:"question"`
	Options          Options `json:"options"`
	Correct          string  `json:"correct"`
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

func (s *Store) GetQuestion(ctx context.Context, a *Ask) (*Question, error) {
	data, err := s.Redis.HGetAll(ctx, a.UserId).Result()
	if err != nil {
		fmt.Println("Redis error: ", err)
		return nil, fmt.Errorf("error getting data from redis: %s", err)

	}
	fmt.Println("Data: ", data)
	fmt.Println("Length: ", len(data))
	if len(data) == 0 {
		fmt.Println("No data found for user:", a.UserId)
		return nil, fmt.Errorf("no data found for user: %s", a.UserId)
	}
	subject := data["subject"]
	questionsAsked := data["questionsAsked"]
	standard := data["standard"]
	fmt.Println("data: ", subject, questionsAsked, standard)
	prompt := fmt.Sprintf(`subject: %s
	standard: %s
	questionsAsked:%s`, subject, standard, questionsAsked)

	llm, err := ollama.New(ollama.WithModel("MrQuizzler"), ollama.WithServerURL("http://ollama:11434"))
	if err != nil {
		fmt.Println("Ollama error: ", err)
		return nil, fmt.Errorf("error communicating with model: %s", err)
	}
	res, err := llm.Call(ctx, prompt)
	if err != nil {
		fmt.Println("Ollama error while calling: ", err)
		return nil, fmt.Errorf("error generating question: %s", err)
	}
	var r Question
	err = json.Unmarshal([]byte(res), &r)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling questions: %s", err)
	}
	if questionsAsked == "" {
		questionsAsked = "[]"
	}
	var askedQuestions []string
	err = json.Unmarshal([]byte(questionsAsked), &askedQuestions)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling questions for redis: %s", err)
	}
	askedQuestions = append(askedQuestions, r.Question)
	updatedQuestion, err := json.Marshal(askedQuestions)
	if err != nil {
		return nil, fmt.Errorf("error marshalling updated questionsAsked: %s", err)
	}
	cmd := s.Redis.HSet(ctx, a.UserId, "questionsAsked", string(updatedQuestion))
	if err := cmd.Err(); err != nil {
		return nil, fmt.Errorf("error saving updated questionsAsked to Redis: %s", err)
	}
	fmt.Println(cmd)
	fmt.Println(r.Dialogue)
	fmt.Println(r.Question)
	fmt.Println(r.Options.A)
	fmt.Println(r.Options.B)
	fmt.Println(r.Options.C)
	fmt.Println(r.Options.D)
	fmt.Println(r.PositiveDialogue)
	fmt.Println(r.NegativeDialogue)
	fmt.Println("Correct: ", r.Correct)
	return &r, nil
}

func (s *Store) GetInitialData(i *InititalPrompt) error {
	ctx := context.Background()
	jsonBytes, err := json.Marshal([]string{})
	if err != nil {
		return fmt.Errorf("error marshalling question: %s", err)
	}
	jsonString := string(jsonBytes)
	cmd := s.Redis.HSet(ctx, i.UserId, map[string]any{
		"questionsAsked": jsonString,
		"standard":       i.Standard,
		"subject":        i.Subject,
	})
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("error saving updated questionsAsked to Redis: %s", err)
	}
	boolCmd := s.Redis.Expire(ctx, i.UserId, 90*time.Minute)
	fmt.Println(cmd)
	fmt.Println(boolCmd)
	return nil
}
