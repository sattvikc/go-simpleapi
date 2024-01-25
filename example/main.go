package main

import (
	"fmt"

	"github.com/sattvikc/go-fastapi"
)

type LoginPOST struct {
	TenantId      string `path:"tenantId"`
	Authorization *int   `header:"Authorization"`

	Role    string `query:"role"`
	Headers struct {
		ContentType   string `header:"Content-Type"`
		ContentLength uint64 `header:"Content-Length"`
		UserAgent     string `header:"User-Agent"`
	}

	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Id       *int   `json:"id"`

		Numbers []int `json:"numbers"`

		Items []struct {
			Name    string   `json:"name"`
			Address *float64 `json:"address"`
		} `json:"items"`
	} `body:"json"`
}

func loginPOST(ctx *fastapi.Context, req LoginPOST) error {
	fmt.Printf("Request: %+v", req)

	ctx.JSON(200, map[string]interface{}{
		"status": "OK",
	})

	return nil
}

func main() {
	server := fastapi.New()
	server.POST("/:tenantId/login", loginPOST)
	server.ListenAndServe(":8000")
}
