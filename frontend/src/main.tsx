import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'

import 'antd/dist/reset.css'

import { LogFromFrontend } from '../wailsjs/go/main/App'

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)

// 重写控制台方法
const original = {
  log: console.log,
  warn: console.warn,
  error: console.error
}

console.log = (...args) => {
  original.log(...args)
  LogFromFrontend('[LOG] ' + args.join(' '))
}

console.warn = (...args) => {
  original.warn(...args)
  LogFromFrontend('[WARN] ' + args.join(' '))
}

console.error = (...args) => {
  original.error(...args)
  LogFromFrontend('[ERROR] ' + args.join(' '))
}
