package server

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"
	"unicode"

	"github.com/kennygrant/sanitize"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// signup is a handler func to create a new user account.
// Validates user information and creates a new user account.
func signup(res http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	var myUser user

	clientMsg := "" // To display message to the user on the client
	username := ""
	password := ""
	cmfpassword := ""
	firstname := ""
	lastname := ""
	email := ""

	// Process form submission
	if req.Method == http.MethodPost {
		// Get form values and sanitize the strings
		username = sanitize.Accents(req.FormValue("username"))
		password = sanitize.Accents(req.FormValue("password"))
		cmfpassword = sanitize.Accents(req.FormValue("cmfpassword"))
		firstname = sanitize.Accents(req.FormValue("firstname"))
		lastname = sanitize.Accents(req.FormValue("lastname"))
		email = sanitize.Accents(req.FormValue("email"))

		// Validates the input fields from the user
		if err := validateUserInput(username, password, cmfpassword, firstname, lastname, email); err != nil {
			clientMsg = "ERROR: " + err.Error()
			log.Error(err)
		} else {
			// Check if username exist i.e. exist means it is already taken
			if _, ok := mapUsers[username]; ok {
				clientMsg = "ERROR: " + "username already taken"
				log.Error("[" + username + "] username already taken")
			} else {
				// Create a new session token
				id, _ := uuid.NewV4()
				// Set an expiry time of 120 seconds for the cookie, the same as the cache
				sessionToken := &http.Cookie{
					Name:     "sessionToken",
					Value:    id.String(),
					Expires:  time.Now().Add(120 * time.Second),
					HttpOnly: true,
					Path:     "/",
					Domain:   "localhost",
					Secure:   true,
				}
				// Set the session token as a cookie on the client
				http.SetCookie(res, sessionToken)

				// Store the session token in a map on the server
				mapSessions[sessionToken.Value] = username

				// Hashed the password from user input before saving
				bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
				if err != nil {
					clientMsg = "WARNING: " + "internal server error"
					log.Warn("internal server error")
				} else {
					myUser = user{username, bPassword, firstname, lastname, email, false, time.Now(), time.Now(), time.Now(), time.Now()}
					mapUsers[username] = myUser

					log.WithFields(logrus.Fields{
						"userName": username,
					}).Infof("[%s] user account created successfully", username)

					// Redirect to the main index page
					http.Redirect(res, req, "/", http.StatusSeeOther)
					return
				}
			}
		}
	}

	data := struct {
		User        user
		UserName    string
		Password    string
		CmfPassword string
		FirstName   string
		LastName    string
		Email       string
		ClientMsg   string
	}{
		myUser,
		username,
		password,
		cmfpassword,
		firstname,
		lastname,
		email,
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "signup.gohtml", data)
}

// edituser is a handler func to edit user account information.
// Redirects to index page if user has not login.
// Validates user input and updates the information.
func edituser(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := "" // To display message to the user on the client

	// Set current user information to be display on the form upon first form load
	username := myUser.UserName
	password := ""
	cmfpassword := ""
	firstname := myUser.FirstName
	lastname := myUser.LastName
	email := myUser.Email
	isAdmin := myUser.IsAdmin
	createdDT := myUser.CreatedDT
	currentLoginDT := myUser.CurrentLoginDT
	lastLoginDT := myUser.LastLoginDT

	// Process the form submission
	if req.Method == http.MethodPost {
		// Get form values and sanitize the strings
		username = sanitize.Accents(req.FormValue("username"))
		password = sanitize.Accents(req.FormValue("password"))
		cmfpassword = sanitize.Accents(req.FormValue("cmfpassword"))
		firstname = sanitize.Accents(req.FormValue("firstname"))
		lastname = sanitize.Accents(req.FormValue("lastname"))
		email = sanitize.Accents(req.FormValue("email"))

		// Validates the input fields from the user
		if err := validateUserInput(username, password, cmfpassword, firstname, lastname, email); err != nil {
			clientMsg = "ERROR: " + err.Error()
			log.Error(err)
		} else {
			// Hashed the password from user input before saving
			bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				clientMsg = "WARNING: " + "internal server error"
				log.WithFields(logrus.Fields{
					"userName": myUser.UserName,
				}).Warn("internal server error")

			} else {
				// Update user info in a new struct and set LastModifiedDT to current date/time
				myUser = user{username, bPassword, firstname, lastname, email, isAdmin, createdDT, time.Now(), currentLoginDT, lastLoginDT}
				// Update map user struct to new user struct
				mapUsers[username] = myUser

				log.WithFields(logrus.Fields{
					"userName": username,
				}).Infof("[%s] user account updated successfully", username)

				// Redirect to the main index page
				http.Redirect(res, req, "/", http.StatusSeeOther)
				return
			}
		}
	}

	data := struct {
		User        user
		UserName    string
		Password    string
		CmfPassword string
		FirstName   string
		LastName    string
		Email       string
		ClientMsg   string
	}{
		myUser,
		username,
		password,
		cmfpassword,
		firstname,
		lastname,
		email,
		clientMsg,
	}
	tpl.ExecuteTemplate(res, "edituser.gohtml", data)
}

