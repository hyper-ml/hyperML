package workspace

import (
  "time"
  "sync"
)

const (
  DefaultBranch = "master"
)

//TODO: consider created, updated dates to all objects
// if not here in at DB level

type Repo struct {
  Name string 
}

func NewRepo(name string) *Repo{
  return &Repo {
    Name: name,
  }
}

func RepoRef(name string) *Repo {
  return &Repo {
    Name: name,
  }
}

type RepoType string
const ( 
  STANDARD_REPO RepoType = "STANDARD"
  DATASET_REPO RepoType = "DATASET"
  MODEL_REPO RepoType = "MODEL"
  OUTPUT_REPO RepoType = "OUTPUT"
)

// TODO: add created, updated etc
type RepoAttrs struct {
  Repo *Repo 
  Type RepoType
  Size_bytes uint64
  Description string
  lock sync.RWMutex
  Branches map[string]*Branch 
  Datasets map[string]*Repo 
}

func (ra *RepoAttrs) AddBranch(branch *Branch) {
  ra.lock.Lock()
  defer ra.lock.Unlock()
  
  if ra.Branches == nil {
    ra.Branches = make(map[string]*Branch)
  }

  ra.Branches[branch.Name] = branch
//  return fm
}


type Branch struct {
  Repo *Repo
  Name string 
}

// maps branch to commits
type BranchAttrs struct {
  Branch *Branch
  Head *Commit  
}

type Commit struct {
  Repo *Repo 
  Id  string 
  //TODO: add author, time etc
}

func (c *Commit) IsEqual(p *Commit) bool{
  if p.Id == c.Id {
    return true
  }
  return false
}

type CommitAttrs struct {
  Commit *Commit 
  Parent *Commit 
  Children []*Commit
  Description string
  Size int64
  Started time.Time
  Finished time.Time
}

func (ci *CommitAttrs) Id() string{
  return ci.Commit.Id
}

func (ci *CommitAttrs) IsNew() bool {
  return ci.Started.IsZero()
}

func (ci *CommitAttrs) IsOpen() bool {
  return ci.Finished.IsZero()
}

func (ci *CommitAttrs) IsClosed() bool {
  return !ci.Finished.IsZero()
}

type FileMap struct {
  Commit *Commit
  Entries map[string]*File
  lock sync.RWMutex
}

func NewFileMap(commit *Commit) *FileMap{
  m := make(map[string] *File)
  return &FileMap{
    Commit: commit,
    Entries: m,
  }
}

func CopyFileMap(commit *Commit, refMap FileMap) *FileMap {
  m := NewFileMap(commit)
  for k,v := range refMap.Entries {
    m.Entries[k] = v
  }
  return m
}

// todo : add glob for search_path
func (fm *FileMap) Filepaths(search_path string) (paths []string) {
  
  if len(fm.Entries) == 0 {
    return paths
  }

  for k, _ := range fm.Entries {
    paths = append(paths, k)
  }

  return paths
}

func (fm *FileMap) Count() int{
  return len(fm.Entries)
}

func (fm *FileMap) GetFile(path string) *File {
  if file, ok := fm.Entries[path]; ok {
    return file
  }
  return nil
}

func (fm *FileMap) Add(file *File) {
  fm.lock.Lock()
  defer fm.lock.Unlock()
  fm.Entries[file.Path] = file
//  return fm
}

func (fm *FileMap) Remove(file *File) {
  fm.lock.Lock()
  defer fm.lock.Unlock()

  _, ok := fm.Entries[file.Path];

  if ok {
    delete(fm.Entries, file.Path)
  }
//  return fm
}

type File struct {
  Commit *Commit 
  Path string
}

type FileAttrsMap struct {
  Commit *Commit
  Entries map[string]*FileAttrs
}

type FileAttrs struct {
  File *File 
  FileType string
  SizeBytes int64 
  Object *Object 
  CheckSum string
}

func (f *FileAttrs) GetObjectHash() string {
  if f.Object != nil {
    return f.Object.Hash 
  } 
  
  return ""
}

func (f *FileAttrs) GetObjectPath() string {
  if f.Object != nil {
    return f.Object.Path
  } 
  
  return ""
}
func (f *FileAttrs) Size() int64{
  return f.SizeBytes
}

type Object struct {
  Path string
  Hash string
  Size int
  CheckSum string
  Parts []string 
}
 
func NewFileAttrs(commit *Commit, filePath string, objectPath string, sizeBytes int64, checkSum string) (*FileAttrs) {

  object := &Object{Path: objectPath, Hash: objectPath}
  file := &File{Commit: commit, Path: filePath}
  file_attrs := &FileAttrs{
    File: file, 
    FileType: "FILE",
    SizeBytes: sizeBytes, 
    Object: object, 
    CheckSum: checkSum} 

  return file_attrs
}

func NewDirInfo(commit *Commit, dirPath string, sizeBytes int64) (*FileAttrs) {
  
  file := &File{Commit: commit, Path: dirPath}
  dir_info := &FileAttrs{
    File: file, 
    FileType: "DIR",
    SizeBytes: sizeBytes} 

  return dir_info
}







