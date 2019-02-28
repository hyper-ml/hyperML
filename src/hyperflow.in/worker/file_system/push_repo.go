package file_system

import(
  "io"
  "fmt"
  hf_utils "hyperflow.in/worker/utils"
  ws "hyperflow.in/server/pkg/workspace"

)


// This method sends an object from local system to a signed URL. The
// signed URL is requested from hflserver and then a direct access to 
// cloud storage is initiated for transfer
// 
func (fs *RepoFs) pushObjectByURL(relative_path string) (sent int64, fnerr error){

  var f_attrs *ws.FileAttrs
  file_path := fs.getAbsolutePath(relative_path)

  f, err := hf_utils.Open(file_path)
  if err != nil {
    return sent, err
  } 

  buf := hf_utils.GetBuffer()
  defer hf_utils.PutBuffer(buf)

  var seqs  []int
  write_seq := 0
  writes    := 0
  
  for {
    sent = sent   
    var b int64
    var err error 

    n, err := f.Read(buf)

    if n == 0 && err != nil {
      if err == io.EOF {
         break 
      }
      return sent, err 
    }

    writes++ 
    if write_seq == 0 {
      b, err = fs.sendFileData(file_path, buf[:n])      
    } else {
      b, err = fs.sendFilePart(write_seq, file_path, buf[:n])
    }

    if err != nil {
      return sent, fmt.Errorf("failed while sending file %s to server, err: ", file_path, err)
    } else {
      seqs = append(seqs, write_seq)
      sent = sent + b
    }
        
    write_seq++
  }

  // if total writes dont equal successful ones
  if writes != len(seqs) {
    return 0, fmt.Errorf("some writes may have failed %s", file_path)
  } 
  
  switch writes {
    case 1:
      //just a single write 
      f_attrs, err = fs.wc.FileCheckIn(fs.repo.Name, fs.branch.Name, fs.commit.Id, file_path, sent)      
    
    case 0: 
      return 0, nil

    default:
      // multiple writes 
      f_attrs, err = fs.wc.MergePartsNCheckIn(seqs[1:], fs.repo.Name, fs.branch.Name, fs.commit.Id, file_path, sent)
  }
  
  if f_attrs == nil || err != nil {
    return 0, fmt.Errorf("failed to check in file %s into repo, err: %v", file_path, err)
  }

  return f_attrs.SizeBytes, err
}

func (fs *RepoFs) sendFilePart(seq int, fpath string, data []byte) (sent int64, fnerr error){
  // get obj path 
  var surl string

  // derive signed url for a new part
  surl, fnerr = fs.wc.SignedPutPartURL(seq, fs.repo.Name, fs.branch.Name, fs.commit.Id, fpath)
  if fnerr != nil {
    return 
  }

  // send data 
  sent, fnerr = fs.wc.SendBytesToURL(surl, data)
  
  return
}

// Send file bytes to signed URL 
//  
func (fs *RepoFs) sendFileData(fpath string, data []byte) (sent int64, fnerr error){

  var surl string

  // request a signed URL  
  surl, fnerr = fs.wc.SignedPutURL(fs.repo.Name, fs.branch.Name, fs.commit.Id, fpath)  
  if fnerr != nil {
    return 
  }

  // send data 
  sent, fnerr = fs.wc.SendBytesToURL(surl, data)

  // next step: check in 
  return  
}
