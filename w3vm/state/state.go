package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/sync/singleflight"
)

type forkState struct {
	Accounts     map[common.Address]*Account    `json:"accounts,omitempty"`
	HeaderHashes map[hexutil.Uint64]common.Hash `json:"headerHashes,omitempty"`
}

func (s *forkState) Merge(s2 *forkState) (changed bool) {
	if s2 == nil || len(s2.Accounts) == 0 && len(s2.HeaderHashes) == 0 {
		return false
	}

	// merge accounts
	if s.Accounts == nil {
		s.Accounts = s2.Accounts
		changed = changed || len(s2.Accounts) > 0
	} else {
		for addrS2, accS2 := range s2.Accounts {
			if accS1, ok := s.Accounts[addrS2]; ok {
				if accS1.Storage == nil {
					accS1.Storage = accS2.Storage
					changed = changed || len(accS2.Storage) > 0
				}

				for slotS2, valS2 := range accS2.Storage {
					if _, ok := accS1.Storage[slotS2]; ok {
						continue
					}

					accS1.Storage[slotS2] = valS2
					changed = true
				}
				continue
			}
			changed = true
			s.Accounts[addrS2] = accS2
		}
	}

	// merge header hashes
	if s.HeaderHashes == nil {
		s.HeaderHashes = s2.HeaderHashes
		changed = changed || len(s2.HeaderHashes) > 0
	} else {
		for blockNumber, hash := range s2.HeaderHashes {
			if _, ok := s.HeaderHashes[blockNumber]; ok {
				continue
			}
			changed = true
			s.HeaderHashes[blockNumber] = hash
		}
	}
	return
}

var readGroup = new(singleflight.Group)

func readTestdataState(fp string) (*forkState, error) {
	forkStateAny, err, _ := readGroup.Do(fp, func() (any, error) {
		f, err := os.Open(fp)
		if errors.Is(err, os.ErrNotExist) {
			return &forkState{}, nil
		} else if err != nil {
			return nil, err
		}
		defer f.Close()

		var s *forkState
		if err := json.NewDecoder(f).Decode(&s); err != nil {
			return nil, err
		}
		return s, nil
	})
	if err != nil {
		return nil, err
	}
	return forkStateAny.(*forkState), nil
}

var writeGroup = new(singleflight.Group)

func writeTestdataState(fp string, s *forkState) error {
Retry:
	_, err, shared := writeGroup.Do(fp, func() (any, error) {
		// read current testdata state
		testdataState, err := readTestdataState(fp)
		if err != nil {
			return nil, err
		}

		// merge states
		if testdataState == nil {
			testdataState = new(forkState)
		}
		if changed := testdataState.Merge(s); !changed {
			return nil, nil
		}

		dirPath := filepath.Dir(fp)
		if _, err := os.Stat(dirPath); errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(dirPath, 0775); err != nil {
				return nil, err
			}
		}

		// persist new state
		f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		dec := json.NewEncoder(f)
		dec.SetIndent("", "\t")
		if err := dec.Encode(testdataState); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	if shared {
		goto Retry
	}
	return nil
}
