import React from 'react'
import { Image, Menu } from 'semantic-ui-react'

const FybrikMenuBar = props => (
  <Menu inverted color='blue' >
      <Menu.Item as='h3' style={{ color: 'yellow' }} header>
        <Image size='tiny' src={require('../images/image.png')} style={{ marginRight: '1.5em' }} />
        Namespace: {props.datauserenv.namespace}
      </Menu.Item>
      <Menu.Item as='h3' style={{ color: 'yellow' }} header>
        Region: {props.datauserenv.geography}
      </Menu.Item>
  </Menu>
)

export default FybrikMenuBar