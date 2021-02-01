package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) createContract(c *gin.Context) {
	if !s.tmutex.TryLock() {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no concurrent allowed"})
		return
	}
	defer s.tmutex.Unlock()

	var input CreateContractInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	output := s.handleCreateContract(input)

	c.JSON(http.StatusOK, output)
}

func (s *Server) callContract(c *gin.Context) {
	if !s.tmutex.TryLock() {
		c.JSON(http.StatusOK, "no concurrent allowed")
		return
	}
	defer s.tmutex.Unlock()

	var input CallContractInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	output := s.handleCallContract(input)

	c.JSON(http.StatusOK, output)
}
