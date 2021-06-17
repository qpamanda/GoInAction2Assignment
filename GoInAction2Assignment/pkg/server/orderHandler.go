package server

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"

	order "GoInAction2Assignment/pkg/order"
	pizza "GoInAction2Assignment/pkg/pizza"
)

// addOrder takes in the OrderItem slice, a pizzaNo and orderQty.
// It creates an OrderItem with the pizzaNo and orderQty, then adds it to the slice.
func addOrder(orderSlice []order.OrderItem, pizzaNo int, orderQty int) []order.OrderItem {
	orderItem := order.OrderItem{
		PizzaNo:  pizzaNo,
		OrderQty: orderQty,
	}

	orderSlice = append(orderSlice, orderItem)

	return orderSlice
}

// addorder is a handler func to add a new order.
// Redirects to index page if user has not login.
func addorder(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := ""     // To display message to the user on the client
	bValidOrder := true // Use to determine if an order entry is valid

	viewOrderItemSlice := make([]viewOrderItem, 0) // To use for display in the template
	orderItemSlice := make([]order.OrderItem, 0)   // To add new OrderItem into the slice from user selected pizzas

	pizzaSlice, err := pizzaList.GetAllPizza() // Get all pizzas from the LinkedList data structure

	if err != nil {
		clientMsg = "No pizza in the menu today"
		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Error("no pizza in the menu")
	} else {
		// Loop through pizzaSlice, create a viewOrderItem struct and append it to viewOrderItemSlice
		// viewOrderItemSlice will be used to display the template form
		for idx1, val1 := range pizzaSlice {
			pizzaOrder := viewOrderItem{idx1 + 1, val1.PizzaNo, val1.PizzaName, fmt.Sprintf("%.2f", val1.PizzaPrice), 0, "", ""}
			viewOrderItemSlice = append(viewOrderItemSlice, pizzaOrder)
		}
	}

	// Process the form submission
	if req.Method == http.MethodPost {
		for _, val1 := range viewOrderItemSlice {
			errMsg := ""

			// Get selectedPizza from form checkbox using viewOrderItem.PizzaNo as its value
			// Since viewOrderItem.PizzaNo is an int, use strconv.Itoa to convert it to a string
			selectedPizza := req.FormValue(strconv.Itoa(val1.PizzaNo))

			// If selectedPizza is not an empty string, the pizza is selected i.e. checkbox is checked in the form
			// Otherwise, no action required since the pizza is not selected
			if selectedPizza != "" {
				// Get associated quantity value from form textbox that is in line with the selected pizza
				selectedQty := req.FormValue("orderqty" + strconv.Itoa(val1.ItemNo))

				// Convert selectedPizza and selectedQty to int values
				pizzaNo, _ := strconv.Atoi(selectedPizza) // blank identifier is used for error as pizzaNo value is assured

				orderQty, err := validateQuantity(selectedQty)

				if err != nil || orderQty == 0 {
					errMsg = "Enter a valid quantity" // Error message to display next to quantity checkbox in the tempplate

					bValidOrder = false // Order entry is invalid
				} else {
					// Add user selected pizza and valid quantity into orderItemSlice
					orderItemSlice = addOrder(orderItemSlice, pizzaNo, orderQty)
				}

				// Update values for display in the template form
				viewOrderItemSlice[val1.ItemNo-1].OrderQty = orderQty
				viewOrderItemSlice[val1.ItemNo-1].Checked = "checked"
				viewOrderItemSlice[val1.ItemNo-1].ErrorMsg = errMsg
			}
		}

		// Continue processing only if there are OrderItem in the orderItemSlice and the order entry is valid
		if len(orderItemSlice) > 0 && bValidOrder {
			// Set generateOrderNo as a go routine and add it into a wait group
			// This is to prevent multiple clients creating an order at the same time and the same order no being used
			wg.Add(1)
			go generateOrderNo(orderItemSlice)
			wg.Wait()

			orderNo := newOrderNo

			// Get the total cost of the order from order.CalOrderTotal func
			totalCost := calOrderTotal(orderItemSlice)

			// Enqueue the order
			orderQueue.Enqueue(orderNo, orderItemSlice, totalCost, myUser.UserName)

			// Print an order receipt on the server
			printOrderReceipt(orderNo, orderItemSlice, totalCost)

			clientMsg = "Order " + strconv.Itoa(orderNo) + " added successfully. Total payment is $" + fmt.Sprintf("%.2f", totalCost)

			log.WithFields(logrus.Fields{
				"userName":       myUser.UserName,
				"orderNo":        orderNo,
				"orderItemSlice": orderItemSlice,
				"totalCost":      totalCost,
			}).Info("new order added successfully")

		} else {
			clientMsg = "No orders made"

			log.WithFields(logrus.Fields{
				"userName": myUser.UserName,
			}).Info("no orders made")

		}
	}

	data := struct {
		User            user
		OrderSlice      []viewOrderItem
		CntCurrentItems int
		MaxOrder        int
		ClientMsg       string
	}{
		myUser,
		viewOrderItemSlice,
		len(viewOrderItemSlice),
		maxOrderQty,
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "addorder.gohtml", data)
}

