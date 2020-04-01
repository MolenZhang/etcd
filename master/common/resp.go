package common

import (
	"net/http"

	"github.com/MolenZhang/log"
	"github.com/gin-gonic/gin"
)

// 统一错误码
const (
	InvalidParamters = "Invalid Parameters"
)

// Result 统一返回字段
type Result struct {
	Header interface{} `json:"responseHeader"`
	Data   interface{} `json:"response"`
}

// RespHeader .
type RespHeader struct {
	Status  int    `json:"status"`
	Version string `json:"version"`
	Time    int64  `json:"time"`
	Msg     string `json:"msg"`
}

// SetHeader set header resp
func (r *Result) SetHeader(h interface{}) *Result {
	r.Header = h
	return r
}

// SetData set data resp
func (r *Result) SetData(d interface{}) *Result {
	r.Data = d
	return r
}

// SetStatus 设置状态
func (r *RespHeader) SetStatus(status int) *RespHeader {
	r.Status = status
	return r
}

// SetMsg 设置返回错误信息
func (r *RespHeader) SetMsg(s string) *RespHeader {
	r.Msg = s
	return r
}

// Done 统一处理返回值
func (r *Result) Done(c *gin.Context) {
	if c == nil {
		return
	}

	if err := recover(); err != nil {
		log.Errorf("got panic: %v, url: %v", err, c.Request.URL.Path)
		if c.Writer.Written() {
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.JSON(200, r)
	}
}

// ResultJob Job得适配通用平台返回
type ResultJob struct {
	Data   interface{} `json:"data"`
	Status int         `json:"-"`
	ErrMsg ErrMsg      `json:"msg"`
}

// ErrMsg 错误信息
type ErrMsg struct {
	Msg map[string]string `json:"msg"`
}

// SetMsg .
func (e *ErrMsg) SetMsg(msg string) *ErrMsg {
	e.Msg = map[string]string{}
	e.Msg["error"] = msg
	return e
}

// GetMsg .
func (e *ErrMsg) GetMsg() map[string]string {
	return e.Msg
}

// SetMsg 出错时返回此信息
func (rj *ResultJob) SetMsg(msg map[string]interface{}) *ResultJob {
	rj.Data = msg
	return rj
}

// SetData ,
func (rj *ResultJob) SetData(d interface{}) *ResultJob {
	rj.Data = d
	return rj
}

// SetStatus .
func (rj *ResultJob) SetStatus(status int) *ResultJob {
	rj.Status = status
	return rj
}

// GetStatus .
func (rj *ResultJob) GetStatus() int {
	return rj.Status
}

// GetJobData .
func (rj *ResultJob) GetJobData() interface{} {
	return rj.Data
}

// Done 处理job相关返回
func (rj *ResultJob) Done(c *gin.Context) {
	if c == nil {
		return
	}
	if err := recover(); err != nil {
		log.Errorf("got panic: %v, url: %v", err, c.Request.URL.Path)
		if c.Writer.Written() {
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	switch rj.GetStatus() {
	case http.StatusOK:
		c.JSON(http.StatusOK, rj.GetJobData())
	default:
		c.JSON(rj.GetStatus(), rj.ErrMsg.GetMsg())
	}
}
