package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"secuarchive-web/internal/models"
	"secuarchive-web/internal/services"

	"github.com/gin-gonic/gin"
)

func GetAIConfig(c *gin.Context) {
	var aiConfig models.AIConfig
	result := services.DB.First(&aiConfig)
	if result.Error != nil {
		aiConfig = models.AIConfig{
			Provider:   "openai",
			Model:      "gpt-3.5-turbo",
			IsEnabled:  false,
			CustomModel: "",
		}
		services.DB.Create(&aiConfig)
	}

	aiConfig.APIKey = ""
	c.JSON(http.StatusOK, aiConfig)
}

func SaveAIConfig(c *gin.Context) {
	var req struct {
		Provider    string `json:"provider"`
		APIKey      string `json:"api_key"`
		APIEndpoint string `json:"api_endpoint"`
		Model       string `json:"model"`
		CustomModel string `json:"custom_model"`
		IsEnabled   bool   `json:"is_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var aiConfig models.AIConfig
	services.DB.First(&aiConfig)

	aiConfig.Provider = req.Provider
	aiConfig.APIEndpoint = req.APIEndpoint
	aiConfig.Model = req.Model
	aiConfig.CustomModel = req.CustomModel
	aiConfig.IsEnabled = req.IsEnabled

	if req.APIKey != "" {
		aiConfig.APIKey = req.APIKey
	}

	services.DB.Save(&aiConfig)
	aiConfig.APIKey = ""
	c.JSON(http.StatusOK, aiConfig)
}

func AIChat(c *gin.Context) {
	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var aiConfig models.AIConfig
	services.DB.First(&aiConfig)

	if !aiConfig.IsEnabled || aiConfig.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI is not configured or disabled"})
		return
	}

	var endpoint string
	switch aiConfig.Provider {
	case "openai":
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/chat/completions"
		} else {
			endpoint = "https://api.openai.com/v1/chat/completions"
		}
	case "anthropic":
		endpoint = "https://api.anthropic.com/v1/messages"
	case "google":
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/models/" + aiConfig.Model + ":generateContent"
		} else {
			endpoint = "https://generativelanguage.googleapis.com/v1/models/" + aiConfig.Model + ":generateContent"
		}
	case "deepseek":
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/chat/completions"
		} else {
			endpoint = "https://api.deepseek.com/v1/chat/completions"
		}
	case "zhipu":
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/chat/completions"
		} else {
			endpoint = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
		}
	case "baidu":
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/chat/completions"
		} else {
			endpoint = "https://qianfan.baidubce.com/v2/chat/completions"
		}
	case "aliyun":
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/services/aigc/text-generation/generation"
		} else {
			endpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
		}
	case "moonshot":
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/chat/completions"
		} else {
			endpoint = "https://api.moonshot.cn/v1/chat/completions"
		}
	case "azure":
		endpoint = aiConfig.APIEndpoint
	case "ollama":
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/api/chat"
		} else {
			endpoint = "http://localhost:11434/api/chat"
		}
	default:
		if aiConfig.APIEndpoint != "" {
			endpoint = aiConfig.APIEndpoint + "/chat/completions"
		} else {
			endpoint = "https://api.openai.com/v1/chat/completions"
		}
	}

	var body map[string]interface{}
	switch aiConfig.Provider {
	case "anthropic":
		body = map[string]interface{}{
			"model":      aiConfig.Model,
			"max_tokens": 4096,
			"messages": []map[string]string{
				{"role": "user", "content": req.Message},
			},
		}
	case "google":
		body = map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]string{
						{"text": req.Message},
					},
				},
			},
			"generationConfig": map[string]interface{}{
				"temperature":     0.9,
				"maxOutputTokens": 2048,
			},
		}
	case "baidu":
		body = map[string]interface{}{
			"model": aiConfig.Model,
			"messages": []map[string]string{
				{"role": "user", "content": req.Message},
			},
		}
	case "aliyun":
		body = map[string]interface{}{
			"model": aiConfig.Model,
			"input": map[string]interface{}{
				"messages": []map[string]string{
					{"role": "user", "content": req.Message},
				},
			},
		}
	case "ollama":
		body = map[string]interface{}{
			"model":  aiConfig.Model,
			"messages": []map[string]string{
				{"role": "user", "content": req.Message},
			},
			"stream": false,
		}
	default:
		body = map[string]interface{}{
			"model": aiConfig.Model,
			"messages": []map[string]string{
				{"role": "user", "content": req.Message},
			},
			"stream": false,
		}
	}

	bodyBytes, _ := json.Marshal(body)

	httpClient := &http.Client{}
	httpReq, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))

	switch aiConfig.Provider {
	case "anthropic":
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", aiConfig.APIKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	case "google":
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
	case "baidu":
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
	case "azure":
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("api-key", aiConfig.APIKey)
	case "aliyun":
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
	case "moonshot":
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
	case "ollama":
		httpReq.Header.Set("Content-Type", "application/json")
	default:
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to AI service: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	switch aiConfig.Provider {
	case "anthropic":
		if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
			if block, ok := content[0].(map[string]interface{}); ok {
				if text, ok := block["text"].(string); ok {
					c.JSON(http.StatusOK, gin.H{"response": text})
					return
				}
			}
		}
	case "google":
		if candidates, ok := result["candidates"].([]interface{}); ok && len(candidates) > 0 {
			if candidate, ok := candidates[0].(map[string]interface{}); ok {
				if content, ok := candidate["content"].(map[string]interface{}); ok {
					if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
						if part, ok := parts[0].(map[string]interface{}); ok {
							if text, ok := part["text"].(string); ok {
								c.JSON(http.StatusOK, gin.H{"response": text})
								return
							}
						}
					}
				}
			}
		}
	case "baidu":
		if resultData, ok := result["result"].(string); ok {
			c.JSON(http.StatusOK, gin.H{"response": resultData})
			return
		}
	case "aliyun":
		if output, ok := result["output"].(map[string]interface{}); ok {
			if text, ok := output["text"].(string); ok {
				c.JSON(http.StatusOK, gin.H{"response": text})
				return
			}
		}
	default:
		if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						c.JSON(http.StatusOK, gin.H{"response": content})
						return
					}
				}
			}
		}
	}

	if errorMsg, ok := result["error"].(map[string]interface{}); ok {
		errMsg := "AI service error"
		if msg, ok := errorMsg["message"].(string); ok {
			errMsg = msg
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "No response from AI service"})
}

func GetAIProviders(c *gin.Context) {
	providers := []map[string]interface{}{
		{
			"id":       "openai",
			"name":     "OpenAI",
			"endpoint": "https://api.openai.com/v1",
			"models":   []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"},
		},
		{
			"id":       "anthropic",
			"name":     "Anthropic (Claude)",
			"endpoint": "https://api.anthropic.com/v1",
			"models":   []string{"claude-sonnet-4-20250514", "claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229", "claude-3-haiku-20240307"},
		},
		{
			"id":       "google",
			"name":     "Google (Gemini)",
			"endpoint": "https://generativelanguage.googleapis.com/v1",
			"models":   []string{"gemini-2.0-flash", "gemini-1.5-pro", "gemini-1.5-flash", "gemini-1.0-pro"},
		},
		{
			"id":       "azure",
			"name":     "Azure OpenAI",
			"endpoint": "",
			"models":   []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-35-turbo"},
		},
		{
			"id":       "deepseek",
			"name":     "DeepSeek",
			"endpoint": "https://api.deepseek.com/v1",
			"models":   []string{"deepseek-chat", "deepseek-coder"},
		},
		{
			"id":       "zhipu",
			"name":     "智谱AI",
			"endpoint": "https://open.bigmodel.cn/api/paas/v4",
			"models":   []string{"glm-4-plus", "glm-4", "glm-4-flash", "glm-3-turbo"},
		},
		{
			"id":       "baidu",
			"name":     "百度AI (ERNIE)",
			"endpoint": "https://qianfan.baidubce.com/v2",
			"models":   []string{"ernie-4.0-8k", "ernie-3.5-8k", "ernie-speed-8k"},
		},
		{
			"id":       "aliyun",
			"name":     "阿里云 (通义千问)",
			"endpoint": "https://dashscope.aliyuncs.com/api/v1",
			"models":   []string{"qwen-plus", "qwen-turbo", "qwen-max"},
		},
		{
			"id":       "moonshot",
			"name":     "月之暗面 (Kimi)",
			"endpoint": "https://api.moonshot.cn/v1",
			"models":   []string{"kimi-k2", "kimi-k1.5", "kimi-k1"},
		},
		{
			"id":       "ollama",
			"name":     "Ollama (本地部署)",
			"endpoint": "",
			"models":   []string{},
		},
		{
			"id":       "custom",
			"name":     "自定义",
			"endpoint": "",
			"models":   []string{},
		},
	}
	c.JSON(http.StatusOK, providers)
}

func TestAIConnection(c *gin.Context) {
	var aiConfig models.AIConfig
	services.DB.First(&aiConfig)

	if !aiConfig.IsEnabled || aiConfig.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI is not configured or disabled"})
		return
	}

	var endpoint string
	if aiConfig.APIEndpoint != "" {
		switch aiConfig.Provider {
		case "openai", "deepseek", "moonshot":
			endpoint = aiConfig.APIEndpoint + "/models"
		case "anthropic":
			endpoint = aiConfig.APIEndpoint
		case "google":
			endpoint = aiConfig.APIEndpoint + "/models"
		case "azure":
			endpoint = aiConfig.APIEndpoint
		case "ollama":
			endpoint = aiConfig.APIEndpoint + "/api/tags"
		default:
			endpoint = aiConfig.APIEndpoint
		}
	} else {
		switch aiConfig.Provider {
		case "openai":
			endpoint = "https://api.openai.com/v1/models"
		case "anthropic":
			endpoint = "https://api.anthropic.com/v1/messages"
		case "deepseek":
			endpoint = "https://api.deepseek.com/v1/models"
		case "google":
			endpoint = "https://generativelanguage.googleapis.com/v1/models"
		case "zhipu":
			endpoint = "https://open.bigmodel.cn/api/paas/v4/models"
		case "baidu":
			endpoint = "https://qianfan.baidubce.com/v2/models"
		case "aliyun":
			endpoint = "https://dashscope.aliyuncs.com/api/v1/models"
		case "moonshot":
			endpoint = "https://api.moonshot.cn/v1/models"
		case "ollama":
			endpoint = "http://localhost:11434/api/tags"
		case "azure":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Azure需要配置自定义端点"})
			return
		default:
			endpoint = "https://api.openai.com/v1/models"
		}
	}

	httpClient := &http.Client{}
	var httpReq *http.Request
	if aiConfig.Provider == "ollama" {
		httpReq, _ = http.NewRequest("GET", endpoint, nil)
	} else {
		httpReq, _ = http.NewRequest("GET", endpoint, nil)
	}

	switch aiConfig.Provider {
	case "anthropic":
		httpReq.Header.Set("x-api-key", aiConfig.APIKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	case "azure":
		httpReq.Header.Set("api-key", aiConfig.APIKey)
	case "google":
		httpReq.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
	default:
		httpReq.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
	} else if resp.StatusCode == 401 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
	} else {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Connection failed with status: " + resp.Status})
	}
}

func ChatWithContext(c *gin.Context) {
	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var reports []models.Report
	services.DB.Limit(10).Order("upload_time DESC").Find(&reports)

	context := "以下是与安全服务报告相关的信息:\n"
	for _, r := range reports {
		context += fmt.Sprintf("- 报告: %s, 分类: %s/%s, 标签: %s\n", 
			r.Name, r.Category, r.SubCategory, r.Tags)
	}

	fullMessage := context + "\n\n用户问题: " + req.Message

	var aiConfig models.AIConfig
	services.DB.First(&aiConfig)

	if !aiConfig.IsEnabled || aiConfig.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI is not configured or disabled"})
		return
	}

	var endpoint string
	if aiConfig.APIEndpoint != "" {
		endpoint = aiConfig.APIEndpoint
	} else {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	body := map[string]interface{}{
		"model": aiConfig.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个安全服务报告分析助手，可以根据提供的报告信息回答用户的问题。"},
			{"role": "user", "content": fullMessage},
		},
	}

	bodyBytes, _ := json.Marshal(body)

	httpClient := &http.Client{}
	httpReq, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to AI service"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					c.JSON(http.StatusOK, gin.H{"response": content})
					return
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"response": "No response from AI service"})
}