// deleteuser is a handler func to delete user account. Redirects to index page if user has not login.
// Only admin user has access to delete users and admin is not allowed to delete oneself.
func deleteuser(res http.ResponseWriter, req *http.Request) {
	myUser := getUser(res, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	clientMsg := ""

	if req.Method == http.MethodPost {
		// Username is retrieved from dropdownlist box thus there is no need to validate for valid username
		username := req.FormValue("username")

		// Check if current user is admin. If so, prompt that account cannot be deleted
		if myUser.UserName == username && myUser.IsAdmin {
			clientMsg = "WARNING: " + "admin account cannot be deleted"

			log.WithFields(logrus.Fields{
				"userName": myUser.UserName,
			}).Warn("admin account cannot be deleted")

		} else {
			// Delete user from server map
			delete(mapUsers, username)

			clientMsg = "[" + username + "] user account deleted successfully. "

			log.WithFields(logrus.Fields{
				"userName": myUser.UserName,
			}).Infof("[%s] user account deleted successfully", username)
		}
	}

	data := struct {
		User      user
		MapUsers  map[string]user
		CntUsers  int
		ClientMsg string
	}{
		myUser,
		mapUsers,
		len(mapUsers),
		clientMsg,
	}

	tpl.ExecuteTemplate(res, "deleteuser.gohtml", data)
}

// logout func is a handler to logout the current user. Redirects to index page if user has not login.
// Otherwise, delete session token from server and client, then redirects to index page.
func logout(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	sessionToken, _ := req.Cookie("sessionToken")

	// Get username before session is deleted
	username := mapSessions[sessionToken.Value]

	// Delete the session token from the server
	delete(mapSessions, sessionToken.Value)
	// Remove the cookie from the client
	sessionToken = &http.Cookie{
		Name:   "sessionToken",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(res, sessionToken)

	log.WithFields(logrus.Fields{
		"userName": username,
	}).Infof("[%s] user logout successfully", username)

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

// getUser func gets the current user. Checks for valid session token.
// Add a new session token cookie to the client if one is not found.
// Return user struct if found.
func getUser(res http.ResponseWriter, req *http.Request) user {
	// Get current session cookie
	sessionToken, err := req.Cookie("sessionToken")
	// No session token found
	if err != nil {
		// Add new session token cookie
		id, _ := uuid.NewV4()
		// Set an expiry time of 120 seconds for the cookie, the same as the cache
		sessionToken = &http.Cookie{
			Name:     "sessionToken",
			Value:    id.String(),
			Expires:  time.Now().Add(120 * time.Second),
			HttpOnly: true,
			Path:     "/",
			Domain:   "localhost",
			Secure:   true,
		}
	}
	http.SetCookie(res, sessionToken)

	// If the user exists already, get user
	var myUser user
	if username, ok := mapSessions[sessionToken.Value]; ok {
		myUser = mapUsers[username]
	}

	return myUser
}

// alreadyLoggedIn func checks if a user has already logged in. Checks for valid session token.
// Returns true if already logged in, false otherwise.
func alreadyLoggedIn(req *http.Request) bool {
	sessionToken, err := req.Cookie("sessionToken")
	if err != nil {
		return false
	}
	// Get username from session map
	username := mapSessions[sessionToken.Value]
	_, ok := mapUsers[username]
	return ok
}

// validateUserInput func checks if a user has already logged in. Checks for valid session token.
func validateUserInput(userName string, password string, cmfPassword string, firstName string, lastName string, email string) error {
	// Validate username
	if userName == "" {
		return errors.New("username cannot be blank")
	} else if len(userName) < minUserName || len(userName) > maxUserName {
		return fmt.Errorf("username must be between %d - %d characters", minUserName, maxUserName)
	}

	// Validate password
	if password == "" {
		return errors.New("password cannot be blank")
	} else if len(password) < minPassword || len(password) > maxPassword {
		return fmt.Errorf("password must be between %d - %d characters", minPassword, maxPassword)
	} else if err := validatePassword(password); err != nil {
		return err
	}

	// Validate confirm password
	if cmfPassword == "" {
		return errors.New("confirm password cannot be blank")
	} else if cmfPassword != password {
		return errors.New("confirm password must be the same as password")
	}

	// Validate first name
	if firstName == "" {
		return errors.New("first name cannot be blank")
	}

	// Validate last name
	if lastName == "" {
		return errors.New("last name cannot be blank")
	}

	// Validate email (email is not mandatory)
	if email != "" && !isValidEmail(email) {
		return errors.New("invalid email")
	}

	return nil
}

// validatePassword validates that the input user password must contain as least
// one upper case, lower case, numeric and special characters.
func validatePassword(password string) error {

next:
	for name, classes := range map[string][]*unicode.RangeTable{
		"upper case": {unicode.Upper, unicode.Title},
		"lower case": {unicode.Lower},
		"numeric":    {unicode.Number, unicode.Digit},
		"special":    {unicode.Space, unicode.Symbol, unicode.Punct, unicode.Mark},
	} {
		for _, r := range password {
			if unicode.IsOneOf(classes, r) {
				continue next
			}
		}
		return fmt.Errorf("password must have at least one %s character", name)
	}

	return nil
}

// isValidEmail validates if the string parameter is a valid email using regexp
func isValidEmail(e string) bool {
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
