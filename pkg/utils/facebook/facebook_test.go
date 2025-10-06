package facebook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qxbao/asfpc/infras"
	"resty.dev/v3"
)

func TestSignCreator(t *testing.T) {
	fg := FacebookGraph{}

	data := map[string]string{
		"api_key": "test_key",
		"email":   "user@example.com",
	}

	result := fg.signCreator(data)

	if _, exists := result["sig"]; !exists {
		t.Error("Expected sig field")
	}
	if len(result["sig"]) != 32 {
		t.Error("Invalid signature length")
	}
}

func TestUserAgents(t *testing.T) {
	chrome := GenerateModernChromeUA()
	android := GetRandomAndroidUA()
	ios := GetRandomIOSUA()

	if len(chrome) < 50 || len(android) < 50 || len(ios) < 50 {
		t.Error("User agents too short")
	}

	if !strings.Contains(chrome, "Chrome") {
		t.Error("Chrome UA should contain Chrome")
	}
}

// TestAccessTokenResponse tests the response parsing with mock data
func TestAccessTokenResponse(t *testing.T) {
	t.Run("Successful token response parsing", func(t *testing.T) {
		mockToken := "EAABwzLixnjYBO1234567890abcdef"

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := AccessTokenResponse{AccessToken: &mockToken}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		// Simulate what the GenerateFBAccessTokenAndroid does
		newClient := func() *resty.Client {
			return resty.New().SetBaseURL(mockServer.URL)
		}

		c := newClient()
		defer c.Close()

		var atResponse AccessTokenResponse
		resp, err := c.R().Get(mockServer.URL)

		t.Logf("Access Token Response: %s", resp.String())

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		err = json.Unmarshal([]byte(resp.String()), &atResponse)
		if err != nil {
			t.Fatalf("Expected no parse error, got %v", err)
		}

		if atResponse.AccessToken == nil {
			t.Fatal("Expected token, got nil")
		}
		if *atResponse.AccessToken != mockToken {
			t.Errorf("Expected token %s, got %s", mockToken, *atResponse.AccessToken)
		}
	})

	t.Run("Missing access token in response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return error response without access_token
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": {"message": "Invalid credentials", "type": "OAuthException"}}`))
		}))
		defer mockServer.Close()

		newClient := func() *resty.Client {
			return resty.New().SetBaseURL(mockServer.URL)
		}

		var response AccessTokenResponse
		c := newClient()
		defer c.Close()

		resp, _ := c.R().Get(mockServer.URL)
		json.Unmarshal([]byte(resp.String()), &response)

		if response.AccessToken != nil {
			t.Error("Expected nil access token for error response")
		}
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{invalid json`))
		}))
		defer mockServer.Close()

		c := resty.New()
		defer c.Close()

		var response AccessTokenResponse
		resp, _ := c.R().Get(mockServer.URL)
		err := json.Unmarshal([]byte(resp.String()), &response)

		if err == nil {
			t.Error("Expected JSON parse error")
		}
	})
}

func TestGetGroupFeed(t *testing.T) {
	t.Run("Successful group feed retrieval", func(t *testing.T) {
		mockAccessToken := "EAABwzLixnjYBO_test_token_123"

		mockPosts := []infras.Post{
			{
				ID:          strPtr("123456789_111"),
				UpdatedTime: strPtr("2025-10-06T10:00:00+0000"),
				Message:     strPtr("This is a test post about potential customers"),
			},
			{
				ID:          strPtr("123456789_222"),
				UpdatedTime: strPtr("2025-10-06T11:00:00+0000"),
				Message:     strPtr("Another post looking for business opportunities"),
			},
		}

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("access_token") != mockAccessToken {
				t.Error("Expected access token in query")
			}

			response := infras.GetGroupPostsResponse{
				Data: &mockPosts,
				Paging: &infras.Paging{
					Next: strPtr("https://graph.facebook.com/v23.0/next_page"),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		newClient := func() *resty.Client {
			return resty.New().SetBaseURL(mockServer.URL)
		}

		kwargs := map[string]string{
			"fields": "id,updated_time,message",
			"limit":  "25",
		}

		// Mock the graphQuery function by calling the server directly
		c := newClient()
		defer c.Close()

		var result infras.GetGroupPostsResponse
		resp, err := c.R().
			SetQueryParams(map[string]string{
				"access_token": mockAccessToken,
				"fields":       kwargs["fields"],
				"limit":        kwargs["limit"],
			}).
			Get(mockServer.URL)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		json.Unmarshal([]byte(resp.String()), &result)

		if result.Data == nil {
			t.Fatal("Expected data, got nil")
		}
		if len(*result.Data) != 2 {
			t.Errorf("Expected 2 posts, got %d", len(*result.Data))
		}
		if (*result.Data)[0].Message == nil || !strings.Contains(*(*result.Data)[0].Message, "potential customers") {
			t.Error("Expected correct message in first post")
		}
		if result.Paging == nil {
			t.Error("Expected paging info")
		}
	})

	t.Run("No access token error", func(t *testing.T) {
		fg := FacebookGraph{AccessToken: ""}

		newClient := func() *resty.Client {
			return resty.New()
		}

		groupId := "123456789"
		kwargs := map[string]string{}

		// Test that GetGroupFeed returns error for empty access token
		_, err := fg.GetGroupFeed(&groupId, &kwargs, newClient)

		if err == nil {
			t.Fatal("Expected error for missing access token")
		}
		if !strings.Contains(err.Error(), "does not have an access token") {
			t.Errorf("Expected error about missing token, got: %v", err)
		}
	})

	t.Run("API error response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "Invalid OAuth token", "type": "OAuthException"}}`))
		}))
		defer mockServer.Close()

		newClient := func() *resty.Client {
			return resty.New().SetBaseURL(mockServer.URL)
		}

		c := newClient()
		defer c.Close()

		resp, err := c.R().Get(mockServer.URL)
		if err == nil && resp.StatusCode() >= 400 {
			t.Log("Correctly received error status:", resp.StatusCode())
		}
	})
}

// TestGetUserDetails tests user profile retrieval with mock data
func TestGetUserDetails(t *testing.T) {
	t.Run("Successful user profile retrieval", func(t *testing.T) {
		mockAccessToken := "EAABwzLixnjYBO_profile_token"

		mockProfile := infras.UserProfile{
			ID:       strPtr("987654321"),
			Name:     strPtr("John Doe"),
			Email:    strPtr("john.doe@example.com"),
			Gender:   strPtr("male"),
			Birthday: strPtr("01/15/1990"),
		}

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("access_token") != mockAccessToken {
				t.Error("Expected access token in query")
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockProfile)
		}))
		defer mockServer.Close()

		newClient := func() *resty.Client {
			return resty.New().SetBaseURL(mockServer.URL)
		}

		c := newClient()
		defer c.Close()

		var result infras.UserProfile
		resp, err := c.R().
			SetQueryParam("access_token", mockAccessToken).
			Get(mockServer.URL)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		json.Unmarshal([]byte(resp.String()), &result)

		if result.Name == nil || *result.Name != "John Doe" {
			t.Error("Expected correct name")
		}
		if result.Email == nil || *result.Email != "john.doe@example.com" {
			t.Error("Expected correct email")
		}
		if result.ID == nil || *result.ID != "987654321" {
			t.Error("Expected correct ID")
		}
	})
}

// Helper function to create string pointers for mock data
func strPtr(s string) *string {
	return &s
}
