package server

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/alecthomas/template"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/tikasan/gae-go-starter/api"
	"github.com/tikasan/gae-go-starter/controller"
	"github.com/tikasan/gae-go-starter/db"
	"github.com/tikasan/gae-go-starter/define"
	"github.com/tikasan/gae-go-starter/model"
)

type Server struct {
	echo *echo.Echo
	db   *gorm.DB
}

func New() *Server {
	return &Server{}
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (s *Server) Setup(env string) {
	cs, err := db.NewConfigsFromFile(define.DB_CONFIG)
	if err != nil {
		log.Fatalf("cannot open database configuration. exit. %s", err)
	}
	db, err := cs.Open(env)
	if err != nil {
		log.Fatalf("db initialization failed: %s", err)
	}
	s.db = db
	s.db.AutoMigrate(model.Comments{})
}

func (s *Server) Run() {
	s.echo = echo.New()
	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.CORS())

	t := &Template{
		templates: template.Must(template.ParseGlob("./views/*.html")),
	}
	s.echo.Renderer = t

	BBS := &controller.BBS{DB: s.db}
	s.echo.GET("/", BBS.Index)
	s.echo.GET("/:id", BBS.Show)
	s.echo.GET("/new", BBS.New)
	s.echo.POST("/save", BBS.Save)
	s.echo.GET("/edit/:id", BBS.Edit)
	s.echo.POST("/update", BBS.Update)
	s.echo.GET("/delete/:id", BBS.DeleteConf)
	s.echo.POST("/delete", BBS.Delete)

	API := &api.Request{DB: s.db}
	s.echo.GET("/api/comments", API.GetAllComments)

	s.echo.Pre(middleware.RemoveTrailingSlash())
	http.Handle("/", s.echo)
}

func init() {
	s := New()
	env := "production"
	if os.Getenv("RUN_WITH_DEVAPPSERVER") == "1" {
		env = "development"
	}
	s.Setup(env)
	s.Run()
}
