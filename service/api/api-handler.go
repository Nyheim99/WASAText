package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.POST("/session", rt.doLogin)
	rt.router.GET("/user", rt.validateAuthorization(rt.getUser))
	rt.router.GET("/users", rt.validateAuthorization(rt.getUsers))

	rt.router.PUT("/user/username", rt.validateAuthorization(rt.setMyUserName))
	rt.router.PUT("/user/photo", rt.validateAuthorization(rt.setMyPhoto))

	rt.router.POST("/conversations", rt.validateAuthorization(rt.createConversation))
	rt.router.GET("/conversations", rt.validateAuthorization(rt.getMyConversations))
	
	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	

	return rt.router
}