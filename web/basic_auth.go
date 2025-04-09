package web

import (
	"backup-x/entity"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// ViewFunc func
type ViewFunc func(http.ResponseWriter, *http.Request)

type Response struct {
	Code    string        `json:"code"`    // 如果字段名为 Code，可以不加 tag
	Message string        `json:"message"` // 解析 "message" 到 Message 字段
	Data    []interface{} `json:"data"`    // data 是数组，元素类型不确定用 interface{}
}

// BasicAuth basic auth
func BasicAuth(f ViewFunc) ViewFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 异常捕获
		defer func() {
			if p := recover(); p != nil {
				w.WriteHeader(http.StatusUnauthorized)
				log.Printf("%s 请求登录保错: %s!\n", r.RemoteAddr, p)
			}
		}()
		// 获取配置
		conf, _ := entity.GetConfigCache()

		// 1. 获取 request header
		jwtAuthPrefix := "JWT "
		auth := r.Header.Get("Authorization")
		if auth == "" {
			// 认证失败，提示 401 Unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("%s 请求登录! 未携带token\n", r.RemoteAddr)
			return
		}
		token := auth[len(jwtAuthPrefix):]
		data := map[string]interface{}{
			"token": token,
		}
		jsonData, _ := json.Marshal(data) // 序列化为 JSON
		// 2. 创建 HTTP 请求
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		req, err := http.NewRequest("POST", conf.Url, bytes.NewBuffer(jsonData))
		if err != nil {
			panic(err)
		}
		// 3. 设置请求头
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", auth)
		// 4. 发送请求
		client := &http.Client{Transport: tr}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		// 5. 读取响应
		body, _ := io.ReadAll(resp.Body)
		var response Response
		err = json.Unmarshal(body, &response)
		if err != nil {
			panic(err)
		}
		if response.Code == "0" {
			// 执行被装饰的函数
			f(w, r)
			return
		}

		// 认证失败，提示 401 Unauthorized
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("%s 请求登录: 认证失败!\n", r.RemoteAddr)
	}
}
