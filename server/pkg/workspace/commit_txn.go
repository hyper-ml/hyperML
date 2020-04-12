package workspace

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"time"
)

const (
	defaultBranch = "master"
)

//TODO: handle multiple branches

type commitTxn struct {
	repoName   string
	branchName string
	attrs      *CommitAttrs
	q          *queryServer
}

func NewCommitTxn(repoName string, branchName string, commitId string, q *queryServer) (*commitTxn, error) {

	if repoName == "" || branchName == "" {
		return nil, fmt.Errorf("Repo and Branch are mandatory")
	}

	txn := &commitTxn{
		repoName:   repoName,
		branchName: branchName,
		q:          q,
	}

	battrs, err := q.GetBranchAttrs(repoName, branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branch attributes: %v", err)
	}

	if commitId != "" {

		// verify commitId is valid for a new txn
		// commit is branch head and closed  -> create new
		// commit parent is branch head and current commit is open -- > continue
		// else raise invalid commit error
		commit_attrs, err := q.GetCommitAttrsById(repoName, commitId)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch commit attributes: %v", err)
		}

		switch {
		// new branch
		case battrs.Head == nil:
			txn.attrs = commit_attrs

		// recent but closed commit
		case battrs.Head.IsEqual(commit_attrs.Commit):
			if commit_attrs.IsClosed() {

				txn.attrs = NewCommit(battrs.Branch.Repo, battrs.Branch, battrs.Head)

			} else {
				return nil, fmt.Errorf("Invalid commit: expected a closed status")
			}

		// open but valid commit
		case battrs.Head.IsEqual(commit_attrs.Parent):
			txn.attrs = commit_attrs

		default:
			return nil, fmt.Errorf("stale commit. ")
		}

	} else {
		// if commit iD is not passed

		// verify branch is open for new txn
		// check if branch has any commits
		// check current head is closed then create a new

		switch {
		case battrs.Head == nil:
			txn.attrs = NewCommit(battrs.Branch.Repo, battrs.Branch, nil)

		default:
			head_attrs, err := q.GetCommitAttrsById(repoName, battrs.Head.Id)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve head commit: %v", err)
			}

			if head_attrs.IsOpen() {
				return nil, fmt.Errorf("Unexpected status of head commit: Open")
			} else {

				txn.attrs = NewCommit(battrs.Branch.Repo, battrs.Branch, battrs.Head)
			}
		}
	}

	return txn, nil
}

// save commit attrs to database
//
func (ct *commitTxn) OpenCommit() (*CommitAttrs, error) {

	// commit is already in db
	if !ct.attrs.IsNew() {
		return ct.attrs, nil
	}

	if ct.attrs.IsClosed() {
		return nil, fmt.Errorf("commit is already closed")
	}

	// add file map
	if err := ct.addFileMap(ct.attrs.Commit, ct.attrs.Parent); err != nil {
		return nil, err
	}

	// insert comit attrs in db
	if err := ct.q.InsertCommitAttrs(ct.repoName, ct.attrs.Commit.Id, ct.attrs); err != nil {
		return nil, err
	}

	return ct.attrs, nil
}

func (ct *commitTxn) CloseCommit() error {
	// check parent is head
	// change status
	// scoopHead
	// catch errors

	if ct.attrs == nil {
		return fmt.Errorf("Txn is invalid. Please initiate again.")
	}

	if !ct.attrs.IsOpen() {
		return fmt.Errorf("Commit is already closed")
	}

	battrs, err := ct.q.GetBranchAttrs(ct.repoName, ct.branchName)
	if err != nil {
		return fmt.Errorf("Invalid branch on commit transaction: %s", ct.branchName)
	}

	// if branch head is empty or parent is still at the head, we are good
	if battrs.Head == nil || battrs.Head.IsEqual(ct.attrs.Parent) {
		if err := ct.scoopHead(battrs, ct.attrs.Commit); err != nil {
			return fmt.Errorf("failed to scoop head to current commit, err: %v", err)
		}

		// close commit not
		ct.attrs.Finished = time.Now()
		ct.attrs.Size = ct.getSize()

		return ct.q.UpdateCommitAttrs(ct.repoName, ct.attrs.Id(), ct.attrs)

	} else {
		return fmt.Errorf("The remote branch contains changes that may be missing in local.")
	}

	return nil
}

func (ct *commitTxn) setCommitAttrs(c *CommitAttrs) {
	ct.attrs = c
}

func (ct *commitTxn) GetCommitAttrs() *CommitAttrs {
	return ct.attrs
}

func (ct *commitTxn) IsOpenCommit() bool {
	if !ct.attrs.IsOpen() {
		return false
	}
	return true
}

func (ct *commitTxn) setCommitAttrsByBranch() error {
	commit_attrs, err := ct.q.GetCommitAttrsByBranch(ct.repoName, ct.branchName)
	if err != nil {
		return err
	}
	ct.attrs = commit_attrs
	return nil

}

func (ct *commitTxn) addFileMap(commit *Commit, parent *Commit) error {
	var err error
	var fm *FileMap = NewFileMap(commit)

	if parent != nil {
		if parent.Id != "" {
			fm, err = ct.q.GetFileMap(ct.repoName, parent.Id)
			if err != nil {
				return fmt.Errorf("error gettign file map of parent commit, err: %v", err)
			}
		}
	}

	return ct.q.InsertFileMap(ct.repoName, commit.Id, fm)
}

