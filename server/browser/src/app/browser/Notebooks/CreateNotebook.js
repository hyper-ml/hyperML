import React from "react"
import Button from 'react-bootstrap/Button'
import './style.scss'

export class CreateNotebook extends React.Component {
    render() {
        console.log('props:', this.props)
        return (
            <div className='create-nb-toolbar'> 
                <Button variant="primary" onClick={this.props.onCreate}> New Notebook</Button>
            </div>
        );
    }
} 

export default CreateNotebook