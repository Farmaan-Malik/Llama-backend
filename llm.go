package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type LlmResponse struct {
	Dialogue          string  `json:"dialogue"`
	Question          string  `json:"question"`
	Options           Options `json:"options"`
	Correct           string  `json:"correct"`
	Positive_Dialogue string  `json:"positive_dialogue"`
	Negative_Dialogue string  `json:"negative_dialogue"`
}

type Options struct {
	A string `json:"a"`
	B string `json:"b"`
	C string `json:"c"`
	D string `json:"d"`
}

func GetQuestion() {
	ctx := context.Background()
	prompt := `Pretend you are a host for a show similar to "who wants to be a millianaire". you need to ask a question and give options out of which one will be correct. The whole response should be in json format. no markdown just pure json. The response should have parameters like:
	dialogue: dialogue from the host(thats you),
	question: ask a question here,
	options: {
	a:some Option,
	b:some Option,
	c:some Option,
	d:some Option
	},
	correct: correct option,
	positive_dialogue: if the user answers correct,
	negative_dialogue: if the user's answer is wrong
	`
	llm, err := ollama.New(ollama.WithModel("llama3.2"))
	if err != nil {
		log.Fatal("Failed to initiate LLM model")
	}
	res, err := llm.Call(ctx, prompt, llms.WithTemperature(0.8))
	if err != nil {
		log.Fatal("Failed to call LLM model")
	}
	var r LlmResponse
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
	fmt.Println(r.Positive_Dialogue)
	fmt.Println(r.Negative_Dialogue)
}
