package types

import "errors"

var ErrMissingConfig = errors.New("missing required configuration")

type Config struct {
	AnthropicAPIKey     string `json:"anthropic_api_key"`
	AirtableAccessToken string `json:"airtable_access_token"`
	AirtableBaseID      string `json:"airtable_base_id"`
	AirtableTableName   string `json:"airtable_table_name"`
	DefaultLanguage     string `json:"default_language"`
}

type TableSchema struct {
	Fields []AirtableField `json:"fields"`
}

type AirtableField struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Contact struct {
	ID              string `json:"id"`
	Fullname        string `json:"fullname"`
	CompanyName     string `json:"company_name"`
	BusinessSegment string `json:"business_segment"`
	Website         string `json:"website"`
	Phone           string `json:"phone"`
	City            string `json:"city"`
	Country         string `json:"country"`
	Email           string `json:"email"`
	OutreachText    string `json:"outreach_text"`
	Error           string `json:"error,omitempty"`
}

type Language struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Selected bool   `json:"selected"`
}
