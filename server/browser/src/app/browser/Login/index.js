import React from "react"
import { connect } from "react-redux"
import { Redirect } from "react-router-dom"
import * as toastActions from "../../toast/actions"
import Toast from '../../toast/Toast'
import { Button, Form } from 'react-bootstrap'

import API from '../../api';
import './style.scss';

export class Login extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            userName: "",
            password: "",
        }
    }

    userNameChange(e) {
        this.setState({userName: e.target.value})
    }

    passwordChange(e) {
        this.setState({password: e.target.value})
    }

    handleSubmit(evt) {
        evt.preventDefault();
        const {showToast, hideToast, history} = this.props;
        let err = ""
        if (this.state.userName === "") {
            err = "User name can not be empty"
        }
        if (this.state.password === "") {
            err = "Password can not be empty"
        }

        if (err) {
            showToast("danger", err)
            return
        }
        API.Login({
            username: this.state.userName,
            password: this.state.password,
        }).then(res => 
            {
                hideToast()
                history.push("/")    
            }
        ).catch(e => {
            console.log(e);
            showToast("danger", e.message)
        })
    }

    componentWillMount() {
        const {hideToast} = this.props;
        hideToast()
        
    }


    render() {
        const { hideToast, toast } = this.props

        if (API.LoggedIn()) {
            return <Redirect To={"/"} />
        }
        let toastDiv = <Toast {...toast} onDismiss={hideToast} />
        if (!toastDiv.message) toastDiv = ""

        return <div className="login">
            {toastDiv}

            <div className="login-form">
            <h4> hyperML</h4>
            <Form>
                <Form.Group controlId="loginUser">
                    <Form.Label>User Name</Form.Label>
                    <Form.Control 
                        value={this.state.userName} 
                        onChange={this.userNameChange.bind(this)} 
                        type="text" 
                        placeholder="Enter User name" />
                    </Form.Group>

                    <Form.Group controlId="loginPassword">
                        <Form.Label>Password</Form.Label>
                        <Form.Control 
                            value={this.state.password} 
                            onChange={this.passwordChange.bind(this)} 
                            type="password" 
                            placeholder="Password"  />
                    </Form.Group>

                <div className="login-submit">
                    <Button  variant="primary" 
                        type="submit" 
                        onSubmit={this.handleSubmit.bind(this)}>
                        Login
                    </Button>
                </div>
            </Form> 
               
            </div>
        </div>
    }
}

const mapDispatchToProps = dispatch => {
    return {
      showToast: (type, message) =>
        dispatch(toastActions.show({ type: type, message: message })),
      hideToast: () => dispatch(toastActions.hide())
    }
  }
  
export default connect(
    state => state,
    mapDispatchToProps
)(Login)
