package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is a Confluence Cloud REST API client using Basic Auth.
type Client struct {
	BaseURL    string // e.g. https://myorg.atlassian.net
	Username   string // Atlassian account email
	APIToken   string // Atlassian API token
	httpClient *http.Client
}

// NewClient creates a new Confluence API client from configuration.
func NewClient(baseURL, username, apiToken string) *Client {
	return &Client{
		BaseURL:  baseURL,
		Username: username,
		APIToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest executes an authenticated HTTP request against the Confluence API.
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	fullURL := c.BaseURL + path
	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.Username, c.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, fmt.Errorf("confluence API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}

// ─────────────────────────────────────────────
// TYPES
// ─────────────────────────────────────────────

// Page represents a Confluence page (v2 API).
type Page struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	SpaceID string `json:"spaceId"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
	Body struct {
		Storage struct {
			Value string `json:"value"`
		} `json:"storage"`
	} `json:"body"`
	Links struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// Space represents a Confluence space.
type Space struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Name       string `json:"name"`
	HomepageID string `json:"homepageId"`
	Links      struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// SearchResult represents one item in a CQL search response.
type SearchResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
	Space struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"space"`
	Excerpt string `json:"excerpt"`
	Links   struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// Comment represents a Confluence inline/footer comment.
type Comment struct {
	ID    string `json:"id"`
	Links struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// ─────────────────────────────────────────────
// READ OPERATIONS
// ─────────────────────────────────────────────

// GetPage fetches a Confluence page by ID, including its storage body.
func (c *Client) GetPage(pageID string) (*Page, error) {
	path := fmt.Sprintf("/wiki/api/v2/pages/%s?body-format=storage", pageID)
	data, _, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var page Page
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}
	return &page, nil
}

// GetPageChildren lists the direct children of a given page.
func (c *Client) GetPageChildren(pageID string) ([]Page, error) {
	path := fmt.Sprintf("/wiki/api/v2/pages/%s/children?limit=50", pageID)
	data, _, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Results []Page `json:"results"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse children response: %w", err)
	}
	return result.Results, nil
}

// GetSpaces lists all available Confluence spaces.
func (c *Client) GetSpaces(limit int) ([]Space, error) {
	if limit <= 0 {
		limit = 25
	}
	path := fmt.Sprintf("/wiki/api/v2/spaces?limit=%d", limit)
	data, _, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Results []Space `json:"results"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse spaces response: %w", err)
	}
	return result.Results, nil
}

// SearchPages runs a CQL search and returns matching pages/content.
func (c *Client) SearchPages(cql string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	path := fmt.Sprintf("/wiki/rest/api/content/search?cql=%s&limit=%d&expand=space,excerpt",
		url.QueryEscape(cql), limit)
	data, _, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Results []SearchResult `json:"results"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}
	return result.Results, nil
}

// ─────────────────────────────────────────────
// WRITE OPERATIONS
// ─────────────────────────────────────────────

// CreatePageRequest is the payload for creating a new Confluence page.
type CreatePageRequest struct {
	SpaceID  string `json:"spaceId"`
	ParentID string `json:"parentId,omitempty"`
	Title    string `json:"title"`
	Body     struct {
		Representation string `json:"representation"`
		Value          string `json:"value"`
	} `json:"body"`
	Status string `json:"status"`
}

// CreatePage creates a new Confluence page under the given space (and optional parent).
// bodyHTML should be valid Confluence storage-format HTML.
func (c *Client) CreatePage(spaceID, parentID, title, bodyHTML string) (*Page, error) {
	req := CreatePageRequest{
		SpaceID:  spaceID,
		ParentID: parentID,
		Title:    title,
		Status:   "current",
	}
	req.Body.Representation = "storage"
	req.Body.Value = bodyHTML

	data, _, err := c.doRequest(http.MethodPost, "/wiki/api/v2/pages", req)
	if err != nil {
		return nil, err
	}
	var page Page
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("failed to parse create page response: %w", err)
	}
	return &page, nil
}

// UpdatePageRequest is the payload for updating an existing page (v2 API).
type UpdatePageRequest struct {
	Title   string `json:"title"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
	Body struct {
		Representation string `json:"representation"`
		Value          string `json:"value"`
	} `json:"body"`
	Status string `json:"status"`
}

// UpdatePage updates the title and body of an existing Confluence page.
// It automatically fetches the current version and increments it.
func (c *Client) UpdatePage(pageID, title, bodyHTML string) (*Page, error) {
	// Fetch current version first (required by Confluence API)
	current, err := c.GetPage(pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current page version: %w", err)
	}

	req := UpdatePageRequest{
		Title:  title,
		Status: "current",
	}
	req.Version.Number = current.Version.Number + 1
	req.Body.Representation = "storage"
	req.Body.Value = bodyHTML

	path := fmt.Sprintf("/wiki/api/v2/pages/%s", pageID)
	data, _, err := c.doRequest(http.MethodPut, path, req)
	if err != nil {
		return nil, err
	}
	var page Page
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("failed to parse update page response: %w", err)
	}
	return &page, nil
}

// AddComment adds a footer comment to a Confluence page.
func (c *Client) AddComment(pageID, commentText string) (*Comment, error) {
	body := map[string]interface{}{
		"type": "comment",
		"container": map[string]interface{}{
			"id":   pageID,
			"type": "page",
		},
		"body": map[string]interface{}{
			"storage": map[string]interface{}{
				"value":          "<p>" + commentText + "</p>",
				"representation": "storage",
			},
		},
	}

	path := fmt.Sprintf("/wiki/rest/api/content/%s/child/comment", pageID)
	data, _, err := c.doRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse comment response: %w", err)
	}
	return &comment, nil
}

// ─────────────────────────────────────────────
// MARKDOWN → CONFLUENCE STORAGE HELPER
// ─────────────────────────────────────────────

// MarkdownToStorage does a best-effort conversion of simple Markdown
// to Confluence storage format HTML. For production use, consider
// the Confluence /wiki/rest/api/contentbody/convert/storage endpoint.
func MarkdownToStorage(md string) string {
	// Wrap raw Markdown in a Confluence macro that renders it.
	// This uses the "noformat" macro as a safe fallback that preserves content.
	// A full converter can be added later via the /convert endpoint.
	return fmt.Sprintf(`<ac:structured-macro ac:name="markdown" ac:schema-version="1">
  <ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body>
</ac:structured-macro>`, md)
}
