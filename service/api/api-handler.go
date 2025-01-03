package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.POST("/session", rt.doLogin)
	rt.router.GET("/users", rt.validateAuthorization(rt.getUsers))
	rt.router.GET("/user/conversations", rt.validateAuthorization(rt.getUserConversations))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}