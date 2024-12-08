package handlers

import (
	"log"
	"outreach-generator/internal/types"
)

func (h *Handlers) loadConfig() (types.Config, error) {
	log.Printf("Loading config")
	var config types.Config

	// Load basic config
	rows, err := h.db.Query("SELECT key, value FROM config")
	if err != nil {
		return config, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return config, err
		}

		switch key {
		case "anthropic_api_key":
			config.AnthropicAPIKey = value
		case "airtable_access_token":
			config.AirtableAccessToken = value
		case "airtable_base_id":
			config.AirtableBaseID = value
		case "airtable_table_name":
			config.AirtableTableName = value
		case "default_language":
			config.DefaultLanguage = value
		}
	}

	log.Printf("Loaded config: %+v", config)
	return config, nil
}

func (h *Handlers) saveConfig(config types.Config) error {
	log.Printf("Saving config: %+v", config)
	tx, err := h.db.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// Save basic config
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	configItems := map[string]string{
		"anthropic_api_key":     config.AnthropicAPIKey,
		"airtable_access_token": config.AirtableAccessToken,
		"airtable_base_id":      config.AirtableBaseID,
		"airtable_table_name":   config.AirtableTableName,
		"default_language":      config.DefaultLanguage,
	}

	for key, value := range configItems {
		if _, err := stmt.Exec(key, value); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}
	log.Printf("Config saved successfully")
	return nil
}

func (h *Handlers) deleteField(id int64) error {
	_, err := h.db.Exec("DELETE FROM airtable_fields WHERE id = ?", id)
	return err
}

func (h *Handlers) getRequiredConfig() (types.Config, error) {
	config, err := h.loadConfig()
	if err != nil {
		return config, err
	}

	if config.AnthropicAPIKey == "" ||
		config.AirtableAccessToken == "" ||
		config.AirtableBaseID == "" ||
		config.AirtableTableName == "" {
		return config, types.ErrMissingConfig
	}

	return config, nil
}
