package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) deploy(c *gin.Context) {
	if !s.tmutex.TryLock() {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no concurrent allowed"})
		return
	}
	defer s.tmutex.Unlock()

	var input DeployInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	output := s.handleDeploy(input)

	c.JSON(http.StatusOK, output)
}

func (s *Server) call(c *gin.Context) {
	if !s.tmutex.TryLock() {
		c.JSON(http.StatusOK, "no concurrent allowed")
		return
	}
	defer s.tmutex.Unlock()

	var input CallInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	output := s.handleCall(input)

	c.JSON(http.StatusOK, output)
}
