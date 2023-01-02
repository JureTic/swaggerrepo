package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

type User struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username" gorm:"unique"`
	Email    string `json:"email" gorm:"unique"`
	Role     int    `json:"role"`
	Password string `json:"-"`
}

type Rating struct {
	Id      uint   `json:"id"`
	UserId  uint   `json:"user_id"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type RatingRequest struct {
	Id     uint   `json:"id"`
	Rating string `json:"rating"`
}

func sendMetrics(timeElapsed string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := strconv.Itoa(int(m.Sys))
	base_url := "http://104.45.183.75/api/metrics/posts/"
	apiURL := base_url + timeElapsed[:len(timeElapsed)-2] + "/" + memoryUsage
	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
}

func authentication() func(c *fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("API_KEY")),
	})
}

func authenticate(c *fiber.Ctx) (r bool) {
	authentication()

	return true
}

func main() {
	var dsn string
	dsn = "postgres://zlqwvdmx:x0tl7AVnX4zi0rsqeKcf8R2dhjvqOpib@ella.db.elephantsql.com/zlqwvdmx"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(Rating{})

	app := fiber.New()
	app.Use(cors.New())

	app.Post("/rate", authentication(), func(c *fiber.Ctx) error {
		start := time.Now()
		// Do api request to another container
		// url := "http://weatherapi:8001/api/test"
		req := new(Rating)
		if err := c.BodyParser(req); err != nil {
			return err
		}
		if req.Rating == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid credentials")
		}
		//save this info in the database
		//db.Model(&Ra{}).Where("id = ?", req.Id).Update("rating", req.Rating)
		var rating Rating
		rating.UserId = req.UserId
		rating.Rating = req.Rating
		rating.Comment = req.Comment
		//save this info in the database
		db.Create(&rating)
		// send to metricsapi
		timeElapsed := time.Since(start).String()
		sendMetrics(timeElapsed)

		return c.SendStatus(fiber.StatusAccepted)
	})

	app.Get("/rating", authentication(), func(c *fiber.Ctx) error {
		//start := time.Now()
		// Do api request to another container
		// url := "http://weatherapi:8001/api/test"
		authentication()

		req := new(User)
		if err := c.BodyParser(req); err != nil {
			return err
		}
		if req.Id == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid credentials")
		}
		//save this info in the database
		//db.Model(&Ra{}).Where("id = ?", req.Id).Update("rating", req.Rating)

		//save this info in the database
		var ratings []Rating
		results := db.Find(&ratings, "user_id = ?", req.Id)
		if results.Error != nil {
			return fiber.NewError(500, "error performing a query")
		}
		average := 0.0
		for i := 0; i < len(ratings); i++ {
			rating := ratings[i]
			average = average + float64(rating.Rating)
		}
		average = average / float64(len(ratings))

		// send to metricsapi
		//timeElapsed := time.Since(start).String()
		//sendMetrics(timeElapsed)

		return c.JSON(fiber.Map{"rating": average})
	})

	app.Get("/users", func(c *fiber.Ctx) error {
		result := []User{}
		db.Find(&result)
		return c.Status(http.StatusOK).JSON(&result)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON([]byte("Posts container working"))
	})

	app.Listen(":8000")
}
