// Пакет database предоставляет функции для работы с базой данных PostgreSQL.
package database

import (
	"database/sql" // Импорт пакета для работы с SQL базами данных
	"fmt"          // Форматированный вывод
	"time"         // Работа со временем
	"log"          // Логирование
	"sync"         // Синхронизация горутин
	"calculatorapi/utility/models" // Структуры данных для калькулятора

	_ "github.com/lib/pq" // Драйвер PostgreSQL
)

// Параметры подключения к базе данных
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123QWEasdf"
	dbname   = "postgres"
)

var (
	db   *sql.DB       // Глобальный объект базы данных
	dbMu sync.Mutex    // Мьютекс для защиты доступа к объекту базы данных
)

// InitializeDB устанавливает новое соединение с базой данных.
func InitializeDB() {
	dbMu.Lock() // Блокировка мьютекса
	defer dbMu.Unlock() // Освобождение мьютекса при выходе из функции

	if db != nil {
		return // Если соединение уже установлено, ничего не делаем
	}

	// Формирование строки подключения
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", psqlInfo) // Открытие соединения
	if err != nil {
		log.Fatalf("Error opening database: %v", err) // Логирование ошибки
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err) // Проверка соединения
	}

	fmt.Println("Database connection established") // Сообщение об успешном соединении
}

// GetDB возвращает глобальное соединение с базой данных, убеждаясь, что оно инициализировано и подключено.
func GetDB() *sql.DB {
	dbMu.Lock() // Блокировка мьютекса
	defer dbMu.Unlock() // Освобождение мьютекса

	if db == nil {
		InitializeDB() // Инициализация соединения, если оно ещё не установлено
	}

	// Проверка, живо ли соединение
	if err := db.Ping(); err != nil {
		fmt.Println("Reconnecting to the database...")
		InitializeDB() // Повторная инициализация при необходимости
	}

	return db
}

// ConnectToDatabase создает и возвращает новое соединение с базой данных (используется для демонстрации; в реальных условиях лучше использовать GetDB).
func ConnectToDatabase() (*sql.DB, error) {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        return nil, err // Возвращение ошибки при неудаче
    }
    return db, nil // Возвращение объекта соединения
}

// SetupDatabase проверяет наличие базы данных и таблицы; при отсутствии создает их.
func SetupDatabase() (*sql.DB, error) {
	// Аналогично ConnectToDatabase, но с дополнительными шагами настройки
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        return nil, err
    }

    exists, err := DatabaseExists(db, dbname)
    if err != nil {
        return nil, err
    }

    if !exists {
        _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
        if err != nil {
            return nil, err
        }
        fmt.Printf("Database '%s' created successfully.\n", dbname)
    }

    err = CreateTableIfNotExists(db)
    if err != nil {
        return nil, err
    }

    return db, nil
}

// DatabaseExists проверяет, существует ли база данных.
func DatabaseExists(db *sql.DB, dbName string) (bool, error) {
	// Выполнение SQL запроса для проверки существования базы данных
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// CreateTableIfNotExists создает таблицу, если она еще не существует.
func CreateTableIfNotExists(db *sql.DB) error {
	// Выполнение SQL запроса для создания таблицы
	var tableExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'calculations')").Scan(&tableExists)
	if err != nil {
		return err
	}

	if tableExists {
		fmt.Println("Table 'calculations' already exists.")
		return nil
	}

	query := `
		CREATE TABLE calculations (
			id SERIAL PRIMARY KEY,
			operation TEXT,
			result DOUBLE PRECISION,
			status TEXT,
			created_time TIMESTAMP,
			start_time TIMESTAMP,
			end_time TIMESTAMP,
			operation_server TEXT,
			server_status TEXT,
			add_duration INTEGER,
			subtract_duration INTEGER,
			multiply_duration INTEGER,
			divide_duration INTEGER,
			inactive_server_time INTEGER
		)
	`

	_, err = db.Exec(query)
	if err != nil {
		return err
	}
	fmt.Println("Table 'calculations' created successfully.")
	return nil
}

