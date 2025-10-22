package handlers

import (
	"github.com/gin-gonic/gin"
)

/*
All clients must include Authorization Header:
{
	"Authorization": "Bearer [access_token]"
}
*/

type projectHandlersStruct struct{}

var ProjectHandlers = &projectHandlersStruct{}

func (s *projectHandlersStruct) CreateProjectHandler(c *gin.Context) {

}

func (s *projectHandlersStruct) GetProjectsHandler(c *gin.Context) {

}

func (s *projectHandlersStruct) ChangeProjectPasswordHandler(c *gin.Context) {

}

func (s *projectHandlersStruct) DeleteProjectHandler(c *gin.Context) {

}
