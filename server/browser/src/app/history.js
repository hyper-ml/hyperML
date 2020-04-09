import { createBrowserHistory } from "history";
import { browserPrefix } from "./constants"

const history = createBrowserHistory({
    basename: browserPrefix
  })
  
  export default history
