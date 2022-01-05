import React from 'react'
import { Button, Table, Dropdown, Input, Icon } from 'semantic-ui-react'

const formatOptions = [
  {
    key: 'parquet',
    text: 'parquet',
    value: 'parquet'
  },
  {
    key: 'table',
    text: 'table',
    value: 'table'
  },
  {
    key: 'csv',
    text: 'csv',
    value: 'csv'
  },
  {
    key: 'json',
    text: 'json',
    value: 'json'
  },
  {
    key: 'avro',
    text: 'avro',
    value: 'avro'
  },
  {
    key: 'orc',
    text: 'orc',
    value: 'orc'
  },
  {
    key: 'binary',
    text: 'binary',
    value: 'binary'
  },
  {
    key: 'arrow',
    text: 'arrow',
    value: 'arrow'
  }
]

const protocolOptions = [
  {
    key: 'fybrik-arrow-flight',
    text: 'fybrik-arrow-flight',
    value: 'fybrik-arrow-flight'
  },
  {
    key: 'db2',
    text: 'db2',
    value: 'db2'
  },
  {
    key: 'kafka',
    text: 'kafka',
    value: 'kafka'
  },
  {
    key: 's3',
    text: 's3',
    value: 's3'
  }
]
  
const InputDataTable = props => {
  // disable remove button if there is only one data
  const isDisabled = (props.applicationData.length === 1)

  return (
    <Table celled >
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell>AssetID
          <Icon.Group><Icon name='asterisk' color='red' corner='top right' /></Icon.Group>
          </Table.HeaderCell>
          <Table.HeaderCell>Format
          <Icon.Group><Icon name='asterisk' color='red' corner='top right' /></Icon.Group>
          </Table.HeaderCell>
          <Table.HeaderCell>Protocol
          <Icon.Group><Icon name='asterisk' color='red' corner='top right' /></Icon.Group>
          </Table.HeaderCell>
          <Table.HeaderCell>Destination Catalog
          </Table.HeaderCell>
          <Table.HeaderCell></Table.HeaderCell>
        </Table.Row>
      </Table.Header>

      <Table.Body>
        {props.applicationData.map(data => (
          <tr key={data.uid}>
            <Table.Cell>
              <Input
                autoComplete='off'
                onChange={(e) => props.handleIDChange(e, data.uid)}
                name='dataSetID'
                value={data.dataSetID}
              />
            </Table.Cell>
            <Table.Cell>
              <Dropdown
                selection
                options={formatOptions}
                name='dataformat'
                value={data.requirements.interface.dataformat}
                onChange={(e, { name, value }) => props.handleDetailsChange(e, data.uid, name, value)}
              />
            </Table.Cell>
            <Table.Cell>
              <Dropdown
                selection
                options={protocolOptions}
                name='protocol'
                value={data.requirements.interface.protocol}
                onChange={(e, { name, value }) => props.handleDetailsChange(e, data.uid, name, value)}
              />
            </Table.Cell>
            <Table.Cell>
              <Input
                autoComplete='off'
                name='catalogID'
                value={data.requirements.copy.catalog.catalogID}
                onChange={(e) => props.handleCatalogChange(e, data.uid)}
              />
            </Table.Cell>
            <Table.Cell textAlign='center'>
              <Button basic icon='add' onClick={() => { props.addDataRow() }} data-tooltip='add' />
              <Button basic icon='remove circle' onClick={() => props.deleteDataRow(data.uid)} data-tooltip='remove' disabled={isDisabled} />
            </Table.Cell>
          </tr>
        ))
        }
      </Table.Body>

    </Table>
  )
}

export default InputDataTable