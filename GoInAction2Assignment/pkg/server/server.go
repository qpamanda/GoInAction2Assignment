/*
Package server implements all the handlers functions and is separated into 5 go files
to segregate the functionalities of the application.

	server.go: Initialises the application variables and starts the server.

	indexHandler.go: Manages the index page and implements the functionalities for user logins.

	orderHandler.go: Implements the functionalities to manage orders.

	pizzaHandler.go: Implements the functionalities to manage pizzas.

	userHandler.go: Implements the functionalities to manage users.
*/
package server

import (
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	filename "github.com/keepeye/logrus-filename"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	order "GoInAction2Assignment/pkg/order"
	pizza "GoInAction2Assignment/pkg/pizza"
)

// Global variables
var (
	wg    sync.WaitGroup
	mutex sync.Mutex

	tpl         *template.Template
	mapUsers    = map[string]user{}
	mapSessions = map[string]string{}

	log  = logrus.New()
	file *os.File

	newPizzaNo    int     // To generate a new Pizza No
	newOrderNo    int     // To generate a new Order No
	maxOrderQty   int     // Set the max order quantity
	standardPrice float64 // Set the standard price of a pizza

	minUserName int // Set the min length for new Username
	maxUserName int // Set the max length for new Username
	minPassword int // Set the min length for new Password
	maxPassword int // Set the max length for new Password

	bFirst = true

	// Create an empty LinkedList
	pizzaList = &pizza.Linkedlist{
		Head: nil,
		Size: 0,
	}

	// Create an empty Queue
	orderQueue = &order.Queue{
		Front: nil,
		Back:  nil,
		Size:  0,
	}

	// Create an empty []viewOrder slice. Use for displaying completed orders.
	completedOrderSlice = make([]viewOrder, 0)
)

// viewOrderItem is used for display in the html templates
type viewOrderItem struct {
	ItemNo     int
	PizzaNo    int
	PizzaName  string
	PizzaPrice string
	OrderQty   int
	Checked    string
	ErrorMsg   string
}

// viewOrder is used for display in the html templates
type viewOrder struct {
	IdxNo          int
	OrderNo        int
	ViewOrderItems []viewOrderItem
	TotalCost      string
	UserName       string
}

// viewPizzaSales is used for display in the html templates
type viewPizzaSales struct {
	PizzaNo     int
	PizzaName   string
	OrderQty    int
	TotalSales  float64
	STotalSales string
}

// viewPizza is used for display in the html templates
type viewPizza struct {
	PizzaNo     int
	PizzaName   string
	PizzaPrice  float64
	SPizzaPrice string
	Selected    string
}

// user struct for storing user account information
type user struct {
	UserName       string
	Password       []byte
	FirstName      string
	LastName       string
	Email          string
	IsAdmin        bool
	CreatedDT      time.Time
	LastModifiedDT time.Time
	CurrentLoginDT time.Time
	LastLoginDT    time.Time
}

// InitServer will start the required workflow before server starts.
// It will complete all initialisation and run only once.
// First it will parse templates. Then it will open/create the file for logging.
// After which, it will load variables from .env and initialise the global variables.
// It will then create an admin user and a standard pizza menu for testing purpose.
func InitServer() {
	// Parse templates
	tpl = template.Must(template.ParseGlob("templates/*"))

	// Log file name is based on current day. Thus a new file is created for each day.
	date := time.Now().Format("20060102")
	logFileName := "log/" + date + "_events.log"

	// Create a new log file for append
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("FATAL: OpenFile - ", err)
	}

	// Log to events.log file
	log.SetOutput(file)
	// Set log formatter
	log.SetFormatter(&logrus.JSONFormatter{})
	// Set log level from Info level onwards
	log.SetLevel(logrus.InfoLevel)

	// Use 3rd party package filename to display filename and line no during logging
	filenameHook := filename.NewHook()
	filenameHook.Field = "line"
	log.AddHook(filenameHook)

	// Load setup.env file from same directory
	err = godotenv.Load("setup.env")
	if err != nil {
		log.Fatal("FATAL: Error loading .env file")
	}

	// Get env variables for admin user (ADMIN_USERNAME, ADMIN_PASSWORD, ADMIN_FIRSTNAME, ADMIN_LASTNAME)
	adminUserName := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminFName := os.Getenv("ADMIN_FIRSTNAME")
	adminLName := os.Getenv("ADMIN_LASTNAME")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	// Encrypt the admin password
	bPassword, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.MinCost)

	// Add the admin user into a map
	mapUsers["admin"] = user{adminUserName, bPassword, adminFName, adminLName, adminEmail, true, time.Now(), time.Now(), time.Now(), time.Now()}

	// Get the default list of pizza in a string
	sPizza := os.Getenv("STANDARD_PIZZA")
	standardPizza := strings.Split(sPizza, "|")

	// Get the standard price of a pizza
	standardPrice, _ = strconv.ParseFloat(os.Getenv("STANDARD_PRICE"), 64)

	// Setup the default pizza menu
	pizzaList.CreateStartMenu(standardPizza, standardPrice)

	// Set the starting pizza no for creation of a new pizza
	newPizzaNo, _ = strconv.Atoi(os.Getenv("PIZZA_NO"))

	// Set the starting order no for creation of a new order
	newOrderNo, _ = strconv.Atoi(os.Getenv("ORDER_NO"))

	// Set the max order quantity for each pizza order
	maxOrderQty, _ = strconv.Atoi(os.Getenv("MAX_ORDER_QTY"))

	// Set the min characters for username
	minUserName, _ = strconv.Atoi(os.Getenv("MIN_USERNAME"))

	// Set the max characters for username
	maxUserName, _ = strconv.Atoi(os.Getenv("MAX_USERNAME"))

	// Set the min characters for password
	minPassword, _ = strconv.Atoi(os.Getenv("MIN_PASSWORD"))

	// Set the max characters for password
	maxPassword, _ = strconv.Atoi(os.Getenv("MAX_PASSWORD"))
}

// StartServer initialise all the handler func then listen and start the server on the given port using https.
// At the end, it will close the logging file.
func StartServer() {

	// Initialise the handle func for each link
	http.HandleFunc("/", index)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/edituser", edituser)
	http.HandleFunc("/deleteuser", deleteuser)
	http.HandleFunc("/addorder", addorder)
	http.HandleFunc("/editorder", editorder)
	http.HandleFunc("/vieworders", vieworders)
	http.HandleFunc("/completeorder", completeorder)
	http.HandleFunc("/pizzasales", pizzasales)
	http.HandleFunc("/addpizza", addpizza)
	http.HandleFunc("/editpizza", editpizza)
	http.HandleFunc("/deletepizza", deletepizza)
	http.HandleFunc("/viewpizza", viewpizza)
	http.HandleFunc("/logout", logout)

	http.Handle("/favicon.ico", http.NotFoundHandler())

	// Set the listen port
	err := http.ListenAndServeTLS(":5221", "certs//cert.pem", "certs//key.pem", nil)
	if err != nil {
		log.Fatal("FATAL: ListenAndServeTLS - ", err)
	}

	// Defer file closure to the end
	defer file.Close()

}
