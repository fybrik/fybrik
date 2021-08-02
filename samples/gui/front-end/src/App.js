import React, { useState, useEffect, useRef } from 'react'
import { Route, Switch } from 'react-router-dom'
import { Divider, Container } from 'semantic-ui-react'
import FybrikMenuBar from './formats/Menu'
import FybrikApplications from './formats/FybrikApplications'
import NewApplication from './formats/NewApplication'
import StoreCredentials from './formats/StoreCredentials'
import NewApplicationEdit from './formats/NewApplicationEdit'

export default function App() {
  // used for cleanup: prevet update state ufter component is unmounted
  const mountedRef = useRef(true)
  // data user env
  const [datauserenv, setDataUserEnv] = useState('NA')

  useEffect(() => {
    const axios = require('axios')
    axios.get(process.env.REACT_APP_BACKEND_ADDRESS + '/v1/env/datauserenv')
      .then(response => {
        if (mountedRef.current) {
          setDataUserEnv(response.data)
        }
        console.log(response)
      }).catch(error => {
        console.log(error);
      })
  }, [])

  useEffect(() => {
    // cleanup
    return () => {
      mountedRef.current = false
    }
  }, [])

  return (
    <div>
      <Container>
        <FybrikMenuBar datauserenv={datauserenv}/>
        <Divider hidden />
          <Switch>
            <Route path="/" exact component={props => <FybrikApplications datauserenv={datauserenv}/>} />
            <Route path="/newapplication" exact component={NewApplication} />
            <Route path="/credentials" exact component={StoreCredentials} />
            <Route path="/newapplicationedit" exact component={NewApplicationEdit} />
          </Switch>
      </Container>
    </div>
  )
}
