package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"outreach-generator/internal/components"
	"outreach-generator/internal/types"
)

func (h *Handlers) HandleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		languages := []types.Language{
			{Code: "en", Name: "English", Selected: true},
			{Code: "pl", Name: "Polish"},
			{Code: "de", Name: "German"},
			{Code: "es", Name: "Spanish"},
			{Code: "fr", Name: "French"},
		}

		component := components.Home(nil, languages)
		component.Render(r.Context(), w)
	}
}

func (h *Handlers) HandleGetCompanies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := h.getRequiredConfig()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		contacts, err := h.fetchAirtableContacts(config)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		component := components.ContactsList(contacts)
		component.Render(r.Context(), w)
	}
}

func (h *Handlers) HandleGenerateOutreach() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := h.getRequiredConfig()
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var req struct {
			Website     string `json:"website"`
			Prompt      string `json:"prompt"`
			RecordID    string `json:"recordId"`
			Language    string `json:"language"`
			ContactInfo struct {
				Name    string `json:"name"`
				Company string `json:"company"`
				Segment string `json:"segment"`
			} `json:"contactInfo"`
		}

		contentType := r.Header.Get("Content-Type")
		log.Printf("Content-Type: %s", contentType)

		if contentType == "application/x-www-form-urlencoded" {
			if err := r.ParseForm(); err != nil {
				log.Printf("Error parsing form: %v", err)
				respondWithError(w, http.StatusBadRequest, "Failed to parse form data")
				return
			}

			req.Website = r.FormValue("website")
			req.Prompt = r.FormValue("prompt")
			req.RecordID = r.FormValue("recordId")
			req.Language = r.FormValue("language")

			// Parse contactInfo from JSON string
			contactInfoStr := r.FormValue("contactInfo")
			if contactInfoStr != "" {
				if err := json.Unmarshal([]byte(contactInfoStr), &req.ContactInfo); err != nil {
					log.Printf("Error parsing contactInfo JSON: %v", err)
					respondWithError(w, http.StatusBadRequest, "Invalid contactInfo format")
					return
				}
			}
		} else {
			// Assume JSON
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				log.Printf("Error decoding JSON request: %v", err)
				respondWithError(w, http.StatusBadRequest, "Invalid request format")
				return
			}
		}

		log.Printf("Decoded request data:")
		log.Printf("- Website: %s", req.Website)
		log.Printf("- Prompt: %s", req.Prompt)
		log.Printf("- RecordID: %s", req.RecordID)
		log.Printf("- Language: %s", req.Language)
		log.Printf("- Contact Info:")
		log.Printf("  - Name: %s", req.ContactInfo.Name)
		log.Printf("  - Company: %s", req.ContactInfo.Company)
		log.Printf("  - Segment: %s", req.ContactInfo.Segment)

		// Fetch website content
		websiteContent, err := h.fetchWebsiteContent(req.Website)
		if err != nil {
			log.Printf("Error fetching website content: %v", err)
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to fetch website content: %v", err))
			return
		}

		// Sprawdź dostępność strony
		if err := checkWebsite(req.Website); err != nil {
			contact := types.Contact{
				ID:              req.RecordID,
				Fullname:        req.ContactInfo.Name,
				CompanyName:     req.ContactInfo.Company,
				BusinessSegment: req.ContactInfo.Segment,
				Website:         req.Website,
				Error:           fmt.Sprintf("Website error: %v", err),
			}
			component := components.ContactCard(contact)
			component.Render(r.Context(), w)
			return
		}

		// Generate outreach text
		outreachText, err := h.generateOutreachText(config, req, websiteContent)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if err := h.updateAirtableOutreach(config, req.RecordID, outreachText); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		contacts, err := h.fetchAirtableContacts(config)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch contact details")
			return
		}

		var contact types.Contact
		for _, c := range contacts {
			if c.ID == req.RecordID {
				contact = c
				contact.OutreachText = outreachText
				break
			}
		}

		component := components.ContactCard(contact)
		component.Render(r.Context(), w)
	}
}

func (h *Handlers) HandleGenerateAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := h.getRequiredConfig()
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		contacts, err := h.fetchAirtableContacts(config)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		for i := range contacts {
			// Sprawdź dostępność strony przed generowaniem
			if err := checkWebsite(contacts[i].Website); err != nil {
				contacts[i].Error = fmt.Sprintf("Website error: %v", err)
				continue
			}

			// Fetch website content
			websiteContent, err := h.fetchWebsiteContent(contacts[i].Website)
			if err != nil {
				contacts[i].Error = fmt.Sprintf("Website error: %v", err)
				continue
			}

			// Generuj outreach tylko dla kontaktów bez błędów
			if contacts[i].Error == "" {
				req := struct {
					Website     string `json:"website"`
					Prompt      string `json:"prompt"`
					RecordID    string `json:"recordId"`
					Language    string `json:"language"`
					ContactInfo struct {
						Name    string `json:"name"`
						Company string `json:"company"`
						Segment string `json:"segment"`
					} `json:"contactInfo"`
				}{
					Website:  contacts[i].Website,
					Prompt:   r.FormValue("prompt"),
					RecordID: contacts[i].ID,
					Language: config.DefaultLanguage,
					ContactInfo: struct {
						Name    string `json:"name"`
						Company string `json:"company"`
						Segment string `json:"segment"`
					}{
						Name:    contacts[i].Fullname,
						Company: contacts[i].CompanyName,
						Segment: contacts[i].BusinessSegment,
					},
				}

				outreachText, err := h.generateOutreachText(config, req, websiteContent)
				if err != nil {
					contacts[i].Error = fmt.Sprintf("Generation error: %v", err)
					continue
				}

				if err := h.updateAirtableOutreach(config, contacts[i].ID, outreachText); err != nil {
					contacts[i].Error = fmt.Sprintf("Update error: %v", err)
					continue
				}

				contacts[i].OutreachText = outreachText
			}
		}

		component := components.ContactsList(contacts)
		component.Render(r.Context(), w)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
