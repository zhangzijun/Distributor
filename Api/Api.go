package Api

import "github.com/gin-gonic/gin"

func ApiServer() {
	initDB()

	r := gin.Default()
	r.GET("/key_get", getUIDKey)
	r.GET("/key_used", keyUsed)
	r.Run(":8123")
}
