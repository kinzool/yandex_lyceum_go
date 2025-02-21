package application

import (
	"encoding/json"
	"errors"
	"net/http"
	"yandexlyceum/yandex_lyceum_go/pkg/calculation"
)

type Request struct {
	Expression string `json:"expression"`
}

type Answer struct {
	Id int `json:"id"`
}

var req Request

var indexes = []int{0}

func CalculateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&req)
		_, err := calculation.Calc(req.Expression)
		if err != nil {
			if errors.Is(err, calculation.ErrorInvalidExpression) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		newId := indexes[len(indexes)] + 1
		indexes = append(indexes, newId)
		resp := Answer{Id: newId}
		answ, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusOK)
		w.Write(answ)
	})
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	result, _ := calculation.Calc(req.Expression)
	response := Answer{Result: result}
	answ, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(answ)
}

func ExpressionsHandler(w http.ResponseWriter, r *http.Request) {}

func IdHandler(w http.ResponseWriter, r *http.Request) {}

func TaskHandler(w http.ResponseWriter, r *http.Request) {}

func RunApplication() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", CalculateMiddleware(CalculateHandler))
	mux.HandleFunc("/api/v1/expressions", ExpressionsHandler)
	mux.HandleFunc("/api/v1/expressions/:id", IdHandler)
	mux.HandleFunc("/internal/task", TaskHandler)
	http.ListenAndServe(":8080", mux)
}
