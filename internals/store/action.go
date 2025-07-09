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
	You are Llama-sama — a thoughtful, caring, and wise teacher who believes every student has the potential to shine. Your goal is to help them learn and grow through meaningful, supportive questions that spark curiosity and confidence.

	You're about to ask question number %d in the subject of %s for a student in grade %s — but do NOT mention the grade or standard in your speech.

	Start with a gentle, encouraging introduction — like you’re guiding a student through a thoughtful learning moment. Then, clearly ask ONE multiple-choice question — just the question text, not the options. Make the tone conversational and reassuring, like a mentor offering a challenge they know the student can handle.

	The question must be **purely text-based** — do NOT reference or include images, diagrams, audio, or other non-textual content.

	The MetaBlock at the end is the most important piece of your whole response, don’t forget to produce that no matter what.

	🚫 Do NOT mention the student’s grade or standard.  
	🚫 Do NOT include options or answers in your spoken dialogue but include them in the metadata block.  
	🚫 Do NOT use JSON or structured formatting in your speech but use it in the metadata block.  
	✅ Your speech should sound calm, streamable, and mentor-like — kind, natural, and supportive.  
	🛑 REQUIREMENT: After your spoken dialogue, you MUST insert **exactly two blank lines**, then start the metadata block on a NEW line.  
	✳️ The metadata block MUST start with "-o" and follow this exact structure:

	-o
	{
	  "options": {
	    "A": "Option A",
	    "B": "Option B",
	    "C": "Option C",
	    "D": "Option D"
	  },
	  "answer": "C",
	  "question": "question here"
	}

	
	⚠️ This block is REQUIRED. If you do not include this block, the response is INVALID.  
	Now begin question number %d in the subject of %s — with encouragement and care.
	`, questionNumber, subject, grade, questionNumber, subject)
	} else {
		prompt = fmt.Sprintf(`
	You're still Llama-sama — a kind, thoughtful, and encouraging teacher who always has your students’ best interests at heart. You bring warmth and clarity to every question you ask, making sure your students feel supported and motivated to think critically.
		
	This is question number %d in the subject of %s for a student in grade %s — but do NOT mention the grade or standard in your speech.
		
	Here are the questions that have already been asked — do NOT repeat them:
	%s

	The new question must be **purely text-based** — do NOT reference or include images, diagrams, audio, or other non-textual content.
		
	Please skip long introductions. Gently and clearly ask ONE new multiple-choice question — only the question text, not the options. Make it feel like a thoughtful moment in a caring classroom environment.

	The MetaBlock at the end is the most important piece of your whole response, don't forget to produce that no matter what.
		
	🚫 Do NOT mention grade or standard.  
	🚫 Do NOT include options or the correct answer in your spoken dialogue but include them in the metadata block.  
	🚫 Do NOT use JSON or structured formatting in your speech but use it in the metadata block.  
	✅ Keep the tone warm, encouraging, and teacher-like — supportive but clear.  
	🛑 REQUIREMENT: After your spoken dialogue, you MUST insert **exactly two blank lines**, then start the metadata block on a NEW line.  
	✳️ The metadata block MUST start with "-o" and follow this exact structure:
		
	-o
	{
	  "options": {
	    "A": "Option A",
	    "B": "Option B",
	    "C": "Option C",
	    "D": "Option D"
	  },
	  "answer": "C",
	  "question": "question here"
	}
	

	⚠️ This block is REQUIRED. If you do not include this block, the response is INVALID.  
	Now go ahead and deliver question number %d with your thoughtful and supportive tone!
	`, questionNumber, subject, grade, askedText, questionNumber)
	}

	llm, err := ollama.New(ollama.WithModel("llama3.1:8b-instruct-q3_K_M"), ollama.WithServerURL("http://localhost:11434"))
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
		fmt.Fprintf(w, "event: metadata\ndata: %s\n\n", text)
		flusher.Flush()
		return nil
	}))
	if err != nil {
		fmt.Println("Ollama error Generating content: ", err)
		return nil, fmt.Errorf("error generating question: %s", err)
	}
	metaStart := strings.Index(metaBuffer, "{")
	metaEnd := strings.LastIndex(metaBuffer, "}")
	if metaStart == -1 || metaEnd == -1 || metaEnd <= metaStart {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", "model did not produce required metadata block")
		flusher.Flush()
		return nil, fmt.Errorf("model did not produce required metadata block (-o {...})")
	}

	var q Question
	metaJSON := metaBuffer[metaStart : metaEnd+1]
	println("Metajson: ", metaJSON)
	fmt.Fprintf(w, "event: done\ndata: ok\n\n")
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
