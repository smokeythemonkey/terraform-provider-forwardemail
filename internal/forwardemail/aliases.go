package forwardemail

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// flexString unmarshals a JSON string or number into a string. The API
// returns max_quota as a number (bytes) in responses but accepts a string
// like "1 GB" on write, so the response field needs a tolerant type.
type flexString string

func (f *flexString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = flexString(n.String())
		return nil
	}
	return fmt.Errorf("%w: %s", ErrUnsupportedJSONValue, data)
}

type Alias struct {
	Account                  Account    `json:"user"`
	Domain                   Domain     `json:"domain"`
	Name                     string     `json:"name"`
	Labels                   []string   `json:"labels"`
	Description              string     `json:"description"`
	IsEnabled                bool       `json:"is_enabled"`
	ErrorCodeIfDisabled      int        `json:"error_code_if_disabled"`
	HasRecipientVerification bool       `json:"has_recipient_verification"`
	HasIMAP                  bool       `json:"has_imap"`
	HasPGP                   bool       `json:"has_pgp"`
	PublicKey                string     `json:"public_key"`
	HasSMIME                 bool       `json:"has_smime"`
	SMIMECertificate         string     `json:"smime_certificate"`
	HasWKDDisabled           bool       `json:"has_wkd_disabled"`
	MaxQuota                 flexString `json:"max_quota"`
	Retention                int        `json:"retention"`
	SMTPLimit                int        `json:"smtp_limit"`
	Recipients               []string   `json:"recipients"`
	ID                       string     `json:"id"`
	Object                   string     `json:"object"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

// UnmarshalJSON tolerates the "user" and "domain" fields being serialized
// as plain strings (the update-alias endpoint does this) while other
// endpoints send objects.
func (a *Alias) UnmarshalJSON(data []byte) error {
	type aliasAlias Alias // avoid recursion
	var raw struct {
		aliasAlias
		Account json.RawMessage `json:"user"`
		Domain  json.RawMessage `json:"domain"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*a = Alias(raw.aliasAlias)
	if len(raw.Account) > 0 && string(raw.Account) != "null" {
		if raw.Account[0] == '"' {
			var email string
			if err := json.Unmarshal(raw.Account, &email); err != nil {
				return fmt.Errorf("%w: user: %s", ErrUnsupportedJSONValue, raw.Account)
			}
			a.Account.Email = email
		} else if err := json.Unmarshal(raw.Account, &a.Account); err != nil {
			return err
		}
	}
	if len(raw.Domain) > 0 && string(raw.Domain) != "null" {
		if raw.Domain[0] == '"' {
			var name string
			if err := json.Unmarshal(raw.Domain, &name); err != nil {
				return fmt.Errorf("%w: domain: %s", ErrUnsupportedJSONValue, raw.Domain)
			}
			a.Domain.Name = name
		} else if err := json.Unmarshal(raw.Domain, &a.Domain); err != nil {
			return err
		}
	}
	return nil
}

type AliasParameters struct {
	Recipients               *[]string
	Labels                   *[]string
	Description              *string
	HasRecipientVerification *bool
	IsEnabled                *bool
	ErrorCodeIfDisabled      *int64
	HasIMAP                  *bool
	HasPGP                   *bool
	PublicKey                *string
	HasSMIME                 *bool
	SMIMECertificate         *string
	HasWKDDisabled           *bool
	MaxQuota                 *string
	Retention                *int64
	SMTPLimit                *int64
}

func (c *Client) GetAliases(ctx context.Context, domain string) ([]Alias, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil, ErrEmptyDomain
	}

	encodedDomain := url.PathEscape(domain)

	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/v1/domains/%s/aliases", encodedDomain), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for GetAliases: %w", err)
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch aliases: %w", err)
	}

	var items []Alias
	if err := json.Unmarshal(res, &items); err != nil {
		return nil, fmt.Errorf("failed to parse aliases response: %w", err)
	}

	return items, nil
}

