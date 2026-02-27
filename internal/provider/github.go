package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GitHubRepo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Language      string `json:"language"`
	UpdatedAt     string `json:"updated_at"`
}

type GitHubUser struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
}

type GitHubClient struct {
	httpClient *http.Client
	token      string
}

func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      token,
	}
}

func (g *GitHubClient) ValidateToken() (*GitHubUser, error) {
	resp, err := g.doRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid token: status %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

func (g *GitHubClient) ListRepos(page, perPage int) ([]GitHubRepo, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}

	url := fmt.Sprintf("https://api.github.com/user/repos?sort=updated&per_page=%d&page=%d", perPage, page)
	resp, err := g.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list repos: status %d, body: %s", resp.StatusCode, string(body))
	}

	var repos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode repos: %w", err)
	}

	return repos, nil
}

func (g *GitHubClient) GetRepo(owner, name string) (*GitHubRepo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, name)
	resp, err := g.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("repo not found: status %d", resp.StatusCode)
	}

	var repo GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("failed to decode repo: %w", err)
	}

	return &repo, nil
}

func (g *GitHubClient) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Aihub-Platform")

	return g.httpClient.Do(req)
}
