package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"practice-2/internal/store"
)

func TestTasks_PostThenGetByID(t *testing.T) {
	st := store.New()
	h := &TasksHandler{Store: st}

	body := bytes.NewBufferString(`{"title":"Test task"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	w := httptest.NewRecorder()
	h.Post(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	var created struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID != 1 || created.Title != "Test task" || created.Done != false {
		t.Fatalf("unexpected created: %+v", created)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/tasks?id=1", nil)
	w2 := httptest.NewRecorder()
	h.Get(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w2.Code, w2.Body.String())
	}

	var got struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&got); err != nil {
		t.Fatalf("decode got: %v", err)
	}
	if got.ID != 1 || got.Title != "Test task" || got.Done != false {
		t.Fatalf("unexpected got: %+v", got)
	}
}

func TestTasks_GetList(t *testing.T) {
	st := store.New()
	h := &TasksHandler{Store: st}

	if _, err := st.Create("A"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Create("B"); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var tasks []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestTasks_PatchDone(t *testing.T) {
	st := store.New()
	h := &TasksHandler{Store: st}

	created, err := st.Create("Patch me")
	if err != nil {
		t.Fatal(err)
	}

	body := bytes.NewBufferString(`{"done":true}`)
	req := httptest.NewRequest(http.MethodPatch, "/tasks?id=1", body)

	req.URL.RawQuery = "id=1"

	w := httptest.NewRecorder()
	h.Patch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	tk, err := st.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if tk.Done != true {
		t.Fatalf("expected Done=true, got %v", tk.Done)
	}
}
