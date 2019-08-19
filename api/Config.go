package api

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql" //needed for side effects
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	"github.com/jmoiron/sqlx"
	logrus "github.com/sirupsen/logrus"
)

//Config is the global configuration object
var Config *ConfigStruct

//ConfigStruct holds the various configuration options
type ConfigStruct struct {
	Environment       string
	dbUser            string
	dbPassword        string
	dbHost            string
	dbPort            string
	dbName            string
	dbString          string
	DbConn            *sqlx.DB
	RootAPIURL        string
	RootAPIPort       string
	WebURL            string
	RateLimit         float64
	MailgunPrivateKey string
	MailgunPublicKey  string
	MailgunDomain     string
	MailShouldSend    bool
	MailFromAddress   string
	JWTSigningString  string
	Logger            *logrus.Logger
}

//ConfigSetup sets up the config struct with data from the environment
func ConfigSetup() *ConfigStruct {
	rand.Seed(time.Now().UnixNano())
	if Config != nil {
		return Config
	}
	c := ConfigStruct{}

	c.RateLimit = 100.0

	port := envHelper("PREGXAS_API_PORT", "8090")
	c.RootAPIPort = fmt.Sprintf(":%s", port)

	c.RootAPIURL = envHelper("PREGXAS_API_URL", fmt.Sprintf("http://localhost:%s/", c.RootAPIPort))
	c.WebURL = envHelper("PREGXAS_WEB_URL", "http://localhost:3000/")

	c.Environment = envHelper("PREGXAS_ENV", "develop")

	c.MailgunPrivateKey = os.Getenv("PREGXAS_EMAIL_PRIVATE")
	c.MailgunPublicKey = os.Getenv("PREGXAS_EMAIL_PUBLIC")
	c.MailgunDomain = os.Getenv("PREGXAS_EMAIL_DOMAIN")
	c.MailFromAddress = os.Getenv("PREGXAS_EMAIL_FROM")

	should := os.Getenv("PREGXAS_EMAIL_SHOULD_SEND")
	if should == "" {
		c.MailShouldSend = false
	} else {
		converted, err := strconv.ParseBool(should)
		if err != nil {
			fmt.Println("Warning: Could not convert PREGXAS_EMAIL_SHOULD_SEND; set as false")
			c.MailShouldSend = false
		} else {
			c.MailShouldSend = converted
		}
	}

	c.JWTSigningString = envHelper("PREGXAS_JWT_SIGNING_STRING", "")
	if c.Environment == "production" && c.JWTSigningString == "" {
		panic("insecure JWT signing token provided, aborting startup")
	} else {
		c.JWTSigningString = "THIS_IS_NOT_SECURE_CHANGE_THIS_ASAP_AND_USE_ONLY_IN_TESTING"
	}

	c.dbUser = envHelper("PREGXAS_DB_USER", "root")
	c.dbPassword = envHelper("PREGXAS_DB_PASSWORD", "password")
	c.dbHost = envHelper("PREGXAS_DB_HOST", "localhost")
	c.dbPort = envHelper("PREGXAS_DB_PORT", "3306")
	c.dbName = envHelper("PREGXAS_DB_NAME", "Pregxas")

	c.dbString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.dbUser,
		c.dbPassword, c.dbHost, c.dbPort, c.dbName)

	//setup the DB
	conn, err := sqlx.Open("mysql", c.dbString)
	if err != nil {
		fmt.Printf("\nERROR FOUND\n%v\n\n", err)
		panic(err)
	}
	conn.SetMaxIdleConns(100)

	//check the db
	_, err = conn.Exec("set session time_zone ='-0:00'")
	maxTries := 10
	secondsToWait := 10
	if err != nil {
		//we try again every X second for Y times, if it's still bad, we panic
		for i := 1; i <= maxTries; i++ {
			fmt.Printf("\n\tDB Error!\n\t%+v\n\t\tthis is attempt %d of %d. Waiting %d seconds...\n", err, i, maxTries, secondsToWait)
			time.Sleep(time.Duration(secondsToWait) * time.Second)
			_, err := conn.Exec("set session time_zone ='-0:00'")
			if err == nil {
				break
			}
			if i == maxTries {
				panic("Could not connect to the database, shutting down")
			}
		}
	}

	c.DbConn = conn

	// if the database is empty, we need to set it up
	_, err = c.DbConn.Exec("SELECT * FROM Users LIMIT 1")
	if err != nil {
		c.DbConn.Exec(fmt.Sprintf("CREATE DATABASE %s", c.dbName))
		err = populateDB(c.dbUser, c.dbPassword, c.dbHost, c.dbPort, c.dbName)
		if err != nil {
			fmt.Printf("\n%+v\n", err)
			panic("no database schema!")
		}
	}

	// now the logger
	c.Logger = logrus.New()
	c.Logger.SetFormatter(&logrus.JSONFormatter{})

	Config = &c
	return Config
}

func populateDB(user, password, host, port, name string) error {
	path := "file://sql"
	dbString := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user,
		password, host, port)
	db, err := sql.Open("mysql", dbString+"?multiStatements=true")
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", name))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("USE %s", name))
	if err != nil {
		return err
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		path,
		"mysql",
		driver,
	)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		return err
	}
	return nil
}

// IsProd is a helper func to determine if we are in production
func IsProd() bool {
	if Config.Environment == "production" || Config.Environment == "prod" {
		return true
	}
	return false
}

func envHelper(variable, defaultValue string) string {
	found := os.Getenv(variable)
	if found == "" {
		found = defaultValue
	}
	return found
}

// SetupApp sets up the Chi Router and basic configuration for the application
func SetupApp() *chi.Mux {
	Config = ConfigSetup()

	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	middlewares := GetMiddlewares()
	for _, m := range middlewares {
		r.Use(m)
	}

	// user routes
	r.Get("/me", GetMyProfileRoute)
	r.Patch("/me", UpdateMyProfileRoute)
	r.Post("/users/login", LoginUserRoute)
	r.Post("/users/signup", SignupUserRoute)
	r.Post("/users/signup/verify", VerifyEmailAndTokenRoute)
	r.Post("/users/login/reset", ResetPasswordStartRoute)
	r.Post("/users/login/reset/verify", ResetPasswordVerifyRoute)

	// prayer requests

	// prayers made

	return r
}
