package mysql

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/yelsukov/otus-ha/backend/errors"
	"github.com/yelsukov/otus-ha/backend/models"
)

type UsersStorage struct {
	db *sql.DB
}

func NewUsersStorage(db *sql.DB) *UsersStorage {
	return &UsersStorage{db}
}

const modelFields = "u.`id`, u.`username`, u.`first_name`, u.`last_name`, u.`age`, u.`gender`, u.`city`, u.`password_hash`, u.`created_at`, u.`interests`"

func (m *UsersStorage) Login(username, password string) (models.User, error) {
	row := m.db.QueryRow("SELECT "+modelFields+" FROM `users` u WHERE u.`username` = ?", username)
	user, err := scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("4031", "invalid username or password")
		}
		return user, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return user, errors.New("4031", "invalid username or password")
	}

	return user, nil
}

func (m *UsersStorage) Create(user *models.User) (int64, error) {
	user.Sanitize()
	if err := m.validate(user, nil); err != nil {
		return 0, err
	}
	user.CreatedAt = time.Now()

	upd := prepareExecStmt(user, nil)
	stmt, err := m.db.Prepare("INSERT `users` SET " + strings.Join(upd.Set, ","))
	defer closeStmt(stmt)
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(upd.Params...)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (m *UsersStorage) Get(id int64) (models.User, error) {
	row := m.db.QueryRow("SELECT "+modelFields+" FROM `users` u WHERE u.`id` = ?", id)
	user, err := scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("4041", "user not found")
		}
		return user, err
	}
	return user, nil
}

func (m *UsersStorage) Fetch(match [][2]string, offset, limit uint32) ([]models.User, error) {
	query := "SELECT " + modelFields + " FROM `users` u"
	where := &QueryStmt{make([]string, 0, 3), make([]interface{}, 0, 3)}
	if len(match) != 0 {
		for i := 0; i < len(match); i++ {
			where.Conditions = append(where.Conditions, match[i][0])
			where.Params = append(where.Params, match[i][1])
		}
	}
	if len(where.Conditions) > 0 {
		query += " WHERE " + strings.Join(where.Conditions[:], " AND ")
	}

	if limit == 0 {
		limit = 25
	}
	query += " LIMIT " + strconv.Itoa(int(offset)) + ", " + strconv.Itoa(int(limit))

	stmt, err := m.db.Prepare(query)
	defer closeStmt(stmt)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(where.Params...)
	defer closeRows(rows)
	if err != nil {
		return nil, err
	}

	return populateUserList(rows, limit)
}

func (m *UsersStorage) FetchFriends(userId int64, offset, limit uint32) ([]models.User, error) {
	query := "SELECT " + modelFields + " FROM `users` u INNER JOIN `friends` f ON f.friend_id = u.id WHERE f.user_id = ? LIMIT ?,?"
	if limit == 0 {
		limit = 25
	}
	stmt, err := m.db.Prepare(query)
	defer closeStmt(stmt)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(userId, offset, limit)
	defer closeRows(rows)
	if err != nil {
		return nil, err
	}

	return populateUserList(rows, limit)
}

func (m *UsersStorage) Update(user *models.User, clean *models.User) error {
	user.Sanitize()
	if err := user.Validate(); err != nil {
		return err
	}

	upd := prepareExecStmt(user, clean)
	if len(upd.Set) == 0 {
		return nil
	}

	query := "UPDATE `users` SET " + strings.Join(upd.Set, ",") + " WHERE " + strings.Join(upd.Conditions, " AND ")
	stmt, err := m.db.Prepare(query)
	defer closeStmt(stmt)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(upd.Params...)

	return err
}

func (m *UsersStorage) PrefixSearch(fnPrefix, lnPrefix string, offset, limit uint32) ([]models.User, error) {
	return m.Fetch([][2]string{
		{"`first_name` LIKE (?)", fnPrefix + "%"},
		{"`last_name` LIKE (?)", lnPrefix + "%"},
	}, offset, limit)
}

