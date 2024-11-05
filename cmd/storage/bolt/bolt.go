package bolt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/boltdb/bolt" //TODO change to bbolt for updates as package
	"github.com/mitchellh/go-homedir"
	"github.com/tommyshem/ogi/cmd/issue"
)

type Store struct {
	Owner  string
	Repo   string
	DBBolt *bolt.DB
}

func (s Store) BucketName() []byte {
	return []byte(fmt.Sprintf("%s-%s", s.Owner, s.Repo))
}

// BucketNameString returns the bucket name as a string for the current store,
// which is a combination of the GitHub owner and repo names separated by a hyphen.
func (s Store) BucketNameString() string {
	return string(fmt.Sprintf("%s-%s", s.Owner, s.Repo))
}

// New creates a new Store instance for the specified GitHub owner and repo.
// It initializes a BoltDB database and creates a bucket for storing issues
// if it does not already exist. Returns the initialized Store and any error
// encountered during the database setup.
func New(owner string, repo string) (*Store, error) {
	s := &Store{Owner: owner, Repo: repo}
	// open the bolt database
	db, err := bolt.Open(Location(), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		slog.Info("New bolt db could not be opened filename :" + Location())
		return s, err
	}
	s.DBBolt = db
	// create bucket if it doesn't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(s.BucketName())
		return err
	})
	return s, err
}

// Clear deletes the entire bucket for the specified owner and repo,
// effectively clearing all locally stored issues. It then recreates the
// bucket to ensure that the bucket is ready for new issues to be saved.
func (s *Store) Clear() error {
	return s.DBBolt.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(s.BucketName())
		_, err := tx.CreateBucketIfNotExists(s.BucketName())
		return err
	})
}

// Save persists the given issue to the local database. It uses a BoltDB bucket
// created from the owner and repo name, and creates two sub-buckets: one for
// the issue data itself, and another for a lookup table mapping issue numbers
// to their respective states.
func (s *Store) Save(is issue.Issue) error {

	return s.DBBolt.Update(func(tx *bolt.Tx) error {

		pb, _ := tx.CreateBucketIfNotExists(s.BucketName())

		b, _ := pb.CreateBucketIfNotExists([]byte(*is.State))

		//data
		data, err := json.Marshal(is)
		if err != nil {
			return err
		}
		//
		err = b.Put([]byte(strconv.Itoa(*is.Number)), data)
		if err != nil {
			return err
		}

		inb, _ := pb.CreateBucketIfNotExists([]byte("_map"))

		return inb.Put([]byte(strconv.Itoa(*is.Number)), []byte(*is.State))
	})
}

// Get retrieves the issue with the specified number from the local database.
// It returns the issue with all associated comments and any error encountered
// during the retrieval process.
func (s *Store) Get(number string) (issue.Issue, error) {

	id := []byte(number)

	i := issue.Issue{}

	err := s.DBBolt.View(func(tx *bolt.Tx) error {

		pb := tx.Bucket(s.BucketName())

		inb := pb.Bucket([]byte("_map"))
		bn := inb.Get(id)

		b := pb.Bucket(bn)
		v := b.Get(id)

		if v == nil {
			return fmt.Errorf("issue #%s was not found!", number)
		}
		err := json.Unmarshal(v, &i)
		return err
	})

	return i, err
}

// All retrieves all of the issues for the specified owner and repo from the
// local database. It returns all issues with all associated comments and any
// error encountered during the retrieval process.
func (s *Store) All() ([]issue.Issue, error) {
	issues := []issue.Issue{}

	err := s.DBBolt.View(func(tx *bolt.Tx) error {
		for _, state := range []string{"open", "closed"} {
			si, err := s.AllByState(state)
			if err != nil {
				return err
			}
			issues = append(issues, si...)
		}
		return nil
	})
	return issues, err
}

// AllByState retrieves all of the issues for the specified owner and repo with
// the specified state from the local database. It returns all issues with all
// associated comments and any error encountered during the retrieval process.
func (s *Store) AllByState(state string) ([]issue.Issue, error) {
	issues := []issue.Issue{}

	s.DBBolt.View(func(tx *bolt.Tx) error {
		pb := tx.Bucket(s.BucketName())
		b := pb.Bucket([]byte(state))
		if b == nil {
			// the bucket doesn't exist, possibly because there are no
			// issues with this state. exit and move on.
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			i := issue.Issue{}
			err := json.Unmarshal(v, &i)
			if err != nil {
				return err
			}
			issues = append(issues, i)
			return nil
		})
	})
	return issues, nil
}

// Location returns the path to the Bolt database file.
// It is in the user's home directory, with a name of ".ogi-issues.db".
func Location() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		return ""
	}
	expandedDir, _ := homedir.Expand(homeDir)
	return fmt.Sprintf("%s/.ogi-issues.db", expandedDir)
}
