package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/Unknwon/macaron"
	"github.com/uget/uget/core"
)

// Server listens for HTTP requests that manipulate files
type Server struct {
	BindAddr  string    `json:"bind_address,omitempty"`
	Port      uint16    `json:"port"`
	StartedAt time.Time `json:"started_at"`
}

var downloader = core.NewClient()

type macaronLog struct{}

func (w macaronLog) Write(p []byte) (int, error) {
	logrus.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

// Run starts the server
func (s *Server) Run() {
	m := macaron.NewWithLogger(macaronLog{})
	m.Use(macaron.Renderer())
	// JSON API
	m.Group("", func() {
		m.Get("/serverinfo", wrapJSON(s))
		m.Group("/containers", func() {
			m.Post("", s.createContainer)
			m.Get("", s.listContainers)
			m.Get("/:id", s.showContainer)
			m.Delete("/:id", s.deleteContainer)
		})
	})
	// CLICK'N'LOAD v2
	cnl(m)
	s.StartedAt = time.Now().Round(time.Minute)
	m.Run(s.BindAddr, int(s.Port))
}

func addLinks(links []string) {
	logrus.Debugf("Added %v links!", len(links))
}

func (s *Server) createContainer(c *macaron.Context) {
	var container struct {
		string `json:"p"`
	}
	decoder := json.NewDecoder(c.Req.Body().ReadCloser())
	if decoder.Decode(&container) != nil {
		c.Render.Error(http.StatusNotFound, "Invalid JSON.")
	}
	c.Render.RawData(http.StatusOK, []byte("okay!"))
}

func (s *Server) listContainers(c *macaron.Context) {

}

func (s *Server) showContainer(c *macaron.Context) {

}

func (s *Server) deleteContainer(c *macaron.Context) {
	fmt.Printf("Deleting %s\n", c.Params("id"))
	c.Status(http.StatusNoContent)
}

func as(ctype string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", ctype))
	}
}

func wrap(v interface{}) func() interface{} {
	return func() interface{} {
		return v
	}
}

// Wraps a static value in a function block
// This is a convenience method to use with macaron
func wrapJSON(v interface{}) func(*macaron.Context) {
	return func(ctx *macaron.Context) {
		ctx.JSON(http.StatusOK, v)
	}
}
