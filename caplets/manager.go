package caplets

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/evilsocket/islazy/fs"
)

var (
	cache     = make(map[string]*Caplet)
	cacheLock = sync.Mutex{}
)

func List() []*Caplet {
	caplets := make([]*Caplet, 0)
	for _, searchPath := range LoadPaths {
		files, _ := filepath.Glob(searchPath + "/*" + Suffix)
		files2, _ := filepath.Glob(searchPath + "/*/*" + Suffix)

		for _, fileName := range append(files, files2...) {
			if _, err := os.Stat(fileName); err == nil {
				base := strings.Replace(fileName, searchPath+"/", "", -1)
				base = strings.Replace(base, Suffix, "", -1)

				if err, caplet := Load(base); err != nil {
					fmt.Fprintf(os.Stderr, "wtf: %v\n", err)
				} else {
					caplets = append(caplets, caplet)
				}
			}
		}
	}

	sort.Slice(caplets, func(i, j int) bool {
		return strings.Compare(caplets[i].Name, caplets[j].Name) == -1
	})

	return caplets
}

func Load(name string) (error, *Caplet) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if caplet, found := cache[name]; found {
		return nil, caplet
	}

	baseName := name
	names := []string{}
	if !strings.HasSuffix(name, Suffix) {
		name += Suffix
	}

	if name[0] != '/' {
		for _, path := range LoadPaths {
			names = append(names, filepath.Join(path, name))
		}
	} else {
		names = append(names, name)
	}

	for _, fileName := range names {
		if stats, err := os.Stat(fileName); err == nil {
			cap := &Caplet{
				Name: baseName,
				Path: fileName,
				Code: make([]string, 0),
				Size: stats.Size(),
			}
			cache[name] = cap

			if reader, err := fs.LineReader(fileName); err != nil {
				return fmt.Errorf("error reading caplet %s: %v", fileName, err), nil
			} else {
				for line := range reader {
					if line == "" || line[0] == '#' {
						continue
					}
					cap.Code = append(cap.Code, line)
				}
			}

			return nil, cap
		}
	}

	return fmt.Errorf("caplet %s not found", name), nil
}
