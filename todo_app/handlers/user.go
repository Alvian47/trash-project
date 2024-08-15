package handler

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5"

	"todo/databases"
	"todo/models"
	"todo/validation"
)

var Session = session.New()

func RegisterUser(c *fiber.Ctx) error {
	user := models.User{}

	err := c.BodyParser(&user)
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

	// buat validasi data dari request user
	err = validation.Validation.Struct(&user)
	if err != nil {
		fieldErrs := []string{}
		for _, fieldErr := range err.(validator.ValidationErrors) {
			fieldErrs = append(fieldErrs, fieldErr.Field())
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": fieldErrs,
		})
	}

	sqlString := `INSERT INTO users(email, username, password) VALUES($1, $2, $3)`
	_, err = databases.DBpools.Exec(c.Context(), sqlString, user.Email, user.Username, user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"msg":  "success",
		"data": "",
	})
}

func LoginUser(c *fiber.Ctx) error {
	user := models.UserLogin{}

	err := c.BodyParser(&user)
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

	err = validation.Validation.Struct(&user)
	if err != nil {
		fieldErrs := []string{}
		for _, fieldErr := range err.(validator.ValidationErrors) {
			fieldErrs = append(fieldErrs, fieldErr.Field())
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": fieldErrs,
		})
	}

	sqlString := `SELECT id, username FROM users WHERE email=$1 AND password=$2`
	row := databases.DBpools.QueryRow(c.Context(), sqlString, user.Email, user.Password)

	var id int
	var username string

	err = row.Scan(&id, &username)
	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"msg": "failed",
			"err": "invalid username or password",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	// session
	sess, err := Session.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	sess.Set("name", username)
	sess.Set("id_user", id)
	sess.SetExpiry(150 * time.Second)

	if err := sess.Save(); err != nil {
		panic(err.Error())
	}

	// cookie
	c.Cookie(&fiber.Cookie{
		Name: "test",
		Value: "ada ada aja",
		MaxAge: 10,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": "success",
		"data": fiber.Map{
			"id":       id,
			"username": username,
		},
	})
}

func UserLogout(c *fiber.Ctx) error {
	sess, err := Session.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	sess.Delete("name")
	sess.Delete("id_user")

	err = sess.Destroy()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	// dont call Save method, it cause create new session id in cookie
	// err = sess.Save()
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"msg": "failed",
	// 		"err": err.Error(),
	// 	})
	// }

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "success",
		"data": "",
	})
}
