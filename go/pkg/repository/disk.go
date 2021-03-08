package repository

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"

	"github.com/replicate/keepsake/go/pkg/errors"
	"github.com/replicate/keepsake/go/pkg/files"
)

type DiskRepository struct {
	rootDir string
}

func NewDiskRepository(rootDir string) (*DiskRepository, error) {
	return &DiskRepository{
		rootDir: rootDir,
	}, nil
}

func (s *DiskRepository) RootURL() string {
	return "file://" + s.rootDir
}

// Get data at path
func (s *DiskRepository) Get(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(pathpkg.Join(s.rootDir, path))
	if err != nil && os.IsNotExist(err) {
		return nil, errors.DoesNotExist(fmt.Sprintf("Get: path does not exist: %v", path))
	}
	return data, err
}

// GetPath recursively copies repoDir to localDir
func (s *DiskRepository) GetPath(repoDir string, localDir string) error {
	if err := copy.Copy(pathpkg.Join(s.rootDir, repoDir), localDir); err != nil {
		return errors.ReadError(fmt.Sprintf("Failed to copy directory from %s to %s: %v", repoDir, localDir, err))
	}
	return nil
}

// GetPathTar extracts tarball `tarPath` to `localPath`
//
// See repository.go for full documentation.
func (s *DiskRepository) GetPathTar(tarPath, localPath string) error {
	fullTarPath := pathpkg.Join(s.rootDir, tarPath)
	exists, err := files.FileExists(fullTarPath)
	if err != nil {
		return err
	}
	if !exists {
		return errors.DoesNotExist(fmt.Sprintf("Path does not exist: " + fullTarPath))
	}
	if err := extractTar(fullTarPath, localPath); err != nil {
		return err
	}
	return nil
}

func (s *DiskRepository) GetPathItemTar(tarPath, itemPath, localPath string) error {
	fullTarPath := pathpkg.Join(s.rootDir, tarPath)
	exists, err := files.FileExists(fullTarPath)
	if err != nil {
		return err
	}
	if !exists {
		return errors.DoesNotExist("Path does not exist: " + fullTarPath)
	}
	return extractTarItem(fullTarPath, itemPath, localPath)
}

// Put data at path
func (s *DiskRepository) Put(path string, data []byte) error {
	fullPath := pathpkg.Join(s.rootDir, path)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return errors.WriteError(err.Error())
	}
	if err := ioutil.WriteFile(fullPath, data, 0644); err != nil {
		return errors.WriteError(err.Error())
	}
	return nil
}

// PutPath recursively puts the local `localPath` directory into path `repoPath` in the repository
func (s *DiskRepository) PutPath(localPath string, repoPath string) error {
	files, err := getListOfFilesToPut(localPath, repoPath)
	if err != nil {
		return errors.WriteError(err.Error())
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(file.Source)
		if err != nil {
			return errors.WriteError(err.Error())
		}
		err = s.Put(file.Dest, data)
		if err != nil {
			return errors.WriteError(err.Error())
		}
	}
	return nil
}

// PutPathTar recursively puts the local `localPath` directory into a tar.gz file `tarPath` in the repository
// If `includePath` is set, only that will be included.
//
// See repository.go for full documentation.
func (s *DiskRepository) PutPathTar(localPath, tarPath, includePath string) error {
	if !strings.HasSuffix(tarPath, ".tar.gz") {
		return errors.WriteError("PutPathTar: tarPath must end with .tar.gz")
	}

	fullPath := pathpkg.Join(s.rootDir, tarPath)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return errors.WriteError(err.Error())
	}

	tarFile, err := os.Create(fullPath)
	if err != nil {
		return errors.WriteError(err.Error())
	}
	defer tarFile.Close()

	if err := putPathTar(localPath, tarFile, filepath.Base(tarPath), includePath); err != nil {
		return err
	}

	// Explicitly call Close() on success to capture error
	if err := tarFile.Close(); err != nil {
		return errors.WriteError(err.Error())
	}
	return nil
}

// Delete deletes path. If path is a directory, it recursively deletes
// all everything under path
func (s *DiskRepository) Delete(pathToDelete string) error {
	if err := os.RemoveAll(pathpkg.Join(s.rootDir, pathToDelete)); err != nil {
		return errors.WriteError(fmt.Sprintf("Failed to delete %s/%s: %v", s.rootDir, pathToDelete, err))
	}
	return nil
}

// List files in a path non-recursively
//
// Returns a list of paths, prefixed with the given path, that can be passed straight to Get().
// Directories are not listed.
// If path does not exist, an empty list will be returned.
func (s *DiskRepository) List(path string) ([]string, error) {
	files, err := ioutil.ReadDir(pathpkg.Join(s.rootDir, path))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, errors.ReadError(err.Error())
	}
	result := []string{}
	for _, f := range files {
		if !f.IsDir() {
			result = append(result, pathpkg.Join(path, f.Name()))
		}
	}
	return result, nil
}

func (s *DiskRepository) ListTarFile(tarPath string) ([]string, error) {
	fullTarPath := pathpkg.Join(s.rootDir, tarPath)
	exists, err := files.FileExists(fullTarPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.DoesNotExist("Path does not exist: " + fullTarPath)
	}

	files, err := getListOfFilesInTar(fullTarPath)
	if err != nil {
		return nil, err
	}

	tarname := filepath.Base(strings.TrimSuffix(tarPath, ".tar.gz"))
	for idx := range files {
		files[idx] = strings.TrimPrefix(files[idx], tarname+"/")
	}

	return files, nil
}

func (s *DiskRepository) ListRecursive(results chan<- ListResult, folder string) {
	err := filepath.Walk(pathpkg.Join(s.rootDir, folder), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(s.rootDir, path)
			if err != nil {
				return err
			}

			md5sum, err := md5File(path)
			if err != nil {
				return err
			}
			results <- ListResult{Path: relPath, MD5: md5sum}
		}
		return nil
	})
	if err != nil {
		// If directory does not exist, treat this as empty. This is consistent with how blob storage
		// would behave
		if os.IsNotExist(err) {
			close(results)
			return
		}
		results <- ListResult{Error: errors.ReadError(err.Error())}
	}
	close(results)
}

func (s *DiskRepository) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	err := filepath.Walk(pathpkg.Join(s.rootDir, folder), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == filename {
			relPath, err := filepath.Rel(s.rootDir, path)
			if err != nil {
				return err
			}
			results <- ListResult{Path: relPath}
		}
		return nil
	})
	if err != nil {
		// If directory does not exist, treat this as empty. This is consistent with how blob storage
		// would behave
		if os.IsNotExist(err) {
			close(results)
			return
		}

		results <- ListResult{Error: errors.ReadError(err.Error())}
	}
	close(results)
}

func md5File(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
