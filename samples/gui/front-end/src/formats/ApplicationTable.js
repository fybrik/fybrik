import React, { useState } from 'react'
import { Button, Table, Modal, Popup, Grid, Segment } from 'semantic-ui-react'
import { Link } from 'react-router-dom'


const ApplicationTable = (props) => {
  // open state of remove question
  const [openQ, setOpen] = useState({ open: false })
  const onClose = () => setOpen({ open: false })
  const onOpen = () => setOpen({ open: true })

  // identifying info for deleting application instance
  const [deleteApplication, setDeleteApplication] = useState({ uid: '', name: '', secret: '' })
  // save identifying info for deleting application instance
  const onDeleteClicked = (uid, name, secret) => {
    setDeleteApplication({ ...deleteApplication, uid: uid, name: name, secret: secret })
  }
  // delete application instance
  const onDelete = () => {
    props.deleteApplication(deleteApplication.uid, deleteApplication.name, deleteApplication.secret)
    onClose()
  }

  // Show status success/in progress/error, and display the data access instructions
  const TableCellStatus = (status) => {
    if (('ready' in status.status) && status.status.ready) {
      let states = []
      for (const key in status.status.assetStates) {
        let assetState = status.status.assetStates[key]
        if (assetState.conditions[1].status === "True") {
          states.push({key: key, value: assetState.conditions[1].message})
        } else {
          let msg = 'Asset is ready. '
          if (('catalogedAsset' in assetState) && (assetState.catalogedAsset !== "")) {
            msg += 'Registered asset: ' + assetState.catalogedAsset
          }
          if ('endpoint' in assetState) {
            msg += 'Endpoint: ' + JSON.stringify(assetState.endpoint)
          }
          states.push({key: key, value: msg})
        }
      }
      return (
        <Table.Cell textAlign='center'>
          <Popup position='left center' pinned on='click' trigger={<Button basic icon='check' flowing='true' color='green'/>}>
            <Grid>
              {(states.map(elem => (
                <Grid.Row key={elem.key}>
                  <Segment attached><b>{elem.key}: </b>{elem.value}</Segment>
                </Grid.Row>
              )))}
            </Grid>
          </Popup>
        </Table.Cell>
      )
    } else {
      let errorMsgs = []
      if ('assetStates' in status.status) {
        if (('errorMessage' in status.status) && (status.status.errorMessage !== '')) {
          errorMsgs.push(status.status.errorMessage)
        }
        for (const key in status.status.assetStates) {
          let assetState = status.status.assetStates[key]
          if (assetState.conditions[2].status === "True") {
            errorMsgs.push(assetState.conditions[2].message)
          }
        }
      }
      if (errorMsgs.length > 0) {
        return (
          <Table.Cell textAlign='center'>
            <Popup position='left center' pinned on='click' trigger={<Button basic icon='exclamation' flowing='true' color='red'/>}>
              <Grid>
                {(errorMsgs.map(elem => (
                  <Grid.Row key={elem}>
                    <Segment attached><b>Error: </b>{elem}</Segment>
                  </Grid.Row>
                )))}
              </Grid>
            </Popup>
          </Table.Cell>
        )
      } else {
        return (
          <Table.Cell textAlign='center'>
            <Button basic icon='hourglass half' data-tooltip='in progress' color='grey'/>
          </Table.Cell>
        )
      }
    }
  }

  // remove/edit/add credentials buttons
  const TableCellActions = (data) => {
    return (<Table.Cell textAlign='center'>
      <Modal trigger={<Button basic icon='remove circle' data-tooltip='delete' onClick={() => onDeleteClicked(data.application.metadata.uid, data.application.metadata.name, data.application.spec.secretRef)} />}
        size={'tiny'}
        open={openQ.open}
        onOpen={onOpen}
        onClose={onClose}>
        <Modal.Header>Delete Application</Modal.Header>
        <Modal.Content>
          <p>Are you sure you want to delete this application</p>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={onClose} negative>No</Button>
          <Button positive icon='checkmark' labelPosition='right' content='Yes' onClick={onDelete}
          />
        </Modal.Actions>
      </Modal>
      <Link to={{ pathname: '/newapplicationedit', state: { application: data.application } }}>
        <Button basic icon='edit' data-tooltip='edit' />
      </Link>
      <Link to={{ pathname: '/credentials', state: { application: data.application } }}>
        <Button basic icon='handshake outline' data-tooltip='add credentials' />
      </Link>
    </Table.Cell>)
  }

  return (
    <Table celled color={'blue'}>
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell>Application environment</Table.HeaderCell >
          <Table.HeaderCell>Role</Table.HeaderCell>
          <Table.HeaderCell>Intent</Table.HeaderCell>
          <Table.HeaderCell>Status</Table.HeaderCell>
          <Table.HeaderCell></Table.HeaderCell>
        </Table.Row>
      </Table.Header>

      <Table.Body>
        {props.applications.length > 0 ? (
          props.applications.map(application => (
            <Table.Row key={application.metadata.uid}>
              <Table.Cell>{application.metadata.name} </Table.Cell>
              <Table.Cell>{application.spec.appInfo["role"]}</Table.Cell>
              <Table.Cell>{application.spec.appInfo["intent"]}</Table.Cell>
              <TableCellStatus status={application.status} />
              <TableCellActions application={application} />
            </Table.Row>
          ))
        ) : (
            <Table.Row>
              <td colSpan={5}>No application environments</td>
            </Table.Row>
          )}
      </Table.Body>

      <Table.Footer fullWidth>
        <Table.Row>
          <Table.HeaderCell colSpan='5'>
            <Link to={{ pathname: '/newapplication', state: { applications: props.applications } }}>
              <Button floated='right' primary size='small'>
                New Application Environment
              </Button>
            </Link>
            <Button floated='left' basic size='small' icon='refresh' data-tooltip='Reload applications' onClick={() => props.updateApplications()}>
            </Button>
          </Table.HeaderCell>
        </Table.Row>
      </Table.Footer>
    </Table>
  )
}

export default ApplicationTable
