package ginApp

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/real-web-world/bdk/fastcurd"
	"github.com/real-web-world/bdk/json"
)

const (
	HeaderReqID       = "X-Request-ID"
	HeaderContentType = "Content-Type"
	ContentTypeJSON   = "application/json; charset=utf-8"
)

var (
	respServerBad    = fastcurd.RetJSON{Code: fastcurd.CodeServerError, Msg: "服务器开小差了~"}
	respBadReq       = fastcurd.RetJSON{Code: fastcurd.CodeBadReq, Msg: "错误的请求"}
	respNoChange     = fastcurd.RetJSON{Code: fastcurd.CodeDefaultError, Msg: "无更新"}
	respNoAuth       = fastcurd.RetJSON{Code: fastcurd.CodeNoAuth, Msg: "未授权"}
	respNoLogin      = fastcurd.RetJSON{Code: fastcurd.CodeNoLogin, Msg: "未登录"}
	respReqFrequency = fastcurd.RetJSON{Code: fastcurd.CodeRateLimitError, Msg: "请求速度太快了~"}
	respSuccess      = fastcurd.RetJSON{Code: fastcurd.CodeOk}
)

type (
	App struct {
		C         *gin.Context
		beginTime time.Time
		endTime   time.Time
	}
)

// NewGin 创建gin实例
func NewGin() *gin.Engine {
	engine := gin.New()
	return engine
}

func GetApp(c *gin.Context) App {
	return App{
		C: c,
	}
}

// Response 最终响应
func (app App) Response(code int, retJson fastcurd.RetJSON) {
	app.C.Writer.WriteHeader(code)
	app.C.Writer.Header().Set(HeaderContentType, ContentTypeJSON)
	if err := json.NewEncoder(app.C.Writer).Encode(retJson); err != nil {
		_ = app.C.Error(err)
	}
	app.C.Abort()
}

// resp helper

func (app App) Ok(msg string, params ...any) {
	var actData any = nil
	if len(params) == 1 {
		actData = params[0]
	}
	app.JSON(fastcurd.RetJSON{Code: fastcurd.CodeOk, Msg: msg, Data: actData})
}
func (app App) Data(data any) {
	app.RetData(data)
}
func (app App) ServerError(err error) {
	app.Response(http.StatusOK, fastcurd.RetJSON{Code: fastcurd.CodeServerError, Msg: err.Error()})
}
func (app App) ServerBad() {
	app.Response(http.StatusOK, respServerBad)
}
func (app App) RetData(data any, msgParam ...string) {
	msg := ""
	if len(msgParam) == 1 {
		msg = msgParam[0]
	}
	app.Ok(msg, data)
}
func (app App) JSON(json fastcurd.RetJSON) {
	app.Response(http.StatusOK, json)
}
func (app App) SendList(list any, count int64) {
	app.Response(http.StatusOK, fastcurd.RetJSON{
		Code:  fastcurd.CodeOk,
		Data:  list,
		Count: count,
	})
}
func (app App) BadReq() {
	app.Response(http.StatusOK, respBadReq)
}
func (app App) String(html string) {
	app.C.String(http.StatusOK, html)
}
func (app App) ValidError(err error) {
	resp := fastcurd.RetJSON{}
	var actErr validator.ValidationErrors
	switch {
	case errors.As(err, &actErr):
		resp.Code = fastcurd.CodeValidError
		resp.Msg = actErr[0].Error()
	default:
		if err.Error() == "EOF" {
			resp.Code = fastcurd.CodeValidError
			resp.Msg = "请求参数必填"
		} else {
			resp.Code = fastcurd.CodeDefaultError
			resp.Msg = actErr.Error()
		}
	}
	app.Response(http.StatusOK, resp)
}
func (app App) NoChange() {
	app.JSON(respNoChange)
}
func (app App) NoAuth() {
	app.Response(http.StatusUnauthorized, respNoAuth)
}
func (app App) NoLogin() {
	app.Response(http.StatusUnauthorized, respNoLogin)
}
func (app App) ErrorMsg(msg string) {
	resp := fastcurd.RetJSON{Code: fastcurd.CodeDefaultError, Msg: msg}
	app.Response(http.StatusOK, resp)
}
func (app App) CommonError(err error) {
	app.ErrorMsg(err.Error())
}
func (app App) RateLimitError() {
	app.Response(http.StatusOK, respReqFrequency)
}
func (app App) Success() {
	app.Response(http.StatusOK, respSuccess)
}
func (app App) SendAffectRows(affectRows int) {
	app.Data(gin.H{
		"affectRows": affectRows,
	})
}

// ctx value helper

func (app App) GetProcBeginTime() time.Time {
	return app.beginTime
}
func (app App) GetReqID() string {
	reqID := app.C.GetHeader(HeaderReqID)
	return reqID
}
func (app App) GetProcTime() time.Duration {
	return app.endTime.Sub(app.beginTime)
}
