package server

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"golang.org/x/image/webp"

	"markhub/internal/assets"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestMain(m *testing.M) {
	vips.LoggingSettings(nil, vips.LogLevelWarning)
	vips.Startup(nil)
	code := m.Run()
	vips.Shutdown()
	os.Exit(code)
}

func TestProxyRoutesSuccess(t *testing.T) {
	remotePNG := mustFallback(t, "github")
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "github by id",
			path:     "/github?id=42",
			expected: "https://avatars.githubusercontent.com/u/42?size=100",
		},
		{
			name:     "github by user",
			path:     "/github/octocat",
			expected: "https://github.com/octocat.png?size=100",
		},
		{
			name:     "gravatar",
			path:     "/gravatar/0123456789abcdef0123456789abcdef",
			expected: "https://secure.gravatar.com/avatar/0123456789abcdef0123456789abcdef?r=g&s=100",
		},
		{
			name:     "qq",
			path:     "/qq/12345",
			expected: "https://q1.qlogo.cn/g?b=qq&nk=12345&s=100",
		},
		{
			name:     "telegram",
			path:     "/telegram/some_user",
			expected: "https://t.me/i/userpic/320/some_user.jpg",
		},
		{
			name:     "opencollective",
			path:     "/opencollective/markhub",
			expected: "https://images.opencollective.com/markhub/avatar.png?width=100&height=100",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var requested string
			router := NewRouter(Options{
				Client: &http.Client{
					Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
						requested = request.URL.String()
						return response(http.StatusOK, remotePNG), nil
					}),
				},
			})

			recorder := performRequest(router, test.path)
			assertImageResponse(t, recorder)
			assertWebPSize(t, recorder.Body.Bytes(), 100, 100)
			if requested != test.expected {
				t.Fatalf("requested URL = %q, want %q", requested, test.expected)
			}
			if bytes.Equal(recorder.Body.Bytes(), remotePNG) {
				t.Fatal("response body was not converted from upstream PNG")
			}
		})
	}
}

func TestProxyRoutesFallback(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		fallback string
	}{
		{name: "github", path: "/github?id=42", fallback: "github"},
		{name: "gravatar", path: "/gravatar/0123456789abcdef0123456789abcdef", fallback: "gravatar"},
		{name: "qq", path: "/qq/12345", fallback: "qq"},
		{name: "telegram", path: "/telegram/some_user", fallback: "telegram"},
		{name: "opencollective", path: "/opencollective/markhub", fallback: "opencollective"},
		{name: "favicon", path: "/favicon/example.com", fallback: "favicon"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := NewRouter(Options{
				Client: &http.Client{
					Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
						return response(http.StatusNotFound, []byte("not found")), nil
					}),
				},
			})

			recorder := performRequest(router, test.path)
			assertImageResponse(t, recorder)
			assertWebPSize(t, recorder.Body.Bytes(), 100, 100)
			if bytes.Equal(recorder.Body.Bytes(), mustFallback(t, test.fallback)) {
				t.Fatalf("response body was not converted from %s fallback PNG", test.fallback)
			}
		})
	}
}

func TestGravatarEmailAndValidation(t *testing.T) {
	var requested *url.URL
	router := NewRouter(Options{
		Client: &http.Client{
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				requested = request.URL
				return response(http.StatusOK, mustFallback(t, "gravatar")), nil
			}),
		},
	})

	recorder := performRequest(router, "/gravatar/test@example.com?size=80&default=identicon&rating=pg&name=Test")
	assertImageResponse(t, recorder)
	assertWebPSize(t, recorder.Body.Bytes(), 100, 100)

	sum := md5.Sum([]byte("test@example.com"))
	expectedHash := hex.EncodeToString(sum[:])
	if requested.Path != "/avatar/"+expectedHash {
		t.Fatalf("gravatar path = %q, want /avatar/%s", requested.Path, expectedHash)
	}
	query := requested.Query()
	assertQuery(t, query, "s", "100")
	assertQuery(t, query, "r", "g")
	assertQuery(t, query, "size", "80")
	assertQuery(t, query, "default", "identicon")
	assertQuery(t, query, "rating", "pg")
	assertQuery(t, query, "name", "Test")

	var called bool
	router = NewRouter(Options{
		Client: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				called = true
				return response(http.StatusOK, nil), nil
			}),
		},
	})

	recorder = performRequest(router, "/gravatar/not-a-valid-hash")
	assertImageResponse(t, recorder)
	assertWebPSize(t, recorder.Body.Bytes(), 100, 100)
	if called {
		t.Fatal("invalid gravatar parameter should not call upstream")
	}
	if bytes.Equal(recorder.Body.Bytes(), mustFallback(t, "gravatar")) {
		t.Fatal("invalid gravatar parameter returned raw fallback PNG")
	}
}

