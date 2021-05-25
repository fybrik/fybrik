import React, { useState, useEffect, useRef } from 'react'
import { Button, Form, Header, Divider, Icon, Label } from 'semantic-ui-react'
import { Link, Redirect } from 'react-router-dom'

const StoreCredentials = props => {
  const array = require('lodash')
  var uniqid = require('uniqid');
  const axios = require('axios')

  // used for cleanup: prevet update state ufter component is unmounted
  const mountedRef = useRef(true)
  // application instance name {metadata: {name : appname}}
  const application = props.history.location.state.application
  // determine if creating new or editing existing application
  const exists = ('uid' in application.metadata)
  // systems
  const systems = array.compact(array.union([process.env.REACT_APP_POLICY_MANAGER_SERVICE_SYSTEM, process.env.REACT_APP_CATALOG_CONNECTOR_SYSTEM, process.env.REACT_APP_CREDENTIALS_MANAGER_SYSTEM]))
  
  // credentials input
  const [creds, setCreds] = useState(
    systems.map(sys => ({
      system: sys, userName: '', userID: '', password: '', uid: uniqid(),
      errors: { userName: true, userID: true, password: true },
      message: '', errorStore: false, disableNext: true,
      labels: [(sys === process.env.REACT_APP_POLICY_MANAGER_SERVICE_SYSTEM) ? 'Policy Manager' : null,
      (sys === process.env.REACT_APP_CATALOG_CONNECTOR_SYSTEM) ? 'Data Catalog' : null,
      (sys === process.env.REACT_APP_CREDENTIALS_MANAGER_SYSTEM) ? 'Credentials Manager' : null]
    }
    )))

  const [next, setNext] = useState(true)
  // if to navigate to nex page
  const [redirectToNext, setRedirectToNext] = useState(false)

  // handle change to input fields
  const handleChange = (event, uid) => {
    const { name, value } = event.target
    setCreds(creds.map(cred => cred.uid === uid ? {
      ...cred, [name]: value,
      message: '', errorStore: false, disableNext: true, errors: { ...cred.errors, [name]: value.length === 0 }
    } : cred))
  }

  // store credentials
  const handleStore = (event, uid) => {
    creds.forEach(cred => {
      if (cred.uid === uid) {
        let secretName = (application.spec.secretRef === undefined || application.spec.secretRef.length === 0) ? 
          application.metadata.name : application.spec.secretRef 
        const credentials = array.omitBy({ username: cred.userName, password: cred.password, ownerId: cred.userID }, array.isEmpty)
        axios({
          method: 'post',
          url: process.env.REACT_APP_BACKEND_ADDRESS + '/v1/creds/usercredentials',
          data: {
            SecretName: secretName,
            System: cred.system,
            Credentials: credentials 
          }
        })
          .then(response => {
            if (mountedRef.current) {
              setCreds(creds => creds.map(c => c.uid === cred.uid ? { ...cred, userName: '', userID: '', password: '', message: response.statusText, errorStore: false, disableNext: false } : c))
            }
            console.log(response)
          })
          .catch(error => {
            if (mountedRef.current) {
              setCreds(creds => creds.map(c => c.uid === cred.uid ? { ...cred, userName: '', userID: '', password: '', message: error.message, errorStore: true, disableNext: true } : c))
            }
            console.log(error);
          })
      }
    })
  }

  useEffect(() => {
    var disable = false
    creds.map(cred => disable = (disable || cred.disableNext))
    setNext(disable)
  }, [creds])

  // cleanup: prevet update state ufter component is unmounted
  useEffect(() => {
    return () => { mountedRef.current = false }
  }, [])

  // Show error/succes for store credentials
  const StoreResultLabel = (props) => {
    return props.cred.message.length > 0 ?
      (<Label
        content={props.cred.message}
        pointing='left'
        basic
        icon={props.cred.errorStore ? 'exclamation' : 'check'}
        color={props.cred.errorStore ? 'red' : 'green'}
      >
      </Label>) : null
  }

  const RedirectToNext = () => {
    return (redirectToNext)
      ? (<Redirect to={{ pathname: '/newapplicationedit', state: { application } }} />)
      : (<Button icon labelPosition='right' color='green' disabled={next} onClick={() => setRedirectToNext(true)}>
        <Icon name='right arrow' />Next
      </Button>)
  }

  return (
    <div>
      <Header as='h3' textAlign='center'>Store Credentials </Header>
      <Divider hidden />
      <Form noValidate>
        <Form.Input
          required
          label='Application instance name'
          disabled={true}
          color='black'
          value={application.metadata.name}
        />
        <Divider hidden />
        {creds.map(cred => (<div key={cred.uid}>
          <Header as='h4'>{cred.system}
            {cred.labels.map(label =>
              label ? <Label key={uniqid()} basic horizontal pointing='left' size='tiny'>{label}</Label> : null
            )}
          </Header>
          <Form.Group widths='equal' >
            <Form.Input
              label='User Name'
              required formNoValidate 
              autoComplete='off'
              fluid icon='user'
              iconPosition='left'
              placeholder='User Name'
              onChange={(e) => handleChange(e, cred.uid)}
              name='userName'
              value={cred.userName}
            />
            <Form.Input
              label='UserID'
              formNoValidate
              autoComplete='off'
              fluid icon='user'
              iconPosition='left'
              placeholder='UserID'
              onChange={(e) => handleChange(e, cred.uid)}
              name='userID'
              value={cred.userID}
            />
            <Form.Input
              label='Password'
              formNoValidate
              autoComplete='off'
              fluid
              icon='lock'
              iconPosition='left'
              placeholder='Password'
              type='password'
              onChange={(e) => handleChange(e, cred.uid)}
              name='password'
              value={cred.password}
            />
          </Form.Group>
          <Divider hidden />
          <Button color='blue' disabled={cred.errors.userName}
            onClick={(e) => handleStore(e, cred.uid)}>
            <Icon name='down arrow' />Store
            </Button>
          <StoreResultLabel cred={cred} />
          <Divider hidden />
        </div>))}
        <Divider hidden />
        <Button icon labelPosition='left' as={Link} to="/" color='red'>
          <Icon name='left arrow' />Back
        </Button>
        {!exists ? (<RedirectToNext />) : (null)}
      </Form>
    </div>
  )
}

export default StoreCredentials
