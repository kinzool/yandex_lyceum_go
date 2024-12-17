package calculation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrorInvalidExpression = errors.New("expression is not valid")

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	if expression == "" || !isValidParentheses(expression) {
		return 0, ErrorInvalidExpression
	}
	rpn, err := infixToRPN(expression)
	if err != nil {
		return 0, err
	}
	result, err := calculateRPN(rpn)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func isValidParentheses(expression string) bool {
	stack := []rune{}
	for _, char := range expression {
		switch char {
		case '(':
			stack = append(stack, char)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	return len(stack) == 0
}

func infixToRPN(expression string) ([]string, error) {
	rpn := []string{}
	operators := []rune{}
	for _, char := range expression {
		switch char {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			rpn = append(rpn, string(char))
		case '+', '-', '*', '/':
			for len(operators) > 0 && getPrecedence(operators[len(operators)-1]) >= getPrecedence(char) {
				rpn = append(rpn, string(operators[len(operators)-1]))
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, char)
		case '(':
			operators = append(operators, char)
		case ')':
			for len(operators) > 0 && operators[len(operators)-1] != '(' {
				rpn = append(rpn, string(operators[len(operators)-1]))
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, ErrorInvalidExpression
			}
			operators = operators[:len(operators)-1]
		default:
			return nil, ErrorInvalidExpression
		}
	}
	for len(operators) > 0 {
		rpn = append(rpn, string(operators[len(operators)-1]))
		operators = operators[:len(operators)-1]
	}
	return rpn, nil
}

func getPrecedence(operator rune) int {
	switch operator {
	case '*', '/':
		return 2
	case '+', '-':
		return 1
	case '(':
		return 0
	}
	return -1
}

func calculateRPN(rpn []string) (float64, error) {
	stack := []float64{}
	for _, token := range rpn {
		switch token {
		case "+", "-", "*", "/":
			if len(stack) < 2 {
				return 0, ErrorInvalidExpression
			}
			operand2 := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			operand1 := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			result, err := calculateOperation(operand1, operand2, token)
			if err != nil {
				return 0, err
			}
			stack = append(stack, result)
		default:
			number, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, err
			}
			stack = append(stack, number)
		}
	}
	if len(stack) != 1 {
		return 0, ErrorInvalidExpression
	}
	return stack[0], nil
}

func calculateOperation(operand1, operand2 float64, operator string) (float64, error) {
	switch operator {
	case "+":
		return operand1 + operand2, nil
	case "-":
		return operand1 - operand2, nil
	case "*":
		return operand1 * operand2, nil
	case "/":
		if operand2 == 0 {
			return 0, fmt.Errorf("деление на ноль")
		}
		return operand1 / operand2, nil
	}
	return 0, ErrorInvalidExpression
}
