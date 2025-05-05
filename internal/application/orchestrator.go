package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"yandexlyceum/internal/database"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Addr                string
	TimeAddition        int
	TimeSubtraction     int
	TimeMultiplications int
	TimeDivisions       int
}

func ConfigFromEnv() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ta, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	if ta == 0 {
		ta = 100
	}
	ts, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	if ts == 0 {
		ts = 100
	}
	tm, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	if tm == 0 {
		tm = 100
	}
	td, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
	if td == 0 {
		td = 100
	}
	return &Config{
		Addr:                port,
		TimeAddition:        ta,
		TimeSubtraction:     ts,
		TimeMultiplications: tm,
		TimeDivisions:       td,
	}
}

type Registration struct {
	Login    string
	Password string
}

type Orchestrator struct {
	Config      *Config
	exprStore   map[string]*Expression
	taskStore   map[string]*Task
	taskQueue   []*Task
	mu          sync.Mutex
	exprCounter int64
	taskCounter int64
	Db          *sql.DB
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		Config:    ConfigFromEnv(),
		exprStore: make(map[string]*Expression),
		taskStore: make(map[string]*Task),
		taskQueue: make([]*Task, 0),
	}
}

type Expression struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Result *float64 `json:"result,omitempty"`
	AST    *ASTNode `json:"-"`
}

type Task struct {
	ID            string   `json:"id"`
	ExprID        string   `json:"-"`
	Arg1          float64  `json:"arg1"`
	Arg2          float64  `json:"arg2"`
	Operation     string   `json:"operation"`
	OperationTime int      `json:"operation_time"`
	Node          *ASTNode `json:"-"`
}

func (o *Orchestrator) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Wrong Method"}`, http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Expression string `json:"expression"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Expression == "" {
		http.Error(w, `{"error":"Invalid Body"}`, http.StatusUnprocessableEntity)
		return
	}
	ast, err := ParseAST(req.Expression)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnprocessableEntity)
		return
	}
	o.mu.Lock()

	claims, ok := r.Context().Value(UserContextKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, `{"error": "Invalid user data"}`, http.StatusInternalServerError)
		return
	}

	userIDFloat, _ := claims["user_id"].(float64)
	userID := int(userIDFloat)

	id, err := database.AddExpression(context.TODO(), userID, req.Expression, o.Db)
	if err != nil {
		http.Error(w, `{"error":"Something went wrong"}`, http.StatusInternalServerError)
		return
	}

	o.exprCounter = int64(id)
	exprID := fmt.Sprintf("%d", o.exprCounter)
	expr := &Expression{
		ID:     exprID,
		Status: "pending",
		AST:    ast,
	}
	o.exprStore[exprID] = expr
	o.ScheduleTasks(expr)
	o.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": exprID})
}

