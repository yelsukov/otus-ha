package nosql

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/go-tarantool"
	"github.com/yelsukov/otus-ha/backend/models"
)

type UsersStorage struct {
	tt *tarantool.Connection
}

func NewUsersStorage(db *tarantool.Connection) *UsersStorage {
	return &UsersStorage{db}
}

func (m *UsersStorage) PrefixSearch(fnPrefix, lnPrefix string, offset, limit uint32) ([]models.User, error) {
	resp, err := m.tt.Call("search", []interface{}{fnPrefix, lnPrefix, offset, limit})
	if err != nil {
		return nil, err
	}

	// if not found
	if len(resp.Data[0].([]interface{})) == 0 {
		return nil, nil
	}

	users := make([]models.User, len(resp.Data), len(resp.Data))
	for i, tuple := range resp.Data {
		row, ok := tuple.([]interface{})
		if !ok {
			log.Errorf("failed to prepare slice for tuple #%d", i)
			continue
		}
		if len(row) != 10 {
			log.Errorf("invalid fields quantity at tuple #%d: %v", i, row)
			continue
		}

		ct, err := time.Parse("2006-01-02 15:04:05", row[8].(string))
		if err != nil {
			log.WithError(err).Error("failed to parse user creation date")
		}
		users[i] = models.User{
			Id:        int64(row[0].(uint64)),
			Username:  row[1].(string),
			FirstName: models.NullString{NullString: sql.NullString{String: row[2].(string), Valid: true}},
			LastName:  models.NullString{NullString: sql.NullString{String: row[3].(string), Valid: true}},
			Age:       models.NullInt32{NullInt32: sql.NullInt32{Int32: int32(row[4].(uint64)), Valid: true}},
			Gender:    row[5].(string),
			City:      models.NullString{NullString: sql.NullString{String: row[6].(string), Valid: true}},
			Interests: models.NullString{NullString: sql.NullString{String: row[9].(string), Valid: true}},
			CreatedAt: ct,
		}
	}

	return users, nil
}
