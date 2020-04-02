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

	c.Environment = envHelper("PREGXAS_ENV", "test")

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

	// now the logger
	c.Logger = logrus.New()
	c.Logger.SetFormatter(&logrus.JSONFormatter{})

	Config = &c

	// if the database is empty, we need to set it up
	_, err = c.DbConn.Exec("SELECT * FROM Sites WHERE status = 'active' LIMIT 1")
	if err != nil && (c.Environment == "production" || c.Environment == "develop") {
		// currently, migrate is not working correctly. This needs to be looked at
		c.DbConn.Exec(fmt.Sprintf("CREATE DATABASE %s", c.dbName))
		err = populateDB(c.dbUser, c.dbPassword, c.dbHost, c.dbPort, c.dbName)
		if err != nil {
			fmt.Printf("\n%+v\n", err)
			panic("no database schema!")
		}
	}
	// now, try to get the site so we can figure out if setup is needed
	err = LoadSite()
	if err != nil || Site.Status == "pending_setup" {
		// if the site isn't ready to go and the secret key is present, redisplay what is in the db
		key := Site.SecretKey
		if Site.SecretKey == "" {
			// now we need to generate a secret key for logging in
			key = GenerateSiteKey()
			err = SetupInitialSite(key)
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("\n\t--------------------------\n\t--------------------------\n\t-- Site Key: %s --\n\t--------------------------\n\t--------------------------\n\n", key)
		fmt.Printf("Using the client of your choice, you must now setup your site. If you were not expecting this message, please ensure you setup your database correctly\n\n")
	}
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

	// site routes
	r.Get("/admin/site", GetSiteInfoRoute)
	r.Post("/admin/site", SetupSiteRoute)
	r.Patch("/admin/site", UpdateSiteRoute) // TODO: needs OAS3 docs

	// user routes
	r.Get("/me", GetMyProfileRoute)
	r.Patch("/me", UpdateMyProfileRoute)
	r.Post("/users/login", LoginUserRoute)
	r.Post("/users/signup", SignupUserRoute)                 // TODO: needs OAS3 docs
	r.Post("/users/signup/verify", VerifyEmailAndTokenRoute) // TODO: needs OAS3 docs
	r.Post("/users/login/reset", ResetPasswordStartRoute)
	r.Post("/users/login/reset/verify", ResetPasswordVerifyRoute)

	// communities
	r.Post("/communities", CreateCommunityRoute)                 // TODO: needs OAS3 docs
	r.Get("/communities", GetCommunitiesForUserRoute)            // TODO: needs OAS3 docs
	r.Get("/communities/public", GetPublicCommunitiesRoute)      // TODO: needs OAS3 docs
	r.Patch("/communities/{communityID}", UpdateCommunityRoute)  // TODO: needs OAS3 docs
	r.Get("/communities/{communityID}", GetCommunityByIDRoute)   // TODO: needs OAS3 docs
	r.Delete("/communities/{communityID}", DeleteCommunityRoute) // TODO: needs OAS3 docs

	r.Post("/communities/{communityID}/subscribe", nil)   // TODO: implement
	r.Delete("/communities/{communityID}/subscribe", nil) // TODO: implement

	// join requests
	r.Get("/communities/{communityID}/users", GetCommunityLinksRoute)                     // this is for listing; TODO: needs OAS3 docs
	r.Put("/communities/{communityID}/users/{userID}", RequestCommunityMembershipRoute)   // this is for requesting access or requesting a user to join; TODO: needs OAS3 docs
	r.Delete("/communities/{communityID}/users/{userID}", RemoveCommunityMembershipRoute) // this is for removing a request; TODO: needs OAS3 docs
	r.Post("/communities/{communityID}/users/{userID}", ProcessCommunityMembershipRoute)  // this is for approving; TODO: needs OAS3 docs

	// prayer requests
	r.Get("/requests", GetGlobalPrayerRequestsRoute)
	r.Post("/requests", CreatePrayerRequestRoute)
	r.Get("/requests/{requestID}", GetPrayerRequestByIDRoute)
	r.Patch("/requests/{requestID}", UpdatePrayerRequestRoute)
	r.Delete("/requests/{requestID}", DeletePrayerRequestRoute)

	r.Get("/users/{userID}/requests", GetUserPrayerRequestsRoute) // TODO: needs OAS3 docs

	r.Get("/communities/{communityID}/requests", GetCommunityPrayerRequestsRoute)                      // TODO: needs OAS3 docs
	r.Put("/communities/{communityID}/requests/{requestID}", AddPrayerRequestToCommunityRoute)         // TODO: needs OAS3 docs
	r.Delete("/communities/{communityID}/requests/{requestID}", RemovePrayerRequestFromCommunityRoute) // TODO: needs OAS3 docs

	// prayers made
	r.Get("/requests/{requestID}/prayers", GetPrayersMadeOnRequestRoute)      // TODO: needs OAS3 docs
	r.Post("/requests/{requestID}/prayers", AddPrayerToRequestRoute)          // TODO: needs OAS3 docs
	r.Delete("/requests/{requestID}/prayers", RemovePrayerMadeOnRequestRoute) // the whenPrayed query param should be added; TODO: needs OAS3 docs

	// lists
	r.Get("/lists/requests", GetPrayerListsForUserRoute)                                     // TODO: needs OAS3 docs
	r.Post("/lists/requests", CreatePrayerListRoute)                                         // TODO: needs OAS3 docs
	r.Get("/lists/requests/{listID}", GetPrayerListByIDRoute)                                // TODO: needs OAS3 docs
	r.Patch("/lists/requests/{listID}", UpdatePrayerListRoute)                               // TODO: needs OAS3 docs
	r.Delete("/lists/requests/{listID}", DeletePrayerListByIDRoute)                          // TODO: needs OAS3 docs
	r.Put("/lists/requests/{listID}/{requestID}", AddPrayerRequestToPrayerListRoute)         // TODO: needs OAS3 docs
	r.Delete("/lists/requests/{listID}/{requestID}", RemovePrayerRequestFromPrayerListRoute) // TODO: needs OAS3 docs

	// reports
	r.Post("/requests/{requestID}/reports", ReportRequestRoute)      // TODO: needs OAS3 docs
	r.Get("/requests/{requestID}/reports", GetReportsOnRequestRoute) // TODO: needs OAS3 docs

	r.Get("/admin/reports", GetReportsOnPlatformRoute)       // TODO: needs OAS3 docs
	r.Get("/admin/reports/reasons", GetReportReasonsRoute)   // TODO: needs OAS3 docs
	r.Get("/admin/reports/statuses", GetReportStatusesRoute) // TODO: needs OAS3 docs
	r.Get("/admin/reports/{reportID}", GetReportRoute)       // TODO: needs OAS3 docs
	r.Patch("/admin/reports/{reportID}", UpdateReportRoute)  // TODO: needs OAS3 docs

	return r
}
