package server

import (
	"fmt"
	"net/http"
	"strconv"

	pizza "GoInAction2Assignment/pkg/pizza"

	"github.com/sirupsen/logrus"
)

// generatePizzaNo increments the global pizza no for new pizza creation
func generatePizzaNo() {

	// Increment PizzaNo global variable by 1
	newPizzaNo = newPizzaNo + 1
}

// addpizza is a handler func to add a new pizza.
// Redirects to index page if user has not login.
func addpizza(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := "" // To display message to the user on the client
	pizzaName := ""
	inputPrice := ""

	if req.Method == http.MethodPost {
		pizzaName = req.FormValue("pizzaname")

		if pizzaName != "" {
			inputPrice = req.FormValue("pizzaprice")

			pizzaPrice, err := validatePrice(inputPrice)

			if err != nil || pizzaPrice == 0 {
				clientMsg = "Please enter a valid Pizza Price"
			} else {
				generatePizzaNo()

				pizzaNo := newPizzaNo
				pizzaList.AddPizza(pizzaNo, pizzaName, pizzaPrice)

				clientMsg = fmt.Sprintf("%s @ $%.2f added successfully.\n", pizzaName, pizzaPrice)

				log.WithFields(logrus.Fields{
					"userName":   myUser.UserName,
					"pizzaNo":    pizzaNo,
					"pizzaName":  pizzaName,
					"pizzaPrice": pizzaPrice,
				}).Info("pizza added successfully")
			}
		} else {
			clientMsg = "Please enter Pizza Name"
		}
	}

	data := struct {
		User       user
		PizzaName  string
		PizzaPrice string
		ClientMsg  string
	}{
		myUser,
		pizzaName,
		inputPrice,
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "addpizza.gohtml", data)
}

// editpizza is a handler func to edit an existing pizza.
// Selected pizza cannot be edited if an order exists with it.
// Redirects to index page if user has not login.
func editpizza(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := ""
	pizzaNo := 0
	fPizzaName := ""
	fPizzaPrice := ""
	pizzaPrice := 0.0

	viewPizzaSlice := make([]viewPizza, 0)
	pizzaSlice, err := pizzaList.GetAllPizza()

	if err != nil {
		clientMsg = "There are no pizza on the menu today"
		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Error(err)

	}

	if req.Method == http.MethodPost {
		fPizzaNo := req.FormValue("pizzano")
		pizzaNo, _ := strconv.Atoi(fPizzaNo)

		bPizzaInOrder, _ := checkPizzaInOrder(pizzaNo)
		if bPizzaInOrder {
			clientMsg = "Orders have been made on the selected pizza. You are not allowed to edit this pizza."
			log.WithFields(logrus.Fields{
				"userName": myUser.UserName,
				"pizzaNo":  pizzaNo,
			}).Warn("orders made on the selected pizza - cannot edit")
		} else {
			selectedPizza, err := pizzaList.SearchPizza(pizzaNo)

			if err != nil {
				clientMsg = "Cannot edit this pizza"

				log.WithFields(logrus.Fields{
					"userName": myUser.UserName,
					"pizzaNo":  pizzaNo,
				}).Error(err)
			} else {
				fPizzaName = req.FormValue("pizzaname")
				fPizzaPrice = req.FormValue("pizzaprice")

				if fPizzaName == "" {
					fPizzaName = selectedPizza.PizzaName
				}

				if fPizzaPrice == "" {
					fPizzaPrice = fmt.Sprintf("%.2f", selectedPizza.PizzaPrice)
				}

				if fPizzaName == selectedPizza.PizzaName && fPizzaPrice == fmt.Sprintf("%.2f", selectedPizza.PizzaPrice) {
					clientMsg = "No changes made on the selected pizza"

					log.WithFields(logrus.Fields{
						"userName":   myUser.UserName,
						"pizzaNo":    pizzaNo,
						"pizzaName":  fPizzaName,
						"pizzaPrice": fPizzaPrice,
					}).Info("no changes made on the selected pizza")

				} else {
					pizzaPrice, err := validatePrice(fPizzaPrice)

					if err != nil || pizzaPrice == 0 {
						clientMsg = "Please enter a valid Pizza Price."
					} else {
						pizzaList.EditPizza(pizzaNo, fPizzaName, pizzaPrice)

						clientMsg = fmt.Sprintf("%s @ $%s updated successfully.\n", fPizzaName, fPizzaPrice)

						log.WithFields(logrus.Fields{
							"userName":   myUser.UserName,
							"pizzaNo":    pizzaNo,
							"pizzaName":  fPizzaName,
							"pizzaPrice": fPizzaPrice,
						}).Info("selected pizza updated successfully")
					}
				}

				for _, v := range pizzaSlice {
					if pizzaNo == v.PizzaNo {
						viewPizza := viewPizza{pizzaNo, fPizzaName, pizzaPrice, fPizzaPrice, "Selected"}
						viewPizzaSlice = append(viewPizzaSlice, viewPizza)
					} else {
						viewPizza := viewPizza{v.PizzaNo, v.PizzaName, v.PizzaPrice, fmt.Sprintf("%.2f", v.PizzaPrice), ""}
						viewPizzaSlice = append(viewPizzaSlice, viewPizza)
					}
				}
			}
		}
	}

	if len(viewPizzaSlice) == 0 {
		for _, v := range pizzaSlice {
			viewPizza := viewPizza{v.PizzaNo, v.PizzaName, v.PizzaPrice, fmt.Sprintf("%.2f", v.PizzaPrice), ""}
			viewPizzaSlice = append(viewPizzaSlice, viewPizza)
		}
	}

	data := struct {
		User           user
		ViewPizzaSlice []viewPizza
		CntPizza       int
		PizzaNo        int
		PizzaName      string
		PizzaPrice     string
		ClientMsg      string
	}{
		myUser,
		viewPizzaSlice,
		len(viewPizzaSlice),
		pizzaNo,
		fPizzaName,
		fPizzaPrice,
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "editpizza.gohtml", data)
}

