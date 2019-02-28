package auth

import( 
  "encoding/json"
  db_pkg  "hyperflow.in/server/pkg/db"
)

type queryServer struct {
  db db_pkg.DatabaseContext
}

 
func newQueryServer(db db_pkg.DatabaseContext) *queryServer {
  return &queryServer {
    db: db,
  }
}

func (qs *queryServer) userKey(name string) string {
  return "user:" + name
}

func (qs *queryServer) InsertUserAttrs(name string, userAttts *UserAttrs) (error) {
  user_key := qs.userKey(name)
  
  if userAttts.PasswordHash == "" {
    return ErrEmptyPasswordHash()
  }

  err := qs.db.Insert(user_key, userAttts)
  if err != nil {
    return err
  }

  return nil
}


func (qs *queryServer) GetUserAttrs(name string) (*UserAttrs, error) {
  user_key := qs.userKey(name)
  raw_attrs, err := qs.db.Get(user_key)
  if err != nil {
    return nil, err
  }

  user_attrs := UserAttrs{}
  err = json.Unmarshal(raw_attrs, &user_attrs)
  if err != nil {
    return nil, err
  }

  return &user_attrs, nil
}