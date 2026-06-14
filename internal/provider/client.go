package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ErrNotFound is returned when the API responds 404.
var ErrNotFound = errors.New("not found")

type Client struct {
	endpoint string
	token    string
	hc       *http.Client
}

func NewClient(endpoint, token, caCertFile string, insecure bool) (*Client, error) {
	endpoint = strings.TrimRight(endpoint, "/")
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if insecure {
		tlsCfg.InsecureSkipVerify = true
	} else if caCertFile != "" {
		pem, err := os.ReadFile(caCertFile)
		if err != nil {
			return nil, fmt.Errorf("read ca_cert_file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("ca_cert_file %q has no valid certificates", caCertFile)
		}
		tlsCfg.RootCAs = pool
	}
	return &Client{
		endpoint: endpoint,
		token:    token,
		hc: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
		},
	}, nil
}

// do performs a JSON request. On 2xx with out!=nil it decodes the body into out.
// Returns ErrNotFound on 404 and a descriptive error on other non-2xx codes.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s: status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func pathEscape(s string) string { return url.PathEscape(s) }

// --- API payload types (mirror the control-plane JSON) ---

type apiDomain struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Enabled     bool              `json:"enabled"`
	Variables   map[string]string `json:"variables"`
	CreatedAt   string            `json:"created_at,omitempty"`
	UpdatedAt   string            `json:"updated_at,omitempty"`
}

