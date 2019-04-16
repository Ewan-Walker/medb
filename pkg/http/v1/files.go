package v1

import (
	"encoder-backend/pkg/http/utils"
	"encoder-backend/pkg/models"
	"encoding/json"
	"github.com/ewanwalk/gorm"
	"github.com/ewanwalk/respond"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func files(mux *mux.Router) {

	files := mux.PathPrefix("/files").Subrouter()
	files.HandleFunc("", getFiles).Methods("GET")
	files.HandleFunc("/{file}", getFile).Methods("GET")
	files.HandleFunc("/{file}", updateFile).Methods("PUT")
	files.HandleFunc("/{file}/encodes", getFileEncodes).Methods("GET")
	files.HandleFunc("/{file}/encodes/{encode}", getFileEncode).Methods("GET")
}

// getFiles
// obtain all files in the system, supports dynamic filtering
func getFiles(w http.ResponseWriter, r *http.Request) {

	params := utils.Vars(r)

	files := make([]models.File, 0)

	if err := db.Scopes(models.Dyanmic(models.File{}, params)).Find(&files).Error; err != nil {
		respond.With(w, r, http.StatusInternalServerError, err)
		return
	}

	respond.With(w, r, http.StatusOK, files)
}

// getFile
// get the provided file by id
func getFile(w http.ResponseWriter, r *http.Request) {

	params := utils.Vars(r)

	id, ok := params["file"]
	if !ok {
		respond.With(w, r, http.StatusBadRequest, errInvalidRequest)
		return
	}

	file := models.File{}

	if err := db.Preload("Path").First(&file, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respond.With(w, r, http.StatusNotFound, errNotFound)
			return
		}

		respond.With(w, r, http.StatusInternalServerError, err)
		return
	}

	respond.With(w, r, http.StatusOK, file)
}

// updateFile
// modify the provided file
//
// Example:
// {
//   "status_encode": 0
// }
func updateFile(w http.ResponseWriter, r *http.Request) {

	params := utils.Vars(r)

	id, ok := params["file"]
	if !ok {
		respond.With(w, r, http.StatusBadRequest, errInvalidRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err)
		return
	}

	update := make(map[string]interface{})

	if err := json.Unmarshal(body, &update); err != nil {
		respond.With(w, r, http.StatusInternalServerError, err)
		return
	}

	file := models.File{}

	valid := models.Columns(file)
	for key := range update {
		if _, ok := valid[key]; !ok {
			delete(update, key)
		}
	}

	// you cannot edit these fields
	for _, field := range []string{"encode", "path", "status", "directory", "name"} {
		delete(update, field)
	}

	if err := db.First(&file, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respond.With(w, r, http.StatusNotFound, errNotFound)
			return
		}

		respond.With(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := db.Model(&file).Updates(update).Error; err != nil {
		respond.With(w, r, http.StatusInternalServerError, err)
		return
	}

	respond.With(w, r, http.StatusOK, file)
}

// getFileEncodes
// obtain all the provided files encode history
func getFileEncodes(w http.ResponseWriter, r *http.Request) {

	params := utils.Vars(r)

	id, ok := params["file"]
	if !ok {
		respond.With(w, r, http.StatusBadRequest, errInvalidRequest)
		return
	}

	file := models.File{}

	if err := db.Preload("Encodes", models.Dyanmic(models.Encode{}, utils.QueryParams(r))).
		First(&file, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respond.With(w, r, http.StatusNotFound, errNotFound)
			return
		}

		respond.With(w, r, http.StatusInternalServerError, err)
		return
	}

	respond.With(w, r, http.StatusOK, file.Encodes)
}

// getFileEncode
// obtain the provided encode for the provided file
func getFileEncode(w http.ResponseWriter, r *http.Request) {

	params := utils.Vars(r)

	id, ok := params["file"]
	if !ok {
		respond.With(w, r, http.StatusBadRequest, errInvalidRequest)
		return
	}

	eId, ok := params["encode"]
	if !ok {
		respond.With(w, r, http.StatusBadRequest, errInvalidRequest)
		return
	}

	encode := models.Encode{}

	if err := db.Where("file_id = ?", id).First(&encode, eId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respond.With(w, r, http.StatusNotFound, errNotFound)
			return
		}

		respond.With(w, r, http.StatusInternalServerError, err)
		return
	}

	respond.With(w, r, http.StatusOK, encode)
}
