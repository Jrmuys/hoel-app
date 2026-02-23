package server

import (
	"fmt"
	"net/http"
	"time"
)

func (s *Server) subscribeEvents() chan string {
	events := make(chan string, 8)

	s.eventSubscribersLock.Lock()
	s.eventSubscribers[events] = struct{}{}
	s.eventSubscribersLock.Unlock()

	return events
}

func (s *Server) unsubscribeEvents(events chan string) {
	s.eventSubscribersLock.Lock()
	if _, exists := s.eventSubscribers[events]; exists {
		delete(s.eventSubscribers, events)
		close(events)
	}
	s.eventSubscribersLock.Unlock()
}

func (s *Server) publishEvent(event string) {
	s.eventSubscribersLock.RLock()
	defer s.eventSubscribersLock.RUnlock()

	for subscriber := range s.eventSubscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (s *Server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	events := s.subscribeEvents()
	defer s.unsubscribeEvents(events)

	_ = writeSSEEvent(w, "ready", "{}")
	flusher.Flush()

	refreshTicker := time.NewTicker(20 * time.Second)
	heartbeatTicker := time.NewTicker(25 * time.Second)
	defer refreshTicker.Stop()
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case eventName, open := <-events:
			if !open {
				return
			}
			if err := writeSSEEvent(w, eventName, "{}"); err != nil {
				return
			}
			flusher.Flush()
		case <-refreshTicker.C:
			if err := writeSSEEvent(w, "refresh", "{}"); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeatTicker.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, eventName, data string) error {
	if _, err := fmt.Fprintf(w, "event: %s\n", eventName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return nil
}