// editorder is a handler func to edit an existing order.
// Redirects to index page if user has not login.
func editorder(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := "" // To display message to the user on the client

	orderNo := newOrderNo

	bValidOrder := true    // Use to determine if an order entry is valid
	bValidOrderNo := false // Use to determine if an order no is valid

	viewOrderItemSlice := make([]viewOrderItem, 0) // To use for display in the template
	orderItemSlice := make([]order.OrderItem, 0)   // To add new OrderItem into the slice from user selected pizzas

	pizzaSlice, err := pizzaList.GetAllPizza() // Get all pizzas from the LinkedList data structure

	if err != nil {
		clientMsg = "No pizza in the menu today"

		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Error("no pizza in the menu")

	} else {
		// Loop through pizzaSlice, create a viewOrderItem struct and append it to viewOrderItemSlice
		// viewOrderItemSlice will be used to display the template form
		for idx1, val1 := range pizzaSlice {
			pizzaOrder := viewOrderItem{idx1 + 1, val1.PizzaNo, val1.PizzaName, fmt.Sprintf("%.2f", val1.PizzaPrice), 0, "", ""}
			viewOrderItemSlice = append(viewOrderItemSlice, pizzaOrder)
		}
	}

	// Process the form submission
	if req.Method == http.MethodPost {
		// Get orderno from form text field and convert its string to int
		orderNo, err = strconv.Atoi(req.FormValue("orderno"))

		if err != nil {
			clientMsg = "Enter a valid order no"
		} else {
			// Use the orderNo to search for a valid order
			myOrder, err := orderQueue.SearchOrder(orderNo)
			myOrderSlice := myOrder.OrderSlice

			if err != nil || len(myOrderSlice) == 0 {
				clientMsg = "Order not found"

				log.WithFields(logrus.Fields{
					"userName": myUser.UserName,
					"orderNo":  orderNo,
				}).Error("order not found")

				bValidOrderNo = false
			} else {
				// Check if valid order belongs to current user and current user is not admin
				// because admin is allowed to edit all orders
				if myOrder.UserName != myUser.UserName && !myUser.IsAdmin {
					clientMsg = "Order not found"

					log.WithFields(logrus.Fields{
						"userName": myUser.UserName,
						"orderNo":  orderNo,
					}).Error("user is not allowed to edit this order")

					bValidOrderNo = false // Order No is invalid
				} else {
					bValidOrderNo = true // Order No is valid

					// bFirst is set to true in the global variables and it will be set to true
					// everytime user access the index page
					// It will set the values of the form to the order found on the first time
					// when user first access edit order only
					if bFirst {
						// Set viewOrderItem in viewOrderSlice to the values of the order found
						// for display in the template
						for _, val1 := range viewOrderItemSlice {
							for _, val2 := range myOrderSlice {
								if val1.PizzaNo == val2.PizzaNo {
									viewOrderItemSlice[val1.ItemNo-1].OrderQty = val2.OrderQty
									viewOrderItemSlice[val1.ItemNo-1].Checked = "checked"
									viewOrderItemSlice[val1.ItemNo-1].ErrorMsg = ""
								}
							}
						}
						bFirst = false
					}

					for _, val1 := range viewOrderItemSlice {
						errMsg := ""

						// Get selectedPizza from form checkbox using viewOrderItem.PizzaNo as its value
						// Since viewOrderItem.PizzaNo is an int, use strconv.Itoa to convert it to a string
						selectedPizza := req.FormValue(strconv.Itoa(val1.PizzaNo))

						// If selectedPizza is not an empty string, the pizza is selected i.e. checkbox is checked in the form
						// Otherwise, no action required since the pizza is not selected
						if selectedPizza != "" {
							// Get associated quantity value from form textbox that is in line with the selected pizza
							selectedQty := req.FormValue("orderqty" + strconv.Itoa(val1.ItemNo))

							// Convert selectedPizza and selectedQty to int values
							pizzaNo, _ := strconv.Atoi(selectedPizza)

							orderQty, err := validateQuantity(selectedQty)

							if err != nil || orderQty == 0 {
								errMsg = "Enter a valid quantity" // Error message to display next to quantity checkbox in the tempplate

								bValidOrder = false // Order entry is invalid
							} else {
								// Add user selected pizza and valid quantity into orderItemSlice
								orderItemSlice = addOrder(orderItemSlice, pizzaNo, orderQty)
							}

							// Update values for display in the template form
							viewOrderItemSlice[val1.ItemNo-1].OrderQty = orderQty
							viewOrderItemSlice[val1.ItemNo-1].Checked = "checked"
							viewOrderItemSlice[val1.ItemNo-1].ErrorMsg = errMsg

						}
					}

					// Continue processing only if there are OrderItem in the orderItemSlice and order entry is valid
					if len(orderItemSlice) > 0 && bValidOrder {
						// Get the total cost of the order from order.CalOrderTotal func
						totalCost := calOrderTotal(orderItemSlice)

						// Set orderQueue.UpdateOrder as a go routine and add it into a wait group
						// This is to prevent multiple clients updating an order at the same time
						wg.Add(1)
						go orderQueue.UpdateOrder(orderNo, orderItemSlice, totalCost, &wg, &mutex)
						wg.Wait()

						// Print an order receipt on the server
						printOrderReceipt(orderNo, orderItemSlice, totalCost)

						// Display message to the client
						clientMsg = "Order [" + strconv.Itoa(orderNo) + "] updated successfully. Total payment is $" + fmt.Sprintf("%.2f", totalCost)

						log.WithFields(logrus.Fields{
							"userName":       myUser.UserName,
							"orderNo":        orderNo,
							"orderItemSlice": orderItemSlice,
							"totalCost":      totalCost,
						}).Info("order updated successfully")
					} else {
						clientMsg = "No orders updated"

						log.WithFields(logrus.Fields{
							"userName": myUser.UserName,
							"orderNo":  orderNo,
						}).Info("no orders updated")
					}
				}
			}
		}
	}

	data := struct {
		User            user
		OrderNo         int
		OrderSlice      []viewOrderItem
		CntCurrentItems int
		MaxOrder        int
		ClientMsg       string
		ValidOrderNo    bool
	}{
		myUser,
		orderNo,
		viewOrderItemSlice,
		len(viewOrderItemSlice),
		maxOrderQty,
		clientMsg,
		bValidOrderNo,
	}

	tpl.ExecuteTemplate(res, "editorder.gohtml", data)
}

