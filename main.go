package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // Імпортуємо драйвер БД
)

// Record - структура, що відповідає таблиці в БД
type Record struct {
	ID         int
	DeviceName string
	Voltage    int
	RecordDate string
}

var db *sql.DB

func initDB() {
	var err error
	// ФОРМАТ: "користувач:пароль@tcp(хост:порт)/назва_бд"
	// ЗМІНИ "твій_пароль" НА СВІЙ ПАРОЛЬ ВІД MYSQL WORKBENCH!
	dsn := "root:1111@tcp(127.0.0.1:3306)/surge_protection"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Помилка підключення до БД:", err)
	}

	// Перевіряємо чи зв'язок дійсно встановлено
	if err = db.Ping(); err != nil {
		log.Fatal("База даних не відповідає (перевір пароль або чи запущений MySQL):", err)
	}
	log.Println("Успішно підключено до MySQL!")
}

func dbHandler(w http.ResponseWriter, r *http.Request) {
	// Якщо прийшли дані з форми - зберігаємо їх в БД
	if r.Method == http.MethodPost {
		deviceName := r.FormValue("deviceName")
		voltage := r.FormValue("voltage")
		date := r.FormValue("date")

		// SQL запит на вставку даних
		_, err := db.Exec("INSERT INTO voltage_logs (device_name, voltage, record_date) VALUES (?, ?, ?)", deviceName, voltage, date)
		if err != nil {
			http.Error(w, "Помилка запису в БД", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		// Перенаправляємо на головну сторінку, щоб оновити таблицю
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// ЧИТАННЯ ДАНИХ З БД
	rows, err := db.Query("SELECT id, device_name, voltage, record_date FROM voltage_logs ORDER BY id DESC")
	if err != nil {
		http.Error(w, "Помилка читання з БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var rec Record
		// Записуємо дані з БД у нашу структуру (важливо дотримуватись порядку стовпців)
		err := rows.Scan(&rec.ID, &rec.DeviceName, &rec.Voltage, &rec.RecordDate)
		if err != nil {
			log.Println("Помилка парсингу рядка:", err)
			continue
		}
		records = append(records, rec)
	}

	// Виводимо HTML
	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, records)
}

func main() {
	initDB()         // Підключаємося до бази
	defer db.Close() // Закриваємо з'єднання при вимкненні сервера

	http.HandleFunc("/", dbHandler)

	log.Println("Сервер запущено: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
