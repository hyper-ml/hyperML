package workspace



func (a *apiServer) CreateDataset(name string) (*RepoAttrs, error) {
    //TODO: auth, validate repo name
  return a.InitTypedRepo(DATASET_REPO, name)
}

 