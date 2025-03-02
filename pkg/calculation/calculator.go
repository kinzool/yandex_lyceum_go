package calculation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Stack представляет собой стек для хранения чисел.
type Stack struct {
	items []float64
}

// Push добавляет элемент в стек.
func (s *Stack) Push(item float64) {
	s.items = append(s.items, item)
}

// Pop удаляет последний элемент из стека.
func (s *Stack) Pop() (float64, error) {
	if len(s.items) < 1 {
		return 0, errors.New("стек пуст")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, nil
}

// OperatorStack представляет собой стек для хранения операторов.
type OperatorStack struct {
	items []string
}

// Push добавляет оператор в стек.
func (s *OperatorStack) Push(item string) {
	s.items = append(s.items, item)
}

// Pop удаляет последний оператор из стека.
func (s *OperatorStack) Pop() (string, error) {
	if len(s.items) < 1 {
		return "", errors.New("стек операторов пуст")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, nil
}

// getPrecedence возвращает приоритет оператора.
func getPrecedence(operator string) int {
	switch operator {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

func toRPN(expression string) (string, error) {
	expression = strings.ReplaceAll(expression, "(", "( ")
	expression = strings.ReplaceAll(expression, ")", " )")
	expression = strings.ReplaceAll(expression, "+", " + ")
	expression = strings.ReplaceAll(expression, "-", " - ")
	expression = strings.ReplaceAll(expression, "*", " * ")
	expression = strings.ReplaceAll(expression, "/", " / ")
	tokens := strings.Fields(expression)

	operatorStack := &OperatorStack{}
	output := []string{}

	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token)
		} else if token == "(" {
			operatorStack.Push(token)
		} else if token == ")" {
			for len(operatorStack.items) > 0 && operatorStack.items[len(operatorStack.items)-1] != "(" {
				output = append(output, operatorStack.items[len(operatorStack.items)-1])
				operatorStack.items = operatorStack.items[:len(operatorStack.items)-1]
			}
			if len(operatorStack.items) == 0 || operatorStack.items[len(operatorStack.items)-1] != "(" {
				return "", errors.New("неверные скобки")
			}
			operatorStack.items = operatorStack.items[:len(operatorStack.items)-1]
		} else {
			for len(operatorStack.items) > 0 && operatorStack.items[len(operatorStack.items)-1] != "(" && getPrecedence(operatorStack.items[len(operatorStack.items)-1]) >= getPrecedence(token) {
				output = append(output, operatorStack.items[len(operatorStack.items)-1])
				operatorStack.items = operatorStack.items[:len(operatorStack.items)-1]
			}
			operatorStack.Push(token)
		}
	}

	for len(operatorStack.items) > 0 {
		if operatorStack.items[len(operatorStack.items)-1] == "(" {
			return "", errors.New("неверные скобки")
		}
		output = append(output, operatorStack.items[len(operatorStack.items)-1])
		operatorStack.items = operatorStack.items[:len(operatorStack.items)-1]
	}

	return strings.Join(output, " "), nil
}

// EvaluateRPN вычисляет выражение в обратной польской нотации и сохраняет подвыражения.
func EvaluateRPN(expression string) ([]string, float64, error) {
	stack := &Stack{}
	subexpressions := []string{}
	tokens := strings.Split(expression, " ")

	for _, token := range tokens {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack.Push(num)
		} else {
			// Обработка операций
			if len(stack.items) < 2 {
				return nil, 0, errors.New("недостаточно операндов")
			}
			operand2, _ := stack.Pop()
			operand1, _ := stack.Pop()

			var subexpression string
			switch token {
			case "+":
				stack.Push(operand1 + operand2)
				subexpression = fmt.Sprintf("%f + %f", operand1, operand2)
			case "-":
				stack.Push(operand1 - operand2)
				subexpression = fmt.Sprintf("%f - %f", operand1, operand2)
			case "*":
				stack.Push(operand1 * operand2)
				subexpression = fmt.Sprintf("%f * %f", operand1, operand2)
			case "/":
				if operand2 == 0 {
					return nil, 0, errors.New("деление на ноль")
				}
				stack.Push(operand1 / operand2)
				subexpression = fmt.Sprintf("%f / %f", operand1, operand2)
			default:
				return nil, 0, fmt.Errorf("неизвестная операция: %s", token)
			}

			subexpressions = append(subexpressions, subexpression)
		}
	}

	if len(stack.items) != 1 {
		return nil, 0, errors.New("неверное выражение")
	}
	res, err := stack.Pop()

	return subexpressions, res, err
}

func ParseExpression(expression string) ([]string, error) {
	rpn, err := toRPN(expression)
	if err != nil {
		return nil, err
	} else {
		subexpressions, _, err := EvaluateRPN(rpn)
		if err != nil {
			return nil, err
		}
		return subexpressions, nil
	}
}

func CalculateSimpleExpression(expression string) (float64, error) {
	for _, op := range []string{"+", "-", "*", "/"} {
		if strings.Contains(expression, op) {
			parts := strings.Split(expression, op)
			parts[0] = strings.TrimSpace(parts[0])
			parts[1] = strings.TrimSpace(parts[1])
			if len(parts) != 2 {
				return 0, errors.New("неверное выражение")
			}
			num1, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, err
			}

			num2, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return 0, err
			}

			switch op {
			case "+":
				return num1 + num2, nil
			case "-":
				return num1 - num2, nil
			case "*":
				return num1 * num2, nil
			case "/":
				if num2 == 0 {
					return 0, errors.New("деление на ноль")
				}
				return num1 / num2, nil
			default:
				return 0, errors.New("неизвестная операция")
			}
		}
	}

	return 0, errors.New("неверное выражение")
}

func AddExpression(expression string) error {
	_, err := ParseExpression(expression)
	if err != nil {
		return err
	}
	return nil
}