type apiUser struct {
	ID        string            `json:"id,omitempty"`
	Domain    string            `json:"domain"`
	Number    string            `json:"number"`
	Enabled   bool              `json:"enabled"`
	Params    map[string]string `json:"params"`
	Variables map[string]string `json:"variables"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

type apiGateway struct {
	ID        string            `json:"id,omitempty"`
	Name      string            `json:"name"`
	Profile   string            `json:"profile"`
	Enabled   bool              `json:"enabled"`
	Username  string            `json:"username,omitempty"`
	Password  string            `json:"password,omitempty"`
	Realm     string            `json:"realm,omitempty"`
	Proxy     string            `json:"proxy"`
	Register  bool              `json:"register"`
	Params    map[string]string `json:"params"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

type apiAction struct {
	Application string `json:"application"`
	Data        string `json:"data,omitempty"`
}

type apiCondition struct {
	Field      string            `json:"field"`
	Expression string            `json:"expression"`
	Time       map[string]string `json:"time,omitempty"`
	Actions    []apiAction       `json:"actions"`
}

type apiExtension struct {
	ID         string         `json:"id,omitempty"`
	Name       string         `json:"name"`
	Domain     string         `json:"domain"`
	Context    string         `json:"context"`
	Priority   int64          `json:"priority"`
	Enabled    bool           `json:"enabled"`
	Conditions []apiCondition `json:"conditions"`
	CreatedAt  string         `json:"created_at,omitempty"`
	UpdatedAt  string         `json:"updated_at,omitempty"`
}

// --- typed API operations ---

func (c *Client) createDomain(ctx context.Context, d apiDomain) (*apiDomain, error) {
	var out apiDomain
	return &out, c.do(ctx, http.MethodPost, "/api/v1/domains", d, &out)
}
func (c *Client) getDomain(ctx context.Context, name string) (*apiDomain, error) {
	var out apiDomain
	return &out, c.do(ctx, http.MethodGet, "/api/v1/domains/"+pathEscape(name), nil, &out)
}
func (c *Client) updateDomain(ctx context.Context, name string, d apiDomain) (*apiDomain, error) {
	var out apiDomain
	return &out, c.do(ctx, http.MethodPut, "/api/v1/domains/"+pathEscape(name), d, &out)
}
func (c *Client) deleteDomain(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/domains/"+pathEscape(name), nil, nil)
}

func (c *Client) createUser(ctx context.Context, u apiUser) (*apiUser, error) {
	var out apiUser
	return &out, c.do(ctx, http.MethodPost, "/api/v1/users", u, &out)
}
func (c *Client) getUser(ctx context.Context, domain, number string) (*apiUser, error) {
	var out apiUser
	return &out, c.do(ctx, http.MethodGet, "/api/v1/users/"+pathEscape(domain)+"/"+pathEscape(number), nil, &out)
}
func (c *Client) updateUser(ctx context.Context, domain, number string, u apiUser) (*apiUser, error) {
	var out apiUser
	return &out, c.do(ctx, http.MethodPut, "/api/v1/users/"+pathEscape(domain)+"/"+pathEscape(number), u, &out)
}
func (c *Client) deleteUser(ctx context.Context, domain, number string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/users/"+pathEscape(domain)+"/"+pathEscape(number), nil, nil)
}

func (c *Client) createGateway(ctx context.Context, g apiGateway) (*apiGateway, error) {
	var out apiGateway
	return &out, c.do(ctx, http.MethodPost, "/api/v1/gateways", g, &out)
}
func (c *Client) getGateway(ctx context.Context, profile, name string) (*apiGateway, error) {
	var out apiGateway
	return &out, c.do(ctx, http.MethodGet, "/api/v1/gateways/"+pathEscape(profile)+"/"+pathEscape(name), nil, &out)
}
func (c *Client) updateGateway(ctx context.Context, profile, name string, g apiGateway) (*apiGateway, error) {
	var out apiGateway
	return &out, c.do(ctx, http.MethodPut, "/api/v1/gateways/"+pathEscape(profile)+"/"+pathEscape(name), g, &out)
}
func (c *Client) deleteGateway(ctx context.Context, profile, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/gateways/"+pathEscape(profile)+"/"+pathEscape(name), nil, nil)
}

type apiReloadResult struct {
	Status  string `json:"status"`
	Command string `json:"command"`
	Message string `json:"message"`
}

type apiGatewayStatus struct {
	Name       string            `json:"name"`
	Profile    string            `json:"profile"`
	Status     string            `json:"status"`
	State      string            `json:"state"`
	Attributes map[string]string `json:"attributes"`
}

type apiRegistration struct {
	User        string `json:"user"`
	Domain      string `json:"domain"`
	Registered  bool   `json:"registered"`
	Contact     string `json:"contact"`
	Agent       string `json:"agent"`
	NetworkIP   string `json:"network_ip"`
	NetworkPort string `json:"network_port"`
	Expires     string `json:"expires"`
}

func (c *Client) reloadXML(ctx context.Context) (*apiReloadResult, error) {
	var out apiReloadResult
	return &out, c.do(ctx, http.MethodPost, "/api/v1/runtime/reloadxml", nil, &out)
}

func (c *Client) gatewayStatus(ctx context.Context, profile, name string) (*apiGatewayStatus, error) {
	var out apiGatewayStatus
	return &out, c.do(ctx, http.MethodGet, "/api/v1/runtime/gateways/"+pathEscape(profile)+"/"+pathEscape(name), nil, &out)
}

func (c *Client) registration(ctx context.Context, domain, user string) (*apiRegistration, error) {
	var out apiRegistration
	return &out, c.do(ctx, http.MethodGet, "/api/v1/runtime/registrations/"+pathEscape(domain)+"/"+pathEscape(user), nil, &out)
}

// ---------- callcenter ----------

type apiCCQueue struct {
	ID                                string            `json:"id,omitempty"`
	Name                              string            `json:"name"`
	Strategy                          string            `json:"strategy"`
	MohSound                          string            `json:"moh_sound"`
	TimeBaseScore                     string            `json:"time_base_score"`
	MaxWaitTime                       int64             `json:"max_wait_time"`
	MaxWaitTimeWithNoAgent            int64             `json:"max_wait_time_with_no_agent"`
	MaxWaitTimeWithNoAgentTimeReached int64             `json:"max_wait_time_with_no_agent_time_reached"`
	TierRulesApply                    bool              `json:"tier_rules_apply"`
	TierRuleWaitSecond                int64             `json:"tier_rule_wait_second"`
	TierRuleWaitMultiplyLevel         bool              `json:"tier_rule_wait_multiply_level"`
	TierRuleNoAgentNoWait             bool              `json:"tier_rule_no_agent_no_wait"`
	DiscardAbandonedAfter             int64             `json:"discard_abandoned_after"`
	AbandonedResumeAllowed            bool              `json:"abandoned_resume_allowed"`
	Params                            map[string]string `json:"params"`
	CreatedAt                         string            `json:"created_at,omitempty"`
	UpdatedAt                         string            `json:"updated_at,omitempty"`
}

type apiCCAgent struct {
	ID                string            `json:"id,omitempty"`
	Name              string            `json:"name"`
	Type              string            `json:"type"`
	Contact           string            `json:"contact"`
	Status            string            `json:"status"`
	MaxNoAnswer       int64             `json:"max_no_answer"`
	WrapUpTime        int64             `json:"wrap_up_time"`
	RejectDelayTime   int64             `json:"reject_delay_time"`
	BusyDelayTime     int64             `json:"busy_delay_time"`
	NoAnswerDelayTime int64             `json:"no_answer_delay_time"`
	Params            map[string]string `json:"params"`
	CreatedAt         string            `json:"created_at,omitempty"`
	UpdatedAt         string            `json:"updated_at,omitempty"`
}

type apiCCTier struct {
	ID        string `json:"id,omitempty"`
	Queue     string `json:"queue"`
	Agent     string `json:"agent"`
	Level     int64  `json:"level"`
	Position  int64  `json:"position"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func (c *Client) createCCQueue(ctx context.Context, q apiCCQueue) (*apiCCQueue, error) {
	var out apiCCQueue
	return &out, c.do(ctx, http.MethodPost, "/api/v1/callcenter/queues", q, &out)
}
func (c *Client) getCCQueue(ctx context.Context, name string) (*apiCCQueue, error) {
	var out apiCCQueue
	return &out, c.do(ctx, http.MethodGet, "/api/v1/callcenter/queues/"+pathEscape(name), nil, &out)
}
func (c *Client) updateCCQueue(ctx context.Context, name string, q apiCCQueue) (*apiCCQueue, error) {
	var out apiCCQueue
	return &out, c.do(ctx, http.MethodPut, "/api/v1/callcenter/queues/"+pathEscape(name), q, &out)
}
func (c *Client) deleteCCQueue(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/callcenter/queues/"+pathEscape(name), nil, nil)
}

func (c *Client) createCCAgent(ctx context.Context, a apiCCAgent) (*apiCCAgent, error) {
	var out apiCCAgent
	return &out, c.do(ctx, http.MethodPost, "/api/v1/callcenter/agents", a, &out)
}
func (c *Client) getCCAgent(ctx context.Context, name string) (*apiCCAgent, error) {
	var out apiCCAgent
	return &out, c.do(ctx, http.MethodGet, "/api/v1/callcenter/agents/"+pathEscape(name), nil, &out)
}
func (c *Client) updateCCAgent(ctx context.Context, name string, a apiCCAgent) (*apiCCAgent, error) {
	var out apiCCAgent
	return &out, c.do(ctx, http.MethodPut, "/api/v1/callcenter/agents/"+pathEscape(name), a, &out)
}
func (c *Client) deleteCCAgent(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/callcenter/agents/"+pathEscape(name), nil, nil)
}

func (c *Client) createCCTier(ctx context.Context, t apiCCTier) (*apiCCTier, error) {
	var out apiCCTier
	return &out, c.do(ctx, http.MethodPost, "/api/v1/callcenter/tiers", t, &out)
}
func (c *Client) getCCTier(ctx context.Context, queue, agent string) (*apiCCTier, error) {
	var out apiCCTier
	return &out, c.do(ctx, http.MethodGet, "/api/v1/callcenter/tiers/"+pathEscape(queue)+"/"+pathEscape(agent), nil, &out)
}
func (c *Client) updateCCTier(ctx context.Context, queue, agent string, t apiCCTier) (*apiCCTier, error) {
	var out apiCCTier
	return &out, c.do(ctx, http.MethodPut, "/api/v1/callcenter/tiers/"+pathEscape(queue)+"/"+pathEscape(agent), t, &out)
}
func (c *Client) deleteCCTier(ctx context.Context, queue, agent string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/callcenter/tiers/"+pathEscape(queue)+"/"+pathEscape(agent), nil, nil)
}

func (c *Client) ccReload(ctx context.Context) (*apiReloadResult, error) {
	var out apiReloadResult
	return &out, c.do(ctx, http.MethodPost, "/api/v1/runtime/callcenter/reload", nil, &out)
}

type apiConfProfile struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name"`
	Rate            int64             `json:"rate"`
	IntervalMs      int64             `json:"interval_ms"`
	EnergyLevel     int64             `json:"energy_level"`
	ComfortNoise    bool              `json:"comfort_noise"`
	MohSound        string            `json:"moh_sound"`
	VideoMode       string            `json:"video_mode"`
	VideoLayout     string            `json:"video_layout"`
	VideoCanvasSize string            `json:"video_canvas_size"`
	VideoFPS        int64             `json:"video_fps"`
	AutoRecord      string            `json:"auto_record"`
	Params          map[string]string `json:"params"`
	CreatedAt       string            `json:"created_at,omitempty"`
	UpdatedAt       string            `json:"updated_at,omitempty"`
}

type apiConfRoom struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Number     string `json:"number"`
	Domain     string `json:"domain"`
	Context    string `json:"context"`
	Profile    string `json:"profile"`
	Pin        string `json:"pin"`
	MaxMembers int64  `json:"max_members"`
	Priority   int64  `json:"priority"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type apiConfMember struct {
	ID             string `json:"id"`
	CallerIDName   string `json:"caller_id_name"`
	CallerIDNumber string `json:"caller_id_number"`
	JoinTime       int64  `json:"join_time"`
	CanHear        bool   `json:"can_hear"`
	CanSee         bool   `json:"can_see"`
	CanSpeak       bool   `json:"can_speak"`
	HasVideo       bool   `json:"has_video"`
	Talking        bool   `json:"talking"`
}

type apiConfStatus struct {
	Name        string          `json:"name"`
	MemberCount int64           `json:"member_count"`
	Rate        int64           `json:"rate"`
	RunTime     int64           `json:"run_time"`
	Members     []apiConfMember `json:"members"`
}

func (c *Client) createConfProfile(ctx context.Context, p apiConfProfile) (*apiConfProfile, error) {
	var out apiConfProfile
	return &out, c.do(ctx, http.MethodPost, "/api/v1/conference/profiles", p, &out)
}
func (c *Client) getConfProfile(ctx context.Context, name string) (*apiConfProfile, error) {
	var out apiConfProfile
	return &out, c.do(ctx, http.MethodGet, "/api/v1/conference/profiles/"+pathEscape(name), nil, &out)
}
func (c *Client) updateConfProfile(ctx context.Context, name string, p apiConfProfile) (*apiConfProfile, error) {
	var out apiConfProfile
	return &out, c.do(ctx, http.MethodPut, "/api/v1/conference/profiles/"+pathEscape(name), p, &out)
}
func (c *Client) deleteConfProfile(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/conference/profiles/"+pathEscape(name), nil, nil)
}

func (c *Client) createConfRoom(ctx context.Context, r apiConfRoom) (*apiConfRoom, error) {
	var out apiConfRoom
	return &out, c.do(ctx, http.MethodPost, "/api/v1/conference/rooms", r, &out)
}
func (c *Client) getConfRoom(ctx context.Context, name string) (*apiConfRoom, error) {
	var out apiConfRoom
	return &out, c.do(ctx, http.MethodGet, "/api/v1/conference/rooms/"+pathEscape(name), nil, &out)
}
func (c *Client) updateConfRoom(ctx context.Context, name string, r apiConfRoom) (*apiConfRoom, error) {
	var out apiConfRoom
	return &out, c.do(ctx, http.MethodPut, "/api/v1/conference/rooms/"+pathEscape(name), r, &out)
}
func (c *Client) deleteConfRoom(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/conference/rooms/"+pathEscape(name), nil, nil)
}

func (c *Client) conferenceStatus(ctx context.Context, name string) (*apiConfStatus, error) {
	var out apiConfStatus
	return &out, c.do(ctx, http.MethodGet, "/api/v1/runtime/conference/"+pathEscape(name), nil, &out)
}

func (c *Client) createExtension(ctx context.Context, e apiExtension) (*apiExtension, error) {
	var out apiExtension
	return &out, c.do(ctx, http.MethodPost, "/api/v1/dialplan/extensions", e, &out)
}
func (c *Client) getExtension(ctx context.Context, id string) (*apiExtension, error) {
	var out apiExtension
	return &out, c.do(ctx, http.MethodGet, "/api/v1/dialplan/extensions/"+pathEscape(id), nil, &out)
}
func (c *Client) updateExtension(ctx context.Context, id string, e apiExtension) (*apiExtension, error) {
	var out apiExtension
	return &out, c.do(ctx, http.MethodPut, "/api/v1/dialplan/extensions/"+pathEscape(id), e, &out)
}
func (c *Client) deleteExtension(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/dialplan/extensions/"+pathEscape(id), nil, nil)
}
