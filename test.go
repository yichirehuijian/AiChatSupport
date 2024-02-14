package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

type Data struct {
	Content string `json:"content"`
}

type Config struct {
	Port int `json:"port"`
}

type Image struct {
	Image string `json:"image"`
}

// readConfig 函数用于读取配置文件
func readConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	config, err := readConfig("config.json")
	if err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}
	port := fmt.Sprintf(":%d", config.Port)
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		//如果不是POST，报错

		fmt.Println("POST检查...")
		if r.Method != "POST" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		//创建一个dataparam
		var datam map[string]interface{}
		//使用ioutil读取请求体的数据，如果读取出错，报错。
		fmt.Println("读取检查...")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		fmt.Println("转换参数检查...")
		// 将读取的body内容转换到datam中，如果转换出错，报错。
		if err := json.Unmarshal(body, &datam); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("属性判断检查...")
		//判断所需的属性是否存在
		apikey, ok := datam["apikey"].(string)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		message, ok := datam["message"].(string)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		expSeconds, ok := datam["exp"].(string)
		if !ok {
			expSeconds_2, ok := datam["exp_seconds"].(string)
			if ok {
				expSeconds = expSeconds_2
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

		}
		model, ok := datam["model"].(string)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//切分apikey为id和secret
		fmt.Println("apikey切分检查...")
		idSecret := strings.Split(apikey, ".")
		if len(idSecret) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id := idSecret[0]
		secret := idSecret[1]
		// 构建jwt
		timestamp := time.Now().Unix()
		expSecondsRe, err := strconv.Atoi(expSeconds)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"api_key":   id,
			"exp":       timestamp + int64(expSecondsRe)*1000,
			"timestamp": timestamp,
		})
		fmt.Println("构建JWT检查")
		token.Header["sign_type"] = "SIGN"
		//在Signed之前，修改对应的header
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			fmt.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 创建请求
		client := &http.Client{}
		//创建一个对应的接口
		requestData := map[string]interface{}{
			"model": model,
			"messages": []map[string]interface{}{
				{"role": "user", "content": message},
			},
		}
		//将json序列化一把
		fmt.Println("尝试序列化检查")
		//fmt.Println(requestData)
		requestDataBytes, err := json.Marshal(requestData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		req, err := http.NewRequest("POST", "https://open.bigmodel.cn/api/paas/v4/chat/completions", strings.NewReader(string(requestDataBytes)))
		fmt.Println("发送测试")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		fmt.Println(tokenString)
		req.Header.Set("Authorization", tokenString) // 添加JWT token到请求头部
		//在Do之前，都是配置部分
		resp, err := client.Do(req)
		//如果出错了，抛出
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//关闭
		defer resp.Body.Close()
		// 读取响应数据
		fmt.Println("读取响应数据")
		responseData, err := io.ReadAll(resp.Body)
		//输出测试的相应数据
		fmt.Println(string(responseData))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 解析响应数据
		var response map[string]interface{}
		fmt.Println("解析响应数据")
		if err := json.Unmarshal(responseData, &response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 提取响应中的消息内容
		fmt.Println("提取内容并回传~")
		result, ok := response["choices"].([]interface{})
		fmt.Println(result)
		if !ok || len(result) == 0 {
			fmt.Println("长度检查失败")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		messageResult, ok := result[0].(map[string]interface{})
		if !ok {
			fmt.Println("成功检查失败")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		messageValue, ok := messageResult["message"].(map[string]interface{})
		if !ok {
			fmt.Println("message 字段类型断言失败")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		messageResponse := messageValue["content"]

		fmt.Println("测试成功了呀！")

		// 将响应消息写入HTTP响应中
		w.WriteHeader(http.StatusOK)
		data := Data{
			Content: messageResponse.(string),
		}
		// 转换为 JSON 格式

		HtmlToImage(MarkdownToHTML(data.Content), message)
		filename := getMD5Hash(message)

		image := "http://localhost/image/" + filename + ".png"

		img := Image{
			Image: image,
		}

		jsonData, err := json.Marshal(img)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 输出 JSON 数据
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	})
	fmt.Println("Server listening on port", config.Port)
	http.ListenAndServe(port, Log(http.DefaultServeMux))
}
