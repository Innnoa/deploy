import {useState} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Greet} from "../wailsjs/go/main/App";
import { PlusOutlined } from '@ant-design/icons';
import 'antd/dist/reset.css';
import { Button, Space, Typography, Card } from 'antd';
import Info from './components/info';
import Welcome from './components/welcome';

const App: React.FC = () => {
    

    return (
        <div id="App" style={{ display: 'flex', width: '100%', height: '100vh' }}>
            <Info />
            <Welcome />
        </div>
        
    )
}

export default App