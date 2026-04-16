package main

import (
	"fmt"
	"net/http"
	"sync"
)

// SSEHub manages Server-Sent Events connections grouped by court ID.
// It allows broadcasting events to all clients watching a specific court.
type SSEHub struct {
	mu      sync.RWMutex
	courts  map[string]map[chan sseEvent]struct{}
	globals map[chan sseEvent]struct{} // global subscribers (court list page)
}

type sseEvent struct {
	Event string
	Data  string
}

var hub = &SSEHub{
	courts:  make(map[string]map[chan sseEvent]struct{}),
	globals: make(map[chan sseEvent]struct{}),
}

func (h *SSEHub) subscribe(courtID string) chan sseEvent {
	ch := make(chan sseEvent, 16)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.courts[courtID] == nil {
		h.courts[courtID] = make(map[chan sseEvent]struct{})
	}
	h.courts[courtID][ch] = struct{}{}
	return ch
}

func (h *SSEHub) unsubscribe(courtID string, ch chan sseEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.courts[courtID]; ok {
		delete(clients, ch)
		if len(clients) == 0 {
			delete(h.courts, courtID)
		}
	}
	close(ch)
}

func (h *SSEHub) subscribeGlobal() chan sseEvent {
	ch := make(chan sseEvent, 16)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.globals[ch] = struct{}{}
	return ch
}

func (h *SSEHub) unsubscribeGlobal(ch chan sseEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.globals, ch)
	close(ch)
}

// broadcast sends an event to all clients subscribed to a court.
func (h *SSEHub) broadcast(courtID string, event sseEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.courts[courtID]; ok {
		for ch := range clients {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

// broadcastGlobal sends an event to all global subscribers (court list watchers).
func (h *SSEHub) broadcastGlobal(event sseEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.globals {
		select {
		case ch <- event:
		default:
		}
	}
}

// broadcastRoomUpdate notifies all clients watching a court that a room has changed.
func broadcastRoomUpdate(courtID, roomID string) {
	if courtID == "" {
		return
	}
	evt := sseEvent{
		Event: "room_updated",
		Data:  fmt.Sprintf(`{"roomId":"%s","courtId":"%s"}`, roomID, courtID),
	}
	hub.broadcast(courtID, evt)
	hub.broadcastGlobal(evt)
}

// broadcastRoomCreated notifies all clients that a new room was created.
func broadcastRoomCreated(courtID, roomID string) {
	if courtID == "" {
		return
	}
	evt := sseEvent{
		Event: "room_created",
		Data:  fmt.Sprintf(`{"roomId":"%s","courtId":"%s"}`, roomID, courtID),
	}
	hub.broadcast(courtID, evt)
	hub.broadcastGlobal(evt)
}

// broadcastScoresSubmitted notifies clients that scores were submitted (triggers profile refresh).
func broadcastScoresSubmitted(courtID, roomID string) {
	if courtID == "" {
		return
	}
	evt := sseEvent{
		Event: "scores_submitted",
		Data:  fmt.Sprintf(`{"roomId":"%s","courtId":"%s"}`, roomID, courtID),
	}
	hub.broadcast(courtID, evt)
	hub.broadcastGlobal(evt)
}

// broadcastCourtUpdated notifies global subscribers that court info changed.
func broadcastCourtUpdated(courtID string) {
	if courtID == "" {
		return
	}
	hub.broadcastGlobal(sseEvent{
		Event: "court_updated",
		Data:  fmt.Sprintf(`{"courtId":"%s"}`, courtID),
	})
}

// getMatchCourtID fetches the court_id for a given match.
func getMatchCourtID(matchID string) string {
	var courtID string
	sqlDB.QueryRow("SELECT COALESCE(court_id, '') FROM matches WHERE id = ?", matchID).Scan(&courtID)
	return courtID
}

// handleSSE serves an SSE stream for a court or global scope.
func handleSSE(w http.ResponseWriter, r *http.Request) {
	courtID := r.URL.Query().Get("courtId")
	scope := r.URL.Query().Get("scope")

	if courtID == "" && scope != "global" {
		writeError(w, 400, "courtId or scope=global required")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	var ch chan sseEvent
	if scope == "global" {
		ch = hub.subscribeGlobal()
		defer hub.unsubscribeGlobal(ch)
	} else {
		ch = hub.subscribe(courtID)
		defer hub.unsubscribe(courtID, ch)
	}

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case evt := <-ch:
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Event, evt.Data)
			flusher.Flush()
		}
	}
}
