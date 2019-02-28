package workspace

import( 
  "strings"
  "path/filepath"
)


// This function generates directory tree from list
// 
// string key should be relative path from root  
// as there could be 2 folders with same name 
// TODO: add glob search 

func ListToD(paths []string) (map[string][]string, error){
  dtree := make(map[string][]string)
  var root string = "/"
  var sep string = "/"

  var path string
  for i := 0; i < len(paths); i++ {
    path = paths[i] 

    dirs := strings.Split(path, sep)
    if len(dirs) == 1 {
      dtree[root] = append(dtree[root], dirs[0])
    } else {
      last_dir := root 
      for p :=0; p < len(dirs); p++ {

        if dirs[p] != "" {
          current_dir := filepath.Join(last_dir, dirs[p])  
        
          if _, ok := dtree[current_dir]; !ok {
            // add parent  missing in tree 
            dtree[last_dir] = append(dtree[last_dir], current_dir)            
          }

          last_dir = current_dir
        }
      }
    }
  }

  return dtree, nil
}

//[folder_name] = child_list []
//each folder->
//loop through list of elements in path:
  //add to dtree[] the child and individual folder 
  //if file add to dtree

