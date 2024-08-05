package Background

import (
	"Distributor/model"
	"database/sql"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"os"
)

// 定义 add 和 sub 函数
func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func BackgroundServer() {
	// 设置 Gin 模式为 release
	gin.SetMode(gin.DebugMode)

	// 初始化数据库连接
	db, err := sql.Open("sqlite3", "./Data/order_management.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建 orders 表
	if err := model.CreateOrderTable(db); err != nil {
		log.Fatal(err)
	}

	// 设置 Gin
	r := gin.Default()

	// 注册自定义模板函数
	funcMap := template.FuncMap{
		"add": add,
		"sub": sub,
	}

	r.SetFuncMap(funcMap)

	// 加载静态文件和模板文件
	staticPath := "./Background/static"
	templatePath := "Background/templates/**/*"

	log.Printf("Static files path: %s", staticPath)
	log.Printf("Templates path: %s", templatePath)

	r.Static("/static", staticPath)
	r.LoadHTMLGlob(templatePath)

	// 注册 API 路由
	// 在此处注册 orders 相关的路由
	RegisterOrderHandlers(r, db)
	RegisterTaskRoutes(r, db)

	// 首页路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8223"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
