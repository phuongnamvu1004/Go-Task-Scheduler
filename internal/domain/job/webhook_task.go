package job

import (
	"errors"
	"fmt"
	"net/url"
)

// HTTPMethod is an HTTP method supported by webhook tasks.
type HTTPMethod string

const (
	MethodGET    HTTPMethod = "GET"
	MethodPOST   HTTPMethod = "POST"
	MethodPUT    HTTPMethod = "PUT"
	MethodPATCH  HTTPMethod = "PATCH"
	MethodDELETE HTTPMethod = "DELETE"
)

func (m HTTPMethod) Valid() bool {
	switch m {
	case MethodGET, MethodPOST, MethodPUT, MethodPATCH, MethodDELETE:
		return true
	}
	return false
}

// WebhookTask sends an HTTP request when its job runs. The JSON tags define
// both the API and the stored format.
type WebhookTask struct {
	URL     string            `json:"url"`
	Method  HTTPMethod        `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

var _ Task = (*WebhookTask)(nil)

func (t *WebhookTask) Type() TaskType {
	return TaskWebhook
}

func (t *WebhookTask) Validate() error {
	u, err := url.Parse(t.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("webhook: invalid URL %q", t.URL)
	}
	if !t.Method.Valid() {
		return fmt.Errorf("webhook: unsupported HTTP method %q", t.Method)
	}
	if t.Method == MethodGET && t.Body != "" {
		return errors.New("webhook: GET requests cannot have a body")
	}
	return nil
}
