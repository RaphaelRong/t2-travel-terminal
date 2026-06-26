package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const ticketmasterDiscoveryURL = "https://app.ticketmaster.com/discovery/v2/events.json"

type TicketmasterClient struct {
	apiKey string
	client *http.Client
}

type TicketmasterSearchRequest struct {
	City      string `json:"city" binding:"required"`
	Country   string `json:"country"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Days      int    `json:"days"`
	Keyword   string `json:"keyword"`
	Category  string `json:"category"`
	Limit     int    `json:"limit"`
}

type CityEvents struct {
	Metadata struct {
		City      string `json:"city"`
		Country   string `json:"country"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		FetchedAt string `json:"fetched_at"`
		Total     int    `json:"total"`
		Source    string `json:"source"`
	} `json:"metadata"`
	Results []Event `json:"results"`
}

type Event struct {
	ID                 string    `json:"id"`
	ObjectType         string    `json:"object_type"`
	SchemaVersion      string    `json:"schema_version"`
	ExternalID         string    `json:"external_id"`
	Title              string    `json:"title"`
	Description        string    `json:"description,omitempty"`
	Category           string    `json:"category"`
	Rank               int       `json:"rank"`
	ExpectedAttendance *int      `json:"expected_attendance,omitempty"`
	Start              string    `json:"start"`
	End                string    `json:"end"`
	Timezone           string    `json:"timezone,omitempty"`
	Coordinates        []float64 `json:"coordinates,omitempty"`
	City               string    `json:"city,omitempty"`
	CountryCode        string    `json:"country_code,omitempty"`
	Venue              string    `json:"venue,omitempty"`
	DataSource         string    `json:"data_source"`
	SourceName         string    `json:"source_name,omitempty"`
	SourceURL          string    `json:"source_url,omitempty"`
}

func NewTicketmasterClient(apiKey string) *TicketmasterClient {
	return &TicketmasterClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *TicketmasterClient) Configured() bool {
	return strings.TrimSpace(c.apiKey) != ""
}

func (c *TicketmasterClient) DefaultAPIKey() string {
	return strings.TrimSpace(c.apiKey)
}

func (c *TicketmasterClient) SearchEvents(ctx context.Context, req TicketmasterSearchRequest) (CityEvents, error) {
	return c.SearchEventsWithAPIKey(ctx, req, c.apiKey)
}

func (c *TicketmasterClient) SearchEventsWithAPIKey(ctx context.Context, req TicketmasterSearchRequest, apiKey string) (CityEvents, error) {
	result := CityEvents{}
	req.City = strings.TrimSpace(req.City)
	req.Country = strings.ToUpper(strings.TrimSpace(req.Country))
	if req.City == "" {
		return result, fmt.Errorf("city is required")
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return result, fmt.Errorf("Ticketmaster provider is not configured")
	}

	startDate, endDate, err := resolveDateRange(req.StartDate, req.EndDate, req.Days)
	if err != nil {
		return result, err
	}

	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	result.Metadata.City = req.City
	result.Metadata.Country = req.Country
	result.Metadata.StartDate = startDate
	result.Metadata.EndDate = endDate
	result.Metadata.Source = "ticketmaster"
	result.Metadata.FetchedAt = time.Now().UTC().Format(time.RFC3339)

	page := 0
	for {
		if page*limit >= 1000 {
			break
		}

		params := url.Values{}
		params.Set("apikey", apiKey)
		params.Set("city", req.City)
		if req.Country != "" {
			params.Set("countryCode", req.Country)
		}
		if req.Keyword != "" {
			params.Set("keyword", req.Keyword)
		}
		if req.Category != "" {
			params.Set("classificationName", req.Category)
		}
		params.Set("startDateTime", startDate+"T00:00:00Z")
		params.Set("endDateTime", endDate+"T23:59:59Z")
		params.Set("size", fmt.Sprintf("%d", limit))
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("sort", "date,asc")

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, ticketmasterDiscoveryURL+"?"+params.Encode(), nil)
		if err != nil {
			return result, err
		}
		httpReq.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(httpReq)
		if err != nil {
			return result, err
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		if readErr != nil {
			return result, readErr
		}
		if resp.StatusCode != http.StatusOK {
			if isPagingDepthError(body) {
				break
			}
			return result, fmt.Errorf("Ticketmaster API returned HTTP %d: %s", resp.StatusCode, string(body))
		}

		var decoded ticketmasterResponse
		if err := json.Unmarshal(body, &decoded); err != nil {
			return result, err
		}
		if decoded.Embedded == nil || len(decoded.Embedded.Events) == 0 {
			break
		}

		for _, upstreamEvent := range decoded.Embedded.Events {
			event := convertTicketmasterEvent(upstreamEvent, req.City, req.Country)
			if event != nil {
				result.Results = append(result.Results, *event)
			}
		}

		if page+1 >= decoded.Page.TotalPages {
			break
		}
		page++
	}

	result.Metadata.Total = len(result.Results)
	return result, nil
}

func resolveDateRange(start string, end string, days int) (string, string, error) {
	if start != "" && end != "" {
		if _, err := time.Parse("2006-01-02", start); err != nil {
			return "", "", fmt.Errorf("invalid start_date: %w", err)
		}
		if _, err := time.Parse("2006-01-02", end); err != nil {
			return "", "", fmt.Errorf("invalid end_date: %w", err)
		}
		return start, end, nil
	}
	if days <= 0 {
		days = 90
	}
	now := time.Now().UTC()
	return now.Format("2006-01-02"), now.AddDate(0, 0, days).Format("2006-01-02"), nil
}

