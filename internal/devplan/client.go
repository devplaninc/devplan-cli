package devplan

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/devplaninc/devplan-cli/internal/version"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/company"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/user"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/integrations"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/worklog"
	"github.com/opensdd/osdd-api/clients/go/osdd/recipes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	BaseURL string
}

type Client struct {
	BaseURL string

	client *http.Client
}

func NewClient(config Config) *Client {
	baseURL := config.BaseURL
	if baseURL == "" {
		domain := prefs.Domain
		if domain == "" {
			domain = os.Getenv("DEVPLAN_API_DOMAIN")
		}
		baseURL = GetBaseURL(domain)
	}
	return &Client{BaseURL: baseURL, client: &http.Client{}}
}

func (c *Client) GetCompanyProjects(companyID int32) (*company.GetProjectsWithDocsResponse, error) {
	result := &company.GetProjectsWithDocsResponse{}
	return result, c.getParsed(projectsPath(companyID), result)
}

func (c *Client) GetProjectDocuments(companyID int32, projectID string) (*company.GetAllProjectDocsResponse, error) {
	result := &company.GetAllProjectDocsResponse{}
	return result, c.getParsed(projectDocsPath(companyID, projectID), result)
}

func (c *Client) GetDocument(companyID int32, documentID string) (*company.GetDocResponse, error) {
	result := &company.GetDocResponse{}
	return result, c.getParsed(documentPath(companyID, documentID), result)
}

func (c *Client) GetProjectTemplates(companyID int32) (*company.GetTemplatesResponse, error) {
	result := &company.GetTemplatesResponse{}
	return result, c.getParsed(templatesPath(companyID), result)
}

func (c *Client) GetGroup(companyID int32, groupID string) (*company.GetGroupResponse, error) {
	result := &company.GetGroupResponse{}
	return result, c.getParsed(groupPath(companyID, groupID), result)
}

func (c *Client) GetSelf() (*user.GetSelfResponse, error) {
	result := &user.GetSelfResponse{}
	return result, c.getParsed(selfPath, result)
}

func (c *Client) GetIntegration(companyID int32, provider string) (*company.GetIntegrationPropertiesResponse, error) {
	result := &company.GetIntegrationPropertiesResponse{}
	return result, c.getParsed(integrationPath(companyID, provider), result)
}

func (c *Client) GetAllRepos(companyID int32) ([]*integrations.GitRepository, error) {
	ghResult := &company.GetIntegrationPropertiesResponse{}
	err := c.getParsed(integrationPath(companyID, "github"), ghResult)
	if err != nil {
		return nil, err
	}
	bbResult := &company.GetIntegrationPropertiesResponse{}
	err = c.getParsed(integrationPath(companyID, "bitbucket"), bbResult)
	if err != nil {
		return nil, err
	}
	result := ghResult.GetInfo().GetGithub().GetRepositories()
	for _, integ := range bbResult.GetInfo().GetBitBucket().GetIntegrations() {
		result = append(result, integ.GetRepositories()...)
	}
	return result, nil
}

func (c *Client) GetDevRule(companyID int32, ruleName string) (*company.GetDevRuleResponse, error) {
	result := &company.GetDevRuleResponse{}
	return result, c.getParsed(devRulePath(companyID, ruleName), result)
}

func (c *Client) GetIDERecipe(companyID int32) (*recipes.Recipe, error) {
	result := &company.GetDevRecipeResponse{}
	if err := c.getParsed(devIDERecipePath(companyID), result); err != nil {
		return nil, err
	}
	return unmarshalRecipe(result.GetJsonRecipe())
}

func (c *Client) GetTaskRecipe(companyID int32, taskID string) (*recipes.Recipe, error) {
	result := &company.GetTaskRecipeResponse{}
	if err := c.getParsed(devTaskRecipePath(companyID, taskID), result); err != nil {
		return nil, err
	}
	return unmarshalRecipe(result.GetJsonRecipe())
}

func (c *Client) SubmitWorklogItem(companyID int32, item *worklog.WorkLogItem) (*company.SubmitWorkLogResponse, error) {
	result := &company.SubmitWorkLogResponse{}
	req := company.SubmitWorkLogRequest_builder{
		Item: item,
	}.Build()
	return result, c.postParsed(submitWorkLogPath(companyID), req, result)
}

func unmarshalRecipe(js string) (*recipes.Recipe, error) {
	recipe := &recipes.Recipe{}
	u := protojson.UnmarshalOptions{DiscardUnknown: true}
	return recipe, u.Unmarshal([]byte(js), recipe)
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
	u := protojson.UnmarshalOptions{DiscardUnknown: true}

	if err := u.Unmarshal(body, msg); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) get(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseURL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", path, err)
	}

	if err := c.setHeaders(req); err != nil {
		return nil, err
	}

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

func (c *Client) post(path string, req proto.Message) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseURL, path)

	m := protojson.MarshalOptions{}
	payload, err := m.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request for %s: %w", path, err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", path, err)
	}
	if err := c.setHeaders(httpReq); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to post %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to post %s: %s [%v]", url, resp.Status, resp.StatusCode)
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

func (c *Client) setHeaders(req *http.Request) error {
	key, err := VerifyAuth()
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+key)
	req.Header.Add("x-devplan-cli-version", version.GetVersion())
	if req.Method == "POST" {
		req.Header.Add("Content-Type", "application/json")
	}
	return nil
}

func (c *Client) postParsed(path string, req proto.Message, msg proto.Message) error {
	body, err := c.post(path, req)
	if err != nil {
		return fmt.Errorf("failed to post response: %w", err)
	}
	u := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := u.Unmarshal(body, msg); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}