// vieworders is a handler func to view orders created by a user or view all orders if user is admin
// Redirects to index page if user has not login.
func vieworders(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := "" // To display message to the user on the client

	// Get slice of the current orders for the user.
	// Admin can view all orders.
	viewOrderSlice, err := getCurrentOrders(myUser.UserName, myUser.IsAdmin)

	if err != nil {
		clientMsg = "No orders found. "

		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Info(err)
	}

	// Get slice of the completed orders for the user.
	// Admin can view all orders.
	myCompletedOrderSlice := getCompletedOrders(myUser.UserName, myUser.IsAdmin)

	if len(myCompletedOrderSlice) == 0 {
		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Info("no completed orders")
	}

	data := struct {
		User                user
		ViewOrderSlice      []viewOrder
		CntCurrentItems     int
		CompletedOrderSlice []viewOrder
		CntCompletedItems   int
		ClientMsg           string
	}{
		myUser,
		viewOrderSlice,
		len(viewOrderSlice),
		myCompletedOrderSlice,
		len(myCompletedOrderSlice),
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "vieworders.gohtml", data)
}

// completeorder is a handler func to display orders that are currently in the queue.
// It allows the admin user to dequeue the first order in the queue.
// Redirects to index page if user has not login.
func completeorder(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := "" // To display message to the user on the client

	// Process the form submission
	if req.Method == http.MethodPost {
		// Open a new channel
		orderChannel := make(chan order.Order)

		// Create a goroutine for orderQueue.Dequeue
		go orderQueue.Dequeue(orderChannel)

		// Receive the Order struct from the channel
		pizzaOrder := <-orderChannel

		// Set updateCompletedOrders as a go routine and add it into a wait group
		// This is to prevent multiple updates at the same time
		wg.Add(1)
		go updateCompletedOrders(pizzaOrder, myUser)
		wg.Wait()
		clientMsg = "Order [" + strconv.Itoa(pizzaOrder.OrderNo) + "] completed and added to pizza sales."

		log.WithFields(logrus.Fields{
			"userName":       myUser.UserName,
			"orderNo":        pizzaOrder.OrderNo,
			"orderItemSlice": pizzaOrder.OrderSlice,
			"totalCost":      pizzaOrder.TotalCost,
			"orderUserName":  pizzaOrder.UserName,
		}).Info("order completed successfully and added to pizza sales")
	}

	// Get slice of all the current orders. Admin can view all orders.
	viewOrderSlice, err := getCurrentOrders(myUser.UserName, myUser.IsAdmin)

	if err != nil {
		clientMsg = "No orders found"
		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Info(err)
	}

	data := struct {
		User            user
		ViewOrderSlice  []viewOrder
		CntCurrentItems int
		ClientMsg       string
	}{
		myUser,
		viewOrderSlice,
		len(viewOrderSlice),
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "completeorder.gohtml", data)
}

