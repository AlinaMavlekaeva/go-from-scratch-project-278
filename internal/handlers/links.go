package handlers

import (
	"code/internal/db"
	"code/internal/links"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	LinkVisitResponse struct {
		ID        int64     `json:"id"`
		LinkID    int64     `json:"link_id"`
		CreatedAt time.Time `json:"created_at"`
		Ip        string    `json:"ip"`
		UserAgent string    `json:"user_agent"`
		Status    int32     `json:"status"`
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

func (h *LinkHandler) RedirectByShortName(c *gin.Context) {
	shortName := c.Param("code")
	ctx := c.Request.Context()
	conn := h.DB
	link, err := links.GetLinkByShortName(ctx, conn, shortName)
	if err != nil {
		if errors.Is(err, links.ErrEmptyShortName) {
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
	const redirectStatus = http.StatusFound
	referer := c.Request.Referer()
	_, err = links.CreateLinkVisit(ctx, conn, db.CreateLinkVisitParams{LinkID: link.ID, Ip: c.ClientIP(), UserAgent: c.Request.UserAgent(), Referer: sql.NullString{String: referer, Valid: referer != ""}, Status: redirectStatus})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	originalUrl := link.OriginalUrl
	c.Redirect(redirectStatus, originalUrl)
}

func (h *LinkHandler) GetLinks(c *gin.Context) {
	conn := h.DB
	ctx := c.Request.Context()
	rawRange, ok := c.GetQuery("range")
	if ok {
		var bounds []int
		if err := json.Unmarshal([]byte(rawRange), &bounds); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(bounds) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "range must contain 2 elements"})
			return
		}
		if bounds[0] < 0 || bounds[1] <= bounds[0] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid range"})
			return
		}
		offset := bounds[0]
		limit := bounds[1] - bounds[0]
		rows, count, err := links.GetLinksPage(ctx, conn, limit, offset)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		var response = make([]LinkResponse, len(rows))
		for i, row := range rows {
			response[i] = LinkResponse(row)
		}
		c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", bounds[0], bounds[1], count))
		c.JSON(http.StatusOK, response)
		return
	}
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

func (h *LinkHandler) GetLinkVisits(c *gin.Context) {
	conn := h.DB
	ctx := c.Request.Context()
	rawRange, ok := c.GetQuery("range")
	if ok {
		var bounds []int
		if err := json.Unmarshal([]byte(rawRange), &bounds); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(bounds) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "range must contain 2 elements"})
			return
		}
		if bounds[0] < 0 || bounds[1] <= bounds[0] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid range"})
			return
		}
		offset := bounds[0]
		limit := bounds[1] - bounds[0]
		rows, count, err := links.GetLinkVisitsPage(ctx, conn, limit, offset)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		var response = make([]LinkVisitResponse, len(rows))
		for i, row := range rows {
			response[i] = LinkVisitResponse(row)
		}
		c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", bounds[0], bounds[1], count))
		c.JSON(http.StatusOK, response)
		return
	}
	rows, err := links.GetLinkVisits(ctx, conn)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	var response = make([]LinkVisitResponse, len(rows))
	for i, row := range rows {
		response[i] = LinkVisitResponse(row)
	}
	c.JSON(http.StatusOK, response)
}
