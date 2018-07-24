package workspace


import (
  "fmt"
  "testing" 
  "hyperview.in/server/core/db"

)



func Test_NewCommit(t *testing.T) {
  
  db, err:= db.NewDatabaseContext(test_db_name, test_db_user, test_db_password)
  if err != nil {
    t.Fatalf("Test_NewCommit failed to create DB: %s", err)
  }

  commit_id:= "0e54042be69d423cb17711ece7582230"
  _, err = NewCommitTxn(TEST_REPO_NAME, commit_id, db)
  if err != nil {
    fmt.Println("Test_NewCommit Failed:", err)
    t.Fatalf("Test_NewCommit Failed: %s", err)
  }  
}   

func Test_AddFile(t *testing.T) {
  db, err:= db.NewDatabaseContext(test_db_name, test_db_user, test_db_password)
  if err != nil {
    t.Fatalf("Test_NewCommit failed to create DB: %s", err)
  }

  //commit_id:= "b2874bdc9cfd4d0cacd8b8091e90fa93"
  var commit_id string

  txn, err := NewCommitTxn(TEST_REPO_NAME, commit_id, db)
  if err != nil {
    fmt.Println("Test_NewCommit Failed:", err)
    t.Fatalf("Test_NewCommit Failed: %s", err)
  }  

  err = txn.FlushCommit()
  
  commit_id, err = txn.Start()
  
  if err != nil {
    t.Fatalf("Failed to.Start commit: %s", err)
  }

  err = txn.AddFile("/workspace/pattern.py", "/objects/34d234dsffedf3d", 11, "32edcdsf23dcsafewrf")
  if err != nil {
    fmt.Println("Test_AddFile Failed:", err)
    t.Fatalf("Test_AddFile Failed: %s", err)
  }  
  err = txn.End()
}



func Test_ListDir(t *testing.T) {
  db, _ := db.NewDatabaseContext(test_db_name, test_db_user, test_db_password)

  txn, _ := NewCommitTxn(TEST_REPO_NAME, "", db)
  //use existing commit
  commit_id, _ := txn.Start()
  fmt.Println("commit_id", commit_id)
  list, _ := txn.ListDir("/workspace")
  fmt.Println("list of files:", list)
}


func Test_LookupFile(t *testing.T) {

  db, _ := db.NewDatabaseContext(test_db_name, test_db_user, test_db_password)

  txn, _ := NewCommitTxn(TEST_REPO_NAME, "", db)
  //use existing commit
  commit_id, _ := txn.Start()
  fmt.Println("commit_id: ", commit_id)

  f, err := txn.LookupFile("/workspace/pattern.py")
  if err != nil {
    t.Fatalf("Error occured picking directory: %s", err)
  } 
  fmt.Println("f:", f)  
 
}


func Test_LookupDir(t *testing.T) {

  db, _ := db.NewDatabaseContext(test_db_name, test_db_user, test_db_password)

  txn, _ := NewCommitTxn(TEST_REPO_NAME, "", db)
  //use existing commit
  commit_id, _ := txn.Start()
  fmt.Println("commit_id: ", commit_id)

  d, err := txn.LookupFile("/workspace")
  if err != nil {
    t.Fatalf("Error occured picking directory: %s", err)
  }
  fmt.Println("d:", d)
 
}

/*
import (
  "testing"   
)

func Test_List(t *testing.T) {
  fm := make(map[string]*File)
  fm["/picaso/tan/3.txt"] = &File{Path: "/picaso/3.txt"}
  fm["/picaso/"] = &File{Path: "/picaso/"} 
  fm["/picaso/34.txt"] = &File{Path: "/picaso/34.txt"} 
  fm["/images/34.txt"] = &File{Path: "/images/34.txt"} 
  fm["/34.txt"] = &File{Path: "/34.txt"} 
  fm["/picaso/tan1/tan2/"] = &File{Path: "/picaso/tan1/tan2/"} 
  fm["/"] = &File{Path: "/"} 
  fm["/a"] = &File{Path: "/a"} 

  _ = list(fm, "/picaso/*")

}*/