// InsertCalculation вставляет новую запись о вычислении в таблицу 'calculations'.
func InsertCalculation(db *sql.DB, operation string, addDuration, subtractDuration, multiplyDuration, divideDuration, inactiveServerTime int) (int, error) {
    // Вставка данных о вычислении и возвращение идентификатора записи
    if err := db.Ping(); err != nil {
        // If not, attempt to reconnect
        fmt.Println("Reconnecting to the database...")
        if err := db.Close(); err != nil {
            return 0, err
        }
        db, err = SetupDatabase()
        if err != nil {
            return 0, err
        }
    }

    // Proceed with the insertion
    query := `
        INSERT INTO calculations (operation, status, created_time, add_duration, subtract_duration, multiply_duration, divide_duration, inactive_server_time)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `
    status := `created`
    createdTime := time.Now().UTC()

    var id int
    err := db.QueryRow(query, operation, status, createdTime, addDuration, subtractDuration, multiplyDuration, divideDuration, inactiveServerTime).Scan(&id)
    if err != nil {
        return 0, err
    }

    fmt.Println("Calculation record inserted successfully.")
    return id, nil
}

// RunCheckCreatedRecords запускает периодическую проверку записей с статусом "created".
func RunCheckCreatedRecords(db *sql.DB) {
    ticker := time.NewTicker(10 * time.Second) // Создание таймера с интервалом в 10 секунд.
    defer ticker.Stop() // Остановка таймера при выходе из функции.

    for {
        select {
        case <-ticker.C: // При каждом тике таймера.
            err := checkAndPrintCreatedRecords(db) // Выполнение проверки записей.
            if err != nil {
                fmt.Printf("Error during checkAndPrintCreatedRecords: %v\n", err)
            }
        }
    }
}

