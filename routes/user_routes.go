package routes

import "github.com/gin-gonic/gin"

func RegisterRoutes(e *gin.Engine) {
	r := e.Group("/user")
	r.POST("/signup", SignupUserHandler)
	r.POST("/login", LoginUserHandler)
}
