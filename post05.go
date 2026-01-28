package post05

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
)

type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

var (
	Hostname = ""
	Port     = 2345
	Username = ""
	Password = ""
	Database = ""
)

func openConnection() (*sql.DB, error) {
	// Строка подключения
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Hostname, Port, Username, Password, Database)

	// Открыть базу данных
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Функция возвращает ID пользователя username
// -1 если пользователь не существует
func exists(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := -1

	statement := fmt.Sprintf(`SELECT "id" FROM "users" WHERE username = '%s'`, username)
	rows, err := db.Query(statement)

	for rows.Next() {
		var id int
		err = rows.Scan(&id)

		if err != nil {
			fmt.Println("Scan:", err)
			return -1
		}

		userID = id
	}

	defer rows.Close()

	return userID
}

// AddUser добавляет нового пользователя в БД.
// Возвращаем новый ID пользователя
// -1, если произошла ошибка
func AddUser(d Userdata) int {
	d.Username = strings.ToLower(d.Username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID != -1 {
		fmt.Println("User already exists", Username)
		return -1
	}

	insertStatement := `INSERT INTO "users" ("username") values ($1)`
	_, err = db.Exec(insertStatement, d.Username)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	userID = exists(d.Username)
	insertStatement = `INSERT INTO "userdata" ("userid", "name", "surname", "description") values ($1, $2, $3, $4)`
	_, err = db.Exec(insertStatement, userID, d.Name, d.Surname, d.Description)
	if err != nil {
		fmt.Println("db.Exec():", err)
		return -1
	}

	return userID
}

func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}

	defer db.Close()

	// Существует ли идентификатор?
	statement := fmt.Sprintf(`SELECT "username" FROM "users" WHERE id = %d`, id)
	rows, err := db.Query(statement)

	var username string

	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return err
		}
	}
	defer rows.Close()

	if exists(username) != id {
		return fmt.Errorf("user with id %d does not exist", id)
	}

	// Удалить из Userdata
	deleteStatement := `DELETE FROM "userdata" WHERE userid = $1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	// Удалить из Users
	deleteStatement = `DELETE FROM "users" WHERE id = $1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func ListUsers() ([]Userdata, error) {
	Data := []Userdata{}

	db, err := openConnection()
	if err != nil {
		return Data, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT 
    	"id", "username", "name", "surname", "description"
		FROM "users", "userdata"
		WHERE users.id = userdata.userid`)

	if err != nil {
		return Data, err
	}

	for rows.Next() {
		var id int
		var username string
		var name string
		var surname string
		var description string

		err = rows.Scan(&id, &username, &name, &surname, &description)
		temp := Userdata{id, username, name, surname, description}
		Data = append(Data, temp)

		if err != nil {
			return Data, err
		}
	}
	defer rows.Close()

	return Data, nil
}

func UpdateUser(d Userdata) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID == -1 {
		return errors.New("user doesnt not exists")
	}

	d.ID = userID

	updateStatement := `UPDATE "userdata" SET name = $1, surname = $2, description = $3 WHERE userid = $4`
	_, err = db.Exec(updateStatement, d.Name, d.Surname, d.Description, userID)
	if err != nil {
		return err
	}

	return nil
}