// deletepizza is a handler func to delete an existing pizza.
// Selected pizza cannot be deleted if an order exists with it.
// Redirects to index page if user has not login.
func deletepizza(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := ""
	pizzaNo := 0

	if req.Method == http.MethodPost {
		fPizzaNo := req.FormValue("pizzano")
		pizzaNo, _ := strconv.Atoi(fPizzaNo)

		bPizzaInOrder, _ := checkPizzaInOrder(pizzaNo)
		if bPizzaInOrder {
			clientMsg = "Orders have been made on the selected pizza. You are not allowed to delete this pizza."
			log.WithFields(logrus.Fields{
				"userName": myUser.UserName,
				"pizzaNo":  pizzaNo,
			}).Warn("orders made on the selected pizza - cannot delete")
		} else {
			selectedPizza, err := pizzaList.SearchPizza(pizzaNo)

			if err != nil {
				clientMsg = "Cannot delete this pizza."

				log.WithFields(logrus.Fields{
					"userName": myUser.UserName,
					"pizzaNo":  pizzaNo,
				}).Error(err)
			} else {
				pizzaList.DeletePizza(pizzaNo)
				clientMsg = fmt.Sprintf("%s @ $%.2f deleted successfully.\n", selectedPizza.PizzaName, selectedPizza.PizzaPrice)

				log.WithFields(logrus.Fields{
					"userName":   myUser.UserName,
					"pizzaNo":    pizzaNo,
					"pizzaName":  selectedPizza.PizzaName,
					"pizzaPrice": selectedPizza.PizzaPrice,
				}).Info("selected pizza deleted successfully")
			}
		}
	}

	pizzaSlice, err := pizzaList.GetAllPizza()

	if err != nil {
		clientMsg = "There are no pizza on the menu today. "

		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Info("no pizza in the menu")
	}

	data := struct {
		User           user
		ViewPizzaSlice []pizza.Pizza
		CntPizza       int
		PizzaNo        int
		ClientMsg      string
	}{
		myUser,
		pizzaSlice,
		len(pizzaSlice),
		pizzaNo,
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "deletepizza.gohtml", data)
}

// viewpizza is a handler func to view all pizzas.
// Redirects to index page if user has not login.
func viewpizza(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := ""

	pizzaSlice, err := pizzaList.GetAllPizza()
	viewPizzaSlice := make([]viewPizza, 0)

	if err != nil {
		clientMsg = "There are no pizza on the menu today. "

		log.WithFields(logrus.Fields{
			"userName": myUser.UserName,
		}).Info(err)
	} else {
		for _, v := range pizzaSlice {
			viewPizza := viewPizza{v.PizzaNo, v.PizzaName, v.PizzaPrice, fmt.Sprintf("%.2f", v.PizzaPrice), ""}
			viewPizzaSlice = append(viewPizzaSlice, viewPizza)
		}
	}

	data := struct {
		User           user
		ViewPizzaSlice []viewPizza
		CntPizza       int
		ClientMsg      string
	}{
		myUser,
		viewPizzaSlice,
		len(viewPizzaSlice),
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "viewpizza.gohtml", data)
}

// validatePrice will parse the price string as float and returns the parsed value
func validatePrice(price string) (float64, error) {
	retValue, err := strconv.ParseFloat(price, 64)

	if err != nil {
		return retValue, err
	}

	return retValue, nil
}

// checkPizzaInOrder checks whether a pizza exists in any orders.
// Returns true if found, otherwise returns false.
func checkPizzaInOrder(pizzaNo int) (bool, error) {
	// If there are no orders made means there are no pizza in any order, thus return false
	if orderQueue.IsEmpty() {
		return false, nil
	} else {
		return orderQueue.SearchPizzaInOrder(pizzaNo)
	}
}
