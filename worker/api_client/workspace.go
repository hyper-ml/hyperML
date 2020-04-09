package api_client


import(
  "fmt"
  "bytes"
  "net/http"
  "io/ioutil"
  "encoding/json" 
  ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
  
)

func (wc *WorkerClient) SignedGetURL(repoId, branchName, commitId, filepath string) (string, error) {
  return wc.getSignedURL(ws.FileMessageURLGet, repoId, branchName, commitId, filepath)
}

func (wc *WorkerClient) SignedPutURL(repoId, branchName, commitId, filepath string) (string, error) {
  return wc.getSignedURL(ws.FileMessageURLPut, repoId, branchName, commitId, filepath)
}

func (wc *WorkerClient) getSignedURL(op ws.FileMessageType, repoId, branchName, commitId, filepath string) (string, error) {

  c := newRestClient(wc.BaseURL, CommitFileURL)
  req := c.Get()
  
  file_msg := ws.FileMessage{
    MessageType: op,
    Repo: &ws.Repo{
        Name: repoId,
      },
    Commit: &ws.Commit{
        Id: commitId,
      },
    File: &ws.File{
        Path: filepath,
      },
    Branch: &ws.Branch{
      Name: branchName, 
    },
  }
 
  rq_msg, _ := json.Marshal(&file_msg) 
  _ = req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(rq_msg)))

  resp := req.Do()
  json_resp, err := resp.Raw()  
  if err != nil { 
    return "", fmt.Errorf("recvd failure from %s err: %v", req.URL(), err)
  }
  
  if err = json.Unmarshal(json_resp, &file_msg); err != nil {
    return "", err
  }

  switch op {
    case ws.FileMessageURLGet: 
      return file_msg.GetURL, nil
    case ws.FileMessageURLPut:
      return file_msg.PutURL, nil
  }

  return "", fmt.Errorf("something unexpected occurred")
}

func (wc *WorkerClient) SignedGetPartURL(seq int, repoId, branchName, commitId, filepath string) (string, error) {
  return wc.getPartSignedURL(seq, ws.FileMessageURLGet, repoId, branchName, commitId, filepath)
}

func (wc *WorkerClient) SignedPutPartURL(seq int, repoId, branchName, commitId, filepath string) (string, error) {
  return wc.getPartSignedURL(seq, ws.FileMessageURLPut, repoId, branchName, commitId, filepath)
}

// op needs to be http.MethodGet or http.MethodPut
func (wc *WorkerClient) getPartSignedURL(seq int, op ws.FileMessageType, repoId, branchName, commitId, filepath string) (string, error) {
  
  c := newRestClient(wc.BaseURL, CommitFilePartURL)
  req := c.Get()
  
  pm := ws.FilePartMessage{
    Seq: seq,
    MessageType: op,
    Repo: &ws.Repo{
        Name: repoId,
      },
    Branch: &ws.Branch{
      Name: branchName, 
    },
    Commit: &ws.Commit{
        Id: commitId,
      },
    File: &ws.File{
        Path: filepath,
      },
  }

  // set req_msg
  rq_msg, _ := json.Marshal(&pm) 
  _ = req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(rq_msg)))

  resp := req.Do()
  json_resp, err := resp.Raw()  
  if err != nil {
    return "", fmt.Errorf("recvd failure from %s err: %v", req.URL(), err)
  }

  if err = json.Unmarshal(json_resp, &pm); err != nil {
    return "", err
  }
  
  switch op {
    case ws.FileMessageURLGet: 
      return pm.GetURL, nil
    case ws.FileMessageURLPut:
      return pm.PutURL, nil
  }

  return "", fmt.Errorf("something unexpected occurred")
}

