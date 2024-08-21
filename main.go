package main

import (
	external_loquela "alpha-echo/external/loquela"
	"alpha-echo/handlers"
	"alpha-echo/models"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Command Line Flagging
var runTask string
var envFile string
var prepLanguage string

func init() {
	flag.StringVar(&runTask, "runTask", "", "Run tasks. Available: MigrateModels, SeedModels")
	flag.StringVar(&envFile, "envFile", ".env.dev", "Environment file name")
	flag.StringVar(&prepLanguage, "prepLanguage", "", "Prepare Language Audios for [Loquela]")
	flag.Parse()
}

func (t *Templ) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templ.ExecuteTemplate(w, name, data)
}

func main() {
	// Logger
	logger := NewLogger()

	// Env Setup
	if err := godotenv.Load(envFile); err != nil {
		logger["ERROR"].Fatalf("Loading env failed. %v\n", err)
	}
	fmt.Printf("MODE: %s\n", os.Getenv("ENV_INFO"))

	// Echo
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ","),
		AllowMethods:     strings.Split(os.Getenv("CORS_ALLOW_METHODS"), ","),
		AllowHeaders:     strings.Split(os.Getenv("CORS_ALLOW_HEADERS"), ","),
		AllowCredentials: true,
		ExposeHeaders:    strings.Split(os.Getenv("CORS_EXPOSE_HEADERS"), ","),
		MaxAge:           12 * 60 * 60,
	}))
	e.Renderer = newTemplate()
	// if os.Getenv("ENV_INFO") == "DEV" {
	// 	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
	// 		LogHost:     true,
	// 		LogLatency:  true,
	// 		LogProtocol: true,
	// 		LogURI:      true,
	// 		LogURIPath:  true,
	// 		LogValuesFunc: func(e echo.Context, v middleware.RequestLoggerValues) error {
	// 			fmt.Printf("REQUEST\nHost: %v\nLatency: %v\nProtocol: %v\nURI: %v\nURI Path: %v\n", v.Host, v.Latency, v.Protocol, v.URI, v.URIPath)
	// 			return nil
	// 		},
	// 	}))
	// }

	// Database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_DATA"),
		os.Getenv("DB_PORT"),
	)
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// Validator
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("gate_name_in_register_only", GateNameRegisterValidation)

	// Handler
	h := handlers.NewHandler(db, validate, logger)

	// Tasks
	if runTask == "all" {
		ts := []string{
			"MigrateModels",
			"SeedModels",
		}
		RunTasks(ts, db, logger)
		return
	} else if runTask != "" {
		ts := strings.Split(runTask, ",")
		RunTasks(ts, db, logger)
		return
	}

	if prepLanguage == "mandarin" {
		ss := external_loquela.NewSpeechSynthesizer("Mandarin", "Female", 1, logger)
		var vocabularies []models.LoquelaVocabulary
		if err := db.Where("loquela_language_id", 1).Find(&vocabularies).Error; err != nil {
			logger["ERROR"].Fatalf("Failed fetching vocabularies for %s language.", prepLanguage)
		}
		for _, vocabulary := range vocabularies {
			logger["LOQUELA_PREPARE"].Printf("Synthesizing audio for %s", vocabulary.Word)
			ss.BuildRequest(vocabulary.Word)
			ss.BuildUrl()
			ss.BuildFile(&vocabulary)
			ss.Synthesize("")
		}
		logger["LOQUELA_PREPARE"].Printf("Done Synthesizing for %s language", prepLanguage)
		return
	}

	// Static Files
	static := e.Group("/static")

	static.Use(h.IPFilterMiddleware)
	static.Static("/", "views/static/")

	// Routes
	e.HTTPErrorHandler = h.ErrorHandler
	e.Use(h.AccessLogMiddleware)
	e.Use(h.AccessMiddleware)

	e.GET("/", h.IndexHandler.Index)
	e.GET("/default", h.IndexHandler.Default)
	e.GET("/about", h.IndexHandler.About)
	e.GET("/projects", h.IndexHandler.Projects)
	e.GET("/gate", h.IndexHandler.Gate)
	e.POST("/gate", h.IndexHandler.GatePassing)
	e.PUT("/gate", h.IndexHandler.GateSwitch)

	regular := e.Group("/r")
	{
		opus := regular.Group("/opus")
		{
			opus.GET("/", h.OpusHandler.Default)
			opus.GET("/tasks", h.OpusHandler.GetTasks)
			opus.GET("/task/:id", h.OpusHandler.GetTaskByID)
			opus.POST("/category", h.OpusHandler.AddCategory)
			opus.POST("/task", h.OpusHandler.AddTask)
			opus.POST("/task-goal", h.OpusHandler.AddTaskGoal)
			opus.PUT("/task", h.OpusHandler.UpdateTask)
			opus.PUT("/state", h.OpusHandler.UpdateState)
			opus.PUT("/goal", h.OpusHandler.UpdateGoal)
			opus.DELETE("/category/:id", h.OpusHandler.DeleteCategory)
			opus.DELETE("/task/:id", h.OpusHandler.DeleteTask)
		}

		chrysus := regular.Group("/chrysus")
		{
			chrysus.GET("/", h.ChrysusHandler.Default)
		}

		vacuus := regular.Group("/vacuus")
		{
			vacuus.GET("/", h.VacuusHandler.Default)
			vacuus.PUT("/animation", h.VacuusHandler.UpdateAnimation)
		}

		loquela := regular.Group("/loquela")
		{
			loquela.GET("/", h.LoquelaHandler.Default)
			loquela.GET("/home", h.LoquelaHandler.Home)
			loquela.GET("/flashcard", h.LoquelaHandler.Flashcard)
		}
	}

	e.Start(os.Getenv("APP_PORT"))
}
