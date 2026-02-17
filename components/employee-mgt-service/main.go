package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Employee represents an employee record
type Employee struct {
	ID         int    `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// EmployeeStore manages employee data in memory
type EmployeeStore struct {
	sync.RWMutex
	employees map[int]Employee
	nextID    int
}

// NewEmployeeStore creates a new employee store
func NewEmployeeStore() *EmployeeStore {
	return &EmployeeStore{
		employees: make(map[int]Employee),
		nextID:    1,
	}
}

// Add adds a new employee to the store
func (s *EmployeeStore) Add(emp Employee) Employee {
	s.Lock()
	defer s.Unlock()
	emp.ID = s.nextID
	s.nextID++
	s.employees[emp.ID] = emp
	return emp
}

// GetAll returns all employees
func (s *EmployeeStore) GetAll() []Employee {
	s.RLock()
	defer s.RUnlock()
	employees := make([]Employee, 0, len(s.employees))
	for _, emp := range s.employees {
		employees = append(employees, emp)
	}
	return employees
}

// GetByID returns an employee by ID
func (s *EmployeeStore) GetByID(id int) (Employee, bool) {
	s.RLock()
	defer s.RUnlock()
	emp, exists := s.employees[id]
	return emp, exists
}

var store = NewEmployeeStore()

func main() {
	mux := http.NewServeMux()

	// Register routes under /employee-mgt context
	mux.HandleFunc("/employee-mgt/employees", employeesHandler)
	mux.HandleFunc("/employee-mgt/employees/", employeeByIDHandler)

	log.Println("Employee Management Service starting on port 9090...")
	log.Println("Context path: /employee-mgt")
	log.Fatal(http.ListenAndServe(":9090", mux))
}

// employeesHandler handles GET (list all) and POST (create) for employees
func employeesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		getAllEmployees(w, r)
	case http.MethodPost:
		addEmployee(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// employeeByIDHandler handles GET for a specific employee
func employeeByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/employee-mgt/employees/")
	id, err := strconv.Atoi(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	getEmployeeByID(w, r, id)
}

// getAllEmployees returns all employees
func getAllEmployees(w http.ResponseWriter, _ *http.Request) {
	employees := store.GetAll()
	json.NewEncoder(w).Encode(employees)
}

// addEmployee creates a new employee
func addEmployee(w http.ResponseWriter, r *http.Request) {
	var emp Employee
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if emp.FirstName == "" || emp.LastName == "" || emp.Email == "" {
		writeError(w, http.StatusBadRequest, "FirstName, LastName, and Email are required")
		return
	}

	created := store.Add(emp)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// getEmployeeByID returns a specific employee by ID
func getEmployeeByID(w http.ResponseWriter, _ *http.Request, id int) {
	emp, exists := store.GetByID(id)
	if !exists {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}
	json.NewEncoder(w).Encode(emp)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
		Code:    code,
	})
}
