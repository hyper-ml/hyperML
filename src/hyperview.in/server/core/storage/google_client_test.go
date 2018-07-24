package storage


import (
  "io"
  "io/ioutil"
  "strings"
  "testing"
  "golang.org/x/net/context"
  gstore "cloud.google.com/go/storage"  
  "hyperview.in/server/base"
)


const (
  TEST_GCS_BUCKET = "hyperview001"
  TEST_GCS_OBJECT_NAME = "test_dir/objects/Test_987654321"
)

func Test_UpdateGCSFile(t *testing.T) {
  ctx := context.Background()

  client, err := newGoogleClient(ctx, TEST_GCS_BUCKET)
  if err != nil {
    base.Log("Test_UpdateGCSFile(): Failed to create GCS Bucket handle")
    t.Fail()
  }

  if client.Exists(TEST_GCS_OBJECT_NAME) {
    base.Log("Test_UpdateGCSFile(): Cleaning test objects.")
    err = client.Delete(TEST_GCS_OBJECT_NAME)
  }

  wh, err := client.Writer(TEST_GCS_OBJECT_NAME)
  
  written, err := io.Copy(wh, strings.NewReader("First Line \n")) 
  err = wh.Close()
  if err == nil {
    base.Log("[Test_UpdateGCSFile] New object created.", TEST_GCS_OBJECT_NAME)
    base.Log("[Test_UpdateGCSFile] bytes: ", written)
  } else {
    base.Log("[Test_UpdateGCSFile]: New object create failed. ", err)
    t.Fail()
  }

  mh, err := client.Writer(TEST_GCS_OBJECT_NAME)
  if err != nil {
    base.Log("[Test_UpdateGCSFile] Failed to create Modify Object Writer: ", err)
  }

  bytes_wr, err := mh.Write([]byte("Second Line. \n Third Line")) 
  //io.Copy(mh, strings.NewReader("Second Line. \n Third Line")) 

  err =  mh.Close()

  if err == nil {
    base.Log("[Test_UpdateGCSFile] Object modified.", TEST_GCS_OBJECT_NAME)
    base.Log("[Test_UpdateGCSFile] bytes written: ", bytes_wr)
    size, _ := client.Size(TEST_GCS_OBJECT_NAME)
    base.Log("[Test_UpdateGCSFile] Size: ", size)

  } else {
    base.Log("[Test_UpdateGCSFile]: object update failed. ", err)
    t.Fail()

  }

  rh, err := client.Reader(TEST_GCS_OBJECT_NAME, 0, 0)
  c, err := ioutil.ReadAll(rh)
  if err != nil {
    base.Log("[Test_UpdateGCSFile]: object read failed. ", err)
    t.Fail()
  }
  base.Log("[Test_UpdateGCSFile]: Content of test object: ", string(c))


}


func Test_DirectGCS(t *testing.T) {
  ctx := context.Background()
  client, err := gstore.NewClient(ctx)
  if err != nil {
    base.Log("failed to create GCS Client")
  }

  wc := client.Bucket("hyperview001").Object(TEST_GCS_OBJECT_NAME).NewWriter(ctx)
  //wc.ContentType = "text/plain"
  //wc.ACL = []storage.ACLRule{{storage.AllUsers, storage.RoleReader}}

  if _, err := wc.Write([]byte("hello world12345633243434")); err != nil {
    base.Log("Error writing:", err)
  }
  if err := wc.Close(); err != nil {
      // TODO: handle error.
    base.Log("Close err:", err)
  }
  base.Log("updated object:", wc.Attrs())

}