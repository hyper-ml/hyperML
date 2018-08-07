package fuse


func errorAsString(err error) string{
  if err == nil {
    return ""
  }

  return err.Error()

}