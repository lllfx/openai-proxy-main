package handler

import (
	"github.com/gin-gonic/gin"
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
		//bearerHeader := req.Header.Get(authHeader)
		//arr := strings.Split(bearerHeader, " ")
		//var key string
		//if len(arr) == 2 {
		//	key = arr[1]
		//}
		//if key == "null" || strings.Contains(key, "null") || strings.Contains(key, "xxx") {
		//	slog.Info(" token is null, using default token")
		//	req.Header.Del(authHeader)
		//	req.Header.Set(authHeader, "Bearer "+defaultToken)
		//}
	}

	req.URL.RawQuery = ""
	//fmt.Println("重写了")
	//fmt.Println(req.URL.String())
	//byteData, err := io.ReadAll(req.Body)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(string(byteData))
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
	slog.Info("defaultToken", defaultToken)
	slog.Info("openAIApiAddr", openAIApiAddr)
	proxy, err := NewProxy(openAIApiAddr)
	if err != nil {
		slog.Error("new proxy error", "error", err)
		return
	}
	openaiProxy = proxy
}

func proxy(c *gin.Context) {
	slog.Info("proxy request",
		"CF-Connecting-IP", c.Request.Header.Get("CF-Connecting-IP"),
		"ua", c.Request.UserAgent(),
		"method", c.Request.Method,
		"path", c.Request.URL.Path)
	openaiProxy.ServeHTTP(c.Writer, c.Request)
}
