package main

import (
	"C"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	_ "runtime/cgo"
	"sync/atomic"
	"time"
)

var db *sql.DB

// для ассинхронной работы запуск первого чекера в main, остальных закинуть в горутины ( go func() {}() )

func main() {
	SqlOpen() // открываем БД и проверяем наличие таблицы, если нет - создаем

	// раскомментировать ниже для проверки работоспособности

	/*
		a := floodControl{Id: 1, limit: 3, seconds: 3} // создаем три объекта для чекера
		b := floodControl{Id: 2, limit: 3, seconds: 3} // limit - переменная K из задания | seconds - переменная N из задания
		c := floodControl{Id: 3, limit: 3, seconds: 3} // ID указываем для корректной записи в БД

		go func() {
			FloodControl.Opener(&b) // с помощью горутин ассинхронно проверяем все объкеты
		}()
		go func() {
			FloodControl.Opener(&c)
		}()

		FloodControl.Opener(&a) // оставляем одну проверку в основной горутине - функции main, чтобы код не оффался

	*/

	defer db.Close() // в конце закрываем ДБ
}

type FloodControl interface {
	Check(ctx context.Context, userID int64) (bool, error)
	Opener()
}

type floodControl struct {
	Id        int64 // Айди пользователя
	counter   int64 // счётчик
	lastCheck int64 // последнее время
	seconds   int64 // секунды (в задаче N)
	limit     int64 // лимит (в задаче K)
}

func (fc *floodControl) Check(ctx context.Context, userID int64) (bool, error) {
	now := time.Now().Unix()
	if now-fc.lastCheck > fc.seconds {
		atomic.StoreInt64(&fc.counter, 0)
		fc.lastCheck = now
	}

	time.Sleep(time.Second)
	atomic.AddInt64(&fc.counter, 1)

	if atomic.LoadInt64(&fc.counter) > fc.limit {
		return false, nil
	}
	return true, nil
}

func (fc *floodControl) Opener() { // функция которая вызывает функцию Check, но также и делает запись в БД, сделана чтобы уменьшить количество кода при ассинхронной работе
	for {
		result, _ := FloodControl.Check(fc, context.Background(), fc.Id)
		if result == false {
			_, err := db.Exec("INSERT INTO checker (Id, Discription) values ($1, 'Нарушены правила флуд-контроля!')", fc.Id)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func SqlOpen() { // функция которая открывает датабазу SQL и создает таблицу, если ее нет
	var err error
	db, err = sql.Open("sqlite3", "floodControl.db")
	if err != nil {
		fmt.Println("Ошибка в открытии Базы Данных")
		log.Fatal(err)
		return
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS checker (Id INTEGER, Discription TEXT)")
	if err != nil {
		fmt.Println("Ошибка в создании таблицы")
		log.Fatal(err)
		return
	}
}
