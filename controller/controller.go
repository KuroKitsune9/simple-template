package controller

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	// "github.com/go-playground/validator/v10/translations/id"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"todos/helpers"
	"todos/model"
)

type Claims struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	jwt.StandardClaims
}

type BulkDelete struct {
	ID []int64 `json:"id"`
}

func GetAllTaskController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var users []model.TaskRes
		claims := helpers.ClaimToken(c)
		id := claims.ID
		fmt.Println(id)

		query := `SELECT tasks.id, tasks.title, tasks.description, tasks.status, tasks.date, tasks.image, tasks.created_at, tasks.updated_at, tasks.id_user, tasks.category_id, category.name_category, tasks.important
		FROM tasks
		LEFT JOIN category
		ON tasks.category_id = category.id
		WHERE tasks.id_user = $1
		ORDER BY tasks.important ASC`

		rows, err := db.Query(query, id)
		if err != nil {
			return err
		}
		for rows.Next() {
			var user model.TaskRes
			err = rows.Scan(
				&user.Id,
				&user.Title,
				&user.Description,
				&user.Status,
				&user.Date,
				&user.Image,
				&user.CreatedAt,
				&user.UpdatedAt,
				&user.IdUser,
				&user.CategoryId,
				&user.CategoryName,
				&user.Important,
			)
			if err != nil {
				return err
			}
			users = append(users, user)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Succes GET data",
			"data":    users,
		})
	}
}

func GetTaskById(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var user model.TaskRes
		claims := helpers.ClaimToken(c)
		id := claims.ID
		taskId := c.Param("id")

		query := `SELECT id, title, description, status, date, image, created_at, updated_at, id_user FROM tasks WHERE id_user = $1 AND id = $2`

		rows, err := db.Query(query, id, taskId)
		if err != nil {
			return err
		}

		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(
				&user.Id,
				&user.Title,
				&user.Description,
				&user.Status,
				&user.Date,
				&user.Image,
				&user.CreatedAt,
				&user.UpdatedAt,
				&user.IdUser,
			)
			if err != nil {
				return err
			}
		} else {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": "Task Not Found",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Message": "Success see Task Detail",
			"data":    user,
		})
	}
}

func AddTaskController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.TaskReq
		var user model.TaskRes // change
		err := c.Bind(&req)
		if err != nil {
			return err
		}

		// Menerima file gambar dari form dengan nama "image"
		image, err := c.FormFile("image")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Tidak dapat memproses file gambar"})
		}

		// Buka file yang diunggah
		src, err := image.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal membuka file gambar"})
		}
		defer src.Close()

		// Lokasi penyimpanan file gambar lokal
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal membuat direktori penyimpanan"})
		}

		// Generate nama file unik
		dstPath := filepath.Join(uploadDir, image.Filename)

		// Membuka file tujuan untuk penyimpanan
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal membuat file gambar"})
		}
		defer dst.Close()

		// Salin isi file dari file asal ke file tujuan
		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyalin file gambar"})
		}

		// Membuat URL ke gambar yang diunggah
		imageURL := "http://localhost:8080/uploads/" + image.Filename

		layout := "2006-01-02 15:04"
		parsedDate, err := time.Parse(layout, req.Date)
		if err != nil {
			return err
		}

		Claims := helpers.ClaimToken(c)
		id := Claims.ID

		validate := validator.New()
		err = validate.Struct(req)
		if err != nil {
			var errorMessage []string
			validationErrors := err.(validator.ValidationErrors)
			for _, err := range validationErrors {
				errorMessage = append(errorMessage, err.Error())
			}
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": errorMessage,
			})
		}

		query := `
		INSERT INTO tasks (title, description, status, date, image, created_at, id_user, category_id, important)
		VALUES ($1, $2, $3, $4, $5, now(), $6, $7, $8)  
		RETURNING id, title, description, status, date, image, created_at, updated_at, id_user, category_id, important
		`
		row := db.QueryRowx(query, req.Title, req.Description, req.Status, parsedDate, imageURL, id, req.CategoryId, req.Important)
		err = row.Scan(&user.Id, &user.Title, &user.Description, &user.Status, &user.Date, &user.Image, &user.CreatedAt, &user.UpdatedAt, &user.IdUser, &user.CategoryId, &user.Important)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success add task",
			"data":    user,
		})
	}
}