func (ct *commitTxn) scoopHead(branchInfo *BranchAttrs, commit *Commit) error {
	branch := branchInfo.Branch
	repo := branchInfo.Branch.Repo

	branchInfo.Head = commit
	err := ct.q.UpdateBranchAttrs(repo.Name, branch.Name, branchInfo)

	return err
}

func (ct *commitTxn) getSize() int64 {
	var size int64
	repo_name := ct.repoName
	branch_name := ct.branchName
	var commit_id string

	if ct.attrs != nil {
		commit_id = ct.attrs.Id()
	}

	if commit_id == "" || repo_name == "" {
		base.Warn("[commitTxn.getCommitSize] Failed to get size of un-initialized commit txn.")
		return size
	}

	file_map, _ := ct.q.GetFileMap(repo_name, commit_id)

	if len(file_map.Entries) == 0 {
		return size
	}

	for fname, _ := range file_map.Entries {
		f_attrs, err := ct.q.GetFileAttrs(repo_name, commit_id, fname)
		if err != nil {
			base.Debug("[commitTxn.GetCommitSize] Failed to find size of file: ", repo_name, commit_id, fname)
			continue
		}
		size = size + f_attrs.Size()
	}

	base.Info("[commitTxn.GetSize] Size of Repo: ", size, repo_name, branch_name, commit_id)
	return size
}

func (ct *commitTxn) End() error {
	var err error
	if ct.attrs == nil {
		base.Log("finishCommit: Could not fetch any open commit for repo %s", ct.repoName)
		return fmt.Errorf("finishCommit: Could not fetch any open commit for repo %s", ct.repoName)
	}

	if ct.attrs.IsOpen() {

		ct.attrs.Finished = time.Now()
		ct.attrs.Size = ct.getSize()

		err = ct.q.UpdateCommitAttrs(ct.repoName, ct.attrs.Id(), ct.attrs)
		return err
	} else {
		base.Log("finishCommit: No open commit for this repo", ct.repoName)
		return fmt.Errorf("No open commit for this repo: %s", ct.repoName)
	}

}

func (ct *commitTxn) insertFileAttrs(filePath string, object string, size int64, cs string) (*FileAttrs, error) {

	var err error
	file_attrs := NewFileAttrs(ct.attrs.Commit, filePath, object, size, cs)

	//TODO: get file info in return
	err = ct.q.UpsertFileAttrs(ct.repoName, ct.attrs.Id(), filePath, file_attrs)
	if err != nil {
		base.Log("Failed to update file map:", filePath, object, size)
		return nil, err
	}

	err = ct.updateFileMap(filePath)
	if err != nil {
		base.Log("Failed to update file map:", filePath, object, size)
		return nil, err
	}

	return file_attrs, nil
}

func (ct *commitTxn) insertDirInfo(filePath string, size int64) (*FileAttrs, error) {
	var err error
	dir_info := NewDirInfo(ct.attrs.Commit, filePath, size)
	err = ct.q.UpsertFileAttrs(ct.repoName, ct.attrs.Id(), filePath, dir_info)

	if err != nil {
		return nil, err
	}
	err = ct.updateFileMap(filePath)

	if err != nil {
		base.Log("Failed to update file map:", filePath, size)
		return nil, err
	}

	return dir_info, nil
}

func (ct *commitTxn) updateFileMap(filePath string) error {
	newfile := &File{Commit: ct.attrs.Commit, Path: filePath}
	return ct.q.AddFileToMap(ct.repoName, ct.attrs.Id(), newfile)
}

func (ct *commitTxn) AddFile(filePath string, objectName string, size int64, cs string) (*FileAttrs, error) {

	if ct.attrs == nil {
		return nil, fmt.Errorf("Please initiate commit txn first.")
	}

	if !ct.IsOpenCommit() {
		return nil, fmt.Errorf("This repo has no open commit. Please initialize commit before adding files.")
	}

	if objectName == "" {
		return ct.insertDirInfo(filePath, size)
	}

	return ct.insertFileAttrs(filePath, objectName, size, cs)
}

func (ct *commitTxn) AddDir(filePath string, size int64) (*FileAttrs, error) {

	// TODO: get the latest commit info to avoid concurrency issues
	if !ct.attrs.Finished.IsZero() {
		return nil, fmt.Errorf("This repo has no open commit. Please initialize commit before adding files.")
	}

	return ct.insertDirInfo(filePath, size)
}

func (ct *commitTxn) FlushCommit() error {
	//delete commit and the file map
	return ct.Delete()
}

func (ct *commitTxn) Delete() error {
	// delete commit
	var err error
	var branch_attr *BranchAttrs

	if ct.attrs == nil {
		if err = ct.setCommitAttrsByBranch(); err != nil {
			return err
		}
	}

	if !ct.IsOpenCommit() {
		return fmt.Errorf("This repo has no open commit to flush")
	}

	if ct.attrs.Parent != nil {
		branch_attr, err = ct.q.GetBranchAttrs(ct.repoName, ct.branchName)
		if err != nil {
			base.Log("Invalid repo or branch name:", ct.repoName, ct.branchName)
			return err
		}
		if err := ct.scoopHead(branch_attr, ct.attrs.Parent); err != nil {
			base.Log("Unable to scoop branch head to parent:", ct.attrs.Parent.Id)
			return err
		}
	}

	return ct.q.DeleteCommitAttrs(ct.repoName, ct.attrs.Id())
}
