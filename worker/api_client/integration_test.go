package api_client


import(
  "fmt"
  "testing"
  hf_test "github.com/hyper-ml/hyperml/worker/test"
)
 

func Test_SignedPutURL(t *testing.T) {
  test_config, err := hf_test.NewTestConfig()
  if err != nil {
    t.Fatalf("failed to generate test config ,err: %v", err)
  }  
 
  wc, err := NewWorkerClient(test_config.ServerURL.String())
  if err != nil {
    fmt.Println("failed to create a worker client, err: ", err)
  }

  tst_filepath := "/var/tmp/main.py"
  
  url_string, err := wc.SignedPutURL(test_config.Repo.Name, test_config.Branch.Name, test_config.Commit.Id, tst_filepath)
  if err != nil {
    fmt.Println("failed to get SignedPutURL, err: ", err)
  } 

  if url_string == "" {
    t.Fatalf("failed to retrieve signed PUT URL")
  }

  test_config.Destroy()
}

func Test_SignedPutPartURL(t *testing.T) {
  test_config, err := hf_test.NewTestConfig()
  if err != nil {
    t.Fatalf("failed to generate test config ,err: %v", err)
  }  

  wc, err := NewWorkerClient(test_config.ServerURL.String())
  if err != nil {
    fmt.Println("failed to create a worker client, err: ", err)
  }

  tst_filepath := "/var/tmp/part.py"
  
  url_string, err := wc.SignedPutPartURL(1, test_config.Repo.Name, test_config.Branch.Name, test_config.Commit.Id, tst_filepath)
  if err != nil {
    fmt.Println("failed to get SignedPutPartURL, err: ", err)
  }

  if url_string == "" {
    t.Fatalf("failed to retrieve signed PUT Part URL")
  }

  test_config.Destroy()
}


func Test_SendBytesToURL(t *testing.T) {
  
  test_config, err := hf_test.NewTestConfig()
  if err != nil {
    t.Fatalf("failed to generate test config ,err: %v", err)
  }  
  
  wc, err := NewWorkerClient(test_config.ServerURL.String())
  if err != nil {
    t.Fatalf("failed to create a worker client, err: %v", err)
  }

  tst_filepath := "/var/tmp/main.py"

  surl, err := wc.SignedPutURL(test_config.Repo.Name, test_config.Branch.Name, test_config.Commit.Id, tst_filepath)
  if err != nil {
    t.Fatalf("failed to get SignedPutURL, err: %v", err)
  }
  
  if surl == "" {
    t.Fatalf("failed to retrieve signed PUT URL")
  }

  sent, err := wc.SendBytesToURL(surl, []byte("test data"))
  if err != nil {
    t.Fatalf("failed to send data to Signed URL: %v", err)
  }

  get_surl, err:= wc.SignedGetURL(test_config.Repo.Name, test_config.Branch.Name, test_config.Commit.Id, tst_filepath)
  if err != nil {
    t.Fatalf("Failed to receive GET Signed URL, err: %v", err)
  }

  if get_surl == "" {
    t.Fatalf("Get Signed URL is empty.")
  }

  data, rcvd, err := wc.ReceiveBytesFromURL(get_surl)
  if err != nil {
    t.Fatalf("failed to receive bytes from URL, err: %v", err)
  }
 
  if sent != rcvd {
    t.Fatalf("failed to send and receive bytes through URL")
  }

  test_config.Destroy()

}


// create a file and check IN 


