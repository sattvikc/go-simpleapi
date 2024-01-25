package main

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/sattvikc/go-fastapi"
)

type LoginPOST struct {
	TenantId      string `path:"tenantId"`
	Authorization string `header:"Authorization"`

	Role    string `query:"role"`
	Headers struct {
		ContentType   string `header:"Content-Type"`
		ContentLength uint64 `header:"Content-Length"`
		UserAgent     string `header:"User-Agent"`
	}

	Body struct {
		Username string         `form:"username"`
		Password string         `form:"password"`
		File     multipart.File `form:"file"`
	} `body:"multipart"`
}

func loginPOST(ctx *fastapi.Context, req LoginPOST) error {
	fmt.Println("Request:", req, ctx)

	ctx.JSON(200, map[string]interface{}{
		"status": "OK",
	})

	bytes, err := io.ReadAll(req.Body.File)
	if err != nil {
		return err
	}
	req.Body.File.Close()
	fmt.Println(string(bytes))

	return nil
}

func main() {
	server := fastapi.New()

	server.POST("/:tenantId/login", loginPOST)
	server.ListenAndServe(":8000")
}
