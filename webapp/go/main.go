package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	log2 "log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

const Limit = 20
const NazotteLimit = 50

var db_chair *sqlx.DB
var db_estate *sqlx.DB
var mySQLConnectionDataChair *MySQLConnectionEnv
var mySQLConnectionDataEstate *MySQLConnectionEnv
var chairSearchCondition ChairSearchCondition
var estateSearchCondition EstateSearchCondition

func NewMySQLConnectionEnv(host_variable string) *MySQLConnectionEnv {
	return &MySQLConnectionEnv{
		Host:     getEnv(host_variable, "127.0.0.1"),
		Port:     getEnv("MYSQL_PORT", "3306"),
		User:     getEnv("MYSQL_USER", "isucon"),
		DBName:   getEnv("MYSQL_DBNAME", "isuumo"),
		Password: getEnv("MYSQL_PASS", "isucon"),
	}
}

//ConnectDB isuumoデータベースに接続する
func (mc *MySQLConnectionEnv) ConnectDB() (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", mc.User, mc.Password, mc.Host, mc.Port, mc.DBName)
	return sqlx.Open("mysql", dsn)
}

func init() {
	jsonText, err := ioutil.ReadFile("../fixture/chair_condition.json")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(jsonText, &chairSearchCondition)

	jsonText, err = ioutil.ReadFile("../fixture/estate_condition.json")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(jsonText, &estateSearchCondition)
}

func initDBChair(c echo.Context, ch chan error) {
	sqlDir := filepath.Join("..", "mysql", "db")
	p := filepath.Join(sqlDir, "3.chair.sql")
	sqlFile, _ := filepath.Abs(p)
	cmdStr := fmt.Sprintf("mysql -h %v -u %v -p%v -P %v %v < %v",
		mySQLConnectionDataChair.Host,
		mySQLConnectionDataChair.User,
		mySQLConnectionDataChair.Password,
		mySQLConnectionDataChair.Port,
		mySQLConnectionDataChair.DBName,
		sqlFile,
	)
	cmd := exec.Command("bash", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stderr = &out
	ch <- cmd.Run()
}

func initDBEstate(c echo.Context, ch chan error) {
	sqlDir := filepath.Join("..", "mysql", "db")
	p := filepath.Join(sqlDir, "3.estate.sql")
	sqlFile, _ := filepath.Abs(p)
	cmdStr := fmt.Sprintf("mysql -h %v -u %v -p%v -P %v %v < %v",
		mySQLConnectionDataEstate.Host,
		mySQLConnectionDataEstate.User,
		mySQLConnectionDataEstate.Password,
		mySQLConnectionDataEstate.Port,
		mySQLConnectionDataEstate.DBName,
		sqlFile,
	)
	cmd := exec.Command("bash", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stderr = &out
	ch <- cmd.Run()
}

func initialize(c echo.Context) error {
	chChair := make(chan error)
	go initDBChair(c, chChair)
	chEstate := make(chan error)
	go initDBEstate(c, chEstate)
	for i := 0; i < 2; i++ {
		select {
		case err := <- chChair:
			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
		case err := <- chEstate:
			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
		}
	}

	return c.JSON(http.StatusOK, InitializeResponse{
		Language: "go",
	})
}

func main() {
	go func() { log2.Println(http.ListenAndServe(":9876", nil)) }()
	// Echo instance
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize
	e.POST("/initialize", initialize)

	// Chair Handler
	e.GET("/api/chair/:id", getChairDetail)
	e.POST("/api/chair", postChair)
	e.GET("/api/chair/search", searchChairs)
	e.GET("/api/chair/low_priced", getLowPricedChair)
	e.GET("/api/chair/search/condition", getChairSearchCondition)
	e.POST("/api/chair/buy/:id", buyChair)

	// Estate Handler
	e.GET("/api/estate/:id", getEstateDetail)
	e.POST("/api/estate", postEstate)
	e.GET("/api/estate/search", searchEstates)
	e.GET("/api/estate/low_priced", getLowPricedEstate)
	e.POST("/api/estate/req_doc/:id", postEstateRequestDocument)
	e.POST("/api/estate/nazotte", searchEstateNazotte)
	e.GET("/api/estate/search/condition", getEstateSearchCondition)
	e.GET("/api/recommended_estate/:id", searchRecommendedEstateWithChair)

	mySQLConnectionDataChair = NewMySQLConnectionEnv("MYSQL_HOST_CHAIR")
	mySQLConnectionDataEstate = NewMySQLConnectionEnv("MYSQL_HOST_ESTATE")

	var err error
	db_chair, err = mySQLConnectionDataChair.ConnectDB()
	if err != nil {
		e.Logger.Fatalf("DB(chair) connection failed : %v", err)
	}
	db_chair.SetMaxOpenConns(10)
	defer db_chair.Close()

	db_estate, err = mySQLConnectionDataEstate.ConnectDB()
	if err != nil {
		e.Logger.Fatalf("DB(estate) connection failed : %v", err)
	}
	db_estate.SetMaxOpenConns(10)
	defer db_estate.Close()

	// Start server
	serverPort := fmt.Sprintf(":%v", getEnv("SERVER_PORT", "1323"))
	e.Logger.Fatal(e.Start(serverPort))
}
