package handlers

import (
	"guycanella-url-shortner/internal/models"
	"guycanella-url-shortner/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	service services.URLService
}

func NewURLHandler(service services.URLService) *URLHandler {
	return &URLHandler{service: service}
}

func (handler *URLHandler) CreateURL(ctx *gin.Context) {
	var req models.CreateURLRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	resp, err := handler.service.CreateURL(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create URL: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

func (handler *URLHandler) Redirect(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")

	url, err := handler.service.GetOriginalURL(shortCode)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "URL not found or expired"})
		return
	}

	ctx.Redirect(http.StatusFound, url.OriginalUrl)
}

func (handler *URLHandler) Stats(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")

	stats, err := handler.service.GetURLStats(shortCode)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

func (handler *URLHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := handler.service.DeleteURL(id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}