func isPagingDepthError(body []byte) bool {
	var errResp struct {
		Errors []struct {
			Code string `json:"code"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false
	}
	for _, item := range errResp.Errors {
		if item.Code == "DIS1035" {
			return true
		}
	}
	return false
}

func convertTicketmasterEvent(ev ticketmasterEvent, city string, country string) *Event {
	if ev.Name == "" {
		return nil
	}

	start := ""
	if ev.Dates.Start.DateTime != "" {
		start = ev.Dates.Start.DateTime
	} else if ev.Dates.Start.LocalDate != "" {
		start = ev.Dates.Start.LocalDate + "T00:00:00Z"
	}
	if start == "" {
		return nil
	}

	end := start
	if ev.Dates.End != nil {
		if ev.Dates.End.DateTime != "" {
			end = ev.Dates.End.DateTime
		} else if ev.Dates.End.LocalDate != "" {
			end = ev.Dates.End.LocalDate + "T23:59:59Z"
		}
	}

	event := &Event{
		ID:            "ticketmaster:" + ev.ID,
		ObjectType:    "event",
		SchemaVersion: "1.0.0",
		ExternalID:    ev.ID,
		Title:         ev.Name,
		Description:   ev.Info,
		Start:         start,
		End:           end,
		Timezone:      ev.Dates.Timezone,
		City:          city,
		CountryCode:   country,
		DataSource:    "ticketmaster",
		SourceName:    "Ticketmaster",
		SourceURL:     ev.URL,
		Category:      "concerts",
	}

	if ev.Embedded != nil && len(ev.Embedded.Venues) > 0 {
		venue := ev.Embedded.Venues[0]
		event.Venue = venue.Name
		if event.City == "" && venue.City.Name != "" {
			event.City = venue.City.Name
		}
		if event.CountryCode == "" && venue.Country.CountryCode != "" {
			event.CountryCode = venue.Country.CountryCode
		}
		if venue.Location.Longitude != "" && venue.Location.Latitude != "" {
			var lon, lat float64
			if _, err := fmt.Sscanf(venue.Location.Longitude, "%f", &lon); err == nil {
				if _, err := fmt.Sscanf(venue.Location.Latitude, "%f", &lat); err == nil && lon != 0 && lat != 0 {
					event.Coordinates = []float64{lon, lat}
				}
			}
		}
		if venue.Capacity > 0 {
			event.ExpectedAttendance = &venue.Capacity
		}
	}

	if len(ev.Classifications) > 0 {
		if segment := ev.Classifications[0].Segment.Name; segment != "" {
			event.Category = mapTicketmasterCategory(segment)
		}
	}
	event.Rank = estimateRank(ev, event.ExpectedAttendance)
	return event
}

func mapTicketmasterCategory(segment string) string {
	switch strings.ToLower(segment) {
	case "music":
		return "concerts"
	case "sports":
		return "sports"
	case "arts & theatre":
		return "performing-arts"
	case "film":
		return "screening"
	case "miscellaneous", "misc":
		return "community"
	default:
		return strings.ToLower(strings.ReplaceAll(segment, " ", "-"))
	}
}

func estimateRank(ev ticketmasterEvent, attendance *int) int {
	if ev.Popularity != "" {
		var popularity float64
		if _, err := fmt.Sscanf(ev.Popularity, "%f", &popularity); err == nil && popularity > 0 {
			return int(popularity * 100)
		}
	}
	if attendance == nil {
		return 30
	}
	switch {
	case *attendance >= 50000:
		return 80
	case *attendance >= 20000:
		return 65
	case *attendance >= 5000:
		return 50
	case *attendance >= 1000:
		return 40
	default:
		return 30
	}
}

type ticketmasterResponse struct {
	Embedded *struct {
		Events []ticketmasterEvent `json:"events"`
	} `json:"_embedded"`
	Page struct {
		TotalPages int `json:"totalPages"`
	} `json:"page"`
}

type ticketmasterEvent struct {
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	URL             string                  `json:"url"`
	Info            string                  `json:"info"`
	Popularity      string                  `json:"popularity"`
	Dates           ticketmasterDates       `json:"dates"`
	Classifications []ticketmasterClass     `json:"classifications"`
	Embedded        *ticketmasterEmbeddedV1 `json:"_embedded"`
}

type ticketmasterDates struct {
	Timezone string                 `json:"timezone"`
	Start    ticketmasterDatePoint  `json:"start"`
	End      *ticketmasterDatePoint `json:"end,omitempty"`
}

type ticketmasterDatePoint struct {
	LocalDate string `json:"localDate"`
	DateTime  string `json:"dateTime"`
}

type ticketmasterClass struct {
	Segment struct {
		Name string `json:"name"`
	} `json:"segment"`
}

type ticketmasterEmbeddedV1 struct {
	Venues []ticketmasterVenue `json:"venues"`
}

type ticketmasterVenue struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	City     struct {
		Name string `json:"name"`
	} `json:"city"`
	Country struct {
		CountryCode string `json:"countryCode"`
	} `json:"country"`
	Location struct {
		Longitude string `json:"longitude"`
		Latitude  string `json:"latitude"`
	} `json:"location"`
}