func (c *Client) GetAlias(ctx context.Context, domain, alias string) (*Alias, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil, ErrEmptyDomain
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, ErrEmptyAlias
	}

	encodedDomain := url.PathEscape(domain)
	encodedAlias := url.PathEscape(alias)

	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/v1/domains/%s/aliases/%s", encodedDomain, encodedAlias), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for GetAlias: %w", err)
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alias: %w", err)
	}

	var item Alias
	if err := json.Unmarshal(res, &item); err != nil {
		return nil, fmt.Errorf("failed to parse alias response: %w", err)
	}

	return &item, nil
}

func (c *Client) CreateAlias(ctx context.Context, domain, alias string, parameters AliasParameters) (*Alias, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil, ErrEmptyDomain
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, ErrEmptyAlias
	}

	encodedDomain := url.PathEscape(domain)

	params := url.Values{}
	params.Add("name", alias)
	encodeAliasParams(params, parameters)

	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/v1/domains/%s/aliases", encodedDomain), strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for CreateAlias: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create alias: %w", err)
	}

	var item Alias
	if err := json.Unmarshal(res, &item); err != nil {
		return nil, fmt.Errorf("failed to parse create alias response: %w", err)
	}

	return &item, nil
}

func (c *Client) UpdateAlias(ctx context.Context, domain, alias string, parameters AliasParameters) (*Alias, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil, ErrEmptyDomain
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, ErrEmptyAlias
	}

	encodedDomain := url.PathEscape(domain)
	encodedAlias := url.PathEscape(alias)

	params := url.Values{}
	params.Add("name", alias)
	encodeAliasParams(params, parameters)

	req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/v1/domains/%s/aliases/%s", encodedDomain, encodedAlias), strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for UpdateAlias: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update alias: %w", err)
	}

	var item Alias
	if err := json.Unmarshal(res, &item); err != nil {
		return nil, fmt.Errorf("failed to parse update alias response: %w", err)
	}

	return &item, nil
}

func encodeAliasParams(params url.Values, parameters AliasParameters) {
	for k, v := range map[string]*string{
		"description":       parameters.Description,
		"public_key":        parameters.PublicKey,
		"smime_certificate": parameters.SMIMECertificate,
		"max_quota":         parameters.MaxQuota,
	} {
		if v != nil {
			params.Add(k, *v)
		}
	}

	for k, v := range map[string]*bool{
		"has_recipient_verification": parameters.HasRecipientVerification,
		"is_enabled":                 parameters.IsEnabled,
		"has_imap":                   parameters.HasIMAP,
		"has_pgp":                    parameters.HasPGP,
		"has_smime":                  parameters.HasSMIME,
		"has_wkd_disabled":           parameters.HasWKDDisabled,
	} {
		if v != nil {
			params.Add(k, strconv.FormatBool(*v))
		}
	}

	for k, v := range map[string]*int64{
		"error_code_if_disabled": parameters.ErrorCodeIfDisabled,
		"retention":              parameters.Retention,
		"smtp_limit":             parameters.SMTPLimit,
	} {
		if v != nil {
			params.Add(k, strconv.FormatInt(*v, 10))
		}
	}

	for k, v := range map[string]*[]string{
		"recipients[]": parameters.Recipients,
		"labels[]":     parameters.Labels,
	} {
		if v == nil {
			continue
		}
		if len(*v) == 0 {
			params.Add(k, "")
			continue
		}
		for _, vv := range *v {
			params.Add(k, vv)
		}
	}
}

func (c *Client) DeleteAlias(ctx context.Context, domain, alias string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return ErrEmptyDomain
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return ErrEmptyAlias
	}

	encodedDomain := url.PathEscape(domain)
	encodedAlias := url.PathEscape(alias)

	req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/v1/domains/%s/aliases/%s", encodedDomain, encodedAlias), nil)
	if err != nil {
		return fmt.Errorf("failed to create request for DeleteAlias: %w", err)
	}

	_, err = c.doRequest(req)
	if err != nil {
		return fmt.Errorf("failed to delete alias: %w", err)
	}

	return nil
}
