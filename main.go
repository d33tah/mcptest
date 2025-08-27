package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.json", handleOpenAPI)
	mux.HandleFunc("/get_current_time", handleGetCurrentTime)
	handlerWithMiddleware := loggingMiddleware(corsMiddleware(serverHeaderMiddleware(mux)))
	port := "8000"
	log.Printf("Starting Go mock server on http://localhost:%s\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, handlerWithMiddleware); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}
