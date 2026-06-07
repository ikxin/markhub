package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"markhub/internal/assets"
	faviconimage "markhub/internal/image"
)

const cacheControl = "max-age=2592000"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Options struct {
	Client HTTPClient
}

type Server struct {
	client HTTPClient
}

func NewRouter(options Options) *gin.Engine {
	client := options.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	server := &Server{client: client}
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Recovery())

	router.GET("/github", server.githubByID)
	router.GET("/github/:user", server.githubByUser)
	router.GET("/gravatar/:hash", server.gravatar)
	router.GET("/qq/:number", server.qq)
	router.GET("/telegram/:user", server.telegram)
	router.GET("/opencollective/:user", server.openCollective)
	router.GET("/favicon/:host", server.favicon)

	return router
}

func (s *Server) githubByID(c *gin.Context) {
	id := c.Query("id")
	s.proxy(c, fmt.Sprintf("https://avatars.githubusercontent.com/u/%s?size=100", id), "github")
}

func (s *Server) githubByUser(c *gin.Context) {
	user := c.Param("user")
	s.proxy(c, fmt.Sprintf("https://github.com/%s.png?size=100", user), "github")
}

func (s *Server) gravatar(c *gin.Context) {
	hash, ok := gravatarHash(c.Param("hash"))
	if !ok {
		writeFallback(c, "gravatar")
		return
	}

	query := gravatarQuery(c)
	fetchURL := fmt.Sprintf("https://secure.gravatar.com/avatar/%s?%s", hash, query.Encode())
	s.proxy(c, fetchURL, "gravatar")
}

func (s *Server) qq(c *gin.Context) {
	size, ok := c.GetQuery("s")
	if !ok {
		size, ok = c.GetQuery("spec")
	}
	if !ok {
		size = "100"
	}

	number := c.Param("number")
	s.proxy(c, fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=%s", number, size), "qq")
}

func (s *Server) telegram(c *gin.Context) {
	user := c.Param("user")
	s.proxy(c, fmt.Sprintf("https://t.me/i/userpic/320/%s.jpg", user), "telegram")
}

func (s *Server) openCollective(c *gin.Context) {
	user := c.Param("user")
	s.proxy(c, fmt.Sprintf("https://images.opencollective.com/%s/avatar.png?width=100&height=100", user), "opencollective")
}

func (s *Server) favicon(c *gin.Context) {
	host, ok := normalizeHost(c.Param("host"))
	if !ok {
		writeFallback(c, "favicon")
		return
	}

	if data, err := s.faviconByLinkTag(c.Request.Context(), host); err == nil {
		writeImage(c, data)
		return
	}

	if data, err := s.faviconByDefaultPath(c.Request.Context(), host); err == nil {
		writeImage(c, data)
		return
	}

	writeFallback(c, "favicon")
}

func (s *Server) proxy(c *gin.Context, fetchURL string, fallback string) {
	data, err := s.fetch(c.Request.Context(), fetchURL)
	if err != nil {
		writeFallback(c, fallback)
		return
	}

	writeImage(c, data)
}

func (s *Server) faviconByLinkTag(ctx context.Context, host string) ([]byte, error) {
	pageURL := httpURL(host, "/")
	page, err := s.fetch(ctx, pageURL)
	if err != nil {
		return nil, err
	}

	href, ok := findIconHref(string(page))
	if !ok {
		return nil, errors.New("icon link not found")
	}

	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}
	iconURL, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	resolved := base.ResolveReference(iconURL)
	resolved.RawQuery = ""

	data, err := s.fetch(ctx, resolved.String())
	if err != nil {
		return nil, err
	}
	return faviconimage.ResizeToPNG(data, 100, 100)
}

func (s *Server) faviconByDefaultPath(ctx context.Context, host string) ([]byte, error) {
	data, err := s.fetch(ctx, httpURL(host, "/favicon.ico"))
	if err != nil {
		return nil, err
	}
	return faviconimage.ResizeToPNG(data, 100, 100)
}

func (s *Server) fetch(ctx context.Context, fetchURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "markhub")

	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("upstream returned %d", response.StatusCode)
	}

	return io.ReadAll(response.Body)
}

func writeImage(c *gin.Context, data []byte) {
	c.Header("Cache-Control", cacheControl)
	c.Data(http.StatusOK, "image/png", data)
}

func writeFallback(c *gin.Context, name string) {
	data, err := assets.Fallback(name)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	writeImage(c, data)
}

var hashPattern = regexp.MustCompile(`(?i)^([a-f0-9]{32}|[a-f0-9]{64})$`)

func gravatarHash(value string) (string, bool) {
	if hashPattern.MatchString(value) {
		return value, true
	}

	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", false
	}

	sum := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:]), true
}

func gravatarQuery(c *gin.Context) url.Values {
	values := url.Values{}
	setQueryDefault(c, values, "s", "100")
	setOptionalQuery(c, values, "size")
	setOptionalQuery(c, values, "default")
	setOptionalQuery(c, values, "f")
	setOptionalQuery(c, values, "forcedefault")
	setQueryDefault(c, values, "r", "g")
	setOptionalQuery(c, values, "rating")
	setOptionalQuery(c, values, "initials")
	setOptionalQuery(c, values, "name")
	return values
}

func setQueryDefault(c *gin.Context, values url.Values, key string, fallback string) {
	if value, ok := c.GetQuery(key); ok {
		values.Set(key, value)
		return
	}
	values.Set(key, fallback)
}

func setOptionalQuery(c *gin.Context, values url.Values, key string) {
	if value, ok := c.GetQuery(key); ok {
		values.Set(key, value)
	}
}

var linkTagPattern = regexp.MustCompile(`(?is)<link\b[^>]*\brel\s*=\s*["']?(icon|shortcut icon|alternate icon|apple-touch-icon)["']?[^>]*>`)
var hrefPattern = regexp.MustCompile(`(?is)\bhref\s*=\s*["']([^"']+)["']`)

func findIconHref(html string) (string, bool) {
	tag := linkTagPattern.FindString(html)
	if tag == "" {
		return "", false
	}

	matches := hrefPattern.FindStringSubmatch(tag)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

var hostnamePattern = regexp.MustCompile(`(?i)^([a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.?$`)

func normalizeHost(host string) (string, bool) {
	host = strings.TrimSpace(host)
	if host == "" || strings.ContainsAny(host, `/\`) || strings.ContainsRune(host, 0) {
		return "", false
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		ip := net.ParseIP(strings.Trim(host, "[]"))
		if ip == nil || ip.To4() != nil {
			return "", false
		}
		return host, true
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() == nil {
			return "[" + host + "]", true
		}
		return host, true
	}

	if len(host) > 253 || !hostnamePattern.MatchString(host) {
		return "", false
	}
	return host, true
}

func httpURL(host string, path string) string {
	return "http://" + host + path
}
