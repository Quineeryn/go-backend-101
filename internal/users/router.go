package users

import "github.com/gin-gonic/gin"

// Pendaftaran routes
func RegisterRoutes(r *gin.Engine, h *Handler) {
	g := r.Group("/v1/users")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		//g.GET("/:id", h.Get)
		//g.PUT("/:id", h.Update)
		//g.DELETE("/:id", h.Delete)
	}
}
