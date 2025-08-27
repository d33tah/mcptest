package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var getCurrentTimePostSpec = map[string]interface{}{
	"summary":     "Get Current Time",
	"description": "Get current time in a specific timezone",
	"operationId": "tool_get_current_time_post",
	"requestBody": map[string]interface{}{
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/get_current_time_form_model",
				},
			},
		},
		"required": true,
	},
	"responses": map[string]interface{}{
		"200": map[string]interface{}{
			"description": "Successful Response",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"title": "Response Tool Get Current Time Post",
					},
				},
			},
		},
		"422": map[string]interface{}{
			"description": "Validation Error",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"$ref": "#/components/schemas/HTTPValidationError",
					},
				},
			},
		},
	},
	"security": []map[string]interface{}{
		{"HTTPBearer": []interface{}{}},
	},
}

var openapiSpecData = map[string]interface{}{
	"openapi": "3.1.0",
	"info": map[string]interface{}{
		"title":       "mcp-time",
		"description": "mcp-time MCP Server",
		"version":     "1.13.1",
	},
	"paths": map[string]interface{}{
		"/get_current_time": map[string]interface{}{
			"post": getCurrentTimePostSpec,
		},
	},
	"components": map[string]interface{}{
		"schemas": map[string]interface{}{
			"HTTPValidationError": map[string]interface{}{
				"properties": map[string]interface{}{
					"detail": map[string]interface{}{
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/ValidationError",
						},
						"type":  "array",
						"title": "Detail",
					},
				},
				"type":  "object",
				"title": "HTTPValidationError",
			},
			"ValidationError": map[string]interface{}{
				"properties": map[string]interface{}{
					"loc": map[string]interface{}{
						"items": map[string]interface{}{
							"anyOf": []map[string]interface{}{
								{"type": "string"},
								{"type": "integer"},
							},
						},
						"type":  "array",
						"title": "Location",
					},
					"msg": map[string]interface{}{
						"type":  "string",
						"title": "Message",
					},
					"type": map[string]interface{}{
						"type":  "string",
						"title": "Error Type",
					},
				},
				"type":     "object",
				"required": []string{"loc", "msg", "type"},
				"title":    "ValidationError",
			},
			"get_current_time_form_model": map[string]interface{}{
				"properties": map[string]interface{}{
					"timezone": map[string]interface{}{
						"type":        "string",
						"title":       "Timezone",
						"description": "timezone name",
					},
				},
				"type":     "object",
				"required": []string{"timezone"},
				"title":    "get_current_time_form_model",
			},
		},
		"securitySchemes": map[string]interface{}{
			"HTTPBearer": map[string]interface{}{
				"type":   "http",
				"scheme": "bearer",
			},
		},
	},
}

type TimeResponse struct {
	Timezone string `json:"timezone"`
	DateTime string `json:"datetime"`
	IsDST    bool   `json:"is_dst"`
}

var getCurrentTimeResponseData = TimeResponse{
	Timezone: "Europe/Warsaw",
	DateTime: "2025-08-26T21:37:15+02:00",
	IsDST:    true,
}

type RequestLog struct {
	Timestamp  string      `json:"timestamp"`
	Method     string      `json:"method"`
	URI        string      `json:"uri"`
	RemoteAddr string      `json:"remote_addr"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body,omitempty"`
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		logEntry := RequestLog{
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Method:     r.Method,
			URI:        r.RequestURI,
			RemoteAddr: r.RemoteAddr,
			Headers:    r.Header,
			Body:       string(bodyBytes),
		}
		logJSON, err := json.Marshal(logEntry)
		if err != nil {
			log.Printf("Error marshalling log entry: %v", err)
		} else {
			log.Println(string(logJSON))
		}
		next.ServeHTTP(w, r)
	})
}

func serverHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "uvicorn")
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openapiSpecData)
}

func handleGetCurrentTime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getCurrentTimeResponseData)
}

//go:embed public/index.html
var indexHTML string

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}

// SessionStore przechowuje aktywne sesje klientów. Dostęp jest synchronizowany.
type SessionStore struct {
	sessions map[string]chan string
	mu       sync.Mutex
}

// Globalna instancja przechowująca sesje.
var store = &SessionStore{
	sessions: make(map[string]chan string),
}

// add tworzy nową sesję, generuje dla niej unikalny ID i kanał do komunikacji.
func (ss *SessionStore) add() (string, chan string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	sessionID := uuid.New().String()
	messageChan := make(chan string, 1) // Buforowany kanał na 1 wiadomość

	ss.sessions[sessionID] = messageChan
	log.Printf("✅ Sesja utworzona: %s", sessionID)
	return sessionID, messageChan
}

// get pobiera kanał komunikacyjny dla danej sesji.
func (ss *SessionStore) get(sessionID string) (chan string, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ch, found := ss.sessions[sessionID]
	return ch, found
}

// remove usuwa sesję po rozłączeniu klienta.
func (ss *SessionStore) remove(sessionID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, sessionID)
	log.Printf("❌ Sesja usunięta: %s", sessionID)
}

// JSONRPCRequest modeluje strukturę przychodzących zapytań JSON-RPC.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	ID      json.RawMessage `json:"id"` // ID może być liczbą lub stringiem, RawMessage zachowuje oryginalny format.
}

