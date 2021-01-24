package main

import (
   "gorilla/mux"
   "net/http"
   "fmt"
   "strconv"
)

const GETALL string = "GETALL"
const GETONE string = "GETONE"
const POST string   = "POST"
const PUT string    = "PUT"
const DELETE string = "DELETE"

type clichePair struct {
   Id      int
   Cliche  string
   Counter string
}

// Message sent to goroutine that accesses the requested resource.
type crudRequest struct {
   verb     string
   cp       *clichePair
   id       int
   cliche   string
   counter  string
   confirm  chan string
}

var clichesList = []*clichePair{}
var masterId = 1
var crudRequests chan *crudRequest

// GET /
// GET /cliches
func ClichesAll(res http.ResponseWriter, req *http.Request) {
   cr := &crudRequest{verb: GETALL, confirm: make(chan string)}
   completeRequest(cr, res, "read all")
}

// GET /cliches/id
func ClichesOne(res http.ResponseWriter, req *http.Request) {
   id := getIdFromRequest(req)
   cr := &crudRequest{verb: GETONE, id: id, confirm: make(chan string)}
   completeRequest(cr, res, "read one")
}

// POST /cliches
func ClichesCreate(res http.ResponseWriter, req *http.Request) {
   cliche, counter := getDataFromRequest(req)
   cp := new(clichePair)
   cp.Cliche = cliche
   cp.Counter = counter
   cr := &crudRequest{verb: POST, cp: cp, confirm: make(chan string)}
   completeRequest(cr, res, "create")
}

// PUT /cliches/id
func ClichesEdit(res http.ResponseWriter, req *http.Request) {
   id := getIdFromRequest(req)
   cliche, counter := getDataFromRequest(req)
   cr := &crudRequest{verb: PUT, id: id, cliche: cliche, counter: counter, confirm: make(chan string)}
   completeRequest(cr, res, "edit")
}

// DELETE /cliches/id
func ClichesDelete(res http.ResponseWriter, req *http.Request) {
   id := getIdFromRequest(req)
   cr := &crudRequest{verb: DELETE, id: id, confirm: make(chan string)}
   completeRequest(cr, res, "delete")
}

func completeRequest(cr *crudRequest, res http.ResponseWriter, logMsg string) {
   crudRequests<-cr
   msg := <-cr.confirm
   res.Write([]byte(msg))
   logIt(logMsg)
}

func main() {
   populateClichesList()

   // From now on, this gorountine alone accesses the clichesList.
   crudRequests = make(chan *crudRequest, 8)
   go func() { // resource manager
      for {
         select {
         case req := <-crudRequests:
         if req.verb == GETALL {
            req.confirm<-readAll()
         } else if req.verb == GETONE {
            req.confirm<-readOne(req.id)
         } else if req.verb == POST {
            req.confirm<-addPair(req.cp)
         } else if req.verb == PUT {
            req.confirm<-editPair(req.id, req.cliche, req.counter)
         } else if req.verb == DELETE {
            req.confirm<-deletePair(req.id)
         }
      }
   }()
   startServer()
}

func startServer() {
   router := mux.NewRouter()

   // Dispatch map for CRUD operations.
   router.HandleFunc("/", ClichesAll).Methods("GET")
   router.HandleFunc("/cliches", ClichesAll).Methods("GET")
   router.HandleFunc("/cliches/{id:[0-9]+}", ClichesOne).Methods("GET")

   router.HandleFunc("/cliches", ClichesCreate).Methods("POST")
   router.HandleFunc("/cliches/{id:[0-9]+}", ClichesEdit).Methods("PUT")
   router.HandleFunc("/cliches/{id:[0-9]+}", ClichesDelete).Methods("DELETE")

   http.Handle("/", router) // enable the router

   // Start the server.
   port := ":8888"
   fmt.Println("\nListening on port " + port)
   http.ListenAndServe(port, router); // mux.Router now in play
}

// Return entire list to requester.
func readAll() string {
   msg := "\n"
   for _, cliche := range clichesList {
      next := strconv.Itoa(cliche.Id) + ": " + cliche.Cliche + "  " + cliche.Counter + "\n"
      msg += next
   }
   return msg
}

// Return specified clichePair to requester.
func readOne(id int) string {
   msg := "\n" + "Bad Id: " + strconv.Itoa(id) + "\n"

   index := findCliche(id)
   if index >= 0 {
      cliche := clichesList[index]
      msg = "\n" + strconv.Itoa(id) + ": " + cliche.Cliche + "  " + cliche.Counter + "\n"
   }
   return msg
}

// Create a new clichePair and add to list
func addPair(cp *clichePair) string {
   cp.Id = masterId
   masterId++
   clichesList = append(clichesList, cp)
   return "\nCreated: " + cp.Cliche + " " + cp.Counter + "\n"
}

// Edit an existing clichePair
func editPair(id int, cliche string, counter string) string {
   msg := "\n" + "Bad Id: " + strconv.Itoa(id) + "\n"
   index := findCliche(id)
   if index >= 0 {
      clichesList[index].Cliche = cliche
      clichesList[index].Counter = counter
      msg = "\nCliche edited: " + cliche + " " + counter + "\n"
   }
   return msg
}

// Delete a clichePair
func deletePair(id int) string {
   idStr := strconv.Itoa(id)
   msg := "\n" + "Bad Id: " + idStr + "\n"
   index := findCliche(id)
   if index >= 0 {
      clichesList = append(clichesList[:index], clichesList[index + 1:]...)
      msg = "\nCliche " + idStr + " deleted\n"
   }
   return msg
}

//*** utility functions
func findCliche(id int) int {
   for i := 0; i < len(clichesList); i++ {
      if id == clichesList[i].Id {
         return i;
      }
   }
   return -1 // not found
}

func getIdFromRequest(req *http.Request) int {
   vars := mux.Vars(req)
   id, _ := strconv.Atoi(vars["id"])
   return id
}

func getDataFromRequest(req *http.Request) (string, string) {
   // Extract the user-provided data for the new clichePair
   req.ParseForm()
   form := req.Form
   cliche := form["cliche"][0]    // 1st and only member of a list
   counter := form["counter"][0]  // ditto
   return cliche, counter
}

func logIt(msg string) {
   fmt.Println(msg)
}

func populateClichesList() {
   var cliches = []string {
      "Out of sight, out of mind.",
      "A penny saved is a penny earned.",
      "He who hesitates is lost.",
   }
   var counterCliches = []string {
      "Absence makes the heart grow fonder.",
      "Penny-wise and dollar-foolish.",
      "Look before you leap.",
   }

   for i := 0; i < len(cliches); i++ {
      cp := new(clichePair)
      cp.Id = masterId
      masterId++
      cp.Cliche = cliches[i]
      cp.Counter = counterCliches[i]
      clichesList = append(clichesList, cp)
   }
}
