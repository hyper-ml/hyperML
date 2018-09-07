package workspace



func (a *apiServer) CreateDataset(name string) error {
    //TODO: auth, validate repo name
    
  newRepo := &Repo {
      Name: name,
  }

  //newBranch := &Branch {Repo: newRepo, Name: "master"}

  newRepoAttrs :=  &RepoAttrs {
    Repo: newRepo,
  }
  
  newRepoAttrs.Description = "New Data Repo"
  newRepoAttrs.Type = DATASET

  err := a.db.Insert(DATA_REPO_KEY_PREFIX + name, newRepoAttrs)

  return err 
}

 