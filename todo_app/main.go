package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/joho/godotenv"

	"todo/databases"
	handler "todo/handlers"
	"todo/validation"
)

func loadENV(key string) string {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	return os.Getenv(key)
}

func main() {
	app := fiber.New()

	validation.SingletonValidation()

	err := databases.DBPGX(context.Background(), loadENV("DATABASE_URL"))
	if err != nil {
		panic(err.Error())
	}

	key := encryptcookie.GenerateKey()
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: key,
	}))

	app.Use("/api", func(c *fiber.Ctx) error {
		sess, err := handler.Session.Get(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"msg": "failed",
				"err": err.Error(),
			})
		}

		id := c.Cookies("session_id")
		test := c.Cookies("test")
		idUser, username := sess.Get("id_user"), sess.Get("name")
		log.Printf("middleware session_id: %s, cookie test: %s", id, test)
		if id == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"msg": "failed",
				"err": "session invalid",
			})
		}

		if idUser == nil && username == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"msg": "failed",
				"err": "session invalid",
			})
		}

		return c.Next()
	})

	app.Post("/register", handler.RegisterUser)
	app.Post("/login", handler.LoginUser)
	app.Get("/api/logout", handler.UserLogout)

	app.Get("/api/test", handler.Test)
	app.Post("/api/create-task", handler.CreateTask)
	app.Get("/api/single-task/:task_id", handler.GetSingleTask)
	app.Get("/api/all-task", handler.GetAllTask)
	app.Patch("/api/edit-task/:task_id", handler.PatchEditTask)
	app.Delete("/api/del-task/:task_id", handler.DeleteTask)

	app.Listen(":3000")
}
