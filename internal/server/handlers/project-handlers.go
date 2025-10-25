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
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
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

type ProjectResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *projectHandlersStruct) GetProjectsHandler(c *gin.Context) {
	user, err := getUserFromContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	projects, err := db.ProjectService.GetProjectsByUser(user.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	projectsResponse := toProjectResponseList(projects)

	c.JSON(http.StatusOK, gin.H{
		"projects": projectsResponse,
	})
}

type ChangeProjectRequest struct {
	ID       string `json:"id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *projectHandlersStruct) ChangeProjectPasswordHandler(c *gin.Context) {
	user, err := getUserFromContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var request ChangeProjectRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = db.ProjectService.ChangeProjectPassword(request.ID, user.UserID, request.Password)
	if err != nil {
		if errors.Is(err, db.ErrProjectNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		} else if errors.Is(err, db.ErrPasswordSpecialCharacters) ||
			errors.Is(err, db.ErrPasswordTooLong) ||
			errors.Is(err, db.ErrPasswordTooShort) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "project password changed successfully",
	})
}

type DeleteProjectRequest struct {
	ID string `json:"id" binding:"required"`
}

func (s *projectHandlersStruct) DeleteProjectHandler(c *gin.Context) {
	user, err := getUserFromContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var request DeleteProjectRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = db.ProjectService.DeleteProject(request.ID, user.UserID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrProjectNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "project successfully deleted",
	})
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

// ToResponseList converts a slice of Projects
func toProjectResponseList(projects []db.Project) []ProjectResponse {
	res := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		res[i] = ProjectResponse{
			ID:   p.ID,
			Name: p.Name,
		}
	}
	return res
}
