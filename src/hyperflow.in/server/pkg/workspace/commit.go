package workspace

import(
  "hyperflow.in/server/pkg/utils"
)

func NewCommit(repo *Repo, branch *Branch, parent *Commit) (*CommitAttrs) {

  var commit_attrs *CommitAttrs 
  commit_id := utils.NewUUID()

  commit := &Commit {
    Repo: repo,
    Id: commit_id,
  }

  commit_attrs = &CommitAttrs{
    Commit: commit,
    Parent: parent,  
  }

  return commit_attrs
}