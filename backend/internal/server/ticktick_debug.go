package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"hoel-app/backend/internal/integration"
)

type tickTickProjectDebug struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type tickTickProjectDebugResponse struct {
	Projects []tickTickProjectDebug `json:"projects"`
}

type tickTickProjectListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Server) tickTickProjectsHandler(w http.ResponseWriter, r *http.Request) {
	if s.integrationClient == nil {
		http.Error(w, "ticktick integration is unavailable", http.StatusServiceUnavailable)
		return
	}

	apiRoot := strings.TrimSpace(s.tickTickAPIRoot)
	if apiRoot == "" {
		http.Error(w, "TICKTICK_API_ROOT is not configured", http.StatusBadRequest)
		return
	}

	token := strings.TrimSpace(s.tickTickToken)
	if token == "" {
		http.Error(w, "TICKTICK_ACCESS_TOKEN is not configured", http.StatusBadRequest)
		return
	}

	requestURL := fmt.Sprintf("%s/project", strings.TrimRight(apiRoot, "/"))
	response, err := s.integrationClient.Do(r.Context(), requestFromTickTick(requestURL, token))
	if err != nil {
		http.Error(w, "unable to fetch ticktick projects", http.StatusBadGateway)
		return
	}

	var rawProjects []tickTickProjectListItem
	if err := json.Unmarshal(response.Body, &rawProjects); err != nil {
		http.Error(w, "unable to parse ticktick projects response", http.StatusBadGateway)
		return
	}

	projects := make([]tickTickProjectDebug, 0, len(rawProjects))
	for _, project := range rawProjects {
		projectID := strings.TrimSpace(project.ID)
		if projectID == "" {
			continue
		}

		projects = append(projects, tickTickProjectDebug{
			ID:   projectID,
			Name: strings.TrimSpace(project.Name),
		})
	}

	writeJSON(w, http.StatusOK, tickTickProjectDebugResponse{Projects: projects})
}

func requestFromTickTick(url, token string) integration.Request {
	return integration.Request{
		Service: "ticktick",
		Method:  http.MethodGet,
		URL:     url,
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": "Bearer " + token,
		},
	}
}
