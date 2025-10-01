package services

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"github.com/qxbao/asfpc/infras"
	lg "github.com/qxbao/asfpc/pkg/logger"
	"resty.dev/v3"
)

type FacebookGraph struct {
	AccessToken string
}

const BaseURLAndroid string = "https://api.facebook.com/restserver.php"
const BaseURLIOS string = "https://b-graph.facebook.com/auth/login"
const GraphURL string = "https://graph.facebook.com"
const LatestAPIVersion string = "v23.0"
const APISecret string = "62f8ce9f74b12f84c123cc23437a4a32"
const APIKeyAndroid string = "882a8490361da98702bf97a021ddc14d"
const APIKeyIOS string = "6628568379|c1e620fa708a1d5696fb991c1bde5662"

type AccessTokenResponse struct {
	AccessToken *string `json:"access_token"`
}

func (fg FacebookGraph) signCreator(data map[string]string) map[string]string {
	sig := ""
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		sig += fmt.Sprintf("%s=%s", key, data[key])
	}
	sig += APISecret

	hash := md5.Sum([]byte(sig))
	data["sig"] = fmt.Sprintf("%x", hash)

	return data
}

func (fg FacebookGraph) GenerateFBAccessTokenAndroid(username string, password string) (*string, error) {
	data := map[string]string{
		"api_key":                  APIKeyAndroid,
		"email":                    username,
		"credentials_type":         "password",
		"format":                   "json",
		"method":                   "auth.login",
		"generate_machine_id":      "1",
		"generate_session_cookies": "1",
		"locale":                   "en_US",
		"password":                 password,
		"return_ssl_resources":     "0",
		"v":                        "1.0",
	}
	data = fg.signCreator(data)
	values := url.Values{}
	for key, value := range data {
		values.Set(key, value)
	}
	url := fmt.Sprintf("%s?%s", BaseURLAndroid, values.Encode())
	c := resty.New()
	defer c.Close()

	var atResponse AccessTokenResponse
	logger := lg.GetLogger("FacebookGraph")

	resp, err := c.R().
		SetQueryParams(data).
		SetHeader("User-Agent", GetRandomAndroidUA()).
		SetResult(&atResponse).
		Get(url)

	if err != nil {
		return nil, err
	}

	if atResponse.AccessToken == nil {
		logger.Errorf(fmt.Sprintf("Android Token Response missing access_token - Full response: %s", resp.String()))
		return nil, fmt.Errorf("(username = %s, token_type = Android) Failed to get access token: Cannot find access_token in response", username)
	}

	return atResponse.AccessToken, nil
}

func (fg FacebookGraph) GenerateFBAccessTokenIOS(username string, password string) (*string, error) {
	data := map[string]string{
		"access_token": APIKeyIOS,
		"email":        username,
		"password":     password,
		"method":       "post",
	}
	data = fg.signCreator(data)
	values := url.Values{}
	for key, value := range data {
		values.Set(key, value)
	}
	url := fmt.Sprintf("%s?%s", BaseURLIOS, values.Encode())
	c := resty.New()
	defer c.Close()

	var atResponse AccessTokenResponse
	logger := lg.GetLogger("FacebookGraph")

	resp, err := c.R().
		SetResult(&atResponse).
		SetHeader("User-Agent", GetRandomIOSUA()).
		Get(url)

	if err != nil {
		return nil, err
	}

	if atResponse.AccessToken == nil {
		logger.Errorf(fmt.Sprintf("iOS Token Response missing access_token %s - Full response: %s", url, resp.String()))
		return nil, fmt.Errorf("(username = %s, token_type = IOS) Failed to get access token: Cannot find access_token in response", username)
	}

	return atResponse.AccessToken, nil
}

func graphQuery[T any](path string, kwargs *map[string]string) (T, error) {
	c := resty.New()
	defer c.Close()

	var response T

	fullURL := fmt.Sprintf("%s/%s", GraphURL, path)

	resp, err := c.R().
		SetQueryParams(*kwargs).
		SetHeader("User-Agent", GetRandomAndroidUA()).
		Get(fullURL)

	if err != nil {
		var empty T
		return empty, err
	}

	if resp.StatusCode() >= 400 {
		var empty T
		return empty, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	err = json.Unmarshal([]byte(resp.String()), &response)
	if err != nil {
		var empty T
		return empty, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return response, nil
}

func (fg FacebookGraph) GetGroupFeed(groupId *string, kwargs *map[string]string) (infras.GetGroupPostsResponse, error) {
	if fg.AccessToken == "" {
		return infras.GetGroupPostsResponse{}, fmt.Errorf("this account does not have an access token")
	}

	path := fmt.Sprintf("%s/feed", *groupId)
	(*kwargs)["access_token"] = fg.AccessToken
	return graphQuery[infras.GetGroupPostsResponse](path, kwargs)
}

func (fg FacebookGraph) GetUserDetails(userId string, kwargs *map[string]string) (infras.UserProfile, error) {
	(*kwargs)["access_token"] = fg.AccessToken
	return graphQuery[infras.UserProfile](userId, kwargs)
}
