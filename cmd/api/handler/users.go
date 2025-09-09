package handler

import (
	"net/http"

	"github.com/Quineeryn/go-backend-101/cmd/api/grpcclient"
	userv1 "github.com/Quineeryn/go-backend-101/pkg/proto/user/v1"
	"github.com/gin-gonic/gin"
)

type Users struct {
	UC *grpcclient.UserClient
}

func (h *Users) Get(c *gin.Context) {
	id := c.Param("id")
	ctx := grpcclient.WithRequestID(c.Request.Context(), c.GetHeader("X-Request-ID"))
	res, err := h.UC.US.GetUser(ctx, &userv1.GetUserRequest{Id: id})
	if err != nil {
		MapGrpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res.GetUser()})
}

func (h *Users) Create(c *gin.Context) {
	var in struct{ Email, Name string }
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	ctx := grpcclient.WithRequestID(c.Request.Context(), c.GetHeader("X-Request-ID"))
	res, err := h.UC.US.CreateUser(ctx, &userv1.CreateUserRequest{Email: in.Email, Name: in.Name})
	if err != nil {
		MapGrpcErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": res.GetUser()})
}

func (h *Users) List(c *gin.Context) {
	ctx := grpcclient.WithRequestID(c.Request.Context(), c.GetHeader("X-Request-ID"))
	res, err := h.UC.US.ListUsers(ctx, &userv1.ListUsersRequest{Page: 1, PageSize: 50})
	if err != nil {
		MapGrpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res.GetUsers()})
}
