package nutsdb

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"github.com/mitchellh/go-homedir"
	"github.com/nutsdb/nutsdb"
	"github.com/tommyshem/ogi/cmd/issue"
)

// NutsStore represents the nuts database for a GitHub repository.
type NutsStore struct {
	Owner    string
	Repo     string
	DBNuts   *nutsdb.DB
	Filepath string
}

// nuts database bucket layout
//
// repo bucket structure
//   bucket name = owner-repo-bucket
//     key   = issue number
//     value = issue json data
//
//   bucket name owner-repo-map-bucket
//     key   = issue number
//     value = issue state
//
//   bucket name = owner-repo-comments-bucket
//     key   = issue number
//     value = comment json data

// BucketNameString returns the bucket name as a string for the current store,
// which is a combination of the GitHub owner and repo names separated by a hyphen.
func (store NutsStore) RepoBucketName() string {
	return string(fmt.Sprintf("%s-%s-%s", store.Owner, store.Repo, "bucket"))
}

// New creates a new NutsStore instance for the specified GitHub owner and repo.
// It initializes a NutsDB database and creates a bucket for storing issues
// if it does not already exist. Returns the initialized NutsStore and any error
// encountered during the database setup.
func New(owner string, repo string) (*NutsStore, error) {
	slog.Info(fmt.Sprintf("Creating NutsDB store for %s/%s", owner, repo))
	store := &NutsStore{Owner: owner, Repo: repo}
	// open the nuts database
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(location()),
	)
	if err != nil {
		slog.Info("Failed to open NutsDB database")
		log.Fatal(err)
		return store, err
	}
	defer db.Close()
	store.DBNuts = db
	// create repo bucket if it doesn't exist
	if err := db.Update(
		func(tx *nutsdb.Tx) error {
			return tx.NewBucket(nutsdb.DataStructureBTree, store.RepoBucketName())
		}); err != nil {
		fmt.Printf("create repo bucket failed")
		log.Fatal(err)
	}
	return store, err
}

// Clear deletes the entire bucket for the specified owner and repo,
// effectively clearing all locally stored issues. It then recreates the
// bucket to ensure that the bucket is ready for new issues to be saved.
func (store *NutsStore) Clear() error {
	return store.DBNuts.Update(
		func(tx *nutsdb.Tx) error {
			err := deleteBucket(tx, store.RepoBucketName())
			err = createBucket(tx, store.RepoBucketName())
			return err
		})
}

// Save persists the given issue to the local database. It uses a BTree bucket
// created from the owner and repo name, and creates two sub-buckets: one for
// the issue data itself, and another for a lookup table mapping issue numbers
// to their respective states.
func (store *NutsStore) Save(currentIssue issue.Issue) error {
	return store.DBNuts.Update(func(tx *nutsdb.Tx) error {
		// create the repo bucket if it doesn't exist
		err := createBucket(tx, store.RepoBucketName())
		if err != nil {
			return err
		}
		// get the data form the issue
		data, err := json.Marshal(currentIssue)
		if err != nil {
			return err
		}
		// save the issue
		err = tx.Put(store.RepoBucketName(), []byte(strconv.Itoa(*currentIssue.Number)), []byte(data), 0)
		if err != nil {
			return err
		}
		//create the _map bucket if it doesn't exist
		err = createBucket(tx, store.RepoBucketName()+"-map-bucket")
		if err != nil {
			return err
		}
		//map the issue number to its state
		return tx.Put(store.RepoBucketName()+"-map-bucket", []byte(strconv.Itoa(*currentIssue.Number)), []byte(*currentIssue.State), 0)

	})
}

// Get retrieves the issue with the specified number from the local database.
// It returns the issue with all associated comments and any error encountered
// during the retrieval process.
func (store *NutsStore) Get(issueNumber string) (issue.Issue, error) {
	//todo redo storage as nested buckets is not supported

	currentIssue := issue.Issue{}

	err := store.DBNuts.View(func(tx *nutsdb.Tx) error {

		value, err := tx.Get(store.RepoBucketName(), []byte(issueNumber))

		if value == nil {
			return fmt.Errorf("issue #%s was not found!", issueNumber)
		}
		err = json.Unmarshal(value, &currentIssue)
		return err
	})

	return currentIssue, err
}

// All retrieves all of the issues for the specified owner and repo from the
// local database. It returns all issues with all associated comments and any
// error encountered during the retrieval process.
func (store *NutsStore) All() ([]issue.Issue, error) {
	issues := []issue.Issue{}

	err := store.DBNuts.View(func(tx *nutsdb.Tx) error {

		iterator := nutsdb.NewIterator(tx, store.RepoBucketName(), nutsdb.IteratorOptions{Reverse: false})
		for {
			value, _ := iterator.Value()
			slog.Info("Key: " + string(iterator.Key()))
			slog.Info("Value: " + string(value))

			currentIssue := issue.Issue{}
			err := json.Unmarshal(value, &currentIssue)
			if err != nil {
				return err
			}
			issues = append(
				issues,
				currentIssue,
			)

			if !iterator.Next() {
				break
			}
		}
		err := tx.Commit()
		if err != nil {
			return err
		}

		return nil
	})
	return issues, err
}

func (store *NutsStore) AllByState(state string) ([]issue.Issue, error) {
	issues := []issue.Issue{}
	/*
		err := store.DBNuts.View(func(tx *nutsdb.Tx) error {

			pb := tx.Bucket(store.RepoBucketName())

			b := pb.Bucket([]byte(state))

			if b == nil {
				// the bucket doesn't exist, possibly because there are no
				// issues with this state. exit and move on.
				return nil
			}
			return b.ForEach(func(key, value []byte) error {
				currentIssue := issue.Issue{}
				err := json.Unmarshal(value, &currentIssue)
				if err != nil {
					return err
				}
				issues = append(issues, currentIssue)
				return nil
			})
		}) */
	return issues, nil
}

// createBucket creates a new bucket with the specified name in the given
// transaction. It is a BTree data structure. If the bucket already exists,
// this function does nothing and returns nil. If there is an error, it is
// logged and returned.
func createBucket(tx *nutsdb.Tx, bucketName string) error {
	if !existBucket(tx, bucketName) {
		err := tx.NewBucket(nutsdb.DataStructureBTree, bucketName)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

// deleteBucket deletes the bucket with the specified name in the given
// transaction. It is a BTree data structure. If there is an error, it
// is logged and returned.
func deleteBucket(tx *nutsdb.Tx, bucketName string) error {
	err := tx.DeleteBucket(nutsdb.DataStructureBTree, bucketName)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
func existBucket(tx *nutsdb.Tx, bucketName string) bool {
	exist := tx.ExistBucket(nutsdb.DataStructureBTree, bucketName)
	return exist
}

// location returns the full path to the nuts database file, which is
// a .db file in the user's home directory with a name of ".ogi-issues.db".
func location() string {
	dir, _ := homedir.Dir()
	dir, _ = homedir.Expand(dir)
	return dir //TODO fmt.Sprintf("%s/.ogi-issues.db", dir)
}

func debugPrintAllBuckets(store *NutsStore) error {
	if err := store.DBNuts.View(
		func(tx *nutsdb.Tx) error {
			return tx.IterateBuckets(nutsdb.DataStructureBTree, "*", func(bucket string) bool {
				fmt.Println("bucket: ", bucket)
				return true
			})
		}); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

// debugPrintLocation prints the location of the database file to the console.
// It uses the location function to retrieve the file path and formats it in a
// human-readable manner for debugging purposes.
func debugPrintLocation() {
	print("The database file location is : %s", location)
}
