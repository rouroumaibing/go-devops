package response

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego/context"

	"github.com/rouroumaibing/go-devops/pkg/apis/versions"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
)

const (
	HTTPContentType     = "Content-Type"
	MIMEApplicationJSON = "application/json"
)

// body json type
func HttpResponse(statusCode int32, body []byte, ctx *context.Context) {
	ctx.Output.SetStatus(int(statusCode))
	ctx.Output.Header(HTTPContentType, MIMEApplicationJSON)
	if err := ctx.Output.Body(body); err != nil {
		return
	}
}

func SetErrorResponse(ctx *context.Context, err error) {
	if err == nil {
		return
	}

	apiError, ok := err.(versions.Response)
	if !ok {
		apiError = errcode.E.Internal.Internal.WithMessage(err.Error())
	}

	body, err := json.Marshal(apiError)
	if err != nil {
		HttpResponse(http.StatusInternalServerError, nil, ctx)
		return
	}

	HttpResponse(apiError.Code, body, ctx)
}

func AddHeader(ctx *context.Context, key, value string) {
	ctx.Output.Header(key, value)
}
