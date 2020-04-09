import React from 'react'

import Toast from 'react-bootstrap/Toast'


export class DisplayError extends React.Component{
   constructor(props) {
     super(props)
     this.state = {
       show: true,
     }
   }  


   setShow() {
     this.setState({show: false});
   }
   render() {
    const { show } = this.state;
    const { msg } = this.props;
    
    return (
        
      <Toast style={{
        position: 'absolute',
        top: '0',
        left: '0',
        }}
        onClose={() => this.setShow(false)} show={show} 
        delay={5000}
        autohide>
            <Toast.Header style={{ background: '#f14', color: '#fff'}}>
              <img
                src="holder.js/20x20?text=%20"
                className="rounded mr-2"
                alt=""
              />
              <strong className="mr-auto">API Error</strong>
              <small>Just now</small>
            </Toast.Header>
            <Toast.Body>{msg}</Toast.Body>
          </Toast> 
    );
   }
    
}