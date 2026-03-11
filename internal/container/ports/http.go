package ports

import (
	"github.com/leadtek-test/q1/common"
	"github.com/leadtek-test/q1/container/app"
)

type HTTPServer struct {
	common.BaseResponse
	App app.Application
}