func EditTaskController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.TaskReq
		var user model.TaskRes
		err := c.Bind(&req)
		if err != nil {
			return err
		}

		taskID := c.Param("id")

		Claims := helpers.ClaimToken(c)
		id := Claims.ID

		layout := "2006-01-02 15:04"
		parseDate, err := time.Parse(layout, req.Date)
		if err != nil {
			return err
		}

		query := `UPDATE tasks SET title = $1, description = $2, status = $3, date = $4, updated_at = now() 
		WHERE id = $5 AND id_user = $6 
		RETURNING id, title, description, status, date, image, created_at, updated_at, id_user`

		rows := db.QueryRowx(query, req.Title, req.Description, req.Status, parseDate, taskID, id)
		err = rows.Scan(
			&user.Id,
			&user.Title,
			&user.Description,
			&user.Status,
			&user.Date,
			&user.Image,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.IdUser,
		)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "sucessfully edited",
		})
	}
}

func DeleteTaskControll(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		taskId := c.Param("id")
		claims := helpers.ClaimToken(c)
		id := claims.ID

		query := `DELETE FROM tasks WHERE id = $1 AND id_user = $2`
		_, err := db.Exec(query, taskId, id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": "Task not found",
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Task sucessfully deleted",
		})
	}
}

func BulkDeleteTask(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req BulkDelete
		err := c.Bind(&req)
		if err != nil {
			return err
		}
		claims := helpers.ClaimToken(c)
		id := claims.ID

		for _, taskID := range req.ID {
			query := `DELETE FROM tasks WHERE id = $1 AND id_user = $2`
			_, err = db.Exec(query, taskID, id)
			if err != nil {
				return err
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "multiple task successfully deleted",
		})

	}
}

func SearchTask(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var users []model.TaskRes
		Claims := helpers.ClaimToken(c)
		id := Claims.ID
		keywoard := c.QueryParam("search")
		date := c.QueryParam("date")
		HitungPage := c.QueryParam("page")
		HitungLimit := c.QueryParam("limit")
		var parsedDate time.Time
		if date != "" {
			layout := "2006-01-02"
			parsedDate, err = time.Parse(layout, date)
			if err != nil {
				return err
			}
		}

		query := `SELECT id, title, description, status, date, image, created_at, updated_at, id_user FROM tasks WHERE id_user = $1`
		keywoard = "%" + keywoard + "%"

		if !parsedDate.IsZero() {
			query += fmt.Sprintf(" AND date::date = '%s'", parsedDate.Format("2006-01-02"))
		}

		if keywoard != "" {
			query += fmt.Sprintf(" AND (title ILIKE '%s' OR description ILIKE '%s')", keywoard, keywoard)
		}

		page, err := strconv.Atoi(HitungPage)
		if err != nil {
			page = 1
		}

		limit, err := strconv.Atoi(HitungLimit)
		if err != nil {
			limit = 10
		}

		offset := (page - 1) * limit

		totalQuery := `SELECT COUNT(*) FROM tasks WHERE id_user = $1`

		if !parsedDate.IsZero() {
			totalQuery += fmt.Sprintf(" AND date::date = '%s'", parsedDate.Format("2006-01-02"))
		}

		if keywoard != "" {
			totalQuery += fmt.Sprintf(" AND (title ILIKE '%s' OR description ILIKE '%s')", keywoard, keywoard)
		}

		var count int
		err = db.Get(&count, totalQuery, id)
		if err != nil {
			return err
		}

		totalPages := count / limit
		if count%limit != 0 {
			totalPages++
		}

		query += " LIMIT $2 OFFSET $3"

		rows, err := db.Query(query, id, limit, offset)
		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var user model.TaskRes
			err = rows.Scan(
				&user.Id,
				&user.Title,
				&user.Description,
				&user.Status,
				&user.Date,
				&user.Image,
				&user.CreatedAt,
				&user.UpdatedAt,
				&user.IdUser,
			)
			if err != nil {
				return err
			}
			users = append(users, user)
		}

		if len(users) == 0 {
			users = []model.TaskRes{}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":     "Pencarian berhasil",
			"data":        users,
			"count":       count,
			"page":        page,
			"limit_page":  limit,
			"total_data":  count,
			"total_pages": totalPages,
		})
	}
}

