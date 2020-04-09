import * as actions from "./actions"

const initialState = {
    show: false,
    type: "danger"
}

export default (state = initialState, action) => {
    switch (action.type) {
      case actions.SHOW:
        return {
          show: true,
          id: action.toast.id,
          type: action.toast.type,
          message: action.toast.message
        }
      case actions.HIDE:
        if (action.toast && action.toast.id !== state.id) {
          return state
        } else {
          return initialState
        }
      default:
        return state
    }
  }