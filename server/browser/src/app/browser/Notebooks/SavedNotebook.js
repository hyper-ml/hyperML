import React from "react"
import { connect } from "react-redux"
import { Button } from 'react-bootstrap'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

export class SavedNotebook extends React.Component {
    constructor(props) {
        super(props)
    }

    renderLaunchNotebook() {
        return <Button variant="outline-primary"> Launch Notebook</Button>
    }


    render() {
        const {savednotebook} = this.props;
        console.log('savednotebook:', savednotebook)
        return (
            <tr className="nb-list-row">
                
                <td> <FontAwesomeIcon style={{color: '#678', marginRight: '1em'}} icon="file-code"/>  {savednotebook.Name} </td>
                <td> {savednotebook.Type}</td>
                <td> <span style={{color: '#999'}}> (last updated {savednotebook.Updated}) </span> </td> 
                <td> {this.renderLaunchNotebook()}</td> 
            </tr>
        )
    }
}


export default connect()(SavedNotebook);