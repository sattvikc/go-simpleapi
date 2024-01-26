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
}
