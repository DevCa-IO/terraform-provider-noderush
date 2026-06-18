package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a thin wrapper over the NodeRush API gateway. Auth is a personal
// access token (PAT) sent as a bearer token, matching the gateway's
// requireApiToken path.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// apiError mirrors the gateway error envelope: { "error": { "code", "message" } }.
type apiError struct {
	Err struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// do performs a request. `out` may be nil for responses with no body to decode.
func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling %s %s: %w", method, path, err)
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var ae apiError
		if json.Unmarshal(raw, &ae) == nil && ae.Err.Message != "" {
			return fmt.Errorf("%s %s: %d %s (%s)", method, path, res.StatusCode, ae.Err.Message, ae.Err.Code)
		}
		return fmt.Errorf("%s %s: %d %s", method, path, res.StatusCode, strings.TrimSpace(string(raw)))
	}

	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// ───────────────────────── SSH keys ─────────────────────────

type SSHKey struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PublicKey   string `json:"publicKey"`
	Fingerprint string `json:"fingerprint"`
}

func (c *Client) CreateSSHKey(ctx context.Context, name, publicKey string) (*SSHKey, error) {
	var out SSHKey
	err := c.do(ctx, http.MethodPost, "/v1/ssh-keys", map[string]string{"name": name, "publicKey": publicKey}, &out)
	return &out, err
}

func (c *Client) GetSSHKey(ctx context.Context, id string) (*SSHKey, error) {
	var list []SSHKey
	if err := c.do(ctx, http.MethodGet, "/v1/ssh-keys", nil, &list); err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == id {
			return &list[i], nil
		}
	}
	return nil, nil // not found
}

func (c *Client) DeleteSSHKey(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/ssh-keys/"+id, nil, nil)
}

// ───────────────────────── Volumes ─────────────────────────

type Volume struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RegionCode string `json:"regionCode"`
	SizeGB     int64  `json:"sizeGB"`
	Status     string `json:"status"`
}

func (c *Client) CreateVolume(ctx context.Context, name, regionCode string, sizeGB int64) (*Volume, error) {
	var out Volume
	err := c.do(ctx, http.MethodPost, "/v1/volumes", map[string]any{"name": name, "regionCode": regionCode, "sizeGB": sizeGB}, &out)
	return &out, err
}

func (c *Client) GetVolume(ctx context.Context, id string) (*Volume, error) {
	var list []Volume
	if err := c.do(ctx, http.MethodGet, "/v1/volumes", nil, &list); err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == id {
			return &list[i], nil
		}
	}
	return nil, nil
}

func (c *Client) ResizeVolume(ctx context.Context, id string, sizeGB int64) error {
	return c.do(ctx, http.MethodPost, "/v1/volumes/"+id+"/resize", map[string]any{"sizeGB": sizeGB}, nil)
}

func (c *Client) DeleteVolume(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/volumes/"+id, nil, nil)
}

// ───────────────────────── Nodes ─────────────────────────

type Node struct {
	ID            string  `json:"id"`
	Hostname      string  `json:"hostname"`
	RegionCode    string  `json:"regionCode"`
	ImageID       string  `json:"imageId"`
	CPU           int64   `json:"cpu"`
	RAMGB         int64   `json:"ramGB"`
	DiskGB        int64   `json:"diskGB"`
	IPv4          *string `json:"ipv4"`
	IPv6          *string `json:"ipv6"`
	Status        string  `json:"status"`
	BillingMode   string  `json:"billingMode"`
	FailureReason *string `json:"failureReason"`
}

// NodeCreate is the create payload. Only non-zero optional fields are sent.
type NodeCreate struct {
	Hostname    string   `json:"hostname"`
	RegionCode  string   `json:"regionCode"`
	ImageID     string   `json:"imageId"`
	CPU         int64    `json:"cpu"`
	RAMGB       int64    `json:"ramGB"`
	DiskGB      int64    `json:"diskGB"`
	BillingMode string   `json:"billingMode,omitempty"`
	SKUID       string   `json:"skuId,omitempty"`
	CloudInit   string   `json:"cloudInit,omitempty"`
	SSHKeyIDs   []string `json:"sshKeyIds,omitempty"`
}

func (c *Client) CreateNode(ctx context.Context, body NodeCreate) (*Node, error) {
	var out Node
	err := c.do(ctx, http.MethodPost, "/v1/nodes", body, &out)
	return &out, err
}

// GetNode returns the node, or nil if it is gone (404 or DESTROYED).
func (c *Client) GetNode(ctx context.Context, id string) (*Node, error) {
	var out Node
	err := c.do(ctx, http.MethodGet, "/v1/nodes/"+id, nil, &out)
	if err != nil {
		if strings.Contains(err.Error(), ": 404 ") {
			return nil, nil
		}
		return nil, err
	}
	if out.Status == "DESTROYED" {
		return nil, nil
	}
	return &out, nil
}

func (c *Client) DeleteNode(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/nodes/"+id, nil, nil)
}

// ───────────────────────── Images / Plans (data sources) ─────────────────────────

type Image struct {
	ID        string `json:"id"`
	OS        string `json:"os"`
	Label     string `json:"label"`
	IsWindows bool   `json:"isWindows"`
	Active    bool   `json:"active"`
}

func (c *Client) ListImages(ctx context.Context) ([]Image, error) {
	var list []Image
	err := c.do(ctx, http.MethodGet, "/v1/images", nil, &list)
	return list, err
}

type Plan struct {
	ID           string `json:"id"`
	Family       string `json:"family"`
	Label        string `json:"label"`
	CPU          int64  `json:"cpu"`
	RAMGB        int64  `json:"ramGB"`
	DiskGB       int64  `json:"diskGB"`
	HourlyCents  int64  `json:"hourlyCents"`
	MonthlyCents int64  `json:"monthlyCents"`
}

func (c *Client) ListPlans(ctx context.Context, regionCode string) ([]Plan, error) {
	path := "/v1/plans"
	if regionCode != "" {
		path += "?regionCode=" + regionCode
	}
	var list []Plan
	err := c.do(ctx, http.MethodGet, path, nil, &list)
	return list, err
}

// ───────────────────────── Regions (data source) ─────────────────────────

type Region struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	CountryCode string `json:"countryCode"`
	Status      string `json:"status"`
}

func (c *Client) ListRegions(ctx context.Context) ([]Region, error) {
	var list []Region
	err := c.do(ctx, http.MethodGet, "/v1/regions", nil, &list)
	return list, err
}