// pizzasales is a handler func to display the sales of all the pizzas, its total quantity and total cost
// Redirects to index page if user has not login.
func pizzasales(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := "" // To display message to the user on the client

	// Get slice of all the current orders. Admin can view all orders.
	viewOrderSlice, err := getCurrentOrders(myUser.UserName, myUser.IsAdmin)

	if err != nil {
		clientMsg = "No orders found"
		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Info(err)
	}

	myCompletedOrderSlice := getCompletedOrders(myUser.UserName, myUser.IsAdmin)

	if len(myCompletedOrderSlice) == 0 {
		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Info("no completed orders")
	}

	// Get current and completed pizza sales as a slice for display
	// Channels are implemented for concurrency
	ch1 := make(chan []viewPizzaSales) // Open a new channel
	ch2 := make(chan []viewPizzaSales) // Open a new channel

	go getPizzaSales(viewOrderSlice, ch1)        // Use goroutine to get pizza sales for viewCurrentPizzaSales
	go getPizzaSales(myCompletedOrderSlice, ch2) // Use goroutine to get pizza sales for viewCompletedPizzaSales

	viewCurrentPizzaSales := <-ch1   // Receive the []viewPizzaSales from the channel
	viewCompletedPizzaSales := <-ch2 // Receive the []viewPizzaSales from the channel

	// Calculate current pizza sales total for display
	currentPizzaSalesTotal := 0.0
	if len(viewCurrentPizzaSales) > 0 {
		currentPizzaSalesTotal = calTotalSales(viewCurrentPizzaSales)
	}

	// Calculate completed pizza sales total for display
	completedPizzaSalesTotal := 0.0
	if len(viewCompletedPizzaSales) > 0 {
		completedPizzaSalesTotal = calTotalSales(viewCompletedPizzaSales)
	}

	data := struct {
		User                     user
		CurrentPizzaSales        []viewPizzaSales
		CntCurrentItems          int
		CurrentPizzaSalesTotal   string
		CompletedPizzaSales      []viewPizzaSales
		CntCompletedItems        int
		CompletedPizzaSalesTotal string
		ClientMsg                string
	}{
		myUser,
		viewCurrentPizzaSales,
		len(viewCurrentPizzaSales),
		fmt.Sprintf("%.2f", currentPizzaSalesTotal),
		viewCompletedPizzaSales,
		len(viewCompletedPizzaSales),
		fmt.Sprintf("%.2f", completedPizzaSalesTotal),
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "pizzasales.gohtml", data)
}

