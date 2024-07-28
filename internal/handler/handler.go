package handler

import (
	"bytes"
	"encoding/json"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	defaultToken  string
	openAIApiAddr = "https://api.openai.com"
	authHeader    = "Authorization"
	openaiProxy   *httputil.ReverseProxy
	modelSet      mapset.Set[string]
)

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		modifyRequest(req)
	}

	proxy.ModifyResponse = modifyResponse()
	proxy.ErrorHandler = errorHandler()
	return proxy, nil
}

func modifyRequest(req *http.Request) {
	if req.Header.Get(authHeader) == "" {
		slog.Info("no token found, using default token")
		req.Header.Set(authHeader, "Bearer "+defaultToken)
	} else {
		slog.Info("token found in request")
		req.Header.Del(authHeader)
		req.Header.Set(authHeader, "Bearer "+defaultToken)
	}
	req.URL.RawQuery = ""
	dealSign := false
	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info(string(body))
	//重写
	reqJson := make(map[string]interface{})
	err = json.Unmarshal(body, &reqJson)
	if err != nil {
		slog.Error(err.Error())
	}
	if v, ok := reqJson["model"]; ok {
		if !modelSet.ContainsOne(v.(string)) {
			reqJson["model"] = "Qwen/Qwen2-7B-Instruct"
			body, err = json.Marshal(reqJson)
			if err != nil {
				slog.Error(err.Error())
			} else {
				slog.Info(string(body))
				req.Body = io.NopCloser(bytes.NewBuffer(body))
				req.ContentLength = int64(len(body))
				dealSign = true
			}
		}
	}
	if !dealSign {
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		slog.Error("Got error while modifying response", "error", err)
		return
	}
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		//byteData, err := io.ReadAll(resp.Body)
		//if err != nil {
		//	panic(err)
		//}
		//fmt.Println("modifyResponse", string(byteData))
		return nil
	}
}

func Router(r *gin.Engine) {
	r.GET("/v1/chat/completions", proxy)
	r.NoRoute(proxy)
}

func Init() {
	defaultToken = os.Getenv("OPENAI_API_KEY")
	openAIApiAddr = os.Getenv("BASE_URL")
	if defaultToken == "" {
		defaultToken = "默认值"
	}
	if openAIApiAddr == "" {
		openAIApiAddr = "默认值"
	}
	slog.Info("defaultToken", defaultToken)
	slog.Info("openAIApiAddr", openAIApiAddr)
	proxy, err := NewProxy(openAIApiAddr)
	if err != nil {
		slog.Error("new proxy error", "error", err)
		return
	}
	openaiProxy = proxy
	modelSet = mapset.NewSet[string]()
	modelSet.Add("Qwen/Qwen2-7B-Instruct")
	modelSet.Add("Qwen/Qwen2-1.5B-Instruct")
	modelSet.Add("Qwen/Qwen1.5-7B-Chat")
	modelSet.Add("THUDM/glm-4-9b-chat")
	modelSet.Add("THUDM/chatglm3-6b")
	modelSet.Add("01-ai/Yi-1.5-9B-Chat-16K")
	modelSet.Add("01-ai/Yi-1.5-6B-Chat")
}

func proxy(c *gin.Context) {
	slog.Info("proxy request",
		"CF-Connecting-IP", c.Request.Header.Get("CF-Connecting-IP"),
		"ua", c.Request.UserAgent(),
		"method", c.Request.Method,
		"path", c.Request.URL.Path)
	openaiProxy.ServeHTTP(c.Writer, c.Request)
}
