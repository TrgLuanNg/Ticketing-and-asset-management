package repository

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-backend/configs"
	"go-backend/internals/core/domains"
	"go-backend/pkg/logger"
	"go-backend/pkg/utils"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUserID(ctx context.Context, userID string) (*domains.USERINFO, error) {
	var user domains.USERINFO
	err := r.db.WithContext(ctx).
		Where("LOWER(\"UserID\") = LOWER(?)", strings.TrimSpace(userID)).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindPermission(ctx context.Context, userID string) (*domains.PERMISSION, error) {
	var perm domains.PERMISSION
	err := r.db.WithContext(ctx).
		Where("LOWER(\"UserID\") = LOWER(?)", strings.TrimSpace(userID)).
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domains.USERINFO) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) CreateRefreshToken(ctx context.Context, token *domains.REFRESH_TOKEN) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *UserRepository) FindRefreshToken(ctx context.Context, tokenStr string) (*domains.REFRESH_TOKEN, error) {
	var token domains.REFRESH_TOKEN
	err := r.db.WithContext(ctx).Where("\"Token\" = ?", tokenStr).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *UserRepository) UpdateRefreshToken(ctx context.Context, token *domains.REFRESH_TOKEN) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *UserRepository) CleanupExpiredTokens(ctx context.Context) {
	cutoff := time.Now().Add(-72 * time.Hour)
	now := time.Now()
	r.db.WithContext(ctx).
		Where("\"ExpiresAt\" < ? OR (\"IsRevoked\" = true AND \"CreatedAt\" < ?)", now, cutoff).
		Delete(&domains.REFRESH_TOKEN{})
}

func (r *UserRepository) GetUserHistory(ctx context.Context, userID string) ([]domains.HISTORY, error) {
	var history []domains.HISTORY
	err := r.db.WithContext(ctx).
		Where("\"UserID\" = ?", userID).
		Order("\"DateTimeAC\" DESC").
		Find(&history).Error
	return history, err
}

func (r *UserRepository) AddHistory(ctx context.Context, userID, action string) {
	now := time.Now()
	h := domains.HISTORY{
		UserID:     userID,
		Action:     action,
		DateTimeAC: &now,
	}
	_ = r.db.WithContext(ctx).Create(&h)
}

// GetSSOUserInfo exchanges authCode with SSO/OIDC Provider and returns extracted claims and raw token
func (r *UserRepository) GetSSOUserInfo(authCode string, cfg *configs.SSOConfig) (map[string]interface{}, string, error) {
	if cfg == nil || cfg.Host == "" {
		return nil, "", fmt.Errorf("SSO host is not configured")
	}

	// 1. Try Authelia OIDC token endpoint
	isOIDC := strings.Contains(cfg.Host, "authelia") ||
		strings.Contains(cfg.Host, "/api/oidc") ||
		strings.Contains(cfg.Host, "9091") ||
		strings.Contains(cfg.Host, "127.0.0.1")

	if isOIDC {
		tokenURL := strings.TrimRight(cfg.Host, "/") + "/api/oidc/token"
		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", authCode)
		form.Set("redirect_uri", "http://localhost:4200/auth")

		req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(cfg.ClientID, cfg.SecretKey)

		if cfg.IssuerHost != "" {
			req.Host = cfg.IssuerHost
		} else if strings.Contains(cfg.Host, "authelia") {
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
		var res map[string]interface{}
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

		claims := make(map[string]interface{})
		if idToken != "" {
			if c, err := utils.DecodeUnverifiedClaims(idToken); err == nil {
				claims = c
			}
		} else if accToken != "" {
			if c, err := utils.DecodeUnverifiedClaims(accToken); err == nil {
				claims = c
			}
		}

		// Also query UserInfo endpoint (/api/oidc/userinfo) using access_token to guarantee preferred_username
		if accToken != "" {
			userInfoURL := strings.TrimRight(cfg.Host, "/") + "/api/oidc/userinfo"
			uReq, err := http.NewRequest("GET", userInfoURL, nil)
			if err == nil {
				uReq.Header.Set("Authorization", "Bearer "+accToken)
				if cfg.IssuerHost != "" {
					uReq.Host = cfg.IssuerHost
				} else if strings.Contains(cfg.Host, "authelia") {
					uReq.Host = "127.0.0.1:9091"
				}
				uResp, err := client.Do(uReq)
				if err == nil {
					defer uResp.Body.Close()
					uBody, _ := io.ReadAll(uResp.Body)
					var uData map[string]interface{}
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

	tokenStr, err := r.GetSSOAccessToken(authCode, cfg)
	if err != nil {
		return nil, "", err
	}
	claims, err := utils.DecodeUnverifiedClaims(tokenStr)
	return claims, tokenStr, err
}

// GetSSOAccessToken exchanges authCode with SSO/OIDC Provider (Authelia or Yusen Portal)
func (r *UserRepository) GetSSOAccessToken(authCode string, cfg *configs.SSOConfig) (string, error) {
	if cfg == nil || cfg.Host == "" {
		return "", fmt.Errorf("SSO host is not configured")
	}

	// 1. Try Authelia OIDC token endpoint if host is Authelia or OIDC port
	isOIDC := strings.Contains(cfg.Host, "authelia") ||
		strings.Contains(cfg.Host, "/api/oidc") ||
		strings.Contains(cfg.Host, "9091") ||
		strings.Contains(cfg.Host, "127.0.0.1")

	if isOIDC {
		tokenURL := strings.TrimRight(cfg.Host, "/") + "/api/oidc/token"
		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", authCode)
		form.Set("redirect_uri", "http://localhost:4200/auth")

		req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(cfg.ClientID, cfg.SecretKey)

		if cfg.IssuerHost != "" {
			req.Host = cfg.IssuerHost
		} else if strings.Contains(cfg.Host, "authelia") {
			req.Host = "127.0.0.1:9091"
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr, Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to call Authelia token endpoint: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var res map[string]interface{}
		if err := json.Unmarshal(body, &res); err == nil {
			if idToken, ok := res["id_token"].(string); ok && idToken != "" {
				return idToken, nil
			}
			if accToken, ok := res["access_token"].(string); ok && accToken != "" {
				return accToken, nil
			}
			if desc, ok := res["error_description"].(string); ok && desc != "" {
				return "", fmt.Errorf("Authelia OIDC error: %s (%v)", desc, res["error"])
			}
			if errCode, ok := res["error"].(string); ok && errCode != "" {
				return "", fmt.Errorf("Authelia OIDC error: %s", errCode)
			}
		}
		return "", fmt.Errorf("Authelia OIDC token response status %d: %s", resp.StatusCode, string(body))
	}

	// 2. Default to Yusen SSO custom REST endpoint: POST {Host}/api/users/get-access-token?authCode={authCode}
	apiURL := fmt.Sprintf("%s/api/users/get-access-token?authCode=%s", strings.TrimRight(cfg.Host, "/"), url.QueryEscape(authCode))
	payload := map[string]string{
		"clientId":  cfg.ClientID,
		"secretKey": cfg.SecretKey,
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
