package web

import (
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"http-win-notice/api"
	"http-win-notice/utils/setting"
	"net/http"
)

//go:embed templates/*
var htmlFS embed.FS

//go:embed static/*
var staticFS embed.FS

func InitRouters(r *gin.Engine) {

	// 打包进exe
	r.Any("/static/*filepath", func(c *gin.Context) {
		staticServer := http.FileServer(http.FS(staticFS))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
	r.StaticFS("/images", http.Dir(setting.ImagesDir))
	r.SetHTMLTemplate(template.Must(template.New("").ParseFS(htmlFS, "templates/*")))

	//外部引用方式
	//r.LoadHTMLGlob("./web/templates/*")
	//r.Static("/static", "./web/static")

	r.GET("/api/toast", api.ToastApi)
	r.GET("/api/config", api.GetConfigApi)
	r.POST("/api/config", api.UpdateConfigApi)
	r.POST("/api/config/icon/upload", api.UploadIconApi)
	r.GET("/api/frp/status", api.GetFrpStatusApi)
	r.GET("/api/frp/logs", api.GetFrpLogsApi)
	r.GET("/", api.HomePage)
	r.GET("/config", api.ConfigPage)
}
