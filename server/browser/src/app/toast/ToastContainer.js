import React from "react"
import { connect } from "react-redux"
import Toast from "./Toast"
import * as toastActions from "./actions"

export const ToastContainer =({toast, hideToast}) => {
    if (!toast.message) {
        return ""
    }
    return <Toast  {...toast} onDismiss={hideAlert} />
}

const mapStateToProps = state => {
    return {
      toast: state.toast
    }
  }
  
  const mapDispatchToProps = dispatch => {
    return {
      hideAlert: () => dispatch(toastActions.clear())
    }
  }
  
  export default connect(mapStateToProps, mapDispatchToProps)(ToastContainer)