func TestGravatarOutputSize(t *testing.T) {
	tests := []struct {
		path         string
		expectedSize int
		expectedS    string
	}{
		{
			path:         "/gravatar/0123456789abcdef0123456789abcdef?s=80&size=120",
			expectedSize: 80,
			expectedS:    "80",
		},
		{
			path:         "/gravatar/0123456789abcdef0123456789abcdef?s=4096",
			expectedSize: 2048,
			expectedS:    "2048",
		},
		{
			path:         "/gravatar/0123456789abcdef0123456789abcdef?s=invalid",
			expectedSize: 100,
			expectedS:    "100",
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			var requested *url.URL
			router := NewRouter(Options{
				Client: &http.Client{
					Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
						requested = request.URL
						return response(http.StatusOK, mustFallback(t, "gravatar")), nil
					}),
				},
			})

			recorder := performRequest(router, test.path)
			assertImageResponse(t, recorder)
			assertWebPSize(t, recorder.Body.Bytes(), test.expectedSize, test.expectedSize)
			assertQuery(t, requested.Query(), "s", test.expectedS)
		})
	}
}

func TestQQSizePriority(t *testing.T) {
	tests := []struct {
		path         string
		expected     string
		expectedSize int
	}{
		{path: "/qq/123?s=640&spec=40", expected: "640", expectedSize: 640},
		{path: "/qq/123?spec=40", expected: "40", expectedSize: 40},
		{path: "/qq/123", expected: "100", expectedSize: 100},
		{path: "/qq/123?s=4096", expected: "2048", expectedSize: 2048},
		{path: "/qq/123?spec=invalid", expected: "100", expectedSize: 100},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			var requested *url.URL
			router := NewRouter(Options{
				Client: &http.Client{
					Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
						requested = request.URL
						return response(http.StatusOK, mustFallback(t, "qq")), nil
					}),
				},
			})

			recorder := performRequest(router, test.path)
			assertImageResponse(t, recorder)
			assertWebPSize(t, recorder.Body.Bytes(), test.expectedSize, test.expectedSize)
			if requested.Query().Get("s") != test.expected {
				t.Fatalf("QQ size = %q, want %q", requested.Query().Get("s"), test.expected)
			}
		})
	}
}

func TestFaviconByLinkTag(t *testing.T) {
	icon := mustFallback(t, "github")
	var requested []string
	router := NewRouter(Options{
		Client: &http.Client{
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				requested = append(requested, request.URL.String())
				switch request.URL.String() {
				case "http://example.com/":
					return response(http.StatusOK, []byte(`<html><head><link rel="icon" href="/favicon.ico?v=1"></head></html>`)), nil
				case "http://example.com/favicon.ico":
					return response(http.StatusOK, icon), nil
				default:
					return response(http.StatusNotFound, nil), nil
				}
			}),
		},
	})

	recorder := performRequest(router, "/favicon/example.com")
	assertImageResponse(t, recorder)
	assertWebPSize(t, recorder.Body.Bytes(), 100, 100)

	expected := []string{"http://example.com/", "http://example.com/favicon.ico"}
	if strings.Join(requested, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("requested URLs = %#v, want %#v", requested, expected)
	}
}

func TestFaviconResizesICO(t *testing.T) {
	icon, err := os.ReadFile("../../test/fixtures/favicon/github.ico")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	router := NewRouter(Options{
		Client: &http.Client{
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				switch request.URL.String() {
				case "http://example.com/":
					return response(http.StatusOK, []byte(`<html></html>`)), nil
				case "http://example.com/favicon.ico":
					return response(http.StatusOK, icon), nil
				default:
					return response(http.StatusNotFound, nil), nil
				}
			}),
		},
	})

	recorder := performRequest(router, "/favicon/example.com")
	assertImageResponse(t, recorder)
	assertWebPSize(t, recorder.Body.Bytes(), 100, 100)
}

func response(status int, data []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     http.Header{},
	}
}

func performRequest(router http.Handler, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

func assertImageResponse(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Header().Get("Cache-Control") != cacheControl {
		t.Fatalf("Cache-Control = %q, want %q", recorder.Header().Get("Cache-Control"), cacheControl)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "image/webp" {
		t.Fatalf("Content-Type = %q, want image/webp", contentType)
	}
}

func assertQuery(t *testing.T, values url.Values, key string, expected string) {
	t.Helper()

	if got := values.Get(key); got != expected {
		t.Fatalf("query %s = %q, want %q", key, got, expected)
	}
}

func assertWebPSize(t *testing.T, data []byte, width, height int) {
	t.Helper()

	config, err := webp.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("data is not a valid WebP: %v", err)
	}
	if config.Width != width || config.Height != height {
		t.Fatalf("WebP size = %dx%d, want %dx%d", config.Width, config.Height, width, height)
	}
}

func mustFallback(t *testing.T, name string) []byte {
	t.Helper()

	data, err := assets.Fallback(name)
	if err != nil {
		t.Fatalf("read fallback %s: %v", name, err)
	}
	return data
}

func Example_gravatarHash() {
	hash, ok := gravatarHash("Test@Example.com")
	fmt.Println(hash, ok)
	// Output: 55502f40dc8b7c769880b10874abc9d0 true
}
