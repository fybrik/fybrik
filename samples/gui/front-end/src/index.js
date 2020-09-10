import React from 'react'
import { render } from 'react-dom'
import { BrowserRouter } from 'react-router-dom'
import 'semantic-ui-less/semantic.less'

import App from './App'

render(
  <BrowserRouter>
    <App />
  </BrowserRouter>,
  document.querySelector('#root')
)