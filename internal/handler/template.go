package handler

import (
	"smtp-lite/internal/model"
	"smtp-lite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TemplateHandler struct {
	templateSvc *service.TemplateService
}

func NewTemplateHandler(templateSvc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateSvc: templateSvc}
}

func (h *TemplateHandler) List(c *gin.Context) {
	templates, err := h.templateSvc.List()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, templates)
}

func (h *TemplateHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	template, err := h.templateSvc.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(200, template)
}

type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Subject     string `json:"subject"`
	Body        string `json:"body" binding:"required"`
	IsHTML      bool   `json:"is_html"`
	Description string `json:"description"`
}

func (h *TemplateHandler) Create(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	template := &model.EmailTemplate{
		Name:        req.Name,
		Subject:     req.Subject,
		Body:        req.Body,
		IsHTML:      req.IsHTML,
		Description: req.Description,
	}

	if err := h.templateSvc.Create(template); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, template)
}

type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
	IsHTML      bool   `json:"is_html"`
	Description string `json:"description"`
}

func (h *TemplateHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Subject != "" {
		updates["subject"] = req.Subject
	}
	if req.Body != "" {
		updates["body"] = req.Body
	}
	updates["is_html"] = req.IsHTML
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if err := h.templateSvc.Update(id, updates); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Updated"})
}

func (h *TemplateHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.templateSvc.Delete(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *TemplateHandler) Duplicate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	template, err := h.templateSvc.Duplicate(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, template)
}