// validateQuantity will parse the qty string as int and returns the parsed value
func validateQuantity(qty string) (int, error) {
	retValue, err := strconv.Atoi(qty)

	if err != nil {
		return retValue, err
	} else {
		if retValue <= 0 || retValue > maxOrderQty {
			return retValue, errors.New("invalid order quantity")
		}
	}

	return retValue, nil
}

// printOrderReceipt prints the receipt of the order made on the server terminal/cmd prompt
func printOrderReceipt(orderNo int, orderSlice []order.OrderItem, totalCost float64) {

	printDividerLine()
	fmt.Println("* RECEIPT *")
	fmt.Println()

	fmt.Println("Order No: ", orderNo)
	fmt.Println()

	pizzaTotal := 0.0

	for _, v := range orderSlice {
		pizzaOrder, _ := pizzaList.SearchPizza(v.PizzaNo)
		pizzaTotal = float64(v.OrderQty) * pizzaOrder.PizzaPrice

		fmt.Printf("%d x %s\t$%.2f\n", v.OrderQty, pizzaOrder.PizzaName, pizzaTotal)
	}

	fmt.Println("\t\t\t--------")
	fmt.Printf("TOTAL PAYMENT\t\t$%.2f\n", totalCost)
	fmt.Println("\t\t\t--------")
}

// calOrderTotal calculates the total amount of an order
func calOrderTotal(orderSlice []order.OrderItem) float64 {

	orderTotal := 0.0
	pizzaTotal := 0.0

	var pizzaOrder pizza.Pizza

	for _, v := range orderSlice {
		pizzaOrder, _ = pizzaList.SearchPizza(v.PizzaNo)
		pizzaTotal = float64(v.OrderQty) * pizzaOrder.PizzaPrice
		orderTotal = orderTotal + pizzaTotal
	}

	return orderTotal
}

// updateCompletedOrders updates the global var for completedOrderSlice for display of orders that have been dequeued.
// Mutex lock and unlock is implemented for concurrency
func updateCompletedOrders(completedOrder order.Order, myUser user) {

	defer wg.Done()

	mutex.Lock()
	orderSlice := completedOrder.OrderSlice // Get []OrderItem from completedOrder into a slice
	viewOrderItemSlice := make([]viewOrderItem, 0)
	pizzaSlice, _ := pizzaList.GetAllPizza()

	for idx1, val1 := range orderSlice {
		for _, val2 := range pizzaSlice {
			if val1.PizzaNo == val2.PizzaNo {
				pizzaOrder := viewOrderItem{idx1 + 1, val1.PizzaNo, val2.PizzaName, fmt.Sprintf("%.2f", val2.PizzaPrice), val1.OrderQty, "", ""}
				viewOrderItemSlice = append(viewOrderItemSlice, pizzaOrder)
			}
		}
	}

	viewOrder := viewOrder{len(completedOrderSlice) + 1, completedOrder.OrderNo, viewOrderItemSlice, fmt.Sprintf("%.2f", completedOrder.TotalCost), completedOrder.UserName}
	completedOrderSlice = append(completedOrderSlice, viewOrder)

	mutex.Unlock()
}

// getCurrentOrders retrieves the current orders into []viewOrder
func getCurrentOrders(userName string, isAdmin bool) ([]viewOrder, error) {

	viewOrderSlice := make([]viewOrder, 0)
	pizzaSlice, _ := pizzaList.GetAllPizza()
	orderQSlice, err := orderQueue.GetAllOrders(userName, isAdmin)

	if err != nil {
		return viewOrderSlice, err
	} else {
		for idx1, val1 := range orderQSlice {
			orderSlice := val1.OrderSlice
			viewOrderItemSlice := make([]viewOrderItem, 0)

			for idx2, val2 := range orderSlice {
				for _, val3 := range pizzaSlice {
					if val2.PizzaNo == val3.PizzaNo {
						pizzaOrder := viewOrderItem{idx2 + 1, val2.PizzaNo, val3.PizzaName, fmt.Sprintf("%.2f", val3.PizzaPrice), val2.OrderQty, "", ""}
						viewOrderItemSlice = append(viewOrderItemSlice, pizzaOrder)
					}
				}
			}

			viewOrder := viewOrder{idx1 + 1, val1.OrderNo, viewOrderItemSlice, fmt.Sprintf("%.2f", val1.TotalCost), val1.UserName}
			viewOrderSlice = append(viewOrderSlice, viewOrder)
		}
	}

	return viewOrderSlice, nil
}

