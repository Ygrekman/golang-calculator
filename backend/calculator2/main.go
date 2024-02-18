package main

import (
    // Импортирование необходимых пакетов
    "encoding/json"    // Для работы с JSON
    "fmt"              // Для форматированного ввода и вывода
    "log"              // Для логирования
    "net/http"         // Для работы с HTTP
    "sync"             // Для синхронизации горутин
    "time"             // Для работы со временем
	"database/sql"     // Для работы с базой данных

    // Импортирование собственных пакетов
    "calculatorapi/utility/calculation"  // Для выполнения вычислений
	"calculatorapi/utility/database"     // Для работы с базой данных
)


// Структура запроса на выполнение операции
type OperationRequest struct {
    ID        int               `json:"id"`          // Идентификатор операции
    Operation string            `json:"operation"`   // Строка операции
    Times     map[string]int    `json:"times"`       // Время выполнения каждой операции
}

var (
    // Глобальные переменные для контроля состояния сервера и горутин
    maxGoroutines    = 5                                 // Максимальное количество горутин
    currentGoroutines = 0                                // Текущее количество работающих горутин
    mu               sync.Mutex                          // Мьютекс для синхронизации доступа к currentGoroutines
    shutdownCh       = make(chan struct{})               // Канал для сигнала остановки сервера
    serverRunning    = true                              // Флаг состояния работы сервера
)

// Преобразование времени выполнения операций из запроса в структуру для вычисления
func ConvertOperationTimes(times map[string]int) calculation.OperationTimes {
    operationTimes := calculation.OperationTimes{}
    for k, v := range times {
        switch k {
        // Заполнение времени выполнения для каждой операции
        case "add_duration":
            operationTimes["+"] = time.Duration(v) * time.Second
        case "subtract_duration":
            operationTimes["-"] = time.Duration(v) * time.Second
        case "multiply_duration":
            operationTimes["*"] = time.Duration(v) * time.Second
        case "divide_duration":
            operationTimes["/"] = time.Duration(v) * time.Second
        }
    }
    return operationTimes
}

// Запуск вычисления на основе полученных данных
func startCalculation(db *sql.DB, id int, operation string, times map[string]int) {
    convertedTimes := ConvertOperationTimes(times)

    // Выполнение вычисления в отдельной горутине
    go func() {
        defer func() {
            // После завершения вычисления уменьшаем количество работающих горутин
            mu.Lock()
            currentGoroutines--
            mu.Unlock()
        }()

        // Обновление статуса вычисления на 'work'
        err := database.UpdateCalculationStatusToWork(db, id)
        if err != nil {
            fmt.Printf("Error updating status to work: %v\n", err)
            return
        }

        // Выполнение вычисления
        operations, result := calculation.EvaluateOperation(operation, convertedTimes)
        for _, op := range operations {
            fmt.Println(op)
        }
        fmt.Printf("Calculation ID %d completed. Result: %.6f\n", id, result)

        // Обновление записи в базе данных на 'completed'
        err = database.UpdateCalculation(db, id, result, "completed")
        if err != nil {
            fmt.Printf("Error updating calculation record to completed: %v\n", err)
        }
    }()
}

// Основная функция сервера
func main() {
    // Инициализация соединения с базой данных
	database.InitializeDB()
	
    // Обработчик запроса на выполнение вычисления
    http.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Method is not supported.", http.StatusNotFound)
            return
        }

        // Проверка на возможность обработки нового запроса
        mu.Lock()
        if !serverRunning {
            http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
            mu.Unlock()
            return
        }
        if currentGoroutines >= maxGoroutines {
            http.Error(w, "Server max capacity reached", http.StatusTooManyRequests)
            mu.Unlock()
            return
        }
        currentGoroutines++
        mu.Unlock()

        // Разбор запроса
        var request OperationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Запуск вычисления
		db := database.GetDB()
		startCalculation(db, request.ID, request.Operation, request.Times)
        w.WriteHeader(http.StatusAccepted)
        fmt.Fprintln(w, "Calculation started successfully.")
    })

    // Обработчик запроса на получение текущего количества горутин
    http.HandleFunc("/goroutines", func(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        fmt.Fprintf(w, "Current number of goroutines: %d\n", currentGoroutines)
        mu.Unlock()
    })

    // Обработчик запроса на остановку сервера
    http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        serverRunning = false
        mu.Unlock()
        close(shutdownCh)
        fmt.Fprintln(w, "Server is shutting down...")
    })

    // Обработчик запроса на проверку состояния сервера
    http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		response := struct {
			Status           string `json:"status"`
			MaxGoroutines    int    `json:"maxGoroutines"`
			CurrentGoroutines int   `json:"currentGoroutines"`
		}{
			Status:           "running",
			MaxGoroutines:    maxGoroutines,
			CurrentGoroutines: currentGoroutines,
		}
	
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

    // Горутина для ожидания остановки сервера и завершения всех операций
    go func() {
        <-shutdownCh
        fmt.Println("Server stopped accepting new requests. Waiting for ongoing operations to complete...")
        for {
            mu.Lock()
            if currentGoroutines == 0 {
                mu.Unlock()
                break
            }
            mu.Unlock()
            time.Sleep(1 * time.Second)
        }
        log.Fatal("Server gracefully shut down")
    }()

    // Запуск сервера на порту 8082
    fmt.Println("Calculator server is starting on port 8082...")
    log.Fatal(http.ListenAndServe(":8082", nil))
}