func populateUserList(rows *sql.Rows, cap uint32) ([]models.User, error) {
	list := make([]models.User, 0, cap)
	for rows.Next() {
		u, err := scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

func scan(row Scanner) (models.User, error) {
	u := models.User{}
	err := row.Scan(&u.Id, &u.Username, &u.FirstName, &u.LastName, &u.Age, &u.Gender, &u.City, &u.PasswordHash, &u.CreatedAt, &u.Interests)
	return u, err
}

func prepareExecStmt(dirty *models.User, clean *models.User) ExecStmt {
	var stmt ExecStmt

	if clean == nil || (dirty.Username != "" && dirty.Username != clean.Username) {
		stmt.Set = append(stmt.Set, "`username` = ?")
		stmt.Params = append(stmt.Params, dirty.Username)
		if clean != nil {
			clean.Username = dirty.Username
		}
	}
	if clean == nil || (dirty.FirstName.Valid && clean.FirstName.String != dirty.FirstName.String) {
		stmt.Set = append(stmt.Set, "`first_name` = ?")
		stmt.Params = append(stmt.Params, dirty.FirstName.String)
		if clean != nil {
			clean.FirstName = dirty.FirstName
		}
	}
	if clean == nil || (dirty.LastName.Valid && clean.LastName.String != dirty.LastName.String) {
		stmt.Set = append(stmt.Set, "`last_name` = ?")
		stmt.Params = append(stmt.Params, dirty.LastName.String)
		if clean != nil {
			clean.LastName = dirty.LastName
		}
	}

	if clean == nil || (dirty.Age.Valid && clean.Age.Int32 != dirty.Age.Int32) {
		stmt.Set = append(stmt.Set, "`age` = ?")
		stmt.Params = append(stmt.Params, dirty.Age.Int32)
		if clean != nil {
			clean.Age = dirty.Age
		}
	}

	if clean == nil || (dirty.Gender != "" && clean.Gender != dirty.Gender) {
		stmt.Set = append(stmt.Set, "`gender` = ?")
		stmt.Params = append(stmt.Params, dirty.Gender)
		if clean != nil {
			clean.Gender = dirty.Gender
		}
	}

	if clean == nil || (dirty.City.Valid && clean.City.String != dirty.City.String) {
		stmt.Set = append(stmt.Set, "`city` = ?")
		stmt.Params = append(stmt.Params, dirty.City.String)
		if clean != nil {
			clean.City = dirty.City
		}
	}

	if clean == nil || (dirty.PasswordHash != nil && string(dirty.PasswordHash) != string(clean.PasswordHash)) {
		stmt.Set = append(stmt.Set, "`password_hash` = ?")
		stmt.Params = append(stmt.Params, dirty.PasswordHash)
	}

	if clean == nil {
		stmt.Set = append(stmt.Set, "`created_at` = ?")
		stmt.Params = append(stmt.Params, dirty.CreatedAt)
	}

	if clean == nil || (dirty.Interests.Valid && clean.Interests.String != dirty.Interests.String) {
		stmt.Set = append(stmt.Set, "`interests` = ?")
		stmt.Params = append(stmt.Params, dirty.Interests.String)
		if clean != nil {
			clean.Interests = dirty.Interests
		}
	}

	if clean != nil {
		stmt.Conditions = append(stmt.Conditions, "`id` = ?")
		stmt.Params = append(stmt.Params, clean.Id)
	}

	return stmt
}

func (m *UsersStorage) validate(upd *models.User, old *models.User) error {
	if err := upd.Validate(); err != nil {
		return err
	}

	if old == nil && upd.Username == "" {
		return errors.New("4001", "`Username` is required")
	}

	if old == nil || old.Username != upd.Username {
		var exists bool
		err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM `users` WHERE `username` = ?)", upd.Username).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("4001", "user with the same username already exists")
		}
	}

	return nil
}
