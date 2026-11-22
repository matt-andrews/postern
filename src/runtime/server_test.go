package runtime

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext(method, target, body string) *gin.Context {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(method, target, reader)
	return c
}

func TestPrepareBindings_AlwaysReturnsCoreKeys(t *testing.T) {
	c := newTestContext(http.MethodGet, "/", "")

	result := prepareBindings(c)

	for _, key := range []string{"query", "route", "headers", "body"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected result to contain key %q", key)
		}
	}
}

func TestPrepareBindings_Query(t *testing.T) {
	c := newTestContext(http.MethodGet, "/search?name=alice&page=2", "")

	result := prepareBindings(c)

	want := map[string]string{"name": "alice", "page": "2"}
	if got := result["query"]; !reflect.DeepEqual(got, want) {
		t.Errorf("query = %v, want %v", got, want)
	}
}

func TestPrepareBindings_QueryRepeatedKeyKeepsFirstValue(t *testing.T) {
	c := newTestContext(http.MethodGet, "/search?tag=go&tag=gin", "")

	result := prepareBindings(c)

	want := map[string]string{"tag": "go"}
	if got := result["query"]; !reflect.DeepEqual(got, want) {
		t.Errorf("query = %v, want %v", got, want)
	}
}

func TestPrepareBindings_Route(t *testing.T) {
	c := newTestContext(http.MethodGet, "/users/42/posts/7", "")
	c.Params = gin.Params{
		{Key: "userId", Value: "42"},
		{Key: "postId", Value: "7"},
	}

	result := prepareBindings(c)

	want := map[string]string{"userId": "42", "postId": "7"}
	if got := result["route"]; !reflect.DeepEqual(got, want) {
		t.Errorf("route = %v, want %v", got, want)
	}
}

func TestPrepareBindings_Headers(t *testing.T) {
	c := newTestContext(http.MethodGet, "/", "")
	c.Request.Header.Set("Authorization", "Bearer abc123")
	c.Request.Header.Set("x-request-id", "req-1")

	result := prepareBindings(c)

	want := map[string]string{
		"Authorization": "Bearer abc123",
		"X-Request-Id":  "req-1",
	}
	if got := result["headers"]; !reflect.DeepEqual(got, want) {
		t.Errorf("headers = %v, want %v", got, want)
	}
}

func TestPrepareBindings_HeaderRepeatedKeepsFirstValue(t *testing.T) {
	c := newTestContext(http.MethodGet, "/", "")
	c.Request.Header.Add("Accept", "text/html")
	c.Request.Header.Add("Accept", "application/json")

	result := prepareBindings(c)

	headers := result["headers"].(map[string]string)
	if got, want := headers["Accept"], "text/html"; got != want {
		t.Errorf("headers[Accept] = %q, want %q", got, want)
	}
}

func TestPrepareBindings_RawBody(t *testing.T) {
	c := newTestContext(http.MethodPost, "/", "hello world")

	result := prepareBindings(c)

	// Note: body is stored as []byte, not string.
	body, ok := result["body"].([]byte)
	if !ok {
		t.Fatalf("body has type %T, want []byte", result["body"])
	}
	if got, want := string(body), "hello world"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestPrepareBindings_EmptyBody(t *testing.T) {
	c := newTestContext(http.MethodGet, "/", "")

	result := prepareBindings(c)

	body, ok := result["body"].([]byte)
	if !ok {
		t.Fatalf("body has type %T, want []byte", result["body"])
	}
	if len(body) != 0 {
		t.Errorf("body = %q, want empty", body)
	}
	if _, ok := result["json"]; ok {
		t.Errorf("expected no json key for an empty body, got %v", result["json"])
	}
}

func TestPrepareBindings_NonJSONBodyHasNoJSONKey(t *testing.T) {
	c := newTestContext(http.MethodPost, "/", "definitely not json")

	result := prepareBindings(c)

	if _, ok := result["json"]; ok {
		t.Errorf("expected no json key for a non-JSON body, got %v", result["json"])
	}
}

func TestPrepareBindings_JSONBody(t *testing.T) {
	c := newTestContext(http.MethodPost, "/", `{"name":"alice","age":30}`)
	c.Request.Header.Set("Content-Type", "application/json")

	result := prepareBindings(c)

	if got, want := string(result["body"].([]byte)), `{"name":"alice","age":30}`; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}

	want := map[string]any{"name": "alice", "age": float64(30)}
	got, ok := result["json"]
	if !ok {
		t.Fatal("expected result to contain key \"json\" for a JSON body")
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("json = %v, want %v", got, want)
	}
}
