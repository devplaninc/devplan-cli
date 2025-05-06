package devplan

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/globals"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/company"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/user"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
)

type ClientConfig struct {
	BaseURL string
}

type Client struct {
	BaseURL string

	client *http.Client
}

func NewClient(config ClientConfig) *Client {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = GetBaseURL(globals.Domain)
	}
	return &Client{BaseURL: baseURL, client: &http.Client{}}
}

func (c *Client) GetCompanyProjects(companyID int32) (*company.GetProjectsWithDocsResponse, error) {
	result := &company.GetProjectsWithDocsResponse{}
	return result, c.getParsed(projectsPath(companyID), result)
}

func (c *Client) GetGroup(companyID int32, groupID string) (*company.GetGroupResponse, error) {
	result := &company.GetGroupResponse{}
	return result, c.getParsed(groupPath(companyID, groupID), result)
}

func (c *Client) GetSelf() (*user.GetSelfResponse, error) {
	result := &user.GetSelfResponse{}
	return result, c.getParsed(selfPath, result)
}

func (c *Client) GetRepoSummaries(companyID int32) (*company.GetRepoSummariesResponse, error) {
	result := &company.GetRepoSummariesResponse{}
	return result, c.getParsed(repoSummariesPath(companyID), result)
}

func (c *Client) getParsed(path string, msg proto.Message) error {
	body, err := c.get(path)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	if err := protojson.Unmarshal(body, msg); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) get(path string) ([]byte, error) {
	key, err := globals.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to init get request, no API key: %w", err)
	}
	if key == "" {
		return nil, fmt.Errorf("failed to init get request, no API key")
	}
	url := fmt.Sprintf("%s/%s", c.BaseURL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", path, err)
	}

	req.Header.Add("Authorization", "Bearer "+key)

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %s: %s [%v]", url, resp.Status, resp.StatusCode)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response %s: %w", url, err)
	}
	return body, nil
}
