package sync

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/siddontang/go-mysql/mysql"
)

type masterInfo struct {
	sync.RWMutex

	Name string
	Pos  uint32

	lastSaveTime time.Time
	filePath     string
}

func loadMasterInfo(dataDir string) (*masterInfo, error) {
	var m masterInfo

	if len(dataDir) == 0 {
		return &m, errors.New("empty data dir for master file load")
	}

	m.filePath = path.Join(dataDir, "master.info")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	f, err := os.Open(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &m, nil
		}
		return nil, err
	}
	defer f.Close()

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	ss := strings.Split(string(bs), "|")
	if len(ss) != 2 {
		return nil, errors.New("invalid format of master file")
	}
	m.Name = ss[0]
	pos, err := strconv.Atoi(ss[1])
	if err != nil {
		return nil, err
	}
	m.Pos = uint32(pos)

	return &m, nil
}

func (m *masterInfo) save(pos mysql.Position, force bool) error {
	m.Lock()
	defer m.Unlock()

	m.Name = pos.Name
	m.Pos = pos.Pos

	if len(m.filePath) == 0 {
		return errors.New("file path is empty")
	}

	n := time.Now()
	// Do not save the position in the file more than once every 500 ms
	if !m.lastSaveTime.IsZero() && !force && n.Sub(m.lastSaveTime) < (500*time.Millisecond) {
		return nil
	}
	m.lastSaveTime = n

	return writeMasterFile(m.filePath, []byte(m.Name+"|"+strconv.Itoa(int(m.Pos))), 0644)
}

func (m *masterInfo) position() mysql.Position {
	m.RLock()
	defer m.RUnlock()
	return mysql.Position{
		Name: m.Name,
		Pos:  m.Pos,
	}
}

// Write file to temp and atomically move when everything else succeeds.
func writeMasterFile(filename string, data []byte, perm os.FileMode) error {
	// Create new file with position
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}

	if err = f.Close(); err != nil {
		return err
	}

	if perm != 0 {
		err = os.Chmod(f.Name(), perm)
	}

	return err
}
