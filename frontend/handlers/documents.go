package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"frontend/templates"

	"github.com/gin-gonic/gin"
)

func apiBase() string {
	if base := os.Getenv("API_BASE_URL"); base != "" {
		return base
	}
	return "http://localhost:8081"
}

func DocumentsPage(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	templates.DocumentsPage().Render(c.Request.Context(), c.Writer)
}

func DocumentPage(c *gin.Context) {
	id := c.Param("id")
	req, _ := http.NewRequest("GET", apiBase()+"/documents/"+id, nil)
	token, _ := c.Get("token")
	if t, ok := token.(string); ok {
		req.Header.Set("Authorization", "Bearer "+t)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("DocumentPage: GET /documents/%s failed: %v", id, err)
		c.String(http.StatusBadGateway, "Failed to reach API")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.String(http.StatusNotFound, "Document not found")
		return
	}

	var doc templates.Document
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		log.Printf("DocumentPage: failed to decode response for document %s: %v", id, err)
		c.String(http.StatusInternalServerError, "Failed to parse document")
		return
	}

	c.Header("Content-Type", "text/html")
	templates.DocumentPage(doc).Render(c.Request.Context(), c.Writer)
}

func UpdateDocumentTitle(c *gin.Context) {
	id := c.Param("id")
	title := c.PostForm("title")

	payload, _ := json.Marshal(map[string]string{"title": title})
	req, _ := http.NewRequest("PATCH", apiBase()+"/documents/"+id, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	token, _ := c.Get("token")
	if t, ok := token.(string); ok {
		req.Header.Set("Authorization", "Bearer "+t)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("UpdateDocumentTitle: PATCH /documents/%s failed: %v", id, err)
		c.String(http.StatusBadGateway, "Failed to reach API")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Printf("UpdateDocumentTitle: PATCH /documents/%s returned %d", id, resp.StatusCode)
		c.String(resp.StatusCode, "Failed to update title")
		return
	}

	var doc templates.Document
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		// Fallback: use the title we sent
		doc.Title = title
	}

	c.Header("Content-Type", "text/html")
	templates.TitlePartial(doc).Render(c.Request.Context(), c.Writer)
}

func AddPage(c *gin.Context) {
	id := c.Param("id")
	indexStr := c.PostForm("index")
	indexInt, _ := strconv.Atoi(indexStr)

	payload, _ := json.Marshal(map[string]int{"page_index": indexInt})
	req, _ := http.NewRequest("POST", apiBase()+"/documents/"+id+"/pages", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	token, _ := c.Get("token")
	if t, ok := token.(string); ok {
		req.Header.Set("Authorization", "Bearer "+t)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("AddPage: POST /documents/%s/pages failed: %v", id, err)
		c.String(http.StatusBadGateway, "Failed to reach API")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("AddPage: POST /documents/%s/pages returned %d", id, resp.StatusCode)
		c.String(resp.StatusCode, "Failed to add page")
		return
	}

	// Re-fetch the document to get updated pages
	docReq, _ := http.NewRequest("GET", apiBase()+"/documents/"+id, nil)
	if t, ok := token.(string); ok {
		docReq.Header.Set("Authorization", "Bearer "+t)
	}
	docResp, err := http.DefaultClient.Do(docReq)
	if err != nil {
		log.Printf("AddPage: GET /documents/%s failed: %v", id, err)
		c.String(http.StatusBadGateway, "Failed to reach API")
		return
	}
	defer docResp.Body.Close()

	var doc templates.Document
	if err := json.NewDecoder(docResp.Body).Decode(&doc); err != nil {
		log.Printf("AddPage: failed to decode response for document %s: %v", id, err)
		c.String(http.StatusInternalServerError, "Failed to parse document")
		return
	}

	c.Header("Content-Type", "text/html")
	templates.PagesPartial(doc).Render(c.Request.Context(), c.Writer)
}

func DocumentsList(c *gin.Context) {
	req, _ := http.NewRequest("GET", apiBase()+"/documents", nil)
	token, _ := c.Get("token")
	if t, ok := token.(string); ok {
		req.Header.Set("Authorization", "Bearer "+t)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("DocumentsList: GET /documents failed: %v", err)
		c.String(http.StatusBadGateway, "Failed to reach API")
		return
	}
	defer resp.Body.Close()

	var docs []templates.Document
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		log.Printf("DocumentsList: failed to decode response: %v", err)
		c.String(http.StatusInternalServerError, "Failed to parse documents")
		return
	}

	c.Header("Content-Type", "text/html")
	templates.DocumentList(docs).Render(c.Request.Context(), c.Writer)
}
