package web

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/aceberg/WatchYourLAN/internal/check"
	"github.com/aceberg/WatchYourLAN/internal/conf"
	"github.com/aceberg/WatchYourLAN/internal/prometheus"
	"github.com/gin-gonic/gin"
)

// Gui - start web server
func Gui(dirPath, nodePath string) {

	confPath := dirPath + "/config_v2.yaml"
	check.Path(confPath)

	appConfig = conf.Get(confPath)

	appConfig.DirPath = dirPath
	appConfig.ConfPath = confPath
	appConfig.DBPath = dirPath + "/scan.db"
	if nodePath != "" {
		appConfig.NodePath = nodePath
	}

	quitScan = make(chan bool)
	updateRoutines()        // routines-upd.go
	go trimHistoryRoutine() // trim-history.go

	slog.Info("Config dir", "path", appConfig.DirPath)

	address := appConfig.Host + ":" + appConfig.Port

	slog.Info("=================================== ")
	slog.Info("Web GUI at http://" + address)
	slog.Info("=================================== ")

	gin.SetMode(gin.ReleaseMode)
	// router := gin.Default()
	router := gin.New()
	router.Use(gin.Recovery())

	templ := template.Must(template.New("").ParseFS(templFS, "templates/*"))
	router.SetHTMLTemplate(templ) // templates

	router.StaticFS("/fs/", http.FS(pubFS)) // public

	router.GET("/", indexHandler)          // index.go
	router.GET("/config", indexHandler)    // index.go
	router.GET("/history", indexHandler)   // index.go
	router.GET("/host/*any", indexHandler) // index.go
	router.GET("/metrics", prometheus.Handler(&appConfig))

	router.GET("/api/all", apiAll)                    // api.go
	router.GET("/api/config", apiGetConfig)           // api.go
	router.GET("/api/edit/:id/:name/*known", apiEdit) // api.go
	router.GET("/api/history/*mac", apiHistory)       // api.go
	router.GET("/api/host/:id", apiHost)              // api.go
	router.GET("/api/host/del/:id", apiHostDel)       // api.go
	router.GET("/api/notify_test", apiNotifyTest)     // api.go
	router.GET("/api/port/:addr/:port", apiPort)      // api.go
	router.GET("/api/status/*iface", apiStatus)       // api.go
	router.GET("/api/version", apiVersion)            // api.go

	router.POST("/api/config/", saveConfigHandler)                // config.go
	router.POST("/api/config_settings/", saveSettingsHandler)     // config.go
	router.POST("/api/config_influx/", saveInfluxHandler)         // config.go
	router.POST("/api/config_prometheus/", savePrometheusHandler) // config.go

	err := router.Run(address)
	check.IfError(err)
}