// Method sends data to cloud storage server through signed url
//
func (wc *WorkerClient) SendBytesToURL(surl string, data []byte) (sent int64, fnerr error){
  var notsent int64  
  request_body := ioutil.NopCloser(bytes.NewReader(data))

  req, err := http.NewRequest("PUT", surl, request_body)
  if err != nil {
    return notsent, fmt.Errorf("failed creating a http request for signed url %s, err: %v", surl, err) 
  }
  
  res, err := http.DefaultClient.Do(req)
  if err != nil {
    return notsent, fmt.Errorf("failed while calling signed url %s, err: %v", surl, err)
  }

  defer res.Body.Close()

  content, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return notsent, fmt.Errorf("failed to read http response: %v", err)
  }

  if res.StatusCode != http.StatusOK {
    return notsent, fmt.Errorf("code= %d body=%s", res.StatusCode, string(content))
  }

  // storage api may not always send bytes, so return len(data)  
  return int64(len(data)), nil
}

func (wc *WorkerClient) ReceiveBytesFromURL(surl string) (data []byte, rcvd int64, fnerr error){
  var err error 
  req, err := http.NewRequest("GET", surl, nil)
  if err != nil {
    fnerr = fmt.Errorf("failed creating a http request for signed url %s, err: %v", surl, err)
    return  
  }
  
  res, err := http.DefaultClient.Do(req)
  if err != nil {
    fnerr = fmt.Errorf("failed while calling signed url %s, err: %v", surl, err)
    return  
  }

  defer res.Body.Close()

  data, err = ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, rcvd, fmt.Errorf("failed to read http response: %v", err)
  }

  if res.StatusCode != http.StatusOK {
    return nil, rcvd, fmt.Errorf("code= %d body=%s", res.StatusCode, string(data))
  }

  // storage api may not always send bytes, so return len(data)  
  return data, int64(len(data)), nil
}

// Method requests server to check object size and update in file metadata 
func (wc *WorkerClient) FileCheckIn(repoId, branchName, commitId, filepath string, size int64) (*ws.FileAttrs, error){

  c := newRestClient(wc.BaseURL, CommitFileCheckIn)
  req := c.Post()

  fa := ws.FileAttrsMessage{ 
    MessageType: ws.FileMessageCheckIn,
    Repo: &ws.Repo{
        Name: repoId,
      },
    Branch: &ws.Branch{
      Name: branchName,
    },
    Commit: &ws.Commit{
        Id: commitId,
      },
    FileAttrs: &ws.FileAttrs{
      File: &ws.File{
        Commit: &ws.Commit{
          Id: commitId,
        },
        Path: filepath,
      },
      SizeBytes: size,
    },
  }

  rq_json, _ := json.Marshal(&fa) 
  _ = req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(rq_json)))

  res := req.Do()
  res_body, err := res.Raw()  
  if err != nil {
    return nil, fmt.Errorf("failed to check in file %s: %v", filepath, err)
  }

  if err = json.Unmarshal(res_body, &fa); err != nil {
    return nil, fmt.Errorf("unformed json response while checking in file %s: %v", string(res_body), err)
  }
  
  return fa.FileAttrs, nil
}

// Method sends part for merging and checking in final object  
func (wc *WorkerClient) MergePartsNCheckIn(parts []int, repoId, branchName, commitId, filepath string, size int64) (*ws.FileAttrs, error){

  c := newRestClient(wc.BaseURL, CommitFilePartMerge)
  req := c.Post()

  p_msg := ws.FilePartsMessage{ 
    Sequences: parts,
    Repo: &ws.Repo{
        Name: repoId,
      },
    Branch: &ws.Branch{
      Name: branchName,
    },
    Commit: &ws.Commit{
        Id: commitId,
      },
    File: &ws.File{
        Commit: &ws.Commit{
          Id: commitId,
        },
        Path: filepath,
      },
    FileAttrs: &ws.FileAttrs{
      File: &ws.File{
        Commit: &ws.Commit{
          Id: commitId,
        },
        Path: filepath,
      },
      SizeBytes: size,
    },
  }

  rq_json, _ := json.Marshal(&p_msg) 
  _ = req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(rq_json)))

  res := req.Do()
  res_body, err := res.Raw()  
  if err != nil {
    return nil, fmt.Errorf("failed to check in file %s: %v", filepath, err)
  }
  if err = json.Unmarshal(res_body, &p_msg); err != nil {
    return nil, fmt.Errorf("unformed json response while checking in file %s: %v", string(res_body), err)
  }
  
  return p_msg.FileAttrs, nil
}








