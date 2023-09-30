package main

import (
	"encoding/json"
	"fmt"
	"io"
	"main/config"
	"main/request"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type (
	Message struct {
		Role    string `json:"role" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	ChatRequest struct {
		Messages []*Message `json:"messages" binding:"required"`
	}

	ChatStream struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
)

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		prefix := "Bearer "

		// 获取token，如果token格式不合法则拦截
		auth := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(auth, prefix) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("auth", auth)
		ctx.Next()
	}
}

func Chat(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	defer c.Writer.Flush()

	// 解析参数
	req := new(ChatRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.SSEvent("error", "参数错误")
		return
	}

	// url
	baseUrl := strings.TrimSuffix(config.PROXY, "/")
	url := baseUrl + "/v1/chat/completions"

	// messages
	var messages []map[string]string
	for _, m := range req.Messages {
		messages = append(messages, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	// request
	r := request.NewRequest(url)
	r.SetHeader("Content-Type", "application/json")
	r.SetHeader("Authorization", c.GetString("auth"))
	r.SetData(map[string]interface{}{
		"model":    config.MODEL,
		"messages": messages,
		"stream":   true,
	})

	// response
	resp, err := r.POST()
	if err != nil {
		c.SSEvent("error", "服务故障")
		return
	}
	defer resp.Body.Close()

	// 流式读取
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			c.SSEvent("error", "服务故障")
			break
		}

		lines := strings.Split(string(buf[:n]), "data: ")
		for _, line := range lines {
			// 处理
			line = strings.TrimSpace(strings.Trim(line, "\n"))
			if len(line) == 0 {
				continue
			}
			if line == "[DONE]" {
				break
			}

			// 解析
			s := new(ChatStream)
			if err := json.Unmarshal([]byte(strings.Trim(line, "\n")), s); err != nil || len(s.Choices) == 0 {
				c.SSEvent("error", "服务故障")
				break
			}

			// 响应
			c.SSEvent("msg", s.Choices[0].Delta.Content)
			c.Writer.Flush()
		}
	}
}

func main() {
	r := gin.Default()

	r.POST("/chat", Auth(), Chat)

	if err := r.Run(fmt.Sprintf("%s:%d", config.HOST, config.PORT)); err != nil {
		panic(err)
	}
}
