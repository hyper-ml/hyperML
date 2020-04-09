export const SHOW = "toast/SHOW"
export const HIDE = "toast/HIDE"

export let toastId = 0

export const show = toast => {
    const id = toastId++
    return (dispatch, getState) => {
      if (toast.type !== "danger" || toastId.autoClear) {
        setTimeout(() => {
          dispatch({
            type: HIDE,
            toast: {
              id
            }
          })
        }, 5000)
      }
      dispatch({
        type: SHOW,
        toast: Object.assign({}, toast, {
          id
        })
      })
    }
  }
  
  export const hide = () => {
    return { type: HIDE }
  }