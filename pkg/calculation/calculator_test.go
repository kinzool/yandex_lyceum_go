package calculation_test

import (
	"testing"
	"yandexlyceum/yandex_lyceum_go/pkg/calculation"
)

func TestCalc(t *testing.T) {
	testCasesSuccess := []struct {
		name           string
		expression     string
		expectedResult float64
	}{
		{
			name:           "simple",
			expression:     "1+1",
			expectedResult: 2,
		},
		{
			name:           "priority with brackets",
			expression:     "(2+2)*2",
			expectedResult: 8,
		},
		{
			name:           "priority",
			expression:     "2+2*2",
			expectedResult: 6,
		},
		{
			name:           "float result",
			expression:     "1/2",
			expectedResult: 0.5,
		},
	}

	for _, testCase := range testCasesSuccess {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := calculation.Calc(testCase.expression)
			if err != nil {
				t.Fatalf("Ошибка при выражении: %s", testCase.expression)
			}
			if val != testCase.expectedResult {
				t.Fatalf("Ожидалось: %f, Результат: %f", val, testCase.expectedResult)
			}
		})
	}

	testCasesFail := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:       "simple",
			expression: "1-1*",
		},
		{
			name:       "priority",
			expression: "2+2**2",
		},
		{
			name:       "priority with brackets",
			expression: "((2+2-*(2",
		},
		{
			name:       "/",
			expression: "",
		},
	}

	for _, testCase := range testCasesFail {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := calculation.Calc(testCase.expression)
			if err == nil {
				t.Fatalf("Выражение %s не валидно, но результат  %f был получен", testCase.expression, val)
			}
		})
	}
}
