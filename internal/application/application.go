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
	Result float64 `json:"result"`
}
type Error struct {
	Error string `json:"error"`
}

var req Request

func ErrorMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&req)
		_, err := calculation.Calc(req.Expression)
		if err != nil {
			if errors.Is(err, calculation.ErrorInvalidExpression) {
				resp := Error{Error: "Expression is not valid"}
				answ, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write(answ)
				return
			} else {
				resp := Error{Error: "Internal server error"}
				answ, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(answ)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	result, _ := calculation.Calc(req.Expression)
	response := Answer{Result: result}
	answ, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(answ)
}

func RunApplication() {
	http.HandleFunc("/api/v1/calculate", ErrorMiddleware(CalculateHandler))
	http.ListenAndServe(":8080", nil)
}
