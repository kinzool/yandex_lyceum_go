package application

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"yandexlyceum/yandex_lyceum_go/pkg/calculation"
)

type Request struct {
	Expression string `json:"expression"`
}

type Answer struct {
	Id int `json:"id"`
}

type ExpressionAnswer struct {
	Id     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result,omitempty"`
}

type ExpressionsListAnswer struct {
	Expression []ExpressionAnswer `json:"expressions"`
}

type ExpressionIdAnswer struct {
	Expression ExpressionAnswer `json:"expression"`
}

var req Request

var countGoroutines int = 3
var expressionId int = 0
var expression string
var DistrubitedAnswer []ExpressionAnswer

func CalculateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&req)
		err := calculation.AddExpression(req.Expression)
		if err != nil {
			if errors.Is(err, calculation.ErrorInvalidExpression) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	expression = req.Expression
	w.WriteHeader(http.StatusCreated)
	log.Printf("Выражение %s добавлено\n", req.Expression)
	expressionId++
	resp := Answer{Id: expressionId}
	answ, _ := json.Marshal(resp)
	w.Write(answ)
}

func ExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	distributedExpression, err := calculation.ParseExpression(expression)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, expr := range distributedExpression {
		_, err := calculation.CalculateSimpleExpression(expr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		DistrubitedAnswer = append(DistrubitedAnswer, ExpressionAnswer{Id: len(DistrubitedAnswer) + 1, Status: "pending"})
	}
	resp := ExpressionsListAnswer{Expression: DistrubitedAnswer}
	answ, _ := json.Marshal(resp)
	w.Write(answ)
}

func IdHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	splitedPath := strings.Split(path, "/")
	strId := splitedPath[len(splitedPath)-1]
	id, err := strconv.Atoi(strId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if id > len(DistrubitedAnswer) || id < 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	resp := ExpressionIdAnswer{DistrubitedAnswer[id-1]}
	answ, _ := json.Marshal(resp)
	w.Write(answ)
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	// 	switch r.Method {
	// 	case http.MethodGet:
	// 		var wg sync.WaitGroup
	// 		for i := 0; i < countGoroutines; i++ {
	// 			wg.Add(1)
	// 			go func() {
	// 				defer wg.Done()
	// 				expressions, _ := calculation.ParseExpression(expression)
	// 				res, _ := calculation.CalculateSimpleExpression(expressions[i])

	// 			}()
	// 		}

	// 	}

}

func RunApplication() {
	defer log.Println("Работа сервера прекращена")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", CalculateMiddleware(CalculateHandler))
	mux.HandleFunc("/api/v1/expressions", ExpressionsHandler)
	mux.HandleFunc("/api/v1/expressions/", IdHandler)
	mux.HandleFunc("/internal/task", TaskHandler)
	log.Println("Сервер запущен")
	http.ListenAndServe(":8080", mux)
}
