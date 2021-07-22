import React, { useState, useEffect, useRef } from 'react'
import { Label } from 'semantic-ui-react'
import ApplicationTable from './ApplicationTable'

const FybrikApplications = props => {
  const array = require('lodash')
  // application instances
  const [applications, setApplications] = useState([])
  // error state
  const [axiosMessage, setAxiosMessage] = useState({ message: '', error: false })
  // used for cleanup: prevet update state ufter component is unmounted
  const mountedRef = useRef(true)
  // update
  const [update, setUpdate] = useState(0)
  const [namespace] = useState(props.datauserenv.namespace)

  useEffect(() => {
    var uniqid = require('uniqid');
    const axios = require('axios')

    const getFybrikApplications = async () => {
      try {
        console.log(process.env.REACT_APP_BACKEND_ADDRESS + '/v1/dma/fybrikapplication')
        return await axios.get(process.env.REACT_APP_BACKEND_ADDRESS + '/v1/dma/fybrikapplication')
      } catch (error) {
        console.error(error)
        if (mountedRef.current) {
          setAxiosMessage(axiosMessage => ({ ...axiosMessage, message: error.message, error: true }))
        }
        return null
      }
    }

    const getApplications = async () => {
      const response = await getFybrikApplications()
      if (response && response.data && response.data.items) {
        console.log(response)
        if (mountedRef.current) {
          const items = response.data.items
          items['geography'] = props.datauserenv.geography
          items.forEach(app => {
            app['geography'] = props.datauserenv.geography
            // NOTE: set unique id for data 
            app.spec.data.forEach(data => {
            data['uid'] = uniqid()
          })})
          setApplications(items)
          setAxiosMessage(axiosMessage => ({ ...axiosMessage, message: response.statusText, error: false }))
        }
      }
    }

    getApplications()

  }, [update, props.datauserenv.geography])

  useEffect(() => {
    // cleanup
    return () => {
      mountedRef.current = false
      setUpdate(0)
    }
  }, [])

  const updateApplications = () => {
    setUpdate(update + 1)
  }

  const deleteCredentials = (name) => {
    const axios = require('axios')
    axios.delete(process.env.REACT_APP_BACKEND_ADDRESS + `/v1/creds/usercredentials/${name}`) 
    .then(async (response) => {
      console.log(response);
        setAxiosMessage({ ...axiosMessage, message: 'Application and credential were deleted', error: false })
      })
      .catch(error => {
        console.error(error)
        setAxiosMessage({ ...axiosMessage, message: 'Credentials were not deleted', error: true })
      })
  }

  // send request to delete application, remove row from table upon success
  const deleteApplication = (uid, name, secret) => {
    const axios = require('axios')
    axios.delete(process.env.REACT_APP_BACKEND_ADDRESS + `/v1/dma/fybrikapplication/${name}`)
      .then(async (response) => {
        console.log(response);
        // remove app from table, does not guarantee that kubernetes deletion
        setApplications(applications.filter((app) => app.metadata.uid !== uid))
        setAxiosMessage({ ...axiosMessage, message: `Application was deleted: ${response.statusText}`, error: false })
        if (secret !== undefined && secret.length > 0) {
          deleteCredentials(secret)
        }
      })
      .catch(error => {
        console.log(error);
        setAxiosMessage({ ...axiosMessage, message: `Application was not deleted: ${error.message}`, error: true })
      })
  }

  // show errors
  const ApplicationsResultLabel = () => {
    return (axiosMessage.message.length > 0 && axiosMessage.error) ?
      (<Label
        content={axiosMessage.message}
        pointing='above'
        basic
        icon='exclamation'
        color='red'
      >
      </Label>) : null
  }

  return (
    <div>
      <ApplicationTable applications={applications} deleteApplication={deleteApplication} updateApplications={updateApplications} />
      <ApplicationsResultLabel />
    </div>
  )
}

export default FybrikApplications
