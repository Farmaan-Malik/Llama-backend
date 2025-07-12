## Llama Sama Backend

The backend server for **Llama Sama** — a personalized, AI-powered quiz app designed to help students learn through factual, multiple-choice questions in a fun and supportive way.

This Go-based backend uses the **Gin** web framework and includes:

- **REST + SSE APIs** to deliver real-time quiz questions and feedback
- **AI-generated content** using Ollama and custom prompt templates
- **MongoDB** for user profiles, quiz history, and game data
- **Redis** for session storage and deduplication of quiz questions
- **JWT-based authentication** for secure access
- **Modular architecture** designed for scalability and maintainability

> The AI follows the persona of “Llama-sama” — a kind, wise teacher who delivers direct, factual questions without unnecessary fluff. Each game session is unique, adaptive, and stored efficiently to ensure the user never sees the same question twice.



---

## Getting Started

### 1. Install Go

Download and install Go:  
[https://go.dev/dl](https://go.dev/dl)

Check your version:
```bash
go version
```
### 2.Install Dependencies
To install project dependencies, run
```bash
go mod tidy
```

### 3.Install Ollama
To run the AI model locally, you'll need to install [Ollama](https://ollama.com/):

#### On macOS:
```bash
brew install ollama
```
#### On Windows:

1. Download the installer from the official Ollama website: [https://ollama.com/download](https://ollama.com/download)
2. Run the installer and follow the on-screen instructions.
3. After installation, ensure that the `ollama` command is available in your terminal (e.g., PowerShell or Command Prompt). You may need to restart your terminal or add Ollama to your system's PATH manually if it isn't recognized.

### 4. Clone the Git Repository
```
git clone https://github.com/Farmaan-Malik/Llama-backend.git
```

### 5. Configure Environment Variables 
Make a .env file in the root of your project with the following content:
```
MONGO_URI=mongodb://root:example@localhost:27017/
REDIS_ADDR=<redisAddress>
REDIS_PW=<redisPassword>
JWT_KEY=<JWT_KEY>
```


### 6. Run Docker & Ollama
Make sure docker is Running and run Ollama
```bash
ollama serve
```
### 7. Create Model using ModelFile (options)
```bash
ollama create llama-sama -f Modelfile
```
If using other models navigate to
>internals/store/action.go

Look for
>llm, err := ollama.New(ollama.WithModel("llama-sama"), ollama.WithServerURL("http://localhost:11434"))


 and replace **"llama-sama"** with your own model's name

### 8. Run the Server and Docker Services
#### If air is installed
```bash
air
```
#### Without air
Run Docker Services
```bash
docker compose up ---build
```
Run the server
```bash
go run cmd/api/. 
```

---

## Concepts & Architecture

- **Clean Architecture** — Organized the code into layers (routes, services, models, etc.) for better maintainability and testing.
- **SSE (Server-Sent Events)** — Delivered real-time quiz questions and feedback without WebSockets or polling.
- **Authentication Flow** — Implemented secure signup/login with hashed passwords and token-based sessions.
- **Session Management** — Handled user quiz progress and expiration using Redis and unique keys.
- **Streaming AI Responses** — Learned to pipe LLM responses token-by-token to the client via SSE, improving UX.
- **Prompt Engineering** — Designed reusable prompts for question generation tailored to student grade and subject.

---

## How It Works

The Llama Sama backend powers the quiz experience by dynamically generating and streaming questions using Go, Gin, Redis, MongoDB, and an Ollama LLM model.

### 1. User Starts a Quiz
- The mobile app sends a request with the selected **grade**, **subject**, and **user ID**.
- Only authenticated users can initiate a quiz.

### 2. Authentication
- The backend uses **JWT-based authentication**.
- On login/signup, the user receives a **JWT token** which is stored on the client.
- All subsequent API requests include the token in the `Authorization` header.
- The server validates the token on each request to authorize access.

### 3. Session Initialization
- A session is created in **Redis**, storing:
  - User ID
  - Previously asked questions (to avoid duplicates)

### 4. Prompt Generation
- A dynamic prompt is constructed in Go based on:
  - The selected grade and subject
  - The current question number
  - The tone and persona of **Llama Sama**
- This prompt is passed to an **Ollama** model for question generation.

### 5. AI-Powered Question Generation
- Ollama returns a factual, multiple-choice question with 4 options.
- The backend parses and stores the question and metadata in Redis.

### 6. Streaming the Question
- The question is streamed to the frontend in real time using **Server-Sent Events (SSE)**.
- Users receive tokens as they are generated for a smooth experience.
- The options and correct answer are also streamed and stored client-side to allow instant feedback upon answer selection.

### 7. Game Progression
- The process repeats for a total of **5 questions** or until **10 minutes** pass.

