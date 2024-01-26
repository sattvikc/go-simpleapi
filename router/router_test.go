package router_test

import (
	"net/http"
	"testing"

	"github.com/sattvikc/go-strapi/router"
	"github.com/stretchr/testify/assert"
)

func TestDifferentMethodsForSameRoute(t *testing.T) {
	r := router.New()
	r.Add("/hello", http.MethodGet, "get_correct", "")
	r.Add("/hello", http.MethodPost, "post_correct", "")
	r.Add("/hello", http.MethodPut, "put_correct", "")

	{
		res, _ := r.FindCall("/hello", http.MethodGet)
		assert.Equal(t, "get_correct", res)
	}
	{
		res, _ := r.FindCall("/hello", http.MethodPost)
		assert.Equal(t, "post_correct", res)
	}
	{
		res, _ := r.FindCall("/hello", http.MethodDelete)
		assert.Nil(t, res)
	}

	{
		res, _ := r.FindCall("/hello/", http.MethodGet)
		assert.Equal(t, "get_correct", res)
	}
	{
		res, _ := r.FindCall("/hello/", http.MethodPost)
		assert.Equal(t, "post_correct", res)
	}
	{
		res, _ := r.FindCall("/hello/", http.MethodDelete)
		assert.Nil(t, res)
	}
}

func TestUrlParams(t *testing.T) {
	r := router.New()

	r.Add("/hello/{name}", http.MethodGet, "get_correct", "")
	r.Add("/{a}/{b}/{c}", http.MethodGet, "abc_get_correct", "")

	{
		res, params := r.FindCall("/hello/sattvik", http.MethodGet)
		assert.Equal(t, "get_correct", res)
		assert.Equal(t, "sattvik", params.ByName("name"))
	}

	{
		res, params := r.FindCall("/hello/sattvik/chakravarthy", http.MethodGet)
		assert.Equal(t, "abc_get_correct", res)
		assert.Equal(t, "hello", params.ByName("a"))
		assert.Equal(t, "sattvik", params.ByName("b"))
		assert.Equal(t, "chakravarthy", params.ByName("c"))
		assert.Equal(t, "", params.ByName("d"))
	}
}

func TestNotSpecifyingMethodIsTreatedAsGet(t *testing.T) {
	r := router.New()

	r.Add("/hello/{name}", "", "get_correct", "")
	r.Add("/{a}/{b}/{c}", "", "abc_get_correct", "")

	{
		res, params := r.FindCall("/hello/sattvik", http.MethodGet)
		assert.Equal(t, "get_correct", res)
		assert.Equal(t, "sattvik", params.ByName("name"))
	}

	{
		res, params := r.FindCall("/hello/sattvik/chakravarthy", http.MethodGet)
		assert.Equal(t, "abc_get_correct", res)
		assert.Equal(t, "hello", params.ByName("a"))
		assert.Equal(t, "sattvik", params.ByName("b"))
		assert.Equal(t, "chakravarthy", params.ByName("c"))
		assert.Equal(t, "", params.ByName("d"))
	}
}

func TestRouteNaming(t *testing.T) {
	r := router.New()

	r.Add("/hello/{name}", "", "get_correct", "hello")
	r.Add("/{a}/{b}/{c}", "", "abc_get_correct", "abc")

	{
		assert.Equal(t, "/hello/{name}/", r.FindPattern("hello"))
	}
}

func TestCatchAllRouteWithParams(t *testing.T) {
	r := router.New()

	r.Add("/hello/{name}", "", "get_correct", "hello")
	r.Add("/{a}/{b}/{c}/*", "", "abc_get_correct", "abc")

	{
		res, params := r.FindCall("/hello/sattvik/c/extra/path", http.MethodGet)
		assert.Equal(t, "abc_get_correct", res)
		assert.Equal(t, "hello", params.ByName("a"))
		assert.Equal(t, "sattvik", params.ByName("b"))
		assert.Equal(t, "c", params.ByName("c"))
	}
}
