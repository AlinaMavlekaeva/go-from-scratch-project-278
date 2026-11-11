package handlers

import (
	"code/internal/links"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LinkHandler struct {
	DB *sql.DB
}

func New(conn *sql.DB) *LinkHandler {
	return &LinkHandler{DB: conn}
}

type (
	CreateLinkRequest struct {
		OriginalUrl string `json:"original_url"`
		ShortName   string `json:"short_name"`
	}

	UpdateLinkRequest struct {
		ID          int64  `json:"id"`
		OriginalUrl string `json:"original_url"`
		ShortName   string `json:"short_name"`
	}

	LinkResponse struct {
		ID          int64  `json:"id"`
		OriginalUrl string `json:"original_url"`
		ShortName   string `json:"short_name"`
		ShortUrl    string `json:"short_url"`
	}
)

func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	conn := h.DB
	link, err := links.CreateLink(ctx, conn, links.CreateLinkParams(req))
	if err != nil {
		if errors.Is(err, links.ErrValidate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, LinkResponse(link))
}

func (h *LinkHandler) UpdateLinkByID(c *gin.Context) {
	var req UpdateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id
	ctx := c.Request.Context()
	conn := h.DB
	link, err := links.UpdateLinkByID(ctx, conn, links.UpdateLinkParams(req))
	if err != nil {
		if errors.Is(err, links.ErrValidate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, links.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, LinkResponse(link))
}

func (h *LinkHandler) GetLinkByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	conn := h.DB
	link, err := links.GetLinkByID(ctx, conn, id)
	if err != nil {
		if errors.Is(err, links.ErrInvalidID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, links.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, LinkResponse(link))
}

func (h *LinkHandler) GetLinks(c *gin.Context) {
	conn := h.DB
	ctx := c.Request.Context()
	rows, err := links.GetLinks(ctx, conn)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	var response = make([]LinkResponse, len(rows))
	for i, row := range rows {
		response[i] = LinkResponse(row)
	}
	c.JSON(http.StatusOK, response)
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	conn := h.DB
	ctx := c.Request.Context()
	err = links.DeleteLink(ctx, conn, id)
	if err != nil {
		if errors.Is(err, links.ErrInvalidID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
