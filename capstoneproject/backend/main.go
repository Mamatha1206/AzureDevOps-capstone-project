package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

const (
	PathGoals            = "/goals"
	SelectAllGoalsQuery  = "SELECT id, goal_name FROM goals"
	DbError              = "DB error"
	DbFile               = "./goals.db"
)

type Goal struct {
	ID   int    `json:"id"`
	Name string `json:"goal_name"`
}

func main() {
	db, err := sql.Open("sqlite3", DbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := setupRouter(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}

func setupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()

	router.GET("/health", healthHandler)
	router.GET(PathGoals, getGoalsHandler(db))
	router.POST(PathGoals, addGoalHandler(db))
	router.DELETE(PathGoals+"/:id", deleteGoalHandler(db))

	return router
}

func healthHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func getGoalsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(SelectAllGoalsQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": DbError})
			return
		}
		defer rows.Close()

		var goals []Goal
		for rows.Next() {
			var goal Goal
			if err := rows.Scan(&goal.ID, &goal.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": DbError})
				return
			}
			goals = append(goals, goal)
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "goals": goals})
	}
}

func addGoalHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			GoalName string `json:"goal_name"`
		}
		if err := c.BindJSON(&requestBody); err != nil || requestBody.GoalName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid goal name"})
			return
		}

		res, err := db.Exec("INSERT INTO goals(goal_name) VALUES (?)", requestBody.GoalName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": DbError})
			return
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": DbError})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"success": true, "goal": gin.H{"ID": lastID, "Name": requestBody.GoalName}})
	}
}

func deleteGoalHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		res, err := db.Exec("DELETE FROM goals WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": DbError})
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil || rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Goal not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Goal deleted successfully"})
	}
}
