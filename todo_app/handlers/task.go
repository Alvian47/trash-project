package handler

import (
	"errors"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"todo/databases"
	"todo/models"
	"todo/validation"
)

func CreateTask(c *fiber.Ctx) error {
	task := models.Task{}
	user_id := c.Query("user_id")

	err := c.BodyParser(&task)
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

	err = validation.Validation.Struct(&task)
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

	SQLstatement := `
	WITH tasks_id 
	AS (INSERT INTO tasks(title, content, status, deadline) VALUES($1, $2, $3, $4) RETURNING id)
	INSERT INTO library(user_id, task_id) 
	VALUES($5, (SELECT id from tasks_id));
	`
	tag, err := databases.DBpools.Exec(c.Context(), SQLstatement, task.Title, task.Content, task.Status, task.Deadline, user_id)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": errors.New("failed to add task").Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"msg":  "success",
		"data": "",
	})
}

func GetSingleTask(c *fiber.Ctx) error {
	task := models.Task{}

	id := c.Params("task_id")

	SQLStatement := `SELECT id, title, content, status, deadline, created_at  FROM tasks WHERE id = $1`
	row := databases.DBpools.QueryRow(c.Context(), SQLStatement, id)
	err := row.Scan(&task.Id, &task.Title, &task.Content, &task.Status, &task.Deadline, &task.Created_at)

	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "success",
		"data": task,
	})
}

func PatchEditTask(c *fiber.Ctx) error {
	task := models.Task{}
	id := c.Params("task_id")

	err := c.BodyParser(&task)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	SQLStatement := `UPDATE tasks SET title=$1, content=$2, deadline=$3 WHERE id=$4`
	_, err = databases.DBpools.Exec(c.Context(), SQLStatement, task.Title, task.Content, task.Deadline, id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "success",
		"data": task,
	})
}

func Test(c *fiber.Ctx) error {
	sess, err := Session.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	err = sess.Regenerate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}
	
	c.Cookie(&fiber.Cookie{
		Name: "test",
		Value: "ada ada aja",
		MaxAge: 300,
	})
	sess_id2 := c.Cookies("session_id")
	log.Println("regenerate: ",sess_id2)
	sess.SetExpiry(120 * time.Second)

	if err := sess.Save(); err != nil {
		panic(err.Error())
	}

	return c.Status(fiber.StatusOK).SendString("aman")
}

func GetAllTask(c *fiber.Ctx) error {
	tasks := []models.Task{}

	sess, err := Session.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	id := sess.Get("id_user")
	sess_id := c.Cookies("session_id")
	log.Println("getALl", sess_id)


	sqlStatement := `select tasks.id, tasks.title, tasks.content, tasks.status, tasks.deadline, tasks.created_at, tasks.updated_at 
	from library
	inner join tasks on tasks.id = library.task_id and library.user_id = $1;`

	rows, err := databases.DBpools.Query(c.Context(), sqlStatement, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}
	defer rows.Close()

	for rows.Next() {
		task := models.Task{}
		err := rows.Scan(&task.Id, &task.Title, &task.Content, &task.Status, &task.Deadline, &task.Created_at, &task.Updated_at)
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"msg": "failed",
				"err": err.Error(),
			})
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "failed",
				"err": err.Error(),
			})
		}

		tasks = append(tasks, task)
	}

	if rows.Err() != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": rows.Err().Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "Success",
		"data": tasks,
	})
}

func DeleteTask(c *fiber.Ctx) error {
	id := c.Params("task_id")

	sqlStatement := `DELETE FROM tasks WHERE id=$1`
	tag, err := databases.DBpools.Exec(c.Context(), sqlStatement, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": err.Error(),
		})
	}

	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "failed",
			"err": errors.New("task tidak ditemukan").Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg":  "Success",
		"data": "",
	})
}
