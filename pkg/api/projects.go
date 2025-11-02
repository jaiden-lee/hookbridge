package api

// REQUEST BODIES
type CreateProjectRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangeProjectRequest struct {
	Password string `json:"password" binding:"required"`
}

// RESPONSE BODIES
type ProjectResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
