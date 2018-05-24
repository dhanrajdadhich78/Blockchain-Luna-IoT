package node

import (
	"os"
	"time"

	"github.com/boltdb/bolt"

	//"wizeBlock/wizeNode/core/log"
	"wizeBlock/wizeNode/core/network"
)

const nodesFileName = "nodeslist.db"
const nodesBucket = "nodes"

type NodesListStorage struct {
	DataDir string
}

func (s NodesListStorage) GetNodes() ([]network.NodeAddr, error) {
	//log.Info.Println("Get nodes...")
	db, err := s.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	nodes := []network.NodeAddr{}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(nodesBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			addr := string(v)
			node := network.NodeAddr{}
			node.LoadFromString(addr)
			nodes = append(nodes, node)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (s NodesListStorage) AddNodeToKnown(addr network.NodeAddr) {
	//log.Info.Printf("Add node [%s] to known...", addr)
	db, err := s.openDB()
	if err != nil {
		return
	}
	defer db.Close()

	err = db.Update(func(txdb *bolt.Tx) error {
		b := txdb.Bucket([]byte(nodesBucket))
		addr := addr.String()
		key := []byte(addr)

		return b.Put(key, key)
	})
}

func (s NodesListStorage) RemoveNodeFromKnown(addr network.NodeAddr) {
	db, err := s.openDB()
	if err != nil {
		return
	}
	defer db.Close()

	err = db.Update(func(txdb *bolt.Tx) error {
		b := txdb.Bucket([]byte(nodesBucket))
		addr := addr.String()
		key := []byte(addr)

		return b.Delete(key)
	})
}

func (s NodesListStorage) GetCountOfKnownNodes() (int, error) {
	//log.Info.Println("Get count of known nodes...")
	db, err := s.openDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	count := 0
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(nodesBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			count++
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s NodesListStorage) openDB() (*bolt.DB, error) {
	f, err := s.getDbFile()
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(f, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, err
	}

	return db, nil

}

func (s NodesListStorage) getDbFile() (string, error) {
	filePath := s.DataDir + nodesFileName

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// create empty DB
		db, err := bolt.Open(filePath, 0600, nil)
		if err != nil {
			return "", err
		}

		err = db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte(nodesBucket))
			if err != nil {
				return err
			}
			return nil
		})
		db.Close()

		if err != nil {
			return "", err
		}
	}
	return filePath, nil
}
