package url

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

func BuildPasswordResetURL(baseURL, email, token string) string {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		// sensible fallback for dev
		base = "http://localhost:3000/reset-password"
	}

	u, err := url.Parse(base)
	if err != nil {
		return base + "?token=" + url.QueryEscape(token)
	}

	q := u.Query()

	// Extra "confusing" query params
	q.Set("email", url.QueryEscape(email))                              // user email
	q.Set("time", fmt.Sprintf("%d", time.Now().Unix()))                 // unix timestamp
	q.Set("nonce", uuid.Must(uuid.NewV4()).String())                    // random nonce
	q.Set("chk", fmt.Sprintf("%x", sha256.Sum256([]byte(token+email)))) // fake checksum

	// Required
	q.Set("token", token)

	u.RawQuery = q.Encode()
	return u.String()
}

func BuildVerifyEmailURL(baseURL, email, token string) string {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = "http://localhost:3000/verify-email"
	}
	u, err := url.Parse(base)
	if err != nil {
		return base + "?token=" + url.QueryEscape(token)
	}
	q := u.Query()
	q.Set("time", fmt.Sprintf("%d", time.Now().Unix()))                 // unix timestamp
	q.Set("nonce", uuid.Must(uuid.NewV4()).String())                    // random nonce
	q.Set("chk", fmt.Sprintf("%x", sha256.Sum256([]byte(token+email)))) // fake checksum
	q.Set("token", token)
	q.Set("email", email)
	u.RawQuery = q.Encode()
	return u.String()
}

func BuildSendInviteURL(baseURL, email, providerID, token, role string) string {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = "http://localhost:3000/verify-email"
	}

	u, err := url.Parse(base)
	if err != nil {
		return base + "?token=" + url.QueryEscape(token)
	}

	q := u.Query()
	q.Set("time", fmt.Sprintf("%d", time.Now().Unix()))                 // unix timestamp
	q.Set("nonce", uuid.Must(uuid.NewV4()).String())                    // random nonce
	q.Set("id", providerID)                                             // provider ID
	q.Set("chk", fmt.Sprintf("%x", sha256.Sum256([]byte(token+email)))) // fake checksum
	q.Set("token", token)
	q.Set("email", email)
	q.Set("role", role)
	u.RawQuery = q.Encode()
	return u.String()
}

func GetHeader(c *gin.Context, key string) string {
	header := ""
	if c.GetHeader(key) != "" {
		header = c.GetHeader(key)
	} else if c.GetHeader(strings.ToLower(key)) != "" {
		header = c.GetHeader(strings.ToLower(key))
	} else if c.GetHeader(strings.ToUpper(key)) != "" {
		header = c.GetHeader(strings.ToUpper(key))
	} else if c.GetHeader(strings.Title(key)) != "" {
		header = c.GetHeader(strings.Title(key))
	}
	return header
}

func URLDecode(encodedString string) (string, error) {
	decoded, err := url.QueryUnescape(encodedString)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

func UrlHasQuery(urlString string) (bool, error) {
	urlS, err := URLDecode(urlString)
	if err != nil {
		return false, err
	}

	u, err := url.Parse(urlS)
	if err != nil {
		panic(err)
	}

	queryParameters := u.Query()
	if len(queryParameters) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func AddQueryParam(urlStr *string, paramKey string, paramValue string) error {
	// Parse the URL
	u, err := url.Parse(*urlStr)
	if err != nil {
		return err
	}

	// Get the query parameters as a map
	queryParams, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}

	// Add or update the parameter with the given key and value
	queryParams.Set(paramKey, paramValue)

	// Encode the query parameters and rebuild the URL
	u.RawQuery = queryParams.Encode()

	*urlStr = u.String()
	return nil
}

func Stripslashes(s string) string {
	return strings.ReplaceAll(s, "\\", "")
}
func GenerateGroupByURL(appUrl, path string, querys map[string]string) string {
	versionPath := "/v2"
	u, _ := url.ParseRequestURI(appUrl + versionPath + path)

	for key, value := range querys {
		queryParams, _ := url.ParseQuery(u.RawQuery)
		queryParams.Set(key, value)
		u.RawQuery = queryParams.Encode()
	}
	return u.String()
}
