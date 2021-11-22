package handler

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gorilla/mux"
)

type Category struct {
	ID int `db:"id"`
	Name string `db:"name"`
	Status bool `db:"status"`
}

type FormCategory struct {
	Cat Category
	Errors map[string]string
}

type ListCategory struct {
	Categories []Category
}

func (c *Category) Validate() error {
	return validation.ValidateStruct(c, validation.Field(
		&c.Name, validation.Required.Error("This field is must be required"),
		validation.Length(3,0).Error("This field is must be grater than 3")))
}

func (h *Handler) createCategories(rw http.ResponseWriter, r *http.Request) {
	vErrs := map[string]string{}
	cat := Category{}
	h.loadCreateCategoryForm(rw, cat, vErrs)
}

func (h *Handler) storeCategories(rw http.ResponseWriter, r *http.Request) {
	if err:= r.ParseForm(); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	var category Category
	if err := h.decoder.Decode(&category, r.PostForm); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := category.Validate(); err != nil {
		vErrors, ok := err.(validation.Errors)
		if ok {
			vErrs := make(map[string]string)
			for key, value := range vErrors {
				vErrs[key] = value.Error()
			}
			h.loadCreateCategoryForm(rw, category, vErrs)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	
	const insertCategory = `INSERT INTO categories(name,status) VALUES($1,$2)`
	res:= h.db.MustExec(insertCategory, category.Name, category.Status )

	if ok, err:= res.RowsAffected(); err != nil || ok == 0 {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(rw, r, "/category/list", http.StatusTemporaryRedirect)
}

func (h *Handler) listCategories(rw http.ResponseWriter, r *http.Request) {
	category := []Category{}
	h.db.Select(&category, "SELECT * FROM categories")
	list := ListCategory{
		Categories: category,
	}
	if err:= h.templates.ExecuteTemplate(rw, "list-category.html", list); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) editCategories(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(rw, "invalid URL", http.StatusInternalServerError)
		return
	}
	const getCategory = `SELECT * FROM categories WHERE id=$1`
	var category Category
	h.db.Get(&category, getCategory, id)
	
	if category.ID == 0 {
		http.Error(rw, "invalid URL", http.StatusInternalServerError)
		return
	}
	h.loadEditCategoryForm(rw, category, map[string]string{})
}

func (h *Handler) updateCategories(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(rw, "invalid URL", http.StatusInternalServerError)
		return
	}

	const getCategory = `SELECT * FROM categories WHERE id = $1`
	var category Category
	h.db.Get(&category, getCategory, id)

	if category.ID == 0 {
		http.Error(rw, "invalid URL", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(rw, "invalid URL", http.StatusInternalServerError)
		return
	}

	if err := h.decoder.Decode(&category, r.PostForm); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := category.Validate(); err != nil {
		vErrors, ok := err.(validation.Errors)
		if ok {
			vErrs := make(map[string]string)
			for key, value := range vErrors {
				vErrs[key] = value.Error()
			}
			h.loadEditCategoryForm(rw, category, vErrs)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	const updateCategories = `UPDATE categories SET name = $2, status = $3 WHERE id = $1`
	res:= h.db.MustExec(updateCategories, id, category.Name, category.Status )
	if ok, err := res.RowsAffected(); err != nil || ok == 0 {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(rw, r, "/category/list", http.StatusTemporaryRedirect)
}

func (h *Handler) deleteCategories(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(rw, "invalid URL", http.StatusInternalServerError)
		return
	}

	const getCategory = `SELECT * FROM categories WHERE id = $1`
	var category Category
	h.db.Get(&category, getCategory, id)

	if category.ID == 0 {
		http.Error(rw, "invalid URL", http.StatusInternalServerError)
		return
	}

	const deleteCategories = `DELETE FROM categories WHERE id = $1`
	res:= h.db.MustExec(deleteCategories, id)
	if ok, err:= res.RowsAffected(); err != nil || ok == 0 {
		http.Error(rw, "invalid URL", http.StatusInternalServerError)
		return
	}
	http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
}

func (h *Handler) loadCreateCategoryForm(rw http.ResponseWriter, cat Category, errs map[string]string) {
	form := FormCategory{
		Cat : cat,
		Errors : errs,
	}
	if err:= h.templates.ExecuteTemplate(rw, "create-category.html", form); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) loadEditCategoryForm(rw http.ResponseWriter, cat Category, errs map[string]string) {
	form := FormCategory{
		Cat : cat,
		Errors : errs,
	}
	if err:= h.templates.ExecuteTemplate(rw, "edit-category.html", form); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) searchCategory(rw http.ResponseWriter, r *http.Request) {
	if err:= r.ParseForm(); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	cat := r.FormValue("search")
	const getSearch = "SELECT * FROM categories WHERE name ILIKE '%%' || $1 || '%%'" 
	category := []Category{}
	h.db.Select(&category, getSearch, cat)
	list := ListCategory{
		Categories: category,
	}
	if err:= h.templates.ExecuteTemplate(rw, "list-category.html", list); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}