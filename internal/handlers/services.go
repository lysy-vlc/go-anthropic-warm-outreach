package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"outreach-generator/internal/types"

	"github.com/gocolly/colly/v2"
)

func (h *Handlers) fetchAirtableContacts(config types.Config) ([]types.Contact, error) {
	baseURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s?view=Grid%%20view",
		config.AirtableBaseID,
		url.PathEscape(config.AirtableTableName))

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.AirtableAccessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("airtable API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Records []struct {
			ID          string `json:"id"`
			CreatedTime string `json:"createdTime"`
			Fields      struct {
				Email           string `json:"email"`
				Fullname        string `json:"fullname"`
				Country         string `json:"country"`
				BusinessSegment string `json:"business segment"`
				City            string `json:"city"`
				CompanyName     string `json:"company name"`
				Phone           string `json:"phone"`
				Website         string `json:"website"`
				OutreachText    string `json:"outreach_text,omitempty"`
			} `json:"fields"`
		} `json:"records"`
		Offset string `json:"offset,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	contacts := make([]types.Contact, len(result.Records))
	for i, record := range result.Records {
		contacts[i] = types.Contact{
			ID:              record.ID,
			Fullname:        record.Fields.Fullname,
			CompanyName:     record.Fields.CompanyName,
			BusinessSegment: record.Fields.BusinessSegment,
			Website:         record.Fields.Website,
			Phone:           record.Fields.Phone,
			City:            record.Fields.City,
			Country:         record.Fields.Country,
			Email:           record.Fields.Email,
			OutreachText:    record.Fields.OutreachText,
		}
	}

	return contacts, nil
}

func (h *Handlers) updateAirtableOutreach(config types.Config, recordID, outreachText string) error {
	baseURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s",
		config.AirtableBaseID,
		url.PathEscape(config.AirtableTableName),
		recordID)

	payload := struct {
		Fields struct {
			OutreachText string `json:"outreach_text"`
		} `json:"fields"`
	}{}
	payload.Fields.OutreachText = outreachText

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+config.AirtableAccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("airtable API error: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (h *Handlers) generateOutreachText(config types.Config, req struct {
	Website     string `json:"website"`
	Prompt      string `json:"prompt"`
	RecordID    string `json:"recordId"`
	Language    string `json:"language"`
	ContactInfo struct {
		Name    string `json:"name"`
		Company string `json:"company"`
		Segment string `json:"segment"`
	} `json:"contactInfo"`
}, websiteContent string) (string, error) {
	log.Printf("Generating outreach with data:")
	log.Printf("- Website: %s", req.Website)
	log.Printf("- Prompt template: %s", req.Prompt)
	log.Printf("- Contact: %s from %s (%s)", req.ContactInfo.Name, req.ContactInfo.Company, req.ContactInfo.Segment)

	// Build the system prompt
	systemPrompt := h.buildSystemPrompt(req, websiteContent)
	log.Printf("Sending prompt to Anthropic:\n%s", systemPrompt)

	// Prepare the request to Anthropic's API
	anthropicURL := "https://api.anthropic.com/v1/messages"
	requestBody := map[string]interface{}{
		"model":      "claude-3-sonnet-20240229",
		"max_tokens": 1000,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": systemPrompt,
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req2, err := http.NewRequest("POST", anthropicURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req2.Header.Set("x-api-key", config.AnthropicAPIKey)
	req2.Header.Set("anthropic-version", "2023-06-01")
	req2.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req2)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// Clean up the response
	outreachText := strings.TrimSpace(result.Content[0].Text)
	return outreachText, nil
}

func (h *Handlers) fetchAirtableSchema(config types.Config) (*types.TableSchema, error) {
	baseURL := fmt.Sprintf("https://api.airtable.com/v0/meta/bases/%s/tables",
		config.AirtableBaseID)

	log.Printf("Fetching Airtable schema from: %s", baseURL)

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.AirtableAccessToken))
	log.Printf("Using Authorization header: Bearer %s...", config.AirtableAccessToken[:10])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Response status: %s", resp.Status)
	log.Printf("Response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("airtable API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Tables []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Fields []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Type        string `json:"type"`
				Description string `json:"description"`
			} `json:"fields"`
		} `json:"tables"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Find our table
	var tableSchema *types.TableSchema
	for _, table := range result.Tables {
		if table.Name == config.AirtableTableName {
			tableSchema = &types.TableSchema{
				Fields: make([]types.AirtableField, len(table.Fields)),
			}
			for i, field := range table.Fields {
				tableSchema.Fields[i] = types.AirtableField{
					ID:          field.ID,
					Name:        field.Name,
					Type:        field.Type,
					Description: field.Description,
				}
			}
			break
		}
	}

	if tableSchema == nil {
		return nil, fmt.Errorf("table %s not found", config.AirtableTableName)
	}

	return tableSchema, nil
}