// sseHandler obsługuje endpoint Server-Sent Events (/sse).
func sseHandler(w http.ResponseWriter, r *http.Request) {
	// Krok 1: Klient próbuje POST, zwracamy błąd (zgodnie z logami)
	if r.Method == "POST" {
		log.Printf("⚠️ Otrzymano POST na /sse, odpowiadam 405 Method Not Allowed")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Krok 2: Klient wysyła GET, aby nawiązać połączenie SSE
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming nie jest wspierany!", http.StatusInternalServerError)
		return
	}

	// Ustawienie nagłówków wymaganych przez SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Tworzenie nowej sesji i kanału do wysyłania wiadomości
	sessionID, messageChan := store.add()
	defer store.remove(sessionID)

	// Krok 3: Wysłanie początkowych wiadomości SSE (zgodnie z logami)
	log.Printf("➡️  Klient połączony z /sse, wysyłam dane inicjalizacyjne dla sesji %s", sessionID)

	// Wiadomość 1: Endpoint dla dalszej komunikacji
	endpointData := fmt.Sprintf("/messages/?session_id=%s", sessionID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointData)
	flusher.Flush()

	// Wiadomość 2: Wynik "wirtualnego" zapytania initialize
	initResultData := `{"jsonrpc":"2.0","id":0,"result":{"protocolVersion":"2025-06-18","capabilities":{},"serverInfo":{"name":"reveal-secret-mock","version":"1.0.0"}}}`
	fmt.Fprintf(w, "event: message\ndata: %s\n\n", initResultData)
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		// Oczekiwanie na wiadomości do wysłania dla tej sesji
		case msg := <-messageChan:
			if _, err := fmt.Fprintf(w, "%s", msg); err != nil {
				log.Printf("Błąd zapisu do strumienia SSE dla sesji %s: %v", sessionID, err)
				return
			}
			flusher.Flush()
		// Wykrycie rozłączenia klienta
		case <-ctx.Done():
			log.Printf("Klient dla sesji %s rozłączył się.", sessionID)
			return
		}
	}
}

// messagesHandler obsługuje zapytania JSON-RPC na dedykowanym endpoincie sesji.
func messagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Brak session_id w zapytaniu", http.StatusBadRequest)
		return
	}

	// Krok 4: Odnalezienie kanału SSE dla danej sesji
	messageChan, found := store.get(sessionID)
	if !found {
		http.Error(w, "Sesja nie znaleziona", http.StatusNotFound)
		return
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Nieprawidłowy JSON-RPC", http.StatusBadRequest)
		return
	}

	log.Printf("📬 Otrzymano metodę '%s' [ID: %s] dla sesji %s", req.Method, string(req.ID), sessionID)

	// Krok 5: Natychmiastowa odpowiedź 202 Accepted
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Accepted"))

	// Krok 6: Asynchroniczne przygotowanie i wysłanie odpowiedzi przez SSE
	go processAndRespond(req, messageChan, sessionID)
}

func processAndRespond(req JSONRPCRequest, messageChan chan string, sessionID string) {
	var responseData string

	switch req.Method {
	case "initialize", "notifications/initialized":
		// Zgodnie z logami, te metody nie generują dodatkowej odpowiedzi przez SSE.
		return
	case "tools/list":
		// Odpowiedź zawiera listę dostępnych narzędzi. ID jest kopiowane z zapytania.
		responseData = fmt.Sprintf(
			`{"jsonrpc":"2.0","id":%s,"result":{"tools":[{"name":"reveal_secret","description":"Returns a simple greeting.","inputSchema":{"properties":{},"title":"reveal_secretArguments","type":"object"},"outputSchema":{"properties":{"result":{"title":"Result","type":"string"}},"required":["result"],"title":"reveal_secretOutput","type":"object"}}]}}`,
			string(req.ID),
		)
	case "tools/call":
		// Odpowiedź na wywołanie narzędzia "reveal_secret".
		responseData = fmt.Sprintf(
			`{"jsonrpc":"2.0","id":%s,"result":{"content":[{"type":"text","text":"no siemanko"}],"structuredContent":{"result":"no siemanko"},"isError":false}}`,
			string(req.ID),
		)
	default:
		log.Printf("Nieobsługiwana metoda: %s", req.Method)
		return
	}

	sseMessage := fmt.Sprintf("event: message\ndata: %s\n\n", responseData)

	// Wysłanie wiadomości do kanału SSE z timeoutem
	select {
	case messageChan <- sseMessage:
		log.Printf("📨 Wysłano odpowiedź na '%s' [ID: %s] do sesji %s", req.Method, string(req.ID), sessionID)
	case <-time.After(2 * time.Second):
		log.Printf("Timeout podczas wysyłania odpowiedzi do sesji %s", sessionID)
	}
}

// pingHandler obsługuje proste sprawdzenie działania serwera.
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, ": pong -", time.Now().Format(time.RFC3339))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/openapi.json", handleOpenAPI)
	mux.HandleFunc("/get_current_time", handleGetCurrentTime)
	mux.HandleFunc("/sse", sseHandler)
	mux.HandleFunc("/messages/", messagesHandler)

	handlerWithMiddleware := loggingMiddleware(corsMiddleware(serverHeaderMiddleware(mux)))
	port := "8000"
	log.Printf("Starting Go mock server on http://localhost:%s\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, handlerWithMiddleware); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}
