package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	catlist             *categories.CategoryList
	statusfile          string
	statusList          interface{}
	domainfile          string
	domainlist          []*domains.DomainDummy
	measurementfile     string
	measurementfilelock sync.RWMutex
	resultfile          *os.File
	domainidlock        sync.RWMutex
	domainidfile        *os.File
	domainidfilename    string
	lastid              = 0
)

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate)
	kingpin.Flag("listenaddress", "Socket for the server to listen on.").Default("0.0.0.0:8080").Short('l').StringVar(&listenaddress)
	kingpin.Flag("catfile", "File that contains the category information.").Default("catfile.json").Short('c').StringVar(&catfile)
	kingpin.Flag("statusfile", "File that contains the status information.").Default("statuses.json").Short('s').StringVar(&statusfile)
	kingpin.Flag("domainfile", "File that contains the status information.").Default("domain.json").Short('d').StringVar(&domainfile)
	kingpin.Flag("measurementfile", "File to write the results to.").Default("measurements.json").Short('m').StringVar(&measurementfile)
	kingpin.Flag("idfile", "File to write the last domain id to.").Default("domainid.txt").Short('i').StringVar(&domainidfilename)

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

	cats, err := categories.LoadCategoriesFromCsvFile(catfile)
	if err != nil {
		log.Panic("Cannot load categories from file: %s - %s", catfile, err.Error())
	}

	catlist, err = categories.CreateCategoryList(cats)
	log.Printf("Loading statuses from file.")
	err = statuses.LoadStatusesFromFile(statusfile, &statusList)
	if err != nil {
		log.Panic("Cannot load statuses from file: %s - %s", statusfile, err.Error())
	}

	domainidfile, err = os.OpenFile(domainidfilename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer domainidfile.Close()
	lastid = ReadLastId()
	log.Printf("Last fetched domain id: %d", lastid)

	domainlist, err = domains.LoadDomainsFromTxtFile(domainfile)
	if err != nil {
		log.Panic(err)
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

func ReadLastId() int {
	content, _ := ioutil.ReadFile(domainidfilename)
	id, _ := strconv.Atoi(string(content))
	return id
}

func UpdateLastId(i int) {
	domainidlock.Lock()
	domainidfile.Truncate(0)
	domainidfile.Seek(0, 0)
	domainidfile.WriteString(strconv.Itoa(i))
	lastid = i + 1
	domainidlock.Unlock()
}

func FetchHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, domainlist[0])
}

func Fetch(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, domainlist[lastid])
	UpdateLastId(lastid)

}

func GetCategories(c *gin.Context) {

	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, catlist)
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

	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-type")
	c.Header("Content-Type", "application/vnd.api+json")
	c.JSON(200, nil)
}