func CountStatus(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		Claims := helpers.ClaimToken(c)
		id := Claims.ID

		query := `
			SELECT 
				SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending,
				SUM(CASE WHEN status = 'progress' THEN 1 ELSE 0 END) AS progress,
				SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END) AS done
			FROM tasks 
			WHERE id_user = $1
		`

		counts := struct {
			Pending  int `json:"pending"`
			Progress int `json:"progress"`
			Done     int `json:"done"`
		}{}
		err := db.Get(&counts, query, id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "count succed",
			"data":    counts,
		})
	}
}

func RegisterController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.RegisRequest
		var user model.RegisReponse
		validate := validator.New()

		err := c.Bind(&req)
		if err != nil {
			return err
		}
		err = validate.Struct(req)
		if err != nil {
			var errorMessage []string
			validationErrors := err.(validator.ValidationErrors)
			for _, err := range validationErrors {
				errorMessage = append(errorMessage, err.Error())
			}
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": errorMessage,
			})
		}

		password, err := helpers.HashPassword(req.Password)
		if err != nil {
			return err
		}

		query := `
		INSERT INTO users ( email,  created_at, password)
		VALUES ($1,  now(), $2)  
		RETURNING id, email, created_at
		`
		row := db.QueryRowx(query, req.Email, password)
		err = row.Scan(&user.Id, &user.Email, &user.CreatedAt)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success add user",
			"data":    user,
		})
	}
}

func LoginController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.LoginRequest
		var user model.UserModel
		var err error
		validate := validator.New()

		err = c.Bind(&req)
		if err != nil {
			return err
		}
		err = validate.Struct(req)
		if err != nil {
			var errorMessage []string
			validationErrors := err.(validator.ValidationErrors)
			for _, err := range validationErrors {
				errorMessage = append(errorMessage, err.Error())
			}
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": errorMessage,
			})
		}
		query := `SELECT id, email, created_at, updated_at, password FROM users WHERE email = $1`
		row := db.QueryRowx(query, req.Email)

		err = row.Scan(&user.Id, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Password)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "Email not found",
				})
			}
			return err
		}

		match, err := helpers.ComparePassword(user.Password, req.Password)
		if err != nil {
			if !match {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "password incorrect",
				})
			}
			return err
		}

		var (
			jwtToken  *jwt.Token
			secretKey = []byte("secret")
		)

		jwtClaims := &Claims{
			ID:    user.Id,
			Email: user.Email,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			},
		}

		jwtToken = jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

		token, err := jwtToken.SignedString(secretKey)
		if err != nil {
			return err
		}

		const query2 = `INSERT INTO user_token (user_id, token) VALUES ($1, $2)`
		_ = db.QueryRowx(query2, user.Id, token)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"token":   token,
			"message": "Login success",
			"data":    user,
		})
	}
}

func LogoutController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var reqToken string
		headerDataToken := c.Request().Header.Get("Authorization")

		splitToken := strings.Split(headerDataToken, "Bearer ")
		if len(splitToken) > 1 {
			reqToken = splitToken[1]
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		query := "DELETE FROM user_token WHERE token = $1"

		_, err := db.Exec(query, reqToken)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]interface{}{
					"message": "Data pengguna tidak ditemukan",
				})
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "successfully logout",
		})
	}
}
