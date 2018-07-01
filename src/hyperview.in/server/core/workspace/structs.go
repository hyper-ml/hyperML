package workspace

import (
  "time"
  "sync"
)

//TODO: consider created, updated dates to all objects
// if not here in at DB level

type Repo struct {
  Name string 
}

// TODO: add created, updated etc
type RepoInfo struct {
  Repo *Repo 
  Size_bytes uint64
  Description string
  Branch *Branch 
}

type Branch struct {
  Repo *Repo
  Name string 
}

// maps branch to commits
type BranchInfo struct {
  Branch *Branch
  Head *Commit  
}

type Commit struct {
  Repo *Repo 
  Id  string 
  //TODO: add author, time etc
}


type CommitInfo struct {
  Commit *Commit 
  Parent_commit *Commit
  Child_commits []*Commit
  Description string
  Size_bytes uint64
  Started time.Time
  Finished time.Time
  Tree []*File
  FileMap map[string]*File
  treeLock sync.RWMutex
}

func (commitInfo *CommitInfo) AddFileToMap(file *File) (*CommitInfo, error) {
  commitInfo.FileMap[file.Path] = file
  return commitInfo, nil
}

func (commitInfo *CommitInfo) AddFile(file *File) (*CommitInfo, error) {
  commitInfo.Tree = append(commitInfo.Tree, file)
   return commitInfo, nil
}


type File struct {
  Commit *Commit 
  Path string
}


type FileInfo struct {
  File *File 
  FileType string
  SizeBytes int64 
  Object *Object 
  CheckSum string
}

type Object struct {
  Path string
  Hash string
}


func NewFileInfo(commit *Commit, filePath string, objectPath string, sizeBytes int64, checkSum string) (*FileInfo) {

  object := &Object{Path: objectPath}
  file := &File{Commit: commit, Path: filePath}
  file_info := &FileInfo{
    File: file, 
    FileType: "FILE",
    SizeBytes: sizeBytes, 
    Object: object, 
    CheckSum: checkSum} 

  return file_info
}








