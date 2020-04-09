import React from "react"
import Button from 'react-bootstrap/Button'
import Dropdown from 'react-bootstrap/Dropdown'
import {CustomToggle, CustomMenu} from './toggle'
import Form from 'react-bootstrap/Form'
import Card from 'react-bootstrap/Card'
import { createNotebook, fetchResourceProfiles } from './actions'
import { DisplayError } from '../error'

export class NewNotebook extends React.Component{
    constructor(props) {
      console.log('newnotebook')
      super(props)
      this.state = {
        selectRprofileId: '',
        selectImage: 'jupyter/minimal-notebook',
        spinner: '',
        error: '',
        profiles: [],
        images: [],
      }

      this.onSelectImage = this.onSelectImage.bind(this);
      this.onImageChange = this.onImageChange.bind(this);
      this.onCreateNotebook = this.onCreateNotebook.bind(this);
    }

    componentDidMount() {
      fetchResourceProfiles().then(data => {
        console.log('received profiles:', data)
        let profiles  = (data && data.ResourceProfiles)? data.ResourceProfiles: []
        let defaultProfileId = profiles.length> 0? profiles[0].ID: '';
        this.setState({profiles: profiles, selectRprofileId: defaultProfileId});
      });
    }

    onSelectImage(eventKey, event) {
      console.log('eventKey: ', eventKey)
      console.log('event:', event);
      this.setState({selectImage: eventKey});
    }

    onImageChange(e) {
      this.setState({selectImage: e.target.value});
    }

    onResourceProfileSelect(id) {
      this.setState({selectRprofileId: id});
    }

    onCreateNotebook() {
      const {selectRprofileId, selectImage} = this.state;

      var checkArgs = (profileId, imageName) => {
        console.log('image :', imageName)
        if ((profileId === '') || !profileId) {
          return 'Please choose a Resource Plan';
        }

        if (!imageName || (imageName === '')) {
          return 'Please enter or select a Container Image';
        } 
        return '';
      }

      var argError = checkArgs(selectRprofileId, selectImage);
      console.log('argERror:', argError)
      if  (argError) {
          this.setState({error: argError});
      } else {
      
        this.setState({spinner: true ,error: ''});
        let nb = {
          ContainerImage: {
            Name: selectImage,
          },  
          ResourceProfileID: selectRprofileId,
        }

        createNotebook(nb).then(nb => {
          console.log('notebook:', nb.length); 
          this.setState({spinner: ''});
          if (!nb && nb.ID) {
            this.props.onDone(nb);
          } else {
            this.setState({spinner: '', error: 'Failed to create notebook instance'})
          }
        });
      }
      
    }

    renderImageChoices() {
      const {selectImage} = this.state;
      let image1 = "mindhash/pytorch";
      let image2 = "mindhash/tf2";
      let image3 = "jupyter/minimal-notebook"
      let image4 = "mindhash/rapids-ai"
      return (
        <Form.Group >
            <Form.Label className="new-nb-form-label">Enter Container Image Name</Form.Label>
            <Form.Control as="input" onChange={this.onImageChange} value={selectImage} type="text" style={{width: '40%'}}/>
            <Dropdown style={{fonSize: '0.9em', marginTop: '0.5em'}}>
              <Dropdown.Toggle as={CustomToggle} id="dropdown-custom-components">
                or select from the list
              </Dropdown.Toggle>
              
              <Dropdown.Menu as={CustomMenu}>
                <Dropdown.Item onSelect={this.onSelectImage} eventKey={image1}>{image1}</Dropdown.Item>
                <Dropdown.Item onSelect={this.onSelectImage} eventKey={image2}>{image2}</Dropdown.Item>
                <Dropdown.Item onSelect={this.onSelectImage} eventKey={image3} active={selectImage===image3? true: false}>
                {image3}  
                </Dropdown.Item>
                <Dropdown.Item onSelect={this.onSelectImage} eventKey={image4}>{image4}</Dropdown.Item>
              </Dropdown.Menu>
          </Dropdown>
        </Form.Group>
        );
    }

    renderResourceProfile(prf) {
      const {selectRprofileId} = this.state;
      
      if (!prf || !prf.ID || !prf.Name) {
        return <p key={0} style={{color:'#f14'}}> Launching a notebook instance requires atleast one resource plan. Please contact your system administrator. </p>
      }

      return (
        <Card 
          key={prf.ID} 
          onClick={prf.ID? this.onResourceProfileSelect.bind(this, prf.ID): null} 
          border={(prf.ID === selectRprofileId) ? "primary" : ""} 
          style={{ width: '18rem' }}>
          <Card.Body>
            <Card.Title style={{fontSize: '1em'}}>{prf.Name}</Card.Title>
            <Card.Subtitle className="mb-2 text-muted"> {prf.Subtitle? prf.Subtitle: 'Resource Plan'} </Card.Subtitle>
            <Card.Text>
              {prf.ShortDesc}
            </Card.Text>

          </Card.Body>
        </Card>
      );
    }

    renderResourceProfiles() {
      const {profiles} = this.state;
      var items = []

      if (profiles.length === 0) {
        items.push(this.renderResourceProfile(null));
      }  else {
        items = profiles.map(profile => {
          return this.renderResourceProfile(profile);
        })
      }
      

      // var profile1 = {ID: 1, Name: 'V100.C8.M64', Subtitle: 'GPU enabled Plan', ShortDesc: 'This resource plan supports V100 GPU (32GB), 8 CPUs and 64GB DDRM Memory. '}
      // var profile2 = {ID: 2, Name: 'V100.C4.M16', Subtitle: 'GPU enabled Plan', ShortDesc: 'This resource plan supports V100 GPU (32GB), 4 CPUs and 16GB DDRM Memory. '}
      // var profile3 = {ID: 3, Name: 'C16.M128', Subtitle: 'CPU Resource Plan', ShortDesc: 'This resource plan supports 16 CPUs and 128GB DDRM Memory. '}

      return (
          <div className="new-nb-rprof-select">
            <Form.Group>
              <Form.Label className="new-nb-form-label">Choose Resource Plan</Form.Label>
              <div className="new-nb-rprofs">
                {items}
              </div>
            </Form.Group>
          </div>
      )
    }

   

    render(){
        const {error} = this.state; 

        return <div className="new-notebook dark">
            {error? <DisplayError msg={error} /> : null}
            {this.renderImageChoices()}
            {this.renderResourceProfiles()}
            
            <div className="new-notebook-actions" style={{display: 'flex', flexDirection: 'row', marginTop: '2em'}}>
                
                <Button 
                  variant="primary" 
                  onClick={this.onCreateNotebook}
                  disabled={this.state.spinner}> Create Notebook </Button>
                <Button variant="light" onClick={this.props.onDone}> Cancel </Button>
                
            </div>
        </div>
    }
}

export default NewNotebook;