func checkWebsite(url string) error {
	// Dodaj http:// jeśli nie ma protokołu
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("website unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("website returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (h *Handlers) fetchWebsiteContent(websiteURL string) (string, error) {
	// Add http:// if no protocol specified
	if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
		websiteURL = "http://" + websiteURL
	}

	log.Printf("Fetching content from: %s", websiteURL)

	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.Async(true),
	)

	// Ustaw timeout dla requestów
	c.SetRequestTimeout(10 * time.Second)

	var texts []string

	// Zbierz tekst z najważniejszych elementów
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Ignoruj elementy nawigacyjne, stopki itp.
		e.ForEach("main, article, .content, #content, .main-content", func(_ int, el *colly.HTMLElement) {
			text := el.Text
			texts = append(texts, text)
		})

		// Jeśli nie znaleziono głównej treści, zbierz tekst z paragrafów i nagłówków
		if len(texts) == 0 {
			e.ForEach("p, h1, h2, h3, h4, h5, h6, .about-us, .about, #about, #about-us", func(_ int, el *colly.HTMLElement) {
				text := el.Text
				if len(strings.TrimSpace(text)) > 50 { // Ignoruj krótkie teksty
					texts = append(texts, text)
				}
			})
		}
	})

	err := c.Visit(websiteURL)
	if err != nil {
		return "", fmt.Errorf("error visiting website: %w", err)
	}

	// Poczekaj na zakończenie wszystkich requestów
	c.Wait()

	// Połącz wszystkie znalezione teksty
	content := strings.Join(texts, "\n")
	content = cleanWebsiteContent(content)

	log.Printf("Fetched content length: %d bytes", len(content))
	if len(content) > 500 {
		log.Printf("First 500 chars: %s", content[:500])
	} else {
		log.Printf("Content: %s", content)
	}

	if content == "" {
		return "", fmt.Errorf("no content found on the website")
	}

	return content, nil
}

func (h *Handlers) buildSystemPrompt(req struct {
	Website     string `json:"website"`
	Prompt      string `json:"prompt"`
	RecordID    string `json:"recordId"`
	Language    string `json:"language"`
	ContactInfo struct {
		Name    string `json:"name"`
		Company string `json:"company"`
		Segment string `json:"segment"`
	} `json:"contactInfo"`
}, websiteContent string) string {
	return fmt.Sprintf(`You are a professional outreach specialist. Generate the outreach email in %s based on this website content about %s:

Website Content:
%s

Contact Information:
- Name: %s
- Company: %s
- Business Segment: %s

Additional Context:
%s

Important formatting rules:
1. Start with the subject line on the first line
2. Add a blank line after the subject
3. Then write the email body
4. Use the actual person's name and company from the contact info
5. Do not use placeholders like [Name] or [Company] - use the actual values
6. Do not include any explanatory text or metadata - just the email subject and body
7. Reference specific details from their website to show personalization
8. Keep the tone professional but friendly
9. Focus on how we can help them, not just what we do
10. Keep it concise - no more than 3-4 paragraphs`,
		req.Language,
		req.ContactInfo.Company,
		websiteContent,
		req.ContactInfo.Name,
		req.ContactInfo.Company,
		req.ContactInfo.Segment,
		req.Prompt)
}

// Helper function to clean up website content
func cleanWebsiteContent(content string) string {
	// Usuń nadmiarowe białe znaki
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	// Usuń znaki specjalne i emoji
	content = regexp.MustCompile(`[^\p{L}\p{N}\p{P}\s]`).ReplaceAllString(content, "")

	// Zamień wielokrotne nowe linie na pojedyncze
	content = regexp.MustCompile(`\n+`).ReplaceAllString(content, "\n")

	// Usuń wielokrotne spacje
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	// Przytnij białe znaki z początku i końca
	content = strings.TrimSpace(content)

	// Ogranicz długość tekstu
	if len(content) > 2000 {
		// Znajdź ostatnią kropkę przed limitem
		lastDot := strings.LastIndex(content[:2000], ".")
		if lastDot > 0 {
			content = content[:lastDot+1]
		} else {
			content = content[:2000] + "..."
		}
	}

	return content
}
