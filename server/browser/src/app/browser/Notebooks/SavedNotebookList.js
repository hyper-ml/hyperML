import React from 'react'
import { connect } from "react-redux"
import { Scrollbars } from "react-custom-scrollbars"

import SavedNotebook from './SavedNotebook'
import './style.scss';
import Table from 'react-bootstrap/Table'

export class SavedNotebookList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            savedNotebooks: [
                {ID: 1, Type: 'Notebook', Name: 'BIRT Sample demo.ipynb', Updated: '3 min ago', Location: 'bucket/directory/BIRT Sample demo.ipynb'},
                {ID: 2, Type: 'Notebook', Name: 'Super-resolution demo.ipynb', Updated: '2 days ago', Location: 'bucket/directory/Super-resolution demo.ipynb'},
                {ID: 3, Type: 'Model', Name: 'saved model.hd5', Updated: '1 week ago', Location: 'bucket/directory/saved model.hd5'},
                {ID: 4, Type: 'Data', Name: 'profile-pics', Updated: '1 month ago', Location: 'bucket/directory/profile-prics'}
            ]
        }
    }

    renderNotebookHeader(){
        return (<thead> <tr>
            <th style={{backgroundColor: '#fef', border: '0'}}> Name </th>    
            <th style={{backgroundColor: '#fef', border: '0'}}> Type </th>
            <th style={{backgroundColor: '#fef', border: '0'}}> Updated </th>
            

            </tr></thead>)
    }

    render() {
        const { savedNotebooks } = this.state
        console.log('savedNotebooks:', savedNotebooks);

        return (
            <div className="nb-list">
                <Scrollbars
                    renderTrackVertical={props => <div className="scrollbar-vertical" />}
                >
                    
                    <Table hover>
                    <tbody>
                        {savedNotebooks.map(notebook => (
                            <SavedNotebook key={notebook.ID} savednotebook={notebook} />
                        ))}
                    </tbody>
                    </Table>
                            
                </Scrollbars>
            </div>
        )
    }
}



  
export default connect()(SavedNotebookList)
