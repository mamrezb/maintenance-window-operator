// File: internal/api/httpserver.go

package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	maintenancev1alpha1 "github.com/mamrezb/maintenance-window-manager/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Server encapsulates dependencies for the HTTP API.
type Server struct {
	Client client.Client
}

// ServiceStatusResponse is the JSON response structure for each service.
type ServiceStatusResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     bool   `json:"ready"`
	Critical  bool   `json:"critical"`
}

// handleServices is the HTTP handler that lists the statuses.
func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// List all ServiceChecker CRs.
	var scList maintenancev1alpha1.ServiceCheckerList
	if err := s.Client.List(ctx, &scList); err != nil {
		http.Error(w, "unable to list service checkers", http.StatusInternalServerError)
		return
	}

	var responses []ServiceStatusResponse
	// For every ServiceChecker, extract each service's status.
	for _, sc := range scList.Items {
		// For each status entry, optionally merge in the "critical" flag from the spec.
		for _, svcStatus := range sc.Status.ServiceStatuses {
			critical := false
			// Loop through the spec services to see if this service is marked as critical.
			for _, specSvc := range sc.Spec.Services {
				if specSvc.Name == svcStatus.Name && specSvc.Namespace == svcStatus.Namespace {
					critical = specSvc.Critical
					break
				}
			}
			responses = append(responses, ServiceStatusResponse{
				Name:      svcStatus.Name,
				Namespace: svcStatus.Namespace,
				Ready:     svcStatus.Ready,
				Critical:  critical,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// StartHTTPServer starts the API server on the given address.
// It sets up the routes and begins listening.
func StartHTTPServer(addr string, cli client.Client) error {
	server := &Server{Client: cli}
	mux := http.NewServeMux()
	// Register the /services endpoint.
	mux.HandleFunc("/services", server.handleServices)

	log.Printf("Starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, mux)
}
