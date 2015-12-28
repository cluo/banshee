// Copyright 2015 Eleme Inc. All rights reserved.

package webapp

import (
	"github.com/eleme/banshee/models"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	"github.com/mattn/go-sqlite3"
	"net/http"
	"strconv"
)

// getProject returns project by id.
func getProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Query db.
	proj := &models.Project{}
	if err := db.Admin.DB().First(proj, id).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	ResponseJSON(w, proj)
}

// createProject request
type createProjectRequest struct {
	Name string `json:"name"`
}

// createProject creates a project.
func createProject(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Request
	req := &createProjectRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Validate
	if len(req.Name) <= 0 {
		ResponseError(w, ErrProjectName)
		return
	}
	// Save.
	if err := db.Admin.DB().Create(&models.Project{Name: req.Name}).Error; err != nil {
		switch err {
		case sqlite3.ErrConstraintNotNull:
			ResponseError(w, ErrNotNull)
			return
		case sqlite3.ErrConstraintPrimaryKey:
			ResponseError(w, ErrPrimaryKey)
			return
		case sqlite3.ErrConstraintUnique:
			ResponseError(w, ErrDuplicateProjectName)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
}

// updateProject request
type updateProjectRequest struct {
	Name string `json:"name"`
}

// updateProject updates a project.
func updateProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Request
	req := &updateProjectRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Find
	proj := &models.Project{}
	if err := db.Admin.DB().First(proj, id).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Patch.
	proj.Name = req.Name
	if err := db.Admin.DB().Save(proj).Error; err != nil {
		switch err {
		case sqlite3.ErrConstraintNotNull:
			ResponseError(w, ErrNotNull)
			return
		case sqlite3.ErrConstraintUnique:
			ResponseError(w, ErrDuplicateProjectName)
			return
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
}

// deleteProject deletes a project.
func deleteProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Delete.
	if err := db.Admin.DB().Delete(&models.Project{ID: id}).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
}

// getProjectRules gets project rules.
func getProjectRules(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Query
	var rules []models.Rule
	if err := db.Admin.DB().Model(&models.Project{ID: id}).Related(&rules).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	ResponseJSON(w, rules)
}

// getProjectUsers gets project users.
func getProjectUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Query
	var users []models.User
	if err := db.Admin.DB().Model(&models.Project{ID: id}).Association("Users").Find(&users).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	// Universals
	var univs []models.User
	if err := db.Admin.DB().Where("universal = ?", true).Find(&univs).Error; err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	users = append(users, univs...)
	ResponseJSON(w, users)
}

// addUser adds a user to a project by name.
type addProjectUserRequest struct {
	Name string `json:name`
}

func addProjectUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	// Request
	req := &addProjectUserRequest{}
	if err := RequestBind(r, req); err != nil {
		ResponseError(w, ErrBadRequest)
		return
	}
	// Find user.
	user := &models.User{}
	if err := db.Admin.DB().Where("name = ?", req.Name).First(user).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrUserNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Find proj
	proj := &models.Project{}
	if err := db.Admin.DB().First(proj, id).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrProjectNotFound)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
	// Append user.
	if err := db.Admin.DB().Model(proj).Association("Users").Append(user).Error; err != nil {
		switch err {
		case gorm.RecordNotFound:
			ResponseError(w, ErrNotFound)
			return
		case sqlite3.ErrConstraintPrimaryKey:
			ResponseError(w, ErrPrimaryKey)
			return
		default:
			ResponseError(w, NewUnexceptedWebError(err))
			return
		}
	}
}