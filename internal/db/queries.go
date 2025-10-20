package db

import (
	"errors"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetProjectByID(projectID string) (*Project, error) {
	database := GetDB()

	var project Project
	err := database.First(&project).Where("id = ?", projectID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return &project, nil
}

func GetProjectsByUser(userID string) ([]Project, error) {
	database := GetDB()

	var projects []Project
	err := database.Find(&projects).Where("user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	// can return nil slices; Go feature; technically not a pointer, just the struct is all zero'd out
	// you can actually append to nil slices; nil slice is just an empty slice
	// where all fields are zero

	return projects, nil
}

func GetProjectByIDAndPassword(projectID string, password string) (*Project, error) {
	project, err := GetProjectByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err // project id doesn't exist
	}

	passwordMatch := checkPasswordHash(password, project.Password)
	if !passwordMatch {
		return nil, ErrInvalidPassword
	}

	return project, nil
}

func CreateProject(userID string, name string, password string) (*Project, error) {
	database := GetDB()

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, ErrPasswordTooLong
	}

	// 10 character nanoid
	projectId, err := gonanoid.New(10)
	if err != nil {
		return nil, err
	}

	project := Project{
		ID:       projectId,
		Name:     name,
		Password: hashedPassword,
		UserID:   userID,
	}

	err = database.Create(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrProjectAlreadyExists
		}
		return nil, err
	}

	return &project, nil
}

func ChangeProjectPassword(projectID string, userID string, newPassword string) error {
	database := GetDB()

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return ErrPasswordTooLong
	}

	result := database.Model(&Project{}).
		Where("id = ? AND user_id = ?", projectID, userID).
		Update("password", hashedPassword)

	if result.Error != nil {
		return result.Error
	}

	// just tell users either 1) projectID doesn't exist or 2) you are not owner
	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}

	return nil // if no error, then success
}

func DeleteProject(projectID string, userID string) error {
	database := GetDB()

	result := database.Where("user_id = ? AND user_id = ?", projectID, userID).
		Delete(&Project{})

	if result.Error != nil {
		return result.Error
	}

	// just tell users either 1) projectID doesn't exist or 2) you are not owner
	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}

	return nil // if no error, then success
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err // password longer than 72 bytes
	}

	return string(hash), nil
}

func checkPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
