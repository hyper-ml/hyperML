package workspace 
 

type RepoMessage struct {
  Repo *Repo  
  Branch *Branch  
  Commit *Commit  
  FileMap *FileMap  
}

type StandardRepoMessage struct {
  RepoMessage
  Output *RepoMessage  
  Model *RepoMessage 
}

type RepoAttrsMessage struct {
  RepoAttrs *RepoAttrs  
  BranchAttrs *BranchAttrs  
  CommitAttrs *CommitAttrs  
  FileMap *FileMap  
}
 
type StdRepoAttrsMessage struct {
  RepoAttrsMessage
  OutputAttrs *RepoAttrsMessage  
  ModelAttrs *RepoAttrsMessage  
}

type FileMessage struct { 
  MessageType FileMessageType  
  Repo *Repo  
  Branch *Branch  
  Commit *Commit  
  File *File  
  GetURL string  
  PutURL string 
}

type FileMessageType string
const (
  FileMessageURLGet   FileMessageType = "GET"
  FileMessageURLPut   FileMessageType = "PUT"
  FileMessageCheckIn  FileMessageType = "CHECK_IN"
  FileMessagePut FileMessageType = "PUT_FILE"
)

type FileAttrsMessage struct {
  MessageType FileMessageType  
  Repo *Repo  
  Branch *Branch  
  Commit *Commit  
  FileAttrs *FileAttrs  
  GetURL string  
  PutURL string  
}

type FilePartsMessage struct {
  Sequences []int  
  FileParts []Object  
  File *File  
  Repo *Repo   
  Branch *Branch   
  Commit *Commit  
  FileAttrs *FileAttrs  
}

type FilePartMessage struct {
  Seq int  
  MessageType FileMessageType  
  File *File  
  FilePart *Object  
  Repo *Repo  
  Branch *Branch  
  Commit *Commit  
  GetURL string  
  PutURL string  
}
  
type PutFileResponse struct {
  FileAttrs *FileAttrs  
  Written int64   
  Error string  
}

type CommitSizeRequest struct {
  Repo Repo  
  Branch Branch  
  CommitId string  
}

type CommitSizeResponse struct {
  Size int64  
}











