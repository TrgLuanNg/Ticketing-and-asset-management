package sso

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-backend/configs"
	"go-backend/pkg/logger"
	"go-backend/pkg/utils"
)

type SSOProvider struct {
	cfg *configs.SSOConfig
}

func NewSSOProvider(cfg *configs.SSOConfig) *SSOProvider {
	return &SSOProvider{cfg: cfg}
}

// GetSSOUserInfo exchanges authCode with SSO/OIDC Provider and returns extracted claims and raw token.
func (p *SSOProvider) GetSSOUserInfo(authCode string) (map[string]any, string, error) {
	if p.cfg == nil || p.cfg.Host == "" {
		return nil, "", fmt.Errorf("SSO host is not configured")
	}

	isOIDC := strings.Contains(p.cfg.Host, "authelia") ||
		strings.Contains(p.cfg.Host, "/api/oidc") ||
		strings.Contains(p.cfg.Host, "9091") ||
		strings.Contains(p.cfg.Host, "127.0.0.1")

	if isOIDC {
		tokenURL := strings.TrimRight(p.cfg.Host, "/") + "/api/oidc/token"
		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", authCode)
		form.Set("redirect_uri", "http://localhost:4200/auth")

		req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(p.cfg.ClientID, p.cfg.SecretKey)

		if p.cfg.IssuerHost != "" {
			req.Host = p.cfg.IssuerHost
		} else if strings.Contains(p.cfg.Host, "authelia") {
			req.Host = "127.0.0.1:9091"
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr, Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("failed to call Authelia token endpoint: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var res map[string]any
		if err := json.Unmarshal(body, &res); err != nil {
			return nil, "", fmt.Errorf("failed to parse Authelia token response: %w", err)
		}

		if desc, ok := res["error_description"].(string); ok && desc != "" {
			return nil, "", fmt.Errorf("Authelia OIDC error: %s (%v)", desc, res["error"])
		}
		if errCode, ok := res["error"].(string); ok && errCode != "" {
			return nil, "", fmt.Errorf("Authelia OIDC error: %s", errCode)
		}

		idToken, _ := res["id_token"].(string)
		accToken, _ := res["access_token"].(string)

		claims := make(map[string]any)
		if idToken != "" {
			if c, err := utils.DecodeUnverifiedClaims(idToken); err == nil {
				claims = c
			}
		} else if accToken != "" {
			if c, err := utils.DecodeUnverifiedClaims(accToken); err == nil {
				claims = c
			}
		}

		if accToken != "" {
			userInfoURL := strings.TrimRight(p.cfg.Host, "/") + "/api/oidc/userinfo"
			uReq, err := http.NewRequest("GET", userInfoURL, nil)
			if err == nil {
				uReq.Header.Set("Authorization", "Bearer "+accToken)
				if p.cfg.IssuerHost != "" {
					uReq.Host = p.cfg.IssuerHost
				} else if strings.Contains(p.cfg.Host, "authelia") {
					uReq.Host = "127.0.0.1:9091"
				}
				uResp, err := client.Do(uReq)
				if err == nil {
					defer uResp.Body.Close()
					uBody, _ := io.ReadAll(uResp.Body)
					var uData map[string]any
					if json.Unmarshal(uBody, &uData) == nil {
						for k, v := range uData {
							claims[k] = v
						}
					}
				}
			}
		}

		tokenStr := idToken
		if tokenStr == "" {
			tokenStr = accToken
		}
		return claims, tokenStr, nil
	}

	tokenStr, err := p.getSSOAccessToken(authCode)
	if err != nil {
		return nil, "", err
	}
	claims, err := utils.DecodeUnverifiedClaims(tokenStr)
	return claims, tokenStr, err
}

func (p *SSOProvider) getSSOAccessToken(authCode string) (string, error) {
	apiURL := fmt.Sprintf("%s/api/users/get-access-token?authCode=%s", strings.TrimRight(p.cfg.Host, "/"), url.QueryEscape(authCode))
	payload := map[string]string{
		"clientId":  p.cfg.ClientID,
		"secretKey": p.cfg.SecretKey,
	}
	jsonBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status bool   `json:"status"`
		Data   string `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err == nil && result.Status && result.Data != "" {
		return result.Data, nil
	}

	logger.Errorf("SSO token exchange failed. Response: %s", string(body))
	return "", fmt.Errorf("failed to retrieve SSO token")
}
