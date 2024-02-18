// Пакет calculation предоставляет простой арифметический калькулятор,
// который имитирует задержки выполнения в зависимости от типа операции.
package calculation

import (
    "fmt"       // Используется для форматированного вывода строк
    "strconv"   // Для преобразования строк в числа и обратно
    "strings"   // Для работы со строками
    "time"      // Для имитации задержек
)

// OperationTimes определяет задержки для каждого типа операции.
type OperationTimes map[string]time.Duration

// EvaluateOperation принимает арифметическую операцию в виде строки
// и operationTimes, определяющий задержки для каждой операции.
// Возвращает срез строк с деталями каждого шага вычисления
// и итоговый результат в виде float64.
func EvaluateOperation(operation string, operationTimes OperationTimes) ([]string, float64) {
    var operations []string // Срез для хранения описания операций
    operands, operators := parseOperation(operation) // Разбор операции на операнды и операторы

    // Слияние операндов и операторов в один срез для последовательного вычисления
    expression := make([]string, 0, len(operands)+len(operators))
    for i, op := range operands {
        expression = append(expression, op)
        if i < len(operators) {
            expression = append(expression, operators[i])
        }
    }

    // Сначала обрабатываем умножение и деление
    for i := 0; i < len(expression)-1; i++ {
        if expression[i] == "*" || expression[i] == "/" {
            left, _ := strconv.ParseFloat(expression[i-1], 64)
            right, _ := strconv.ParseFloat(expression[i+1], 64)
            result := performOperation(left, right, expression[i], operationTimes)
            // Запись выполненной операции
            operations = append(operations, fmt.Sprintf("%s %s %s = %.6f", expression[i-1], expression[i], expression[i+1], result))

            // Обновление среза expression
            expression[i+1] = fmt.Sprintf("%.6f", result)
            expression = append(expression[:i-1], expression[i+1:]...)
            i = i - 2 // Корректировка индекса после изменения среза
        }
    }

    // Затем обрабатываем сложение и вычитание
    var result float64
    if len(expression) > 0 {
        result, _ = strconv.ParseFloat(expression[0], 64)
    }
    for i := 1; i < len(expression); i += 2 {
        right, _ := strconv.ParseFloat(expression[i+1], 64)
        result = performOperation(result, right, expression[i], operationTimes)
        // Запись текущей операции с результатом
        operations = append(operations, fmt.Sprintf("%.6f %s %s = %.6f", result, expression[i], expression[i+1], result))
    }

    return operations, result // Возврат истории операций и результата
}

// Разбор операции на операнды и операторы
func parseOperation(operation string) ([]string, []string) {
    // Извлечение операндов
    operands := strings.FieldsFunc(operation, func(c rune) bool {
        return c == '+' || c == '-' || c == '*' || c == '/'
    })

    // Извлечение операторов
    operators := make([]string, 0)
    for _, c := range operation {
        if strings.ContainsRune("+-*/", c) {
            operators = append(operators, string(c))
        }
    }

    return operands, operators // Возврат операндов и операторов
}

// Выполнение операции с учетом задержки
func performOperation(left, right float64, operator string, operationTimes OperationTimes) float64 {
    // Имитация времени выполнения операции
    if duration, ok := operationTimes[operator]; ok {
        fmt.Printf("Performing %s operation, waiting for %v\n", operator, duration)
        time.Sleep(duration) // Задержка
    } else {
        fmt.Println("Unknown operation, no delay applied")
    }

    // Выполнение арифметической операции
    switch operator {
    case "+":
        return left + right
    case "-":
        return left - right
    case "*":
        return left * right
    case "/":
        if right == 0 {
            fmt.Println("Error: Division by zero")
            return 0
        }
        return left / right
    default:
        fmt.Println("Unknown operator", operator)
        return 0
    }
}