import React, { useState, useRef, useEffect } from 'react'
import { Button, Icon } from 'semantic-ui-react'
import { Link, Redirect } from 'react-router-dom'
import { Form, Divider, Label } from 'semantic-ui-react'
import InputDataTable from './InputDataTable'

const NewApplicationEdit = props => {
  // uniq ids for data 
  var uniqid = require('uniqid');
  const axios = require('axios')
  const array = require('lodash')

  // used for cleanup: prevet update state ufter component is unmounted
  const mountedRef = useRef(true)
  // determine if creating new or editing existing application
  const exists = ('uid' in props.history.location.state.application.metadata)
  // parse matchLabels
  const parseMatchLabels = lables => {
    var result = {};
    if (lables.length > 0) {
      array.chunk(array.split(array.split(lables, ':'), ','), 2).forEach(a => {
        result[a[0].trim()] = a[1].trim()
      })
    }
    return result
  }

  const parseSelector = selectorMap => {
    var result = ''
    for (const [key, value] of Object.entries(selectorMap)) {
      result += key + ':' + value + ','
    }
    return result.slice(0,result.length-1)
  }
 
  // application instance
  const [application, setApplication] = useState({
    name: props.history.location.state.application.metadata.name,
    labels: exists ? props.history.location.state.application.spec.selector.workloadSelector : parseMatchLabels(props.history.location.state.application.spec.selector.workloadSelector.matchLabels),
    resourceVersion:  exists ? props.history.location.state.application.metadata.resourceVersion : '',
    role: exists ? props.history.location.state.application.spec.appInfo.role : '',
    purpose: exists ? props.history.location.state.application.spec.appInfo.purpose : '',
    geography: exists ? props.history.location.state.application.geography : props.location.state.application.geography,  
  })
  //application instance data
  const [applicationData, setApplicationData] = useState(
    (exists
      ? props.history.location.state.application.spec.data
      : [{ uid: uniqid(), dataSetID: '', requirements: { interface: { protocol: 's3', dataformat: 'parquet' } }}]))
  // error state
  const [errors, setErrors] = useState({ purpose: !exists, role: !exists, data: !exists })
  // message from request
  const [axiosMessage, setAxiosMessage] = useState({ message: '', error: true })

  // add new application instance data with default initial values
  const addDataRow = () => {
 //   setApplicationData([...applicationData, { uid: uniqid(), dataSetID: '', ifDetails: { protocol: 's3', dataformat: 'parquet' } }])
    setApplicationData([...applicationData, { uid: uniqid(), dataSetID: '', requirements: { interface: { protocol: 's3', dataformat: 'parquet' } }}])
    setErrors({ ...errors, data: true })
  }

  // delete application instance data
  const deleteDataRow = (uid) => {
    setApplicationData(applicationData.filter((data) => data.uid !== uid))
  }

  const handleIDChange = (event, uid) => {
    const { name, value } = event.target
    setApplicationData(applicationData.map(d => d.uid === uid ? { ...d, [name]: value } : d))
  }

  const handleDetailsChange = (event, uid, name, value) => {
     setApplicationData(applicationData.map(d => d.uid === uid ? { ...d, requirements: {interface: { ...d.requirements.interface, [name]: value } }} : d))
  }

  // handle change to input fields
  const handleChange = event => {
    const { name, value } = event.target
    setApplication({ ...application, [name]: value })
    setErrors({ ...errors, [name]: value.length === 0 })
    if (axiosMessage.message.length > 0) {
      setAxiosMessage({ ...axiosMessage, message: '', error: true })
    }
  }

  // submit request to new application instance or update existing one
  const handleSubmit = () => {
    // clean data
    const dataToSend = applicationData.map(data => array.omit(data, ['uid']))
    axios({
      method: exists ? 'put' : 'post',
      url: process.env.REACT_APP_BACKEND_ADDRESS + `/v1/dma/m4dapplication${exists ? `/${application.name}` : ''}`,
      data: {
        apiVersion: 'app.m4d.ibm.com/v1alpha1',
        kind: 'M4DApplication',
        metadata: {  
          name: application.name,
          resourceVersion: application.resourceVersion,
          labels: application.labels,
        },
        spec: {
          appInfo: {
            purpose: application.purpose,
            role: application.role,
          },
          selector: { 
            clusterName: application.geography,
            workloadSelector: {
              matchLabels: application.labels
            },
          },
          data: dataToSend
        }
      }
    })
      .then(response => {
        if (mountedRef.current) {
          setAxiosMessage({ ...axiosMessage, message: response.statusText, error: false })
        }
        console.log(response);
      })
      .catch(error => {
        if (mountedRef.current) {
          setAxiosMessage({ ...axiosMessage, message: error.message, error: true })
        }
        console.log(error);
      })
  }

  useEffect(() => {
    if (mountedRef.current) {
      var error = false
      applicationData.forEach(d => {
        error = error || (d.dataSetID.length === 0)
      })
      setErrors(errors => ({ ...errors, data: error }))
      setAxiosMessage(axiosMessage => ({ ...axiosMessage, message: '', error: true }))
    }
  }, [applicationData])

  useEffect(() => {
    // cleanup: prevet update state ufter component is unmounted
    return () => { mountedRef.current = false }
  }, [])

  // Show error for submit/update request
  const SubmitResultLabel = () => {
    return axiosMessage.message.length > 0 ?
      (<Label
        content={axiosMessage.message}
        pointing='left'
        basic
        icon={axiosMessage.error ? 'exclamation' : 'check'}
        color={axiosMessage.error ? 'red' : 'green'}
      >
      </Label>) : null
  }

  return ((axiosMessage.error) ?
    (<Form noValidate>
      <Form.Input
        label='Application instance name'
        required formNoValidate
        disabled={true}
        defaultValue={application.name}
      />
     <Form.Input
        label='Workload selector'
        required formNoValidate
        disabled={true}
        defaultValue={parseSelector(application.labels)}
      />
      <Form.Input
        label='Purpose'
        required formNoValidate
        autoComplete='off'
        fluid
        placeholder='purpose'
        onChange={handleChange}
        name='purpose'
        value={application.purpose}
      />
      <Form.Input
        label='Role'
        required formNoValidate
        autoComplete='off'
        fluid
        placeholder='role'
        onChange={handleChange}
        name='role'
        value={application.role}
      />

      <Divider hidden />
      <InputDataTable applicationData={applicationData} deleteDataRow={deleteDataRow} addDataRow={addDataRow}
        handleIDChange={handleIDChange} handleDetailsChange={handleDetailsChange} />
      <Divider hidden />

      <Button icon labelPosition='left' as={Link} to="/" color='red'>
        <Icon name='left arrow' />Back
      </Button>
      <Button disabled={errors.purpose || errors.role || errors.data} icon labelPosition='right' primary onClick={handleSubmit}>
        <Icon name='down arrow' />Submit
      </Button>
      <SubmitResultLabel />
    </Form>)
    : (<Redirect to='/' />)
  )
}

export default NewApplicationEdit
