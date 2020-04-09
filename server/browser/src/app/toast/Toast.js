import React from "react"
import Alert from 'react-bootstrap/Alert'

const Toast = ({ show, type, message, onDismiss }) => (
    <Alert
      className={"alert animated " + (show ? "fadeInDown" : "fadeOutUp")}
      bsStyle={type}
      onDismiss={onDismiss}
    >
      <div className="text-center">{message}</div>
    </Alert>
  )
  
  export default Toast;
  