func (o *Orchestrator) ExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Wrong Method"}`, http.StatusMethodNotAllowed)
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	exprs := make([]*Expression, 0, len(o.exprStore))
	for _, expr := range o.exprStore {
		if expr.AST != nil && expr.AST.IsLeaf {
			expr.Status = "completed"
			expr.Result = &expr.AST.Value
		}
		exprs = append(exprs, expr)
	}
	for _, expression := range exprs {
		id, _ := strconv.Atoi(expression.ID)
		database.AddAnswer(context.TODO(), id, float64(*expression.Result), o.Db)
	}

	claims, ok := r.Context().Value(UserContextKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, `{"error": "Invalid user data"}`, http.StatusInternalServerError)
		return
	}

	userIDFloat, _ := claims["user_id"].(float64)
	userID := int(userIDFloat)

	answ, err := database.GetExpressions(userID, o.Db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if len(answ) == 0 {
		http.Error(w, `{"error":"No expressions"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": answ})
}

func (o *Orchestrator) ExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Wrong Method"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/api/v1/expressions/"):]

	claims, ok := r.Context().Value(UserContextKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid user data", http.StatusInternalServerError)
		return
	}
	userIDFloat, _ := claims["user_id"].(float64)
	userID := int(userIDFloat)

	intId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid id"}`, http.StatusInternalServerError)
		return
	}

	res_expr, err := database.GetExpressionByID(context.TODO(), userID, intId, o.Db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": res_expr})
}

func (o *Orchestrator) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Wrong Method"}`, http.StatusMethodNotAllowed)
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.taskQueue) == 0 {
		http.Error(w, `{"error":"No task available"}`, http.StatusNotFound)
		return
	}
	task := o.taskQueue[0]
	o.taskQueue = o.taskQueue[1:]
	if expr, exists := o.exprStore[task.ExprID]; exists {
		expr.Status = "in_progress"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (o *Orchestrator) PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Wrong Method"}`, http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ID == "" {
		http.Error(w, `{"error":"Invalid Body"}`, http.StatusUnprocessableEntity)
		return
	}
	o.mu.Lock()
	task, ok := o.taskStore[req.ID]
	if !ok {
		o.mu.Unlock()
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	}
	task.Node.IsLeaf = true
	task.Node.Value = req.Result
	delete(o.taskStore, req.ID)
	if expr, exists := o.exprStore[task.ExprID]; exists {
		o.ScheduleTasks(expr)
		if expr.AST.IsLeaf {
			expr.Status = "completed"
			expr.Result = &expr.AST.Value
		}
	}
	o.mu.Unlock()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"Result accepted"}`))
}

func (o *Orchestrator) ScheduleTasks(expr *Expression) {
	var traverse func(node *ASTNode)
	traverse = func(node *ASTNode) {
		if node == nil || node.IsLeaf {
			return
		}
		traverse(node.Left)
		traverse(node.Right)
		if node.Left != nil && node.Right != nil && node.Left.IsLeaf && node.Right.IsLeaf {
			if !node.TaskScheduled {
				o.taskCounter++
				taskID := fmt.Sprintf("%d", o.taskCounter)
				var opTime int
				switch node.Operator {
				case "+":
					opTime = o.Config.TimeAddition
				case "-":
					opTime = o.Config.TimeSubtraction
				case "*":
					opTime = o.Config.TimeMultiplications
				case "/":
					opTime = o.Config.TimeDivisions
				default:
					opTime = 100
				}
				task := &Task{
					ID:            taskID,
					ExprID:        expr.ID,
					Arg1:          node.Left.Value,
					Arg2:          node.Right.Value,
					Operation:     node.Operator,
					OperationTime: opTime,
					Node:          node,
				}
				node.TaskScheduled = true
				o.taskStore[taskID] = task
				o.taskQueue = append(o.taskQueue, task)
			}
		}
	}
	traverse(expr.AST)
}

func (o *Orchestrator) AgentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		o.GetTaskHandler(w, r)
	} else if r.Method == http.MethodPost {
		o.PostTaskHandler(w, r)
	} else {
		http.Error(w, `{"error":"Wrong Method"}`, http.StatusMethodNotAllowed)
	}
}

func (o *Orchestrator) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Wrong Method"}`, http.StatusMethodNotAllowed)
	} else if r.Method == http.MethodPost {
		var data Registration
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, `{"error":"Error decoding JSON"}`, http.StatusBadRequest)
			return
		}
		generatedPassword, err := GeneratePassword(data.Password)
		if err != nil {
			http.Error(w, `{"error":"Error encoding password"}`, http.StatusInternalServerError)
		}
		err = database.InsertUsers(context.TODO(), data.Login, generatedPassword, o.Db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			log.Printf("User with login '%s' successfully added\n", data.Login)
			http.Error(w, "successful registration", http.StatusOK)
		}
	}
}

func (o *Orchestrator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Wrong Method"}`, http.StatusMethodNotAllowed)
	} else if r.Method == http.MethodPost {
		var data Registration
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, `{"error":"Error decoding JSON"}`, http.StatusBadRequest)
			return
		}
		id := database.GetUserID(context.TODO(), data.Login, o.Db)
		authentificated := database.IsAuth(context.TODO(), data.Login, data.Password, o.Db)
		if !authentificated {
			http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
			return
		}
		log.Printf("Successfull login for %s\n", data.Login)
		generatedToken, err := GenerateJWT(id, data.Login)
		if err != nil {
			http.Error(w, `{"error":"Error generating jwt"}`, http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    generatedToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   86400,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": generatedToken,
		})
	}
}
func (o *Orchestrator) RunServer() error {
	mux := http.NewServeMux()
	db, err := database.InitDB("finalTask.db")
	o.Db = db
	defer db.Close()
	if err != nil {
		log.Println("Error opening database")
	}
	mux.HandleFunc("/api/v1/register", o.RegisterHandler)
	mux.HandleFunc("/api/v1/login", o.LoginHandler)
	mux.HandleFunc("/api/v1/calculate", AuthMiddleware(o.CalculateHandler))
	mux.HandleFunc("/api/v1/expressions", AuthMiddleware(o.ExpressionsHandler))
	mux.HandleFunc("/api/v1/expressions/", AuthMiddleware(o.ExpressionByIDHandler))
	mux.HandleFunc("/internal/task", o.AgentHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"Not Found"}`, http.StatusNotFound)
	})
	go func() {
		for {
			time.Sleep(2 * time.Second)
		}
	}()
	return http.ListenAndServe(":"+o.Config.Addr, mux)
}
