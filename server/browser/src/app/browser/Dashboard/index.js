import React from "react"
import { connect } from "react-redux"
import './style.scss'
import { Tabs, Tab, Button } from 'react-bootstrap'
import NotebookList from '../Notebooks/NotebookList'
import CreateNotebook from '../Notebooks/CreateNotebook'
import SavedNotebookList from '../Notebooks/SavedNotebookList'
import NewNotebook from '../Notebooks/NewNotebook';
import { connectSocket } from './actions';
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

export class Dashboard extends React.Component {

    constructor(props) {
        super(props)
        this.state = {
            activeTab: "notebooks",
            socket: '',
        }

        this.addNewNotebook = this.addNewNotebook.bind(this);
        this.hideNewNotebook = this.hideNewNotebook.bind(this);
        this.onCreateNotebook = this.onCreateNotebook.bind(this);
    }

    componentDidMount(){
        const {socket} = this.state;
        
        if (!socket) {
            let ws = connectSocket(socket, null)
            this.setState({socket: ws});
        }   
    }

    componentWillUnmount() {
        if (this.socket) {
            console.log('componentWillUnmount:closing socket')
            this.socket.close();
        }
    }

    setActiveTab(k) {
        this.setState({activeTab: k});
    } 

    addNewNotebook() {
        this.setState({activeTab: "newnotebook"});
    }

    hideNewNotebook() {
        this.setState({activeTab: "notebooks"});
    }

    onCreateNotebook(nb) {
        // add new nb to list 
        this.setState({activeTab: "notebooks"});
    }
    
    renderNewNotebook() {
        const {activeTab} = this.state; 

        if (activeTab !== "newnotebook") {
            return null
        }

        return (
            <Tab eventKey="newnotebook" title="New Notebook">
                <NewNotebook onDone={this.onCreateNotebook} /> 
            </Tab>
        );
    }

    renderSavedNotebooks() {
        return (
            <Tab style={{paddingTop: '2em'}} eventKey="savednotebooks" title="My Workbench">
                <Button variant="primary" style={{marginRight: '0.5em', marginBottom: '1em', fontSize: '0.9em'}}>  Import Notebook</Button>
                <div style={{clear: 'both'}}> </div>
                <SavedNotebookList /> 
            </Tab>
        );
    }

    renderNotebooks() {
        return (
            <Tab style={{paddingTop: '2em'}} eventKey="notebooks" title="Notebooks">
                <CreateNotebook onCreate={this.addNewNotebook} />
                <NotebookList socket={this.state.socket} /> 
            </Tab>
        );
    }

    renderJobs() {
        const {activeTab} = this.state;
        if (activeTab === "newnotebook") {
            return null
        }

        return (
            <Tab style={{padding: '1em'}} eventKey="jobs" title="Jobs">
                <p> You have no jobs at the moment.</p>
            </Tab>
        );
    }

    renderStorage() {
        return (
            <Tab style={{padding: '1em'}} eventKey="storage" title="Storage Disks">
                <p> This section displays your storage disks</p>        
            </Tab>
        );
    }

    renderDStore() {
        return (
            <Tab style={{padding: '1em'}} eventKey="dstore" title="Deployments">
                <p> This section displays your model API services</p>        
            </Tab>
        );
    }

    renderBrowse() {
        return (
            <Tab style={{padding: '1em'}} eventKey="browse" title="Browse">
                <p> This section lists out ready-to-use container images, Notebooks and Git Repos</p>        
            </Tab>
        );
    }

    renderSettings() {
        return (
            <Tab style={{padding: '1em'}} eventKey="settings" title="Settings">
                <p> No settings yet. </p>
            </Tab>
        )
    }

    
    render() {
        const {activeTab} = this.state;
        return <div className="dashboard">
            <Tabs  
                defaultActiveKey="notebooks" 
                activeKey={activeTab} 
                id="dashboard"
                onSelect={k => this.setActiveTab(k)}>
                {this.renderNotebooks()}
                {this.renderNewNotebook()}
                {this.renderJobs()}
                {this.renderSettings()}

            </Tabs>
        </div>
    }
}

export default connect(state => state)(Dashboard)