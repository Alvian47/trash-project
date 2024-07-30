package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var dbpools *pgxpool.Pool

type Article struct {
	Id             int       `json:"id"`
	Name           string    `json:"name"`
	Content        string    `json:"content"`
	Tags           string    `json:"tags"`
	Published_date time.Time `json:"published_date"` // gunakan method const date = new Date() const rfc = date.toISOString(); karena golang menggunakan rfc 3339
}

func loadENV(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	return os.Getenv(key)
}

func dbPgx(ctx context.Context, dbUrl string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool %v\n", err)
		os.Exit(1)
	}
	if err := dbpool.Ping(ctx); err != nil {
		panic(err)
	}
	return dbpool
}

func main() {
	app := fiber.New()
	ctx := context.Background()

	db := dbPgx(ctx, loadENV("DATABASE_URL"))
	defer db.Close()
	dbpools = db

	app.Get("/api/all-article/", AllArticle) // /api/all-article/?tags=healt&dates=2024-07-13T06:11:04.435+07:00 and dates must be encoded first
	app.Get("/api/article/:id", singleArticle)
	app.Post("/api/add-article", AddArticle)
	app.Delete("/api/del-article/:id", DeleteArticle)
	app.Put("/api/put-article/:id", UpdateArticle)
	app.Patch("/api/patch-article/:id", UpdateContentTagsArticle)

	app.Listen(":3000")
}

func AllArticle(c *fiber.Ctx) error {
	articles := []Article{}
	tags := c.Query("tags")
	dates := c.Query("dates")

	// Helper function to handle row scanning and error checking
	scanRows := func(rows pgx.Rows) error {
		for rows.Next() {
			article := Article{}
			err := rows.Scan(&article.Id, &article.Name, &article.Content, &article.Tags, &article.Published_date)
			if err != nil {
				return err
			}
			articles = append(articles, article)
		}
		if rows.Err() != nil {
			return rows.Err()
		}
		return nil
	}

	var sqlStatement string
	var rows pgx.Rows
	var err error

	// tags and dates available
	if tags != "" && dates != "" {
		sqlStatement = `SELECT * FROM article WHERE tags = $1 AND published_at = $2`
		rows, err = dbpools.Query(c.Context(), sqlStatement, tags, dates)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "failed",
				"err": err.Error(),
			})
		}
		err = scanRows(rows)
	} else if tags != "" {
		sqlStatement = `SELECT * FROM article WHERE tags = $1`
		rows, err = dbpools.Query(c.Context(), sqlStatement, tags)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "failed",
				"err": err.Error(),
			})
		}
		err = scanRows(rows)
	} else if dates != "" {
		sqlStatement = `SELECT * FROM article WHERE published_at = $1`
		rows, err = dbpools.Query(c.Context(), sqlStatement, dates)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "failed",
				"err": err.Error(),
			})
		}
		err = scanRows(rows)
	} else {
		sqlStatement = `SELECT * FROM article`
		rows, err = dbpools.Query(c.Context(), sqlStatement)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "failed",
				"err": err.Error(),
			})
		}
		err = scanRows(rows)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "success",
		"data": articles,
	})
}

func singleArticle(c *fiber.Ctx) error {
	article := Article{}
	id := c.Params("id")

	sqlStatement := `SELECT * FROM article WHERE id = $1`
	row := dbpools.QueryRow(c.Context(), sqlStatement, id)
	err := row.Scan(&article.Id, &article.Name, &article.Content, &article.Tags, &article.Published_date)

	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "success",
		"data": article,
	})

}

func AddArticle(c *fiber.Ctx) error {
	article := Article{}

	err := c.BodyParser(&article)
	if err == fiber.ErrUnprocessableEntity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	SqlStatement := `INSERT INTO article (name, content, tags, published_at) VALUES($1, $2, $3, $4)`
	tag, err := dbpools.Exec(c.Context(), SqlStatement, article.Name, article.Content, article.Tags, article.Published_date)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": errors.New("failed to add article"),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"msg":  "success",
		"data": "",
	})
}

func DeleteArticle(c *fiber.Ctx) error {
	id := c.Params("id")

	sqlStatement := `DELETE FROM article WHERE id = $1`
	tag, err := dbpools.Exec(c.Context(), sqlStatement, id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": errors.New("article tidak ditemukan").Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "success",
		"data": "",
	})
}

func UpdateArticle(c *fiber.Ctx) error {
	id := c.Params("id")
	article := Article{}

	err := c.BodyParser(&article)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	sqlStatement := `UPDATE article SET name = $1, content = $2, tags = $3 WHERE id = $4`
	_, err = dbpools.Exec(c.Context(), sqlStatement, &article.Name, &article.Content, &article.Tags, id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "success",
		"data": "",
	})
}
func UpdateContentTagsArticle(c *fiber.Ctx) error {
	id := c.Params("id")
	article := Article{}

	err := c.BodyParser(&article)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	sqlStatement := `UPDATE article SET content = $1, tags = $2 WHERE id = $3`
	_, err = dbpools.Exec(c.Context(), sqlStatement, &article.Content, &article.Tags, id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "success",
		"data": "",
	})
}
