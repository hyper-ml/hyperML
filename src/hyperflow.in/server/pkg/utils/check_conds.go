package utils 

import( 
  "reflect"
)

func isNil(val interface{}) (retval bool) {
  defer func() {
    if r := recover(); r != nil {
      retval = false
    }
  }()

  switch val.(type) {
      case string: 
        if val == "" { 
          return true
        }
      case int, int64, int32, uint64, uint: 
        if val == 0 {
          return true
        }  
        
      default:
        if val == "" {
          return true
        }

        if reflect.ValueOf(val).IsNil(){ 
          return true
        }

        if val == nil {  
          return true
        }
  }

  return false
}

func CheckNilParams(args map[string]interface{}) []string {
  var empty_params []string
  
  for key, val := range args { 
    if isNil(val) {  
      empty_params = append(empty_params, key)
    } 
  }

  return empty_params  
}

func IfNilAny(args []interface{}) bool{

  for _, val := range args {
    if isNil(val) { 
      return true
    } 
  }
  return false
}
