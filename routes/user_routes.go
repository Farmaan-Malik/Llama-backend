package routes

import "github.com/gin-gonic/gin"

func (a *Api) RegisterRoutes(e *gin.Engine) {

	r := e.Group("/user")
	r.POST("/signup", a.SignupUserHandler)
	r.POST("/login", a.LoginUserHandler)
}
