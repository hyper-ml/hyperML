import React from 'react'
import { connect } from "react-redux"
import { Scrollbars } from "react-custom-scrollbars"

import * as notebookActions from './actions'
import Notebook from './Notebook'
import { getVisibleNotebooks } from './filters'
import './style.scss';
import Table from 'react-bootstrap/Table'

export class NotebookList extends React.Component {
    constructor(props) {
        super(props)
        const {fetchNotebooks, updateNotebookList, socket} = props;
        // if (api.LoggedIn()){
            fetchNotebooks()
            if (socket) {
                socket.onmessage = evt => {
                    const message = JSON.parse(evt.data)
                    updateNotebookList(message);
                    
                }
            }
        // } else {
        //    history.replace("/login")
        //}
    }

    componentWillReceiveProps(nextProps) {
        const {updateNotebookList} = nextProps;

        if (!this.props.socket && nextProps.socket) {
            nextProps.socket.onmessage = evt => {
                const message = JSON.parse(evt.data)
                updateNotebookList(message);
            }
        }
    }

    renderNotebookHeader(){
        return (<thead> <tr>
            <th style={{backgroundColor: '#fef', border: '0'}}> ID </th>
            <th style={{backgroundColor: '#fef', border: '0'}}> Request Type </th>    
            <th style={{backgroundColor: '#fef', border: '0'}}> Image </th>
            <th style={{backgroundColor: '#fef', border: '0'}}> Phase </th>
            <th style={{backgroundColor: '#fef', border: '0'}}> Status </th>

            </tr></thead>)
    }
     
    render() {
        
        const { visibleNotebooks } = this.props

        if (!visibleNotebooks ) {
            if (visibleNotebooks.length === 0) {
                return <div className="nb-list"> <p> You currently have no notebooks. </p> </div>;
            } 
            if (visibleNotebooks.error) {
                return <div className="nb-list"> <p>An error has occurred. Reason: {visibleNotebooks.reason}</p></div>;
            }
        }
        
        return (
            <div className="nb-list">
                <Scrollbars
                    renderTrackVertical={props => <div className="scrollbar-vertical" />}
                >
                    
                    <Table hover>
                    <tbody>
                        {visibleNotebooks.map(notebook => (
                            <Notebook key={notebook.ID} notebook={notebook} />
                        ))}
                    </tbody>
                    </Table>
                            
                </Scrollbars>
            </div>

        );
    }
}

const mapStateToProps = state => {
    return {
      visibleNotebooks: getVisibleNotebooks(state.notebooks.list, ""),
      filter: state.notebooks.filter,
    }
}
  
const mapDispatchToProps = dispatch => {
    return {
      fetchNotebooks: () => dispatch(notebookActions.fetchNotebooks()),
      setNotebookList: notebooks => dispatch(notebookActions.setList(notebooks)),
      updateNotebookList: notebook => dispatch(notebookActions.updateList(notebook))
    }
}
  
export default connect(
    mapStateToProps,
    mapDispatchToProps
)(NotebookList)
  