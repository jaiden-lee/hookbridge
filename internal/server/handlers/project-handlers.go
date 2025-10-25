package handlers

import (
	"net/http"

	"log"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jaiden-lee/hookbridge/internal/db"
	"github.com/jaiden-lee/hookbridge/internal/server/utils"
)

/*
All clients must include Authorization Header:
{
	"Authorization": "Bearer [access_token]"
}
*/

type projectHandlersStruct struct{}

var ProjectHandlers = &projectHandlersStruct{}

type CreateProjectRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *projectHandlersStruct) CreateProjectHandler(c *gin.Context) {
	user, err := getUserFromContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var request CreateProjectRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	project, err := db.ProjectService.CreateProject(user.UserID, request.Name, request.Password)
	if err != nil {
		if errors.Is(err, db.ErrProjectAlreadyExists) {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("Project %s created with name %s\n", project.ID, project.Name)
	c.JSON(http.StatusCreated, gin.H{
		"message": "project created successfully",
	})
}

func (s *projectHandlersStruct) GetProjectsHandler(c *gin.Context) {

}

func (s *projectHandlersStruct) ChangeProjectPasswordHandler(c *gin.Context) {

}

func (s *projectHandlersStruct) DeleteProjectHandler(c *gin.Context) {

}

func getUserFromContext(c *gin.Context) (*utils.UserData, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, ErrInvalidUserInContext
	}

	userData, ok := user.(*utils.UserData)
	if !ok {
		return nil, ErrInvalidUserInContext
	}

	return userData, nil
}
