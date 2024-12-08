package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"outreach-generator/internal/components"
	"outreach-generator/internal/types"
)

func (h *Handlers) HandleConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := h.loadConfig()
		if err != nil {
			http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
			return
		}

		var schema *types.TableSchema
		if config.AirtableAccessToken != "" && config.AirtableBaseID != "" && config.AirtableTableName != "" {
			schema, err = h.fetchAirtableSchema(config)
			if err != nil {
				log.Printf("Warning: Failed to fetch Airtable schema: %v", err)
			}
		}

		component := components.Config(config, schema)
		component.Render(r.Context(), w)
	}
}

func (h *Handlers) HandleGetConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := h.loadConfig()
		if err != nil {
			http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	}
}

func (h *Handlers) HandleSaveConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			respondWithError(w, http.StatusBadRequest, "Failed to parse form data")
			return
		}

		config := types.Config{
			AnthropicAPIKey:     r.FormValue("anthropic_api_key"),
			AirtableAccessToken: r.FormValue("airtable_access_token"),
			AirtableBaseID:      r.FormValue("airtable_base_id"),
			AirtableTableName:   r.FormValue("airtable_table_name"),
			DefaultLanguage:     r.FormValue("default_language"),
		}

		if err := h.saveConfig(config); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to save configuration")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Configuration saved successfully",
		})
	}
}