// getCompletedOrders retrieves the completed orders into []viewOrder
func getCompletedOrders(userName string, isAdmin bool) []viewOrder {

	myCompletedOrderSlice := make([]viewOrder, 0)

	if !isAdmin {
		i := 0
		for _, val1 := range completedOrderSlice {
			if val1.UserName == userName {
				myCompletedOrderSlice = append(myCompletedOrderSlice, val1)
				myCompletedOrderSlice[i].IdxNo = i + 1
				i++
			}
		}
	} else {
		return completedOrderSlice
	}

	return myCompletedOrderSlice
}

// getPizzaSales takes in a channel and received a []viewPizzaSales slice
func getPizzaSales(viewOrderSlice []viewOrder, ch chan<- []viewPizzaSales) {

	defer close(ch)

	viewPizzaSalesSlice := make([]viewPizzaSales, 0)

	for _, val1 := range viewOrderSlice {
		viewOrderItemSlice := val1.ViewOrderItems
		for _, val2 := range viewOrderItemSlice {
			viewPizzaSalesSlice = updatePizzaInSlice(val2, viewPizzaSalesSlice)
		}
	}

	ch <- viewPizzaSalesSlice
}

// updatePizzaInSlice updates the total quantity and total sales of each type of pizzas that were ordered
func updatePizzaInSlice(vOrderItem viewOrderItem, viewPizzaSalesSlice []viewPizzaSales) []viewPizzaSales {
	bUpdate := false
	pizzaPrice, _ := strconv.ParseFloat(vOrderItem.PizzaPrice, 64)
	totalSales := float64(vOrderItem.OrderQty) * pizzaPrice

	if len(viewPizzaSalesSlice) > 0 {
		for i, v := range viewPizzaSalesSlice {
			if v.PizzaNo == vOrderItem.PizzaNo {
				viewPizzaSalesSlice[i].OrderQty = viewPizzaSalesSlice[i].OrderQty + vOrderItem.OrderQty
				viewPizzaSalesSlice[i].TotalSales = viewPizzaSalesSlice[i].TotalSales + totalSales
				viewPizzaSalesSlice[i].STotalSales = fmt.Sprintf("%.2f", viewPizzaSalesSlice[i].TotalSales)
				bUpdate = true
			}
		}
	}

	if !bUpdate {
		viewPizzaSales := viewPizzaSales{vOrderItem.PizzaNo, vOrderItem.PizzaName, vOrderItem.OrderQty, totalSales, fmt.Sprintf("%.2f", totalSales)}
		viewPizzaSalesSlice = append(viewPizzaSalesSlice, viewPizzaSales)
	}

	return viewPizzaSalesSlice
}

// calTotalSales calculates the total sales of all pizzas that were ordered
func calTotalSales(viewPizzaSalesSlice []viewPizzaSales) float64 {

	totalSales := 0.0
	for _, v := range viewPizzaSalesSlice {
		totalSales = totalSales + v.TotalSales
	}

	return totalSales
}

// generateOrderNo takes []OrderItem slice and checks if any OrderItem exists.
// If so, it will increment the global value of newOrderNo by 1.
// A mutex lock is implemented to prevent multiple orders being generated at the same time
func generateOrderNo(orderSlice []order.OrderItem) {

	defer wg.Done()

	// Increment newOrderNo global variable by 1 if there are OrderItem in the slice
	if len(orderSlice) > 0 {
		mutex.Lock()
		runtime.Gosched()
		newOrderNo = newOrderNo + 1
		mutex.Unlock()
	}
}

// printDividerLine prints a divider line to segregate sections for easy viewing
// when printing on the terminal/cmd prompt
func printDividerLine() {
	fmt.Println("------------------------------------------------------------")
}
