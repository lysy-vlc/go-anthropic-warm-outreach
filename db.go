package main

func (s *Server) saveConfig(config Config) error {
	tx, err := s.db.Begin()
	if err != nil {
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

	// Clear existing fields
	if _, err := tx.Exec("DELETE FROM airtable_fields"); err != nil {
		return err
	}

	// Insert new fields
	for _, field := range config.AirtableFields {
		if _, err := tx.Exec(
			"INSERT INTO airtable_fields (name, airtable_name, type, required) VALUES (?, ?, ?, ?)",
			field.Name,
			field.AirtableName,
			field.Type,
			field.Required,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Server) loadConfig() (Config, error) {
	var config Config

	// Load basic config
	rows, err := s.db.Query("SELECT key, value FROM config")
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

	// Load fields
	fields, err := s.db.Query("SELECT id, name, airtable_name, type, required FROM airtable_fields")
	if err != nil {
		return config, err
	}
	defer fields.Close()

	for fields.Next() {
		var field AirtableField
		if err := fields.Scan(&field.ID, &field.Name, &field.AirtableName, &field.Type, &field.Required); err != nil {
			return config, err
		}
		config.AirtableFields = append(config.AirtableFields, field)
	}

	return config, nil
}

func (s *Server) getRequiredConfig() (Config, error) {
	config, err := s.loadConfig()
	if err != nil {
		return config, err
	}

	if config.AnthropicAPIKey == "" ||
		config.AirtableAccessToken == "" ||
		config.AirtableBaseID == "" ||
		config.AirtableTableName == "" {
		return config, ErrMissingConfig
	}

	return config, nil
}
