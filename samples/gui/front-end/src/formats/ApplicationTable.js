import React, { useState } from 'react'
import { Button, Table, Modal, Popup, Grid, Segment } from 'semantic-ui-react'
import { Link } from 'react-router-dom'


const ApplicationTable = (props) => {
  // open state of remove question
  const [openQ, setOpen] = useState({ open: false })
  const onClose = () => setOpen({ open: false })
  const onOpen = () => setOpen({ open: true })

  // identifying info for deleting application instance
  const [deleteApplication, setDeleteApplication] = useState({ uid: '', name: '' })
  // save identifying info for deleting application instance
  const onDeleteClicked = (uid, name) => {
    setDeleteApplication({ ...deleteApplication, uid: uid, name: name })
  }
  // delete application instance
  const onDelete = () => {
    props.deleteApplication(deleteApplication.uid, deleteApplication.name)
    onClose()
  }

  // Show status success/in progress/error, and display the data access instructions
  const TableCellStatus = (status) => {
    if (('ready' in status.status) && status.status.ready) {
      let feedback = ''
      if (status.status.dataAccessInstructions) {
        feedback += status.status.dataAccessInstructions
      }
      if (status.status.catalogedAssets) {
        feedback += "Cataloged assets:\n"
        for (const [key, value] of Object.entries(status.status.catalogedAssets)) {
          feedback += "\n" + value 
        }
      }
      let t0 = (<span style={{whiteSpace: "pre-line"}}>{feedback}</span>)
      return (
        <Table.Cell positive textAlign='center'>
          <Popup position='left center' pinned on='click' content={<div className="description">
                {t0}
                </div>} 
          trigger={<Button flowing='true' basic size='small' icon='check' color='green'/>}>
          </Popup>
        </Table.Cell>
      )
    } else {
      if ('conditions' in status.status && (status.status.conditions[0].status === "True" || status.status.conditions[1].status === "True")) {
        return (
          <Table.Cell textAlign='center'>
            <Popup position='left center' pinned on='click' trigger={<Button basic icon='exclamation' flowing='true' color='red'/>}>
              <Grid>
                {(status.status.conditions.map((condition, index) => (
                  <Grid.Row key={index}>
                    <Segment secondary attached><b>{condition.type}</b></Segment>
                    <Segment attached><b>Message: </b>{condition.message}</Segment>
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
      <Modal trigger={<Button basic icon='remove circle' data-tooltip='delete' onClick={() => onDeleteClicked(data.application.metadata.uid, data.application.metadata.name)} />}
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
          <Table.HeaderCell>Purpose</Table.HeaderCell>
          <Table.HeaderCell>Status</Table.HeaderCell>
          <Table.HeaderCell></Table.HeaderCell>
        </Table.Row>
      </Table.Header>

      <Table.Body>
        {props.applications.length > 0 ? (
          props.applications.map(application => (
            <Table.Row key={application.metadata.uid}>
              <Table.Cell>{application.metadata.name} </Table.Cell>
              <Table.Cell>{application.spec.appInfo.role}</Table.Cell>
              <Table.Cell>{application.spec.appInfo.purpose}</Table.Cell>
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