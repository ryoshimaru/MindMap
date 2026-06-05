package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/ryoshimaru/MindMap/internal/ai"
	"github.com/ryoshimaru/MindMap/internal/domain"
	apihttp "github.com/ryoshimaru/MindMap/internal/http"
	"github.com/ryoshimaru/MindMap/internal/http/handlers"
	"github.com/ryoshimaru/MindMap/internal/service"
	"github.com/ryoshimaru/MindMap/internal/store"
)

func main() {
	store := configureStore()
	defaultProvider, factories := configureAIProviders()

	services := handlers.Services{
		Auth:      service.NewAuthService(store),
		Goals:     service.NewGoalService(store),
		Tasks:     service.NewTaskService(store),
		AI:        service.NewAIService(store, defaultProvider, factories),
		Analytics: service.NewAnalyticsService(store),
		Versions:  service.NewVersionService(store),
	}

	handler := handlers.New(services)
	router := apihttp.NewRouter(handler)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	log.Printf("GoalMind API listening on :%s with AI_PROVIDER=%s", port, defaultProvider)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("error trying start the server: ", err)
	}
}

func configureStore() store.Store {
	dbURL := strings.TrimSpace(os.Getenv("DB_URL"))
	if dbURL == "" {
		dbURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dbURL == "" {
		log.Print("WARNING: DB_URL/DATABASE_URL is not set; using in-memory store. Data will be lost after backend restart.")
		return store.NewMemoryStore()
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	log.Print("Using PostgreSQL store")
	return store.NewPostgresStore(db)
}

func configureAIProviders() (domain.AIProviderName, map[domain.AIProviderName]service.ProviderFactory) {
	client := &http.Client{Timeout: 45 * time.Second}
	geminiEnvKey := os.Getenv("GEMINI_API_KEY")
	deepSeekEnvKey := os.Getenv("DEEPSEEK_API_KEY")

	factories := map[domain.AIProviderName]service.ProviderFactory{
		domain.AIProviderMock: func(apiKey string) service.Provider {
			return ai.NewMockProvider()
		},
		domain.AIProviderGemini: func(apiKey string) service.Provider {
			if strings.TrimSpace(apiKey) == "" {
				apiKey = geminiEnvKey
			}
			return ai.NewGeminiProvider(client, apiKey, getenv("GEMINI_MODEL", "gemini-1.5-flash"))
		},
		domain.AIProviderDeepSeek: func(apiKey string) service.Provider {
			if strings.TrimSpace(apiKey) == "" {
				apiKey = deepSeekEnvKey
			}
			return ai.NewDeepSeekProvider(client, apiKey, getenv("DEEPSEEK_MODEL", "deepseek-chat"))
		},
	}

	switch strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER"))) {
	case "mock":
		return domain.AIProviderMock, factories
	case "gemini":
		return domain.AIProviderGemini, factories
	case "deepseek":
		return domain.AIProviderDeepSeek, factories
	default:
		return domain.AIProviderGemini, factories
	}
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
