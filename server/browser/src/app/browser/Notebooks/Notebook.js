import React from "react"
import { connect } from "react-redux"

import Button from 'react-bootstrap/Button'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Spinner } from "react-bootstrap";
import * as notebookActions from './actions'
import { DisplayError } from '../error'

export class Notebook extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            spinner: '',
            error: ''
        }
    }
    canOpen() {
        const {notebook} = this.props;

        if (!notebook) {
            return ''
        }
        if ((notebook.Phase.toLowerCase().indexOf('run')) === -1) {
            return ''
        }

        return true;
    }

    canTerminate() {
        const {notebook} = this.props;

        if (!notebook) {
            return ''
        }
        
        if (notebook.Phase.toLowerCase().indexOf("shut") === -1) {
            return true
        }

        return ''
    }

    handleTerminate(){
        const {notebook, updateNotebookList} = this.props;

        this.setState({spinner: true});
        notebookActions.stopNotebook(notebook.ID).then(data => {
            this.setState({spinner: ''});
            
            if (!data && data.ID) {
                updateNotebookList(data)
            } else {
                console.log('Stop Notebook Failure: ', data);
                this.setState({spinner: '', error: 'Failed to initiate delete request'})
            }
          });
    }

    renderStopNotebook() {
        const {spinner, error} = this.state;
        if (this.canTerminate()) {
            
            if (spinner) {
                return <div> {error? <DisplayError msg={error} /> : null} <Spinner /></div>
            }

            return (
                <div>
                    {error? <DisplayError msg={error} /> : null}
                    <Button disabled={this.spinner} onClick={this.handleTerminate} className="btn btn-outline2 btn-outline-danger">
                        <FontAwesomeIcon icon="power-off" /> 
                    </Button>
                </div>
            );
        }
        return null;
    }

    handleOpenNotebook() {
        const {notebook} = this.props;
        if (notebook.POD) {
            // "http://hyperml.com/" + notebook.POD.UserKey + "/lab?token=" + notebook.POD.AuthToken
            window.open( notebook.POD.endpoint, "_blank")    
        }
        
    }

    renderOpenNotebook() {
        if (this.canOpen()) { 
            return <Button onClick={this.handleOpenNotebook} variant="outline-primary"> Open Notebook</Button>
        }
        return null;
    }

    render() {
        const {notebook} = this.props;
        const {error} = this.state; 
        console.log('notebook:', notebook)
        return (
            <tr className="nb-list-row">
                
                <td> {notebook.ID} </td>
                <td> {notebook.POD.PodType} </td>
                <td> {notebook.ContainerImage? notebook.ContainerImage.Name: ''} </td>
                <td> {notebook.Phase}</td>
                <td> {notebook.Status}</td>
                <td> {this.renderOpenNotebook()}</td>
                <td> {this.renderStopNotebook()}</td>
            </tr>
        );
    }
} 

const mapDispatchToProps = dispatch => {
    return {
      updateNotebookList: notebook => dispatch(notebookActions.updateList(notebook))
    }
}
  

export default connect(
    (state)=> state,
    mapDispatchToProps
  )(Notebook);