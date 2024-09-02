package controller

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"todos/model"

)

func GetAllCategory(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var categorys []model.CategoryRes

		query := `SELECT id, name_category, created_at, updated_at FROM category`

		rows, err := db.Query(query)
		if err != nil {
			return err
		}

		for rows.Next() {
			var category model.CategoryRes
			err = rows.Scan(
				&category.Id,
				&category.CategoryName,
				&category.CreatedAt,
				&category.UpdatedAt,
			)
			if err != nil {
				return err
			}
			categorys = append(categorys, category)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success get data",
			"data":    categorys,
		})
	}
}

func GETcategoryById(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var category model.CategoryRes
		Id := c.Param("id")

		query := `SELECT id, name_category, created_at, updated_at FROM category WHERE id = $1`

		rows, err := db.Query(query, Id)
		if err != nil {
			return err
		}

		if rows.Next() {
			err = rows.Scan(
				&category.Id,
				&category.CategoryName,
				&category.CreatedAt,
				&category.UpdatedAt,
			)
			if err != nil {
				return err
			}
		} else {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "category not found",
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "category found",
			"data":    category,
		})
	}
}

func AddCategoryController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var categorys model.CategoryRes
		var req model.CategoryReq
		err := c.Bind(&req)
		if err != nil {
			return err
		}

		query := `INSERT INTO category (name_category, created_at) VALUES ($1, now()) RETURNING id, name_category, created_at, updated_at`

		rows := db.QueryRowx(query, req.CategoryName)

		err = rows.Scan(
			&categorys.Id,
			&categorys.CategoryName,
			&categorys.CreatedAt,
			&categorys.UpdatedAt,
		)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "data updated",
			"data":    categorys,
		})
	}
}

func UpdateCategoryController(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.CategoryReq
		var category model.CategoryRes
		Id := c.Param("id")
		err := c.Bind(&req)
		if err != nil {
			return err
		}

		query := `UPDATE category SET name_category = $1, updated_at = now() WHERE id = $2 RETURNING id, name_category, created_at, updated_at`

		rows := db.QueryRowx(query, req.CategoryName, Id)
		err = rows.Scan(
			&category.Id,
			&category.CategoryName,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "sucessfully updated",
		})
	}
}

func DeleteCategorycontroller(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.CategoryReq
		Id := c.Param("id")
		err := c.Bind(&req)
		if err != nil {
			return err
		}

		query := `DELETE FROM category WHERE id = $1`

		_, err = db.Exec(query, Id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Category not found",
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Category successfully deleted",
		})
	}

}

func BulkDeletecategory(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.CategoryReq
		err := c.Bind(&req)
		if err != nil {
			return err
		}

		for _, Id := range req.Id {
			query := `DELET FROM category WHERE id = $1`
			_, err = db.Exec(query, Id)
			if err != nil {
				return err
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Multiple item deleted successfully",
		})
	}
}
