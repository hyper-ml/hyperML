package git


import (
  "gopkg.in/src-d/go-git.v4"
  "gopkg.in/src-d/go-billy.v4/osfs"
  "gopkg.in/src-d/go-git.v4/plumbing/cache"

  "gopkg.in/src-d/go-git.v4/storage/filesystem"
)



func Init(dir string) (*git.Repository, error) {
  fs := osfs.New(dir)
  dot, _ := fs.Chroot(".git")
  storage := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())
  return git.Init(storage, fs) 
}

  