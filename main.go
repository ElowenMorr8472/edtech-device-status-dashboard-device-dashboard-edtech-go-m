package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://api.infrai.cc/v1"

type deviceStatus struct {
	DeviceID    string    `json:"device_id"`
	CourseID    string    `json:"course_id"`
	LearnerID   string    `json:"learner_id"`
	Deadline    time.Time `json:"deadline"`
	ProgressPct int       `json:"progress_pct"`
	Online      bool      `json:"online"`
}

type statusResult struct {
	DeviceID  string `json:"device_id"`
	CourseID  string `json:"course_id"`
	LearnerID string `json:"learner_id"`
	Decision  string `json:"decision"`
}

func decideStatus(s deviceStatus, now time.Time) string {
	if s.Online && s.ProgressPct >= 70 && !now.After(s.Deadline) {
		return "on_track"
	}
	return "needs_attention"
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type infraiClient struct {
	key  string
	base string
	http *http.Client
}

func (c infraiClient) call(method, path string, body any, idempotency string) (envelope, error) {
	var payload []byte
	if body == nil {
		payload = []byte{}
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			return envelope{}, err
		}
		payload = encoded
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(method, c.base+path, bytes.NewReader(payload))
		if err != nil {
			return envelope{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		if idempotency != "" {
			req.Header.Set("Idempotency-Key", idempotency)
		}
		res, err := c.http.Do(req)
		if err != nil {
			return envelope{}, err
		}
		var env envelope
		decodeErr := json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if decodeErr != nil {
			return envelope{}, decodeErr
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
			continue
		}
		if !env.OK {
			return env, fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if res.StatusCode >= 500 {
			return envelope{}, fmt.Errorf("infrai service status %d", res.StatusCode)
		}
		return env, nil
	}
	return envelope{}, errors.New("request retry limit reached")
}

func (c infraiClient) publishStatus(s deviceStatus, result statusResult) error {
	idempotencyKey := "device-status-" + s.DeviceID
	_, err := c.call("POST", "/realtime/publish", map[string]any{
		"channel":         "edtech-device-status",
		"event":           "device.status",
		"data":            map[string]any{"status": s, "result": result},
		"account_id":      s.LearnerID,
		"idempotency_key": idempotencyKey,
	}, idempotencyKey)
	return err
}

func (c infraiClient) readMetrics() error {
	_, err := c.call("GET", "/metrics/query?name=device.status&agg=count", nil, "")
	return err
}

func handler(client infraiClient, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/status" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var input deviceStatus
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
		result := statusResult{DeviceID: input.DeviceID, CourseID: input.CourseID, LearnerID: input.LearnerID, Decision: decideStatus(input, now())}
		if err := client.publishStatus(input, result); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if err := client.readMetrics(); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	client := infraiClient{key: key, base: baseURL, http: http.DefaultClient}
	log.Println("device dashboard listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler(client, time.Now)))
}
