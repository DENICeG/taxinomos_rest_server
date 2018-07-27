package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/DENICeG/taxinomos_rest_server/categories"
	"github.com/DENICeG/taxinomos_rest_server/domains"
	"github.com/DENICeG/taxinomos_rest_server/logging"
	"github.com/DENICeG/taxinomos_rest_server/statuses"

	"github.com/alecthomas/kingpin"
	"github.com/gin-gonic/gin"
)

var (
	builddate           string
	revision            string
	version             string
	lifetime            int
	listenaddress       string
	debuglevel          int
	configfile          string
	wg                  sync.WaitGroup
	categoryList        interface{}
	catfile             string
	statusfile          string
	statusList          interface{}
	domainfile          string
	domainList          interface{}
	measurementfile     string
	measurementfilelock sync.RWMutex
	resultfile          *os.File
)

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate)
	kingpin.Flag("listenaddress", "Socket for the server to listen on.").Default("0.0.0.0:8080").Short('l').StringVar(&listenaddress)
	kingpin.Flag("catfile", "File that contains the category information.").Default("catfile.json").Short('c').StringVar(&catfile)
	kingpin.Flag("statusfile", "File that contains the status information.").Default("statuses.json").Short('s').StringVar(&statusfile)
	kingpin.Flag("domainfile", "File that contains the status information.").Default("domain.json").Short('d').StringVar(&domainfile)
	kingpin.Flag("measurementfile", "File to write the results to.").Default("measurements.json").Short('m').StringVar(&measurementfile)

	kingpin.Parse()

	//

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logging.GinLogger())

	log.Println("=====  Taxinomos REST API Server  =====")
	log.Printf("Builddate: %s", builddate)
	log.Printf("Version  : %s", version)
	log.Printf("Revision : %s", revision)
	log.Println(" ---")
	log.Println("ServerConfig:")
	log.Printf("  listenaddress: %s", listenaddress)

	log.Printf("Loading categories from file. ")

	err := categories.LoadCategoriesFromFile(catfile, &categoryList)
	if err != nil {
		log.Panic("Cannot load categories from file: %s - %s", catfile, err.Error())
	}

	log.Printf("Loading statuses from file.")
	err = statuses.LoadStatusesFromFile(statusfile, &statusList)
	if err != nil {
		log.Panic("Cannot load statuses from file: %s - %s", statusfile, err.Error())
	}

	log.Printf("Loading domains from file")
	err = domains.LoadDomainsFromFile(domainfile, &domainList)
	if err != nil {
		log.Panic("Cannot load domains from file: %s - %s", domainfile, err.Error())
	}

	resultfile, err = os.OpenFile(measurementfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Panic("Cannot open resultfile: %s - %s", measurementfile, err.Error())
	}

	apiGroup := router.Group("/api/v1")
	{
		apiGroup.GET("/fetch", Fetch)
		apiGroup.GET("/categories", GetCategories)
		apiGroup.GET("/statuses", GetStatuses)
		apiGroup.POST("/measurements", Measurements)

		apiGroup.OPTIONS("/categories", GetCategories)
		apiGroup.OPTIONS("/statuses", GetStatuses)
		apiGroup.OPTIONS("/fetch", FetchHeaders)
		apiGroup.OPTIONS("/measurements", Measurements)
	}

	httpsrv := &http.Server{
		Addr:    listenaddress,
		Handler: router,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Fatal(httpsrv.ListenAndServe())
	}()
	log.Println("HTTP server started.")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	log.Println("Shutdown signals registered.")
	<-signalChan
	log.Println("Shutdown signal received, exiting.")
	httpsrv.Shutdown(context.Background())
	wg.Wait()
	log.Println("Server exiting")
}

func FetchHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, domainList)
}

func Fetch(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, domainList)

}

func GetCategories(c *gin.Context) {

	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, categoryList)
}

func GetStatuses(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, statusList)
}

func Measurements(c *gin.Context) {

	body, _ := ioutil.ReadAll(c.Request.Body)

	if c.Request.Method == "POST" {
		measurementfilelock.Lock()
		resultfile.WriteString(string(body))
		resultfile.WriteString("\n")
		measurementfilelock.Unlock()
	}

	log.Printf("%#v", c.Request)
	log.Printf("%#v", string(body))
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, nil)
}
