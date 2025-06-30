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
	fmt.Println("Data: ", data)
	fmt.Println("Length: ", len(data))
	if len(data) == 0 {
		fmt.Println("No data found for user:", a.UserId)
		return nil, fmt.Errorf("no data found for user: %s", a.UserId)
	}
	subject := data["subject"]
	questionsAsked := data["questionsAsked"]
	grade := data["standard"]
	var askedQuestions []string
	err = json.Unmarshal([]byte(questionsAsked), &askedQuestions)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling questionsAsked: %w", err)
	}
	questionNumber := len(askedQuestions) + 1
	askedText := ""
	for i, q := range askedQuestions {
		askedText += fmt.Sprintf("%d. %s\n", i+1, q)
	}
	fmt.Println(subject, "Here")
	var prompt string
	if questionNumber == 1 {
		prompt = fmt.Sprintf(`
You are a fun and energetic game show host speaking directly to a contestant.

You're about to ask question number %d in the subject of %s for a student in grade %s — but do NOT mention the grade or standard in your speech.

Start with an exciting and friendly introduction, like you're kicking off a quiz show! Then, ask ONE multiple-choice question — just the question text, not the options.

🚫 Do NOT mention the contestant's grade or standard.
🚫 Do NOT include options or answers in your spoken dialogue but include them in the meta data block.
🚫 Do NOT use JSON or structured formatting in your speech but use it in the metadata block.
✅ Your speech should sound natural and be streamable line-by-line, like a live quiz host.
🛑 REQUIREMENT: After your spoken dialogue, you MUST insert **exactly two blank lines**, then start the metadata block on a NEW line.
✳️ The metadata block MUST start with "-o" and follow this exact structure:

-o
{
  "options": {
  //these should be the actual options and one of them should be the right answer for the question.
    "A": "Option A", // replace option A with the actual option
    "B": "Option B", // replace option B with the actual option
    "C": "Option C", // replace option C with the actual option
    "D": "Option D" // replace option D with the actual option
  },
  "answer": "C", //this should be the right answer of the question and should be one of the options,
	"question":"question here" // this should be the question that was asked 

}
⚠️ This block is REQUIRED. If you do not include this block, the response is INVALID.
Now begin question number %d in the subject of %s.
`, questionNumber, subject, grade, questionNumber, subject)

	} else {
		prompt = fmt.Sprintf(`
You're a lively game show host, and the quiz is already underway!

This is question number %d in the subject of %s for a student in grade %s — but do NOT mention the grade or standard in your speech.

Here are the questions that have already been asked — do NOT repeat them:
%s

Skip introductions. Jump straight into a new question, keeping the energy high!  
Ask ONE new multiple-choice question — only the question text, not the options.

🚫 Do NOT mention grade or standard.
🚫 Do NOT include options or the correct answer in your spoken dialogue but include them in the metadata block.
🚫 Do NOT use JSON or structured formatting in your speech but use it in the metadata block.
✅ Keep it natural, streamable, and spoken like a quiz host.
🛑 REQUIREMENT: After your spoken dialogue, you MUST insert **exactly two blank lines**, then start the metadata block on a NEW line.
✳️ The metadata block MUST start with "-o" and follow this exact structure:

-o
{
  "options": {
   //these should be the actual options and one of them should be the right answer for the question.
    "A": "Option A", // replace option A with the actual option
    "B": "Option B", // replace option B with the actual option
    "C": "Option C", // replace option C with the actual option
    "D": "Option D" // replace option D with the actual option
  },
  "answer": "C", //this should be the right answer of the question and should be one of the options,
  "question":"question here" // this should be the question that was asked 
}
⚠️ This block is REQUIRED. If you do not include this block, the response is INVALID.
Now go ahead and deliver question number %d!
`, questionNumber, subject, grade, askedText, questionNumber)
	}
	llm, err := ollama.New(ollama.WithModel("llama3.2"), ollama.WithServerURL("http://localhost:11434"))
	if err != nil {
		fmt.Println("Ollama error: ", err)
		return nil, fmt.Errorf("error communicating with model: %s", err)
	}
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
		llms.TextParts(llms.ChatMessageTypeHuman, "Generate a question"),
	}
	var metaBuffer string
	var capturingMeta bool
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return nil, fmt.Errorf("client does not support streaming")
	}
	println("Calling Generate")
	_, err = llm.GenerateContent(ctx, messages, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		text := string(chunk)
		if !capturingMeta {
			if strings.Contains(text, "-o") {
				capturingMeta = true
				return nil
			}
			// dialogueBuffer += text
			fmt.Println(text)
			fmt.Fprintf(w, "event: question\ndata: %s\n\n", text)
			flusher.Flush()
			return nil
		}
		fmt.Println(text)
		metaBuffer += text
		return nil
	}))
	if err != nil {
		fmt.Println("Ollama error Generating content: ", err)
		return nil, fmt.Errorf("error generating question: %s", err)
	}
	metaStart := strings.Index(metaBuffer, "{")
	metaEnd := strings.LastIndex(metaBuffer, "}")
	if metaStart == -1 || metaEnd == -1 || metaEnd <= metaStart {
		return nil, fmt.Errorf("model did not produce required metadata block (-o {...})")
	}
	var q Question
	metaJSON := metaBuffer[metaStart : metaEnd+1]
	fmt.Fprintf(w, "event: metadata\ndata: %s\n\n", metaJSON)
	flusher.Flush()
	err = json.Unmarshal([]byte(metaJSON), &q)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling questions: %s", err)
	}
	askedQuestions = append(askedQuestions, q.Question)
	updatedQuestion, err := json.Marshal(askedQuestions)
	if err != nil {
		return nil, fmt.Errorf("error marshalling updated questionsAsked: %s", err)
	}
	cmd := s.Redis.HSet(ctx, a.UserId, "questionsAsked", string(updatedQuestion))
	if err := cmd.Err(); err != nil {
		return nil, fmt.Errorf("error saving updated questionsAsked to Redis: %s", err)
	}
	if q.Question == "" || q.Options.A == "" || q.Answer == "" {
		return nil, fmt.Errorf("incomplete metadata fields")
	}
	fmt.Printf("META DATA: %s", metaJSON)
	return &q, nil
}

func (s *ModelStore) GetInitialData(i *InititalPrompt) error {
	ctx := context.Background()
	jsonBytes, err := json.Marshal([]string{})
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error marshalling question: %s", err)
	}
	jsonString := string(jsonBytes)
	cmd := s.Redis.HSet(ctx, i.UserId, map[string]any{
		"questionsAsked": jsonString,
		"standard":       i.Standard,
		"subject":        i.Subject,
	})
	if err := cmd.Err(); err != nil {
		fmt.Println(err)
		return fmt.Errorf("error saving updated questionsAsked to Redis: %s", err)
	}
	boolCmd := s.Redis.Expire(ctx, i.UserId, 90*time.Minute)
	fmt.Println(cmd)
	fmt.Println(boolCmd)
	return nil
}
