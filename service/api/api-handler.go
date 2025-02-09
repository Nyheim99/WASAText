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

	rt.router.GET("/conversations/:conversationID", rt.validateAuthorization(rt.getConversation))
	
	rt.router.PUT("/conversations/:conversationID/photo", rt.validateAuthorization(rt.setGroupPhoto))
	rt.router.PUT("/conversations/:conversationID/name", rt.validateAuthorization(rt.setGroupName))

	rt.router.POST("/conversations/:conversationID/members", rt.validateAuthorization(rt.addToGroup))
	rt.router.DELETE("/conversations/:conversationID/members/me", rt.validateAuthorization(rt.leaveGroup))

	rt.router.POST("/conversations/:conversationID/messages", rt.validateAuthorization(rt.sendMessage))
	rt.router.POST("/conversations/:conversationID/messages/:messageID/forward", rt.validateAuthorization(rt.forwardMessage))
	rt.router.DELETE("/conversations/:conversationID/messages/:messageID", rt.validateAuthorization(rt.deleteMessage))
	
	rt.router.POST("/conversations/:conversationID/messages/:messageID/reactions", rt.validateAuthorization(rt.commentMessage))
	rt.router.DELETE("/conversations/:conversationID/messages/:messageID/reactions/me", rt.validateAuthorization(rt.uncommentMessage))

	rt.router.PUT("/conversations/:conversationID/messages/read", rt.validateAuthorization(rt.markMessagesAsRead))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