// checkAndPrintCreatedRecords проверяет и выводит записи с статусом "created".
func checkAndPrintCreatedRecords(db *sql.DB) error {
	// SQL-запроc.
	query := `
		SELECT id, operation, created_time, add_duration, subtract_duration, multiply_duration, divide_duration, inactive_server_time
		FROM calculations
		WHERE status = 'created'
	`

	rows, err := db.Query(query) // Выполнение запроса.
	if err != nil {
		return fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close() // Закрытие результата запроса при выходе из функции.

	fmt.Println("Records with status 'created':")
	for rows.Next() { // Перебор всех полученных записей.
		var (
			id                   int
			operation            string
			createdTime          time.Time
			addDuration          int
			subtractDuration     int
			multiplyDuration     int
			divideDuration       int
			inactiveServerTime   int
		)

		// Считывание значений текущей записи.
		if err := rows.Scan(&id, &operation, &createdTime, &addDuration, &subtractDuration, &multiplyDuration, &divideDuration, &inactiveServerTime); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		// Вывод информации о записи.
		fmt.Printf("ID: %d, Operation: %s, Created Time: %s, Add Duration: %d, Subtract Duration: %d, Multiply Duration: %d, Divide Duration: %d, Inactive Server Time: %d\n",
			id, operation, createdTime.Format("2006-01-02 15:04:05"), addDuration, subtractDuration, multiplyDuration, divideDuration, inactiveServerTime)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over rows: %w", err)
	}
	return nil // Возвращение nil в случае успешного выполнения функции.
}

// UpdateCalculation обновляет запись о вычислении в таблице 'calculations' по ID.
func UpdateCalculation(db *sql.DB, id int, result float64, status string) error {
    // SQL-запрос для обновления записи.
    query := `
        UPDATE calculations
        SET result = $1, status = $2, end_time = $3
        WHERE id = $4
    `
    endTime := time.Now().UTC()

    // Выполнение запроса.
    _, err := db.Exec(query, result, status, endTime, id)
    if err != nil {
        return err
    }

    fmt.Printf("Calculation record with ID %d updated successfully.\n", id)
    return nil
}

// UpdateCalculationStatusToWork обновляет статус вычисления на 'work' и устанавливает start_time.
func UpdateCalculationStatusToWork(db *sql.DB, id int) error {
    // SQL-запрос для обновления статуса и времени начала.
    query := `
        UPDATE calculations
        SET status = 'work', start_time = timezone('UTC', NOW())
        WHERE id = $1
    `

    _, err := db.Exec(query, id)
    if err != nil {
        return fmt.Errorf("error updating calculation status to work and setting start time: %w", err)
    }

    fmt.Printf("Calculation record with ID %d status updated to work and start time set.\n", id)
    return nil
}

// FetchCalculationsToProcess извлекает максимум 5 строк со статусом "created".
func FetchCalculationsToProcess(db *sql.DB) ([]models.CalculationRequest, error) {
    var calculations []models.CalculationRequest // Слайс для хранения результатов.

    query := `SELECT id, operation, add_duration, subtract_duration, multiply_duration, divide_duration FROM calculations WHERE status = 'created' LIMIT 5`
    rows, err := db.Query(query) // Выполнение запроса.
	// Возврат ошибки в случае ее возникновения.
    if err != nil {
        return nil, err
    }
    defer rows.Close() // Закрытие результата запроса при выходе из функции.

    for rows.Next() { // Перебор всех полученных записей.
        var calc models.CalculationRequest
        if err := rows.Scan(&calc.ID, &calc.Operation, &calc.AddDuration, &calc.SubtractDuration, &calc.MultiplyDuration, &calc.DivideDuration); err != nil {
            return nil, err // Возврат ошибки при возникновении.
        }
        calculations = append(calculations, calc) // Добавление записи в слайс.
    }

    if err = rows.Err(); err != nil {
        return nil, err // Возврат ошибки при возникновении.
    }

    return calculations, nil// Возвращение слайса с результатами и nil в случае успешного выполнения функции.
}

// GetCalculationResultByID извлекает результат вычисления по его ID.
func GetCalculationResultByID(db *sql.DB, id int) (*models.CalculationResponse, error) {
    var (
        result sql.NullFloat64 // Использование sql.NullFloat64 для обработки NULL значений.
        status string
    )
    query := `SELECT result, status FROM calculations WHERE id = $1` // SQL-запрос для выборки.
    err := db.QueryRow(query, id).Scan(&result, &status) // Выполнение запроса и считывание результатов.
    if err != nil {
        return nil, err // Возврат ошибки при возникновении.
    }

    calcResult := &models.CalculationResponse{
        ID:     id,
        Status: status,
    }

    if result.Valid {
        calcResult.Result = result.Float64 // Присвоение результата, если он не NULL.
    }

    return calcResult, nil // Возвращение ответа и nil в случае успешного выполнения функции.
}

// FetchAllCalculations извлекает все вычисления из базы данных.
func FetchAllCalculations(db *sql.DB) ([]models.OperationResponse, error) {
    var calculations []models.OperationResponse // Слайс для хранения результатов.

    query := `SELECT id, operation, result, status FROM calculations` // SQL-запрос для выборки всех записей.
    rows, err := db.Query(query) // Выполнение запроса.
    if err != nil {
        return nil, fmt.Errorf("querying calculations: %w", err)
    }
    defer rows.Close() // Закрытие результата запроса при выходе из функции.

    for rows.Next() { // Перебор всех полученных записей.
        var calc models.OperationResponse
        var result sql.NullFloat64 // Использование sql.NullFloat64 для обработки NULL значений.

        if err := rows.Scan(&calc.ID, &calc.Operation, &result, &calc.Status); err != nil {
            return nil, fmt.Errorf("scanning calculation: %w", err)
        }

        if result.Valid {
            calc.Result = result.Float64 // Присвоение результата, если он не NULL.
        } else {
			// При желании можно обработать NULL результаты иначе, например, присвоить значение по умолчанию или опустить поле.
        }

        calculations = append(calculations, calc) // Добавление записи в слайс.
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("iterating over calculations results: %w", err)
    }

    return calculations, nil // Возвращение слайса с результатами и nil в случае успешного выполнения функции.
}

// ClearAllCalculations удаляет все строки из таблицы 'calculations'.
func ClearAllCalculations(db *sql.DB) error {
    // SQL statement to delete all rows
    query := `DELETE FROM calculations` // SQL-запрос для удаления всех строк.
    _, err := db.Exec(query) // Выполнение запроса.
    if err != nil {
        return fmt.Errorf("clearing all calculations: %w", err)
    }
    fmt.Println("All calculations cleared successfully.")
    return nil // Возвращение nil в случае успешного выполнения функции.
}