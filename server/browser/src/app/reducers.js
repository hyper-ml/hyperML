import { combineReducers } from "redux"
import notebooks from './browser/Notebooks/reducer'
import toast from './toast/reducer'

const rootReducer = combineReducers({ 
  notebooks,
  toast,
  })
  
export default